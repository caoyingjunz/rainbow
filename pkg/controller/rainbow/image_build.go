package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

const (
	BuildWaitStatus = "调度中"
)

func (s *ServerController) CreateBuild(ctx context.Context, req *types.CreateBuildRequest) error {
	_, err := s.factory.Build().Create(ctx, &model.Build{
		Name:       req.Name,
		Arch:       req.Arch,
		Dockerfile: req.Dockerfile,
		RegistryId: req.RegistryId,
		Status:     BuildWaitStatus,
		Namespace:  req.Namespace,
		UserModel: rainbow.UserModel{
			UserId: req.UserId,
		},
		Repo: req.Repo,
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
	updates["dockerfile"] = req.Dockerfile
	updates["repo"] = req.Repo
	updates["status"] = req.Status

	return s.factory.Build().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) ListBuilds(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	listOption.SetDefaultPageOption()

	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}
	opts := []db.Options{
		db.WithUser(listOption.UserId),
		db.WithNameLike(listOption.NameSelector),
		db.WithNamespace(listOption.Namespace),
		db.WithAgent(listOption.Agent),
	}
	var err error
	pageResult.Total, err = s.factory.Build().Count(ctx, opts...)
	if err != nil {
		klog.Errorf("获取构建总数失败 %v", err)
		pageResult.Message = err.Error()
	}
	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithModifyOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)

	pageResult.Items, err = s.factory.Build().List(ctx, opts...)
	if err != nil {
		klog.Errorf("获取构建列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}

func (s *ServerController) GetBuild(ctx context.Context, buildId int64) (interface{}, error) {
	return s.factory.Build().Get(ctx, buildId)
}

func (s *ServerController) CreateBuildMessage(ctx context.Context, req types.CreateBuildMessageRequest) error {
	return s.factory.Build().CreateBuildMessage(ctx, &model.BuildMessage{
		Message: req.Message,
		BuildId: req.Id,
	})
}

func (s *ServerController) ListBuildMessages(ctx context.Context, buildId int64) (interface{}, error) {
	return s.factory.Build().ListBuildMessages(ctx, db.WithBuild(buildId))
}

func (s *ServerController) UpdateBuildStatus(ctx context.Context, req *types.UpdateBuildStatusRequest) error {
	return s.factory.Build().UpdateBy(ctx, map[string]interface{}{"status": req.Status}, db.WithId(req.BuildId))
}
