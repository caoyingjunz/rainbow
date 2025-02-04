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
}

type ServerController struct {
	factory db.ShareDaoFactory
}

func NewServer(f db.ShareDaoFactory) *ServerController {
	return &ServerController{
		factory: f,
	}
}
