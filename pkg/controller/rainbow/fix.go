package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"k8s.io/klog/v2"
)

func (s *ServerController) Fix(ctx context.Context, req *types.FixRequest) error {
	switch req.Type {
	case "image":
		return s.fixImages(ctx, req.UserId, req.Image)

	}
	return nil
}

func (s *ServerController) fixImages(ctx context.Context, userId string, imageSpec types.FixImageSpec) error {
	images, err := s.factory.Image().ListImagesWithTag(ctx, db.WithUser(userId), db.WithName(imageSpec.Name), db.WithNamespace(imageSpec.Namespace))
	if err != nil {
		return err
	}
	if len(images) <= 1 {
		klog.Infof("镜像: %s(%s)仅发现一个，无需订正", imageSpec.Namespace, imageSpec.Name)
		return nil
	}

	retainImage := images[0]
	ImageId := retainImage.Id
	oldTagMap := make(map[string]bool)
	for _, oldTag := range retainImage.Tags {
		oldTagMap[oldTag.Name] = true
	}

	for _, image := range images[1:] {
		klog.Infof("镜像%s(%s) id(%d) 将被移除", image.Namespace, image.Name, image.Id)
		for _, tag := range image.Tags {
			// 如果已经在老的tag里，则删除，如果不在老的里，则更新到老的里
			if oldTagMap[tag.Name] {
				klog.Infof("将删除版本 %s", tag.Name)
			} else {
				klog.Infof("将合并版本 %s 至 %d", tag.Name, ImageId)
			}
		}
	}

	return nil
}
