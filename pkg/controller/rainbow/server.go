package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"k8s.io/klog/v2"
	"time"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error
	GetRegistry(ctx context.Context, registryId int64) (interface{}, error)
	ListRegistries(ctx context.Context) (interface{}, error)

	CreateTask(ctx context.Context, req *types.CreateTaskRequest) error
	UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error

	GetAgent(ctx context.Context, agentId int64) (interface{}, error)
	ListAgents(ctx context.Context) (interface{}, error)

	CreateImage(ctx context.Context, req *types.CreateImageRequest) error
	UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error
	GetImage(ctx context.Context, imageId int64) (interface{}, error)
	ListImages(ctx context.Context, taskId int64) (interface{}, error)

	UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error

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

func (s *ServerController) ListAgents(ctx context.Context) (interface{}, error) {
	return s.factory.Agent().List(ctx)
}

func (s *ServerController) Run(ctx context.Context, workers int) error {
	go s.monitor(ctx)

	return nil
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
				err = s.factory.Agent().UpdateByName(ctx, agent.Name, map[string]interface{}{"status": model.UnknownAgentType, "message": "Agent stopped posting status."})
				if err != nil {
					klog.Error("failed to sync agent %s status %v", agent.Name, err)
				}
			}
		}
	}
}
