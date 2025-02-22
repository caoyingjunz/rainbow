package rainbow

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateTask(ctx context.Context, req *types.CreateTaskRequest) error {
	object, err := s.factory.Task().Create(ctx, &model.Task{
		Name:       req.Name,
		UserId:     req.UserId,
		RegisterId: req.RegisterId,
		AgentName:  req.AgentName,
		Mode:       req.Mode,
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
			TaskId:   taskId,
			TaskName: req.Name,
			UserId:   req.UserId,
			Name:     i,
		})
	}

	if err = s.factory.Image().CreateInBatch(ctx, images); err != nil {
		_ = s.DeleteTaskWithImages(ctx, taskId)
		return fmt.Errorf("failed to create tasks images %v", err)
	}

	return nil
}

func (s *ServerController) UpdateTask(ctx context.Context, req *types.UpdateTaskRequest) error {
	if err := s.factory.Task().Update(ctx, req.Id, req.ResourceVersion, map[string]interface{}{
		"register_id": req.RegisterId,
		"mode":        req.Mode,
	}); err != nil {
		return err
	}

	old, err := s.factory.Image().ListWithTask(ctx, req.Id)
	if err != nil {
		return err
	}
	var oldImages []string
	for _, o := range old {
		oldImages = append(oldImages, o.Name)
	}
	oldImageMap := sets.NewString(oldImages...)

	var addImages []string
	for _, n := range req.Images {
		if oldImageMap.Has(n) {
			continue
		}
		addImages = append(addImages, n)
	}
	var images []model.Image
	for _, i := range addImages {
		images = append(images, model.Image{
			TaskId: req.Id,
			Name:   i,
		})
	}
	if err = s.factory.Image().CreateInBatch(ctx, images); err != nil {
		return fmt.Errorf("failed to create tasks images %v", err)
	}

	return nil
}

func (s *ServerController) ListTasks(ctx context.Context, userId string) (interface{}, error) {
	if len(userId) == 0 {
		return s.factory.Task().List(ctx)
	}

	return s.factory.Task().ListWithUser(ctx, userId)
}

func (s *ServerController) UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error {
	return s.factory.Task().UpdateDirectly(ctx, req.TaskId, map[string]interface{}{"status": req.Status, "message": req.Message, "process": req.Process})
}

func (s *ServerController) DeleteTask(ctx context.Context, taskId int64) error {
	return s.DeleteTaskWithImages(ctx, taskId)
}

func (s *ServerController) DeleteTaskWithImages(ctx context.Context, taskId int64) error {
	_ = s.factory.Image().SoftDeleteInBatch(ctx, taskId)
	_ = s.factory.Task().Delete(ctx, taskId)
	return nil
}
