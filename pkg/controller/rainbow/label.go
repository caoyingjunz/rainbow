package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"k8s.io/klog/v2"
)

func (s *ServerController) CreateLabel(ctx context.Context, req *types.CreateLabelRequest) error {
	_, err := s.factory.Label().Create(ctx, &model.Label{
		Name: req.Name,
	})
	if err != nil {
		klog.Errorf("创建标签失败 %v", err)
		return err
	}

	return nil
}

func (s *ServerController) DeleteLabel(ctx context.Context, labelId int64) error {
	err := s.factory.Label().Delete(ctx, labelId)
	if err != nil {
		klog.Errorf("删除失败 %v", err)
		return err
	}

	return nil
}

func (s *ServerController) UpdateLabel(ctx context.Context, req *types.UpdateLabelRequest) error {
	updates := make(map[string]interface{})
	updates["name"] = req.Name
	return s.factory.Label().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) ListLabels(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	list, err := s.factory.Label().List(ctx, db.WithNameLike(listOption.NameSelector))
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ServerController) CreateLogo(ctx context.Context, req *types.CreateLogoRequest) error {
	_, err := s.factory.Label().CreateLogo(ctx, &model.Logo{
		Name: req.Name,
		Logo: req.Logo,
	})
	return err
}

func (s *ServerController) UpdateLogo(ctx context.Context, req *types.UpdateLogoRequest) error {
	updates := make(map[string]interface{})
	updates["logo"] = req.Logo
	return s.factory.Label().UpdateLogo(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) DeleteLogo(ctx context.Context, logoId int64) error {
	return s.factory.Label().DeleteLogo(ctx, logoId)
}

func (s *ServerController) ListLogos(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	return s.factory.Label().ListLogos(ctx, db.WithNameLike(listOption.NameSelector))
}
