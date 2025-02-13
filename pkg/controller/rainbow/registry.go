package rainbow

import (
	"context"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

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

func (s *ServerController) UpdateRegistry(ctx context.Context, req *types.UpdateRegistryRequest) error {
	return s.factory.Registry().Update(ctx, req.Id, req.ResourceVersion, map[string]interface{}{
		"user_id":    req.UserId,
		"repository": req.Repository,
		"namespace":  req.Namespace,
		"username":   req.Username,
		"password":   req.Password,
	})
}

func (s *ServerController) DeleteRegistry(ctx context.Context, registryId int64) error {
	return s.factory.Registry().Delete(ctx, registryId)
}

func (s *ServerController) ListRegistries(ctx context.Context) (interface{}, error) {
	return s.factory.Registry().List(ctx)
}

func (s *ServerController) GetRegistry(ctx context.Context, registryId int64) (interface{}, error) {
	return s.factory.Registry().Get(ctx, registryId)
}
