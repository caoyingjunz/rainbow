package rainbow

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error
	UpdateRegistry(ctx context.Context, req *types.UpdateRegistryRequest) error
	DeleteRegistry(ctx context.Context, registryId int64) error
	GetRegistry(ctx context.Context, registryId int64) (interface{}, error)
	ListRegistries(ctx context.Context, userId string) (interface{}, error)

	CreateTask(ctx context.Context, req *types.CreateTaskRequest) error
	UpdateTask(ctx context.Context, req *types.UpdateTaskRequest) error
	ListTasks(ctx context.Context, userId string) (interface{}, error)
	DeleteTask(ctx context.Context, taskId int64) error
	UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error

	GetAgent(ctx context.Context, agentId int64) (interface{}, error)
	ListAgents(ctx context.Context) (interface{}, error)
	UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error

	CreateImage(ctx context.Context, req *types.CreateImageRequest) error
	UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error
	SoftDeleteImage(ctx context.Context, imageId int64) error
	GetImage(ctx context.Context, imageId int64) (interface{}, error)
	ListImages(ctx context.Context, taskId int64, userId string) (interface{}, error)

	UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error
	CreateImages(ctx context.Context, req *types.CreateImagesRequest) error

	Run(ctx context.Context, workers int) error
}

type ServerController struct {
	factory db.ShareDaoFactory
}

func NewServer(f db.ShareDaoFactory) *ServerController {
	return &ServerController{
		factory: f,
	}
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
