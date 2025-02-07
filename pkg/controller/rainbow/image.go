package rainbow

import (
	"context"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateImage(ctx context.Context, req *types.CreateImageRequest) error {
	_, err := s.factory.Image().Create(ctx, &model.Image{})

	return err
}

func (s *ServerController) UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error {
	return nil
}

func (s *ServerController) ListImages(ctx context.Context) (interface{}, error) {
	return s.factory.Image().List(ctx)
}

func (s *ServerController) GetImage(ctx context.Context, imageId int64) (interface{}, error) {
	return s.factory.Image().Get(ctx, imageId)
}
