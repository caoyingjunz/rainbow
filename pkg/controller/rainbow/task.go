package rainbow

import (
	"context"
	"fmt"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateTask(ctx context.Context, req *types.CreateTaskRequest) error {
	object, err := s.factory.Task().Create(ctx, &model.Task{
		UserId:     req.UserId,
		RegisterId: req.RegisterId,
		AgentName:  req.AgentName,
	})
	if err != nil {
		return err
	}

	if len(req.Images) == 0 {
		return nil
	}
	taskId := object.Id

	var images []model.Image
	for _, i := range req.Images {
		images = append(images, model.Image{
			TaskId: taskId,
			Name:   i,
		})
	}

	if err = s.factory.Image().CreateInBatch(ctx, images); err != nil {
		_ = s.DeleteTaskWithImages(ctx, taskId)
		return fmt.Errorf("failed to create tasks images %v", err)
	}

	return nil
}

func (s *ServerController) UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error {
	return s.factory.Task().UpdateDirectly(ctx, req.TaskId, map[string]interface{}{"status": req.Status, "message": req.Message})
}

func (s *ServerController) DeleteTaskWithImages(ctx context.Context, taskId int64) error {
	_ = s.factory.Image().DeleteInBatch(ctx, taskId)
	_ = s.factory.Task().Delete(ctx, taskId)
	return nil
}
