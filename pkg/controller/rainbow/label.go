package rainbow

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
)

func (s *ServerController) preCreateLabel(ctx context.Context, req *types.CreateLabelRequest) error {
	_, err := s.factory.Label().Get(ctx, db.WithName(req.Name))
	if err == nil {
		return fmt.Errorf("标签%s已存在，无法重复创建", req.Name)
	}
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

func (s *ServerController) CreateLabel(ctx context.Context, req *types.CreateLabelRequest) error {
	if err := s.preCreateLabel(ctx, req); err != nil {
		return err
	}

	if _, err := s.factory.Label().Create(ctx, &model.Label{
		Name: req.Name,
	}); err != nil {
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
	// 初始化分页属性
	listOption.SetDefaultPageOption()

	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}
	opts := []db.Options{
		db.WithNameLike(listOption.NameSelector),
	}

	var err error
	pageResult.Total, err = s.factory.Label().Count(ctx, opts...)
	if err != nil {
		klog.Errorf("获取标签总数失败 %v", err)
		pageResult.Message = err.Error()
	}
	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithModifyOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Label().List(ctx, opts...)
	if err != nil {
		klog.Errorf("获取标签列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
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
	// 初始化分页属性
	listOption.SetDefaultPageOption()

	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}
	opts := []db.Options{
		db.WithNameLike(listOption.NameSelector),
	}

	var err error
	pageResult.Total, err = s.factory.Label().CountLogos(ctx, opts...)
	if err != nil {
		klog.Errorf("获取 Logo 总数失败 %v", err)
		pageResult.Message = err.Error()
	}
	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithModifyOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Label().ListLogos(ctx, opts...)
	if err != nil {
		klog.Errorf("获取 Logo 列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}

func (s *ServerController) ListLabelImages(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	labelIds, err := s.parseLabelIds(listOption.LabelIds)
	if err != nil {
		return nil, err
	}

	listOption.SetDefaultPageOption()
	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}
	images, total, err := s.factory.Label().ListLabelImages(ctx, labelIds, listOption.Page, listOption.Limit)
	if err != nil {
		klog.Errorf("获取 label image 失败 %v", err)
		return nil, err
	}

	pageResult.Items = images
	pageResult.Total = total
	return pageResult, nil
}

func (s *ServerController) parseLabelIds(labelIdsStr string) ([]int64, error) {
	l := strings.TrimSpace(labelIdsStr)
	if len(l) == 0 {
		return nil, fmt.Errorf("no label ids found")
	}
	parts := strings.Split(l, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("no label ids found")
	}

	var labelIds []int64
	for _, p := range parts {
		labelId, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("id %s 不合规", p)
		}
		labelIds = append(labelIds, labelId)
	}

	return labelIds, nil
}
