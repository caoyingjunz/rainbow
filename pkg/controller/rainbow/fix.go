package rainbow

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) Fix(ctx context.Context, req *types.FixRequest) (interface{}, error) {
	switch req.Type {
	case "image":
		return s.fixImages(ctx, req.UserId, req.Image)
	case "getImages":
		return s.getImages(ctx, req)
	case "allImages":
		return s.fixAllImages(ctx)
	}
	return nil, nil
}

func (s *ServerController) fixAllImages(ctx context.Context) (interface{}, error) {
	images, err := s.factory.Image().ListImagesWithTag(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string][]model.Image)
	for _, image := range images {
		key := fmt.Sprintf("%s|||%s|||%s|||%s", image.UserId, image.UserName, image.Namespace, image.Name)
		old, ok := m[key]
		if ok {
			m[key] = append(old, image)
		} else {
			m[key] = []model.Image{image}
		}
	}
	for k, v := range m {
		if len(v) <= 1 {
			continue
		}

		parts := strings.Split(k, "|||")
		userId, _, ns, imageName := parts[0], parts[1], parts[2], parts[3]
		_, err = s.fixImages(ctx, userId, types.FixImageSpec{
			Name:      imageName,
			Namespace: ns,
		})
		if err != nil {
			return nil, nil
		}
	}

	return nil, nil
}

func (s *ServerController) getImages(ctx context.Context, req *types.FixRequest) (interface{}, error) {
	images, err := s.factory.Image().ListImagesWithTag(ctx)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]model.Image)
	for _, image := range images {
		key := fmt.Sprintf("%s|||%s|||%s|||%s", image.UserId, image.UserName, image.Namespace, image.Name)
		old, ok := m[key]
		if ok {
			m[key] = append(old, image)
		} else {
			m[key] = []model.Image{image}
		}
	}

	for k, v := range m {
		if len(v) <= 1 {
			continue
		}

		parts := strings.Split(k, "|||")
		userId, userName, ns, imageName := parts[0], parts[1], parts[2], parts[3]
		fmt.Println("userId:", userId)
		fmt.Println("userName:", userName)
		fmt.Println("namespace", ns)
		fmt.Println("imageName:", imageName)
		fmt.Println("版本数:", len(v))
	}

	return nil, nil
}

func (s *ServerController) fixImages(ctx context.Context, userId string, imageSpec types.FixImageSpec) (interface{}, error) {
	images, err := s.factory.Image().ListImagesWithTag(ctx, db.WithUser(userId), db.WithName(imageSpec.Name), db.WithNamespace(imageSpec.Namespace))
	if err != nil {
		return nil, err
	}
	if len(images) <= 1 {
		klog.Infof("镜像: %s(%s)仅发现一个，无需订正", imageSpec.Namespace, imageSpec.Name)
		return nil, nil
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
				klog.Infof("将删除版本 %d(%s)", tag.Id, tag.Name)
				if err = s.factory.Image().DeleteTag(ctx, tag.Id); err != nil {
					return nil, err
				}
			} else {
				klog.Infof("将合并版本 %d(%s) 至 %d", tag.Id, tag.Name, ImageId)
				if err = s.factory.Image().UpdateTag(ctx, tag.ImageId, tag.Name, map[string]interface{}{
					"image_id": ImageId,
				}); err != nil {
					return nil, err
				}
			}
		}
		klog.Infof("镜像%d将将被删除", image.Id)
		if err = s.factory.Image().Delete(ctx, image.Id); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
