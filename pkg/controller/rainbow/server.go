package rainbow

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	swr "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util/huaweicloud"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error
	UpdateRegistry(ctx context.Context, req *types.UpdateRegistryRequest) error
	DeleteRegistry(ctx context.Context, registryId int64) error
	GetRegistry(ctx context.Context, registryId int64) (interface{}, error)
	ListRegistries(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	CreateTask(ctx context.Context, req *types.CreateTaskRequest) error
	UpdateTask(ctx context.Context, req *types.UpdateTaskRequest) error
	ListTasks(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	DeleteTask(ctx context.Context, taskId int64) error
	GetTask(ctx context.Context, taskId int64) (interface{}, error)
	UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error

	ListTaskImages(ctx context.Context, taskId int64) (interface{}, error)

	GetAgent(ctx context.Context, agentId int64) (interface{}, error)
	ListAgents(ctx context.Context) (interface{}, error)
	UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error

	CreateImage(ctx context.Context, req *types.CreateImageRequest) error
	UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error
	DeleteImage(ctx context.Context, imageId int64) error
	GetImage(ctx context.Context, imageId int64) (interface{}, error)
	ListImages(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error
	CreateImages(ctx context.Context, req *types.CreateImagesRequest) error
	DeleteImageTag(ctx context.Context, imageId int64, name string) error

	GetCollection(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	AddDailyReview(ctx context.Context, page string) error

	CreateLabel(ctx context.Context, req *types.CreateLabelRequest) error
	DeleteLabel(ctx context.Context, labelId int64) error
	UpdateLabel(ctx context.Context, req *types.UpdateLabelRequest) error
	ListLabels(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	CreateLogo(ctx context.Context, req *types.CreateLogoRequest) error
	DeleteLogo(ctx context.Context, logoId int64) error
	ListLogos(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	Overview(ctx context.Context) (interface{}, error)
	Downflow(ctx context.Context) (interface{}, error)
	Store(ctx context.Context) (interface{}, error)

	Run(ctx context.Context, workers int) error
}

var (
	SwrClient  *swr.SwrClient
	RegistryId *int64
)

type ServerController struct {
	factory db.ShareDaoFactory
	cfg     rainbowconfig.Config
}

func NewServer(f db.ShareDaoFactory, cfg rainbowconfig.Config) *ServerController {
	sc := &ServerController{
		factory: f,
		cfg:     cfg,
	}

	if SwrClient == nil || RegistryId == nil {
		reg, err := f.Registry().GetDefaultRegistry(context.TODO())
		if err == nil {
			if len(reg.Ak) == 0 || len(reg.Sk) == 0 || len(reg.RegionId) == 0 {
				klog.Errorf("默认华为仓库未设置必要配置, ak(%s) sk(%s) regionId(%s)", reg.Ak, reg.Sk, reg.RegionId)
			} else {
				client, err := huaweicloud.NewHuaweiCloudClient(huaweicloud.HuaweiCloudConfig{
					AK:       reg.Ak,
					SK:       reg.Sk,
					RegionId: reg.RegionId,
				})
				if err == nil {
					SwrClient = client
					RegistryId = &reg.Id
					klog.Infof("创建华为仓库客户端成功，仓库名称: %s(%d) ", reg.Name, *RegistryId)
				} else {
					klog.Errorf("创建为仓库客户端失败 %v", err)
				}
			}
		} else {
			klog.Errorf("获取默认华为仓库失败: %v", err)
		}
	}

	return sc
}

func (s *ServerController) GetAgent(ctx context.Context, agentId int64) (interface{}, error) {
	return s.factory.Agent().Get(ctx, agentId)
}

func (s *ServerController) UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error {
	return s.factory.Agent().UpdateByName(ctx, req.AgentName, map[string]interface{}{"status": req.Status, "message": fmt.Sprintf("Agent has been set to %s", req.Status)})
}

func (s *ServerController) ListAgents(ctx context.Context) (interface{}, error) {
	return s.factory.Agent().List(ctx)
}

func (s *ServerController) Run(ctx context.Context, workers int) error {
	go s.monitor(ctx)
	go s.schedule(ctx)

	return nil
}

func (s *ServerController) schedule(ctx context.Context) {
	klog.Infof("starting scheduler controller")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.doSchedule(ctx); err != nil {
			klog.Error("failed to do schedule %v", err)
		}
	}
}

func (s *ServerController) doSchedule(ctx context.Context) error {
	item, err := s.factory.Task().GetOneForSchedule(ctx)
	if err != nil {
		return err
	}
	if item == nil {
		return err
	}

	targetAgent, err := s.assignAgent(ctx)
	if err != nil {
		return err
	}
	if targetAgent == "" {
		return nil
	}
	if err = s.factory.Task().Update(ctx, item.Id, item.ResourceVersion, map[string]interface{}{
		"agent_name": targetAgent,
	}); err != nil {
		return err
	}

	return nil
}

func (s *ServerController) assignAgent(ctx context.Context) (string, error) {
	agents, err := s.factory.Agent().ListForSchedule(ctx)
	if err != nil {
		return "", err
	}
	if len(agents) == 0 {
		klog.Warningf("不存在可用工作节点，等待下一次调度")
		return "", nil
	}

	var agentNames []string
	agentMap := make(map[string]int)
	for _, agent := range agents {
		agentNames = append(agentNames, agent.Name)
		agentMap[agent.Name] = 0
	}
	agentSet := sets.NewString(agentNames...)

	runningTasks, err := s.factory.Task().GetRunningTask(ctx)
	if err != nil {
		return "", err
	}

	if len(runningTasks) == 0 {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(len(agentNames))
		agent := agentNames[x]
		klog.Infof("当前节点均空闲，工作节点 %s 被随机选中", agent)
		return agent, nil
	} else {
		for _, t := range runningTasks {
			if !agentSet.Has(t.AgentName) {
				continue
			}
			old, ok := agentMap[t.AgentName]
			if ok {
				agentMap[t.AgentName] = old + 1
			} else {
				continue
			}
		}

		min := len(runningTasks)
		agent := ""
		for k, v := range agentMap {
			if min >= v {
				min = v
				agent = k
			}
		}
		// 一个 agent 最大并发为 10
		if min > 10 {
			klog.Warningf("工作节点已满负载，等待下一次调度")
			return "", nil
		}
		if agent == "" {
			klog.Warningf("未选中工作节点，等待下一次调度")
			return "", nil
		}

		klog.Infof("工作节点 %s 已选中", agent)
		return agent, nil
	}
}

func (s *ServerController) monitor(ctx context.Context) {
	klog.Infof("starting agent monitor")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		agents, err := s.factory.Agent().List(ctx)
		if err != nil {
			klog.Error("failed to get agents %v", err)
			continue
		}

		for _, agent := range agents {
			diff := time.Now().Sub(agent.LastTransitionTime)
			if diff > time.Minute*5 {
				if agent.Status == model.UnknownAgentType {
					continue
				}
				err = s.factory.Agent().UpdateByName(ctx, agent.Name, map[string]interface{}{"status": model.UnknownAgentType, "message": "Agent stopped posting status"})
				if err != nil {
					klog.Error("failed to sync agent %s status %v", agent.Name, err)
				}
			}
		}
	}
}
