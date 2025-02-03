package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db/model"

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
}

type ServerController struct {
	factory db.ShareDaoFactory
}

func (s *ServerController) CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error {
	_, err := s.factory.Registry().Create(ctx, &model.Registry{
		UserId:     req.UserId,
		Repository: req.Repository,
		Namespace:  req.Namespace,
		Username:   req.Username,
		Password:   req.Password,
	})

	return err
}

func (s *ServerController) ListRegistries(ctx context.Context) (interface{}, error) {
	return s.factory.Registry().List(ctx)
}

func (s *ServerController) GetRegistry(ctx context.Context, registryId int64) (interface{}, error) {
	return s.factory.Registry().Get(ctx, registryId)
}

func NewServer(f db.ShareDaoFactory) *ServerController {
	return &ServerController{
		factory: f,
	}
}
