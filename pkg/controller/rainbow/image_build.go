package rainbow

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateBuild(ctx context.Context, req *types.CreateBuildRequest) error {
	_, err := s.factory.Build().Create(ctx, &model.Build{
		Name: req.Name,
	})
	if err != nil {
		klog.Errorf("创建镜像失败 %v", err)
	}

	return nil
}

func (s *ServerController) DeleteBuild(ctx context.Context, buildId int64) error {
	err := s.factory.Build().Delete(ctx, buildId)
	if err != nil {
		klog.Errorf("删除失败 %v", err)
		return err
	}

	return nil
}

func (s *ServerController) UpdateBuild(ctx context.Context, req *types.UpdateBuildRequest) error {
	updates := make(map[string]interface{})
	return s.factory.Build().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) ListBuilds(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	list, err := s.factory.Build().List(ctx, db.WithNameLike(listOption.NameSelector))
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ServerController) GetBuild(ctx context.Context, buildId int64) (interface{}, error) {
	return s.factory.Build().Get(ctx, buildId)
}

func (s *ServerController) UpdateBuildStatus(ctx context.Context, req *types.UpdateBuildStatusRequest) error {
	return s.factory.Build().UpdateBy(ctx, map[string]interface{}{"status": req.Status}, db.WithId(req.BuildId))
}
