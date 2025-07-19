package rainbow

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) ListKubernetesVersions(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	// 初始化分页属性
	listOption.SetDefaultPageOption()
	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}

	var err error
	// 先获取总数
	pageResult.Total, err = s.factory.Task().GetKubernetesVersionCount(ctx)
	if err != nil {
		klog.Errorf("获取镜像总数失败 %v", err)
		pageResult.Message = err.Error()
	}

	offset := (listOption.Page - 1) * listOption.Limit
	opts := []db.Options{
		db.WithOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	} // 先写条件，再写排序，再偏移，再设置每页数量
	pageResult.Items, err = s.factory.Task().ListKubernetesVersions(ctx, opts...)
	if err != nil {
		klog.Errorf("获取镜像列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}

func (s *ServerController) SyncKubernetesVersions(ctx context.Context) (interface{}, error) {
	return nil, nil
}
