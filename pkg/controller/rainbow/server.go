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
}

type ServerController struct {
	factory db.ShareDaoFactory
}

func (s *ServerController) CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error {
	return nil
}

func (s *ServerController) ListRegistries(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func NewServer(f db.ShareDaoFactory) *ServerController {
	return &ServerController{
		factory: f,
	}
}
