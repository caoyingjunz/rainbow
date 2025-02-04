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

	taskId := object.Id
	fmt.Println("taskId", taskId)

	return err
}
