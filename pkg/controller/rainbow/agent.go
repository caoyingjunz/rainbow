package rainbow

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/pixiulib/strutil"
	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
)

type AgentGetter interface {
	Agent() Interface
}
type Interface interface {
	Run(ctx context.Context, workers int) error
}

type AgentController struct {
	factory db.ShareDaoFactory
	cfg     rainbowconfig.Config

	queue workqueue.RateLimitingInterface

	name     string
	callback string
	baseDir  string
}

func NewAgent(f db.ShareDaoFactory, cfg rainbowconfig.Config) *AgentController {
	return &AgentController{
		factory:  f,
		cfg:      cfg,
		name:     cfg.Agent.Name,
		baseDir:  cfg.Agent.DataDir,
		callback: cfg.Plugin.Callback,
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rainbow-agent"),
	}
}

func (s *AgentController) Run(ctx context.Context, workers int) error {
	// 注册 rainbow 代理
	if err := s.RegisterAgentIfNotExist(ctx); err != nil {
		return err
	}

	go s.report(ctx)

	go s.getNextWorkItems(ctx)

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 1*time.Second)
	}

	return nil
}

func (s *AgentController) report(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		newAgent, err := s.factory.Agent().GetByName(ctx, s.name)
		if err != nil {
			klog.Error("failed to get agent status %v", err)
			continue
		}

		updates := map[string]interface{}{"last_transition_time": time.Now()}
		if newAgent.Status == model.UnknownAgentType {
			updates["status"] = model.RunAgentType
			updates["message"] = "Agent started posting status"
		}

		err = s.factory.Agent().UpdateByName(ctx, s.name, updates)
		if err != nil {
			klog.Error("failed to sync agent status %v", err)
		}
	}
}

func (s *AgentController) getNextWorkItems(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取未处理
		tasks, err := s.factory.Task().ListWithAgent(ctx, s.name, 0)
		if err != nil {
			klog.Error("failed to list tasks %v", err)
			continue
		}
		if len(tasks) == 0 {
			continue
		}

		for _, task := range tasks {
			s.queue.Add(fmt.Sprintf("%d/%d", task.Id, task.ResourceVersion))
		}
	}
}

func (s *AgentController) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *AgentController) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	taskId, resourceVersion, err := KeyFunc(key)
	if err != nil {
		s.handleErr(ctx, err, key)
	} else {
		_ = s.factory.Task().UpdateDirectly(ctx, taskId, map[string]interface{}{"status": "镜像初始化", "message": "初始化环境中", "process": 1})
		s.handleErr(ctx, s.sync(ctx, taskId, resourceVersion), key)
	}
	return true
}

func (s *AgentController) GetOneAdminRegistry(ctx context.Context) (*model.Registry, error) {
	regs, err := s.factory.Registry().GetAdminRegistries(ctx)
	if err != nil {
		klog.Errorf("获取默认镜像仓库失败: %v", err)
		return nil, err
	}
	if len(regs) == 0 {
		klog.Errorf("no admin or default registry found")
		return nil, fmt.Errorf("no admin or default registry found")
	}

	// 随机分，暂时不考虑负载情况，后续优化
	rand.Seed(time.Now().UnixNano())
	x := rand.Intn(len(regs))
	t := regs[x]
	return &t, err
}

func (s *AgentController) makePluginConfig(ctx context.Context, task model.Task) (*rainbowconfig.PluginTemplateConfig, error) {
	taskId := task.Id

	var (
		registry *model.Registry
		err      error
	)
	// 未指定自定义参考时，使用默认仓库
	if task.RegisterId == 0 {
		registry, err = s.GetOneAdminRegistry(ctx)
	} else {
		registry, err = s.factory.Registry().Get(ctx, task.RegisterId)
	}
	if err != nil {
		klog.Error("failed to get registry %v", err)
		return nil, fmt.Errorf("failed to get registry %v", err)
	}

	pluginTemplateConfig := &rainbowconfig.PluginTemplateConfig{
		Default: rainbowconfig.DefaultOption{
			Time: time.Now().Unix(), // 注入时间戳，确保每次内容都不相同
		},
		Plugin: rainbowconfig.PluginOption{
			Callback:   s.callback,
			TaskId:     taskId,
			RegistryId: registry.Id,
			Synced:     true,
			Driver:     task.Driver,
		},
		Registry: rainbowconfig.Registry{
			Repository: registry.Repository,
			Namespace:  registry.Namespace,
			Username:   registry.Username,
			Password:   registry.Password,
		},
	}

	// 根据type判断是镜像列表推送还是k8s镜像组推送
	switch task.Type {
	case 0:
		tags, err := s.factory.Image().ListTags(ctx, db.WithTask(taskId))
		if err != nil {
			klog.Errorf("获取任务所属 tags 失败 %v", err)
			return nil, err
		}

		var imageIds []int64
		imageMap := make(map[int64][]model.Tag)
		for _, tag := range tags {
			imageIds = append(imageIds, tag.ImageId)
			old, ok := imageMap[tag.ImageId]
			if ok {
				imageMap[tag.ImageId] = append(old, tag)
			} else {
				imageMap[tag.ImageId] = []model.Tag{tag}
			}
		}
		images, err := s.factory.Image().List(ctx, db.WithIDIn(imageIds...))
		if err != nil {
			klog.Errorf("获取任务所属镜像失败 %v", err)
			return nil, err
		}

		var img []rainbowconfig.Image
		for _, i := range images {
			ts, ok := imageMap[i.Id]
			if !ok {
				klog.Warningf("未能找到镜像(%d)的tags", i.Name)
				continue
			}
			var tagStr []string
			for _, tt := range ts {
				tagStr = append(tagStr, tt.Name)
			}
			img = append(img, rainbowconfig.Image{
				Name: i.Name,
				Id:   i.Id,
				Path: i.Path,
				Tags: tagStr,
			})
		}
		pluginTemplateConfig.Default.PushImages = true
		pluginTemplateConfig.Images = img
	case 1:
		pluginTemplateConfig.Default.PushKubernetes = true
		pluginTemplateConfig.Kubernetes.Version = task.KubernetesVersion
	}

	return pluginTemplateConfig, err
}

func (s *AgentController) sync(ctx context.Context, taskId int64, resourceVersion int64) error {
	task, err := s.factory.Task().GetOne(ctx, taskId, resourceVersion)
	if err != nil {
		if errors.IsNotUpdated(err) {
			return nil
		}
		return fmt.Errorf("failted to get one task %d %v", taskId, err)
	}

	tplCfg, err := s.makePluginConfig(ctx, *task)
	cfg, err := yaml.Marshal(tplCfg)
	if err != nil {
		return err
	}

	taskIdStr := fmt.Sprintf("%d", taskId)

	destDir := filepath.Join(s.baseDir, taskIdStr)
	if err = util.EnsureDirectoryExists(destDir); err != nil {
		return err
	}
	if !util.IsDirectoryExists(destDir + "/plugin") {
		if err = util.Copy(s.baseDir+"/plugin", destDir); err != nil {
			return err
		}
	}

	git := util.NewGit(destDir+"/plugin", taskIdStr, taskIdStr+"-"+time.Now().String())
	if err = git.Checkout(); err != nil {
		return err
	}
	if err = util.WriteIntoFile(string(cfg), destDir+"/plugin/config.yaml"); err != nil {
		return err
	}
	if err = git.Push(); err != nil {
		return err
	}
	return nil
}

// TODO
func (s *AgentController) handleErr(ctx context.Context, err error, key interface{}) {
	if err == nil {
		return
	}
	klog.Error(err)
}

func (s *AgentController) RegisterAgentIfNotExist(ctx context.Context) error {
	if len(s.name) == 0 {
		return fmt.Errorf("agent name missing")
	}

	var err error
	_, err = s.factory.Agent().GetByName(ctx, s.name)
	if err == nil {
		return nil
	}
	_, err = s.factory.Agent().Create(ctx, &model.Agent{Name: s.name, Status: model.RunAgentType, Type: model.PublicAgentType, Message: "Agent started posting status"})
	return err
}

func KeyFunc(key interface{}) (int64, int64, error) {
	str, ok := key.(string)
	if !ok {
		return 0, 0, fmt.Errorf("failed to convert %v to string", key)
	}
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("parts length not 2")
	}

	taskId, err := strutil.ParseInt64(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to Parse taskId to Int64 %v", err)
	}
	resourceVersion, err := strutil.ParseInt64(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to Parse resourceVersion to Int64 %v", err)
	}

	return taskId, resourceVersion, nil
}
