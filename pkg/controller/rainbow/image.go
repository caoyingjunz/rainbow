package rainbow

import (
	"context"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateImage(ctx context.Context, req *types.CreateImageRequest) error {
	_, err := s.factory.Image().Create(ctx, &model.Image{
		TaskId: req.TaskId,
		Name:   req.Name,
		Status: req.Status,
	})

	return err
}

func (s *ServerController) UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error {
	updates := make(map[string]interface{})
	updates["task_id"] = req.TaskId
	updates["name"] = req.Name
	updates["status"] = req.Status
	updates["message"] = req.Message
	return s.factory.Image().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) ListImages(ctx context.Context, taskId int64) (interface{}, error) {
	if taskId == 0 {
		return s.factory.Image().List(ctx)
	}

	return s.factory.Image().ListWithTask(ctx, taskId)
}

func (s *ServerController) GetImage(ctx context.Context, imageId int64) (interface{}, error) {
	return s.factory.Image().Get(ctx, imageId)
}
