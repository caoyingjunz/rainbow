package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"k8s.io/klog/v2"
)

func (s *ServerController) CreateDockerfile(ctx context.Context, req *types.CreateDockerfileRequest) error {
	_, err := s.factory.Dockerfile().Create(ctx, &model.Dockerfile{
		Name:       req.Name,
		Dockerfile: req.Dockerfile,
	})
	if err != nil {
		klog.Errorf("创建镜像失败 %v", err)
	}

	return nil
}

func (s *ServerController) DeleteDockerfile(ctx context.Context, dockerfileId int64) error {
	err := s.factory.Dockerfile().Delete(ctx, dockerfileId)
	if err != nil {
		klog.Errorf("删除失败 %v", err)
		return err
	}

	return nil
}

func (s *ServerController) UpdateDockerfile(ctx context.Context, req *types.UpdateDockerfileRequest) error {
	updates := make(map[string]interface{})
	updates["dockerfile"] = req.Dockerfile
	return s.factory.Dockerfile().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) ListDockerfile(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	list, err := s.factory.Dockerfile().List(ctx, db.WithNameLike(listOption.NameSelector))
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ServerController) GetDockerfile(ctx context.Context, dockerfileId int64) (interface{}, error) {
	return s.factory.Dockerfile().Get(ctx, dockerfileId)
}
