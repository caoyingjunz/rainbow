package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error
	ListRegistries(ctx context.Context) (interface{}, error)
	GetRegistry(ctx context.Context, registryId int64) (interface{}, error)

	CreateTask(ctx context.Context, req *types.CreateTaskRequest) error

	GetAgent(ctx context.Context, agentId int64) (interface{}, error)
	ListAgents(ctx context.Context) (interface{}, error)
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
