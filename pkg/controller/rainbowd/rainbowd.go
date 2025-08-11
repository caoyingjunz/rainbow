package rainbowd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

type RainbowdGetter interface {
	Rainbowd() Interface
}

type Interface interface {
	Run(ctx context.Context, workers int) error
}

type rainbowdController struct {
	name    string
	factory db.ShareDaoFactory
	cfg     rainbowconfig.Config

	queue workqueue.RateLimitingInterface
}

func New(f db.ShareDaoFactory, cfg rainbowconfig.Config) *rainbowdController {
	return &rainbowdController{
		factory: f,
		cfg:     cfg,
		name:    cfg.Rainbowd.Name,
		queue:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rainbowd"),
	}
}

func (s *rainbowdController) Run(ctx context.Context, workers int) error {
	if err := s.RegisterIfNotExist(ctx); err != nil {
		klog.Errorf("register rainbowd failed: %v", err)
		return err
	}

	go s.getNextWorkItems(ctx)

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, s.worker, 1*time.Second)
	}

	return nil
}

func (s *rainbowdController) RegisterIfNotExist(ctx context.Context) error {
	if len(s.name) == 0 {
		return fmt.Errorf("rainbowd name is empty")
	}

	var err error
	_, err = s.factory.Rainbowd().GetByName(ctx, s.name)
	if err == nil {
		return nil
	}
	_, err = s.factory.Rainbowd().Create(ctx, &model.Rainbowd{
		Name:   s.name,
		Status: model.RunAgentType,
	})
	return nil
}

func (s *rainbowdController) worker(ctx context.Context) {
	for s.processNextWorkItem(ctx) {
	}
}

func (s *rainbowdController) processNextWorkItem(ctx context.Context) bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	agentId, resourceVersion, err := util.KeyFunc(key)
	if err != nil {
		s.handleErr(ctx, err, key)
	} else {
		if err = s.sync(ctx, agentId, resourceVersion); err != nil {
			s.handleErr(ctx, err, key)
		}
	}

	return true
}

func (s *rainbowdController) getNextWorkItems(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取未处理
		agents, err := s.factory.Agent().List(ctx, db.WithRainbowdName(s.name))
		if err != nil {
			klog.Error("failed to list my agents %v", err)
			continue
		}
		if len(agents) == 0 {
			continue
		}
		for _, agent := range agents {
			s.queue.Add(fmt.Sprintf("%d/%d", agent.Id, agent.ResourceVersion))
		}
	}
}

// TODO
func (s *rainbowdController) handleErr(ctx context.Context, err error, key interface{}) {
	if err == nil {
		return
	}
	klog.Error(err)
}

// 1. 获取 agent 期望状态 (数据库状态)
// 2. 获取 agent 实际运行状态（容器状态）
// 3. 调整容器状态为数据库状态
func (s *rainbowdController) sync(ctx context.Context, agentId int64, resourceVersion int64) error {
	old, err := s.factory.Agent().Get(ctx, agentId)
	if err != nil {
		klog.Errorf("获取 agent %s 失败", err)
		return err
	}

	return s.reconcileAgent(old)
}

func (s *rainbowdController) startAgentContainer(agent *model.Agent) error {
	return nil
}

func (s *rainbowdController) reconcileAgent(agent *model.Agent) error {
	runContainer, err := s.getAgentContainer(agent)
	if err != nil {
		return err
	}

	// 容器不存在，需要创建
	if runContainer == nil {
		if err = s.prepareConfig(agent); err != nil {
			klog.Errorf("prepare agent Config 失败 %v", err)
			return err
		}
		if err = s.startAgentContainer(agent); err != nil {
			klog.Errorf("start agent Config 失败 %v", err)
			return err
		}

		return nil
	}

	// 存在，则重构
	// TODO: 检查是否需要重构
	if agent.Status == model.RunAgentType && runContainer.Status == "Running" {
		return nil
	}

	return nil
}

func (s *rainbowdController) getAgentContainer(agent *model.Agent) (*types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	cs, err := cli.ContainerList(context.TODO(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for _, c := range cs {
		for _, name := range c.Names {
			if name == agent.Name {
				return &c, nil
			}
		}
	}
	return nil, nil
}

func (s *rainbowdController) prepareConfig(agent *model.Agent) error {
	agentName := agent.Name
	// 准备工作文件夹
	destDir := filepath.Join(s.cfg.Rainbowd.DataDir, agentName)
	if err := util.EnsureDirectoryExists(destDir); err != nil {
		return err
	}

	// 拷贝 plugin
	if !util.IsDirectoryExists(destDir + "/plugin") {
		if err := util.Copy(s.cfg.Rainbowd.TemplateDir+"/plugin", destDir); err != nil {
			return err
		}
	}
	// 拷贝 agent，每次都重置最新
	if err := util.Copy(s.cfg.Rainbowd.TemplateDir+"/agent", destDir); err != nil {
		return err
	}
	// 配置文件 config.yaml
	data, err := util.ReadFromFile(s.cfg.Rainbowd.TemplateDir + "/config.yaml")
	if err != nil {
		return err
	}
	var cfg rainbowconfig.Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	cfg.Agent.Name = agentName
	cfgData, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err = util.WriteIntoFile(string(cfgData), destDir+"/config.yaml"); err != nil {
		return err
	}

	// 渲染 .git/config
	gc := struct{ URL string }{URL: fmt.Sprintf("https://github.com/%s/plugin.git", agent.GithubUser)}
	tpl := template.New(agentName)
	t := template.Must(tpl.Parse(GitConfig))

	var buf bytes.Buffer
	if err = t.Execute(&buf, gc); err != nil {
		return err
	}
	if err = ioutil.WriteFile(destDir+"/plugin/.git/config", buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

const GitConfig = `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = {{ .URL }}
	fetch = +refs/heads/*:refs/remotes/origin/*
`
