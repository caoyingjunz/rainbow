package rainbow

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) GetCollection(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	var (
		taskNum   int64
		imageNum  int64
		reviewNum int64
		err       error
	)

	taskNum, err = s.factory.Task().Count(ctx)
	if err != nil {
		klog.Error("获取任务数量失败: %v", err)
	}
	imageNum, err = s.factory.Image().Count(ctx)
	if err != nil {
		klog.Error("获取镜像数量失败: %v", err)
	}

	reviews, err := s.factory.Task().ListReview(ctx)
	if err != nil {
		klog.Error("获取浏览数量失败: %v", err)
	}
	for _, review := range reviews {
		reviewNum = +review.Count
	}

	return map[string]int64{
		"tasks":  taskNum,
		"images": imageNum,
		"review": reviewNum,
	}, nil
}

func (s *ServerController) AddDailyReview(ctx context.Context, page string) error {
	return s.factory.Task().AddDailyReview(ctx, &model.Daily{
		Page: page,
	})
}
