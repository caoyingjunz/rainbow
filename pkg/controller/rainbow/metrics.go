package rainbow

import (
	"context"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) syncMetrics(ctx context.Context) {
	now := time.Now()

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	recordDay := yesterdayStart.Format("2006-01-02")

	pullCount, err := s.factory.Image().PullAllCount(ctx)
	if err != nil {
		klog.Errorf("获取 pull count 失败 %v", err)
		return
	}
	taskCount, err := s.factory.Task().Count(ctx)
	if err != nil {
		klog.Errorf("获取日活任务失败 %v", err)
		return
	}
	imageCount, err := s.factory.Image().Count(ctx)
	if err != nil {
		klog.Errorf("获取镜像数失败 %v", err)
		return
	}
	tagCount, err := s.factory.Image().TagCount(ctx)
	if err != nil {
		return
	}
	_, err = s.factory.Metrics().Create(ctx, &model.Metrics{Pull: pullCount, Task: taskCount, Image: imageCount, RecordDay: recordDay, Tags: tagCount})
	if err != nil {
		klog.Errorf("创建日活数据失败 %v", err)
		return
	}
}

func (s *ServerController) ListMetrics(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
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
	}

	var err error
	pageResult.Total, err = s.factory.Metrics().CountMetrics(ctx, opts...)
	if err != nil {
		klog.Errorf("获取Metric总数失败 %v", err)
		pageResult.Message = err.Error()
	}
	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithModifyOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Metrics().List(ctx, opts...)
	if err != nil {
		klog.Errorf("获取Metric列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}
