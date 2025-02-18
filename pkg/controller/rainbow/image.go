package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateImage(ctx context.Context, req *types.CreateImageRequest) error {
	_, err := s.factory.Image().Create(ctx, &model.Image{
		Name:     req.Name,
		TaskId:   req.TaskId,
		TaskName: req.TaskName,
		Status:   req.Status,
	})

	return err
}

func (s *ServerController) UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error {
	updates := make(map[string]interface{})
	updates["task_id"] = req.TaskId
	updates["task_name"] = req.TaskName
	updates["name"] = req.Name
	updates["status"] = req.Status
	updates["message"] = req.Message
	return s.factory.Image().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error {
	return s.factory.Image().UpdateDirectly(ctx, req.Name, req.TaskId, map[string]interface{}{
		"status":  req.Status,
		"message": req.Message,
		"target":  req.Target,
	})
}

func (s *ServerController) ListImages(ctx context.Context, taskId int64, userId string) (interface{}, error) {
	if taskId == 0 && len(userId) == 0 {
		return s.factory.Image().List(ctx)
	}
	if taskId != 0 && len(userId) != 0 {
		return s.factory.Image().ListWithUserAndTask(ctx, taskId, userId)
	}

	if taskId != 0 {
		return s.factory.Image().ListWithTask(ctx, taskId)
	}
	return s.factory.Image().ListWithUser(ctx, userId)
}

func (s *ServerController) GetImage(ctx context.Context, imageId int64) (interface{}, error) {
	return s.factory.Image().Get(ctx, imageId)
}
