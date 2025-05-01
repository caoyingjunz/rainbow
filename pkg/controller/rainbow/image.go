package rainbow

import (
	"context"
	"fmt"
	"strings"
	"time"

	swrmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/model"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateImage(ctx context.Context, req *types.CreateImageRequest) error {
	_, err := s.factory.Image().Create(ctx, &model.Image{
		Name:       req.Name,
		RegisterId: req.RegisterId,
		IsPublic:   req.IsPublic,
	})
	if err != nil {
		klog.Errorf("创建镜像失败 %v", err)
	}

	return err
}

func (s *ServerController) UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error {
	updates := make(map[string]interface{})
	updates["name"] = req.Name
	updates["status"] = req.Status
	updates["message"] = req.Message
	updates["is_public"] = req.IsPublic
	return s.factory.Image().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) TryUpdateRemotePublic(ctx context.Context, req *types.UpdateImageStatusRequest, old *model.Image) error {
	if RegistryId == nil {
		return nil
	}
	reg := RegistryId
	if *reg != req.RegistryId {
		return nil
	}
	if req.Status != "同步完成" {
		return nil
	}

	name := req.Name
	namespace := "pixiu-public"
	newPublic := true

	for i := 0; i < 3; i++ {
		response, err := SwrClient.ShowRepository(&swrmodel.ShowRepositoryRequest{
			Namespace:  namespace,
			Repository: name,
		})
		if err != nil {
			klog.Errorf("获取远端镜像 %s 失败 %v", name, err)
			time.Sleep(1 * time.Second)
			continue
		}

		oldPublic := *response.IsPublic
		if oldPublic == newPublic {
			return nil
		}

		_, err = SwrClient.UpdateRepo(&swrmodel.UpdateRepoRequest{
			Namespace:  namespace,
			Repository: name,
			Body:       &swrmodel.UpdateRepoRequestBody{IsPublic: newPublic},
		})
		if err != nil {
			klog.Errorf("更新远端镜像 %s 失败 %v", name, err)
			time.Sleep(1 * time.Second)
			continue
		} else {
			_ = s.factory.Image().Update(ctx, req.ImageId, old.ResourceVersion, map[string]interface{}{"public_updated": true})
			return nil
		}
	}

	return fmt.Errorf("更新 更新远端镜像 %s 失败", name)
}

func (s *ServerController) UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error {
	old, err := s.factory.Image().Get(ctx, req.ImageId, false)
	if err != nil {
		klog.Errorf("获取镜像(%d)失败: %v", req.ImageId, err)
		return err
	}

	if !old.PublicUpdated {
		if err := s.TryUpdateRemotePublic(ctx, req, old); err != nil {
			klog.Errorf("尝试设置华为仓库为 public 失败: %v", err)
		}
	}

	parts := strings.Split(req.Target, ":")
	tag := parts[1]

	return s.factory.Image().UpdateTag(ctx, req.ImageId, tag, map[string]interface{}{"status": req.Status, "message": req.Message})
}

// DeleteImage 删除镜像和对应的tags
func (s *ServerController) DeleteImage(ctx context.Context, imageId int64) error {
	if err := s.factory.Image().Delete(ctx, imageId); err != nil {
		return fmt.Errorf("删除镜像 %d 失败 %v", imageId, err)
	}

	delImage, err := s.factory.Image().Get(ctx, imageId, true)
	if err != nil {
		klog.Errorf("获取已删除镜像 %d 失败: %v", imageId, s, err)
		return nil
	}

	if !s.isDefaultRepo(delImage.RegisterId) {
		return nil
	}

	_, err = SwrClient.DeleteRepo(&swrmodel.DeleteRepoRequest{
		Namespace:  HuaweiNamespace,
		Repository: delImage.Name,
	})
	if err != nil {
		klog.Warningf("删除远端镜像失败 %v", err)
	}

	return nil
}

func (s *ServerController) ListImages(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	if listOption.Limits == 0 {
		parts := make([]string, 0)
		if len(listOption.LabelSelector) != 0 {
			parts = strings.Split(listOption.LabelSelector, ",")
		}
		return s.factory.Image().ListImagesWithTag(ctx, db.WithUser(listOption.UserId), db.WithNameLike(listOption.NameSelector), db.WithLabelIn(parts...))
	}

	// TODO: 临时实现，后续再优化
	return s.factory.Image().ListImagesWithTag(ctx, db.WithStatus("同步完成"), db.WithLimit(listOption.Limits))
}

func (s *ServerController) isDefaultRepo(regId int64) bool {
	return regId == *RegistryId
}

func (s *ServerController) GetImage(ctx context.Context, imageId int64) (interface{}, error) {
	object, err := s.factory.Image().Get(ctx, imageId, false)
	if err != nil {
		return nil, err
	}

	if !s.isDefaultRepo(object.RegisterId) {
		return object, nil
	}

	req := &swrmodel.ShowRepositoryRequest{}
	req.Namespace = object.Namespace
	req.Repository = object.Name
	resp2, err := SwrClient.ShowRepository(req)
	if err != nil {
		return object, nil
	}
	object.Pull = *resp2.NumDownload

	request := &swrmodel.ListRepositoryTagsRequest{}
	request.Namespace = object.Namespace
	request.Repository = object.Name
	resp, err := SwrClient.ListRepositoryTags(request)
	if err != nil {
		return object, nil
	}

	tags := *resp.Body
	tagMap := make(map[string]swrmodel.ShowReposTagResp)
	for _, tag := range tags {
		object.Size = object.Size + tag.Size
		tagMap[tag.Tag] = tag
	}

	objTags := object.Tags
	for i, oldTag := range objTags {
		name := oldTag.Name
		exists, ok := tagMap[name]
		if !ok {
			continue
		}
		if oldTag.Status != types.SyncImageComplete {
			continue
		}

		oldTag.Size = exists.Size
		oldTag.Message = exists.Manifest
		objTags[i] = oldTag
	}
	object.Tags = objTags
	return object, nil
}

func (s *ServerController) CreateImages(ctx context.Context, req *types.CreateImagesRequest) error {
	task, err := s.factory.Task().Get(ctx, req.TaskId)
	if err != nil {
		klog.Errorf("未传任务名，通过任务ID获取任务详情失败 %v", err)
		return err
	}

	for _, r := range req.Names {
		_, err = s.factory.Image().Create(ctx, &model.Image{
			Name:   r,
			UserId: task.UserId,
		})
		if err != nil {
			klog.Errorf("创建镜像失败 %v", err)
			return err
		}
	}

	return nil
}

func (s *ServerController) DeleteImageTag(ctx context.Context, imageId int64, name string) error {
	err := s.factory.Image().DeleteTag(ctx, imageId, name)
	if err != nil {
		klog.Errorf("删除镜像(%d) tag %s 失败:%v", imageId, name, err)
		return fmt.Errorf("删除镜像(%d) tag %s 失败:%v", imageId, name, err)
	}

	delTag, err := s.factory.Image().GetTag(ctx, imageId, name, true)
	if err != nil {
		klog.Errorf("获取已删除镜像(%d)的tag(%s) 失败: %v", imageId, name, err)
		return nil
	}
	image, err := s.factory.Image().Get(ctx, imageId, false)
	if err != nil {
		klog.Errorf("获取镜像(%d)的失败: %v", imageId, err)
		return nil
	}

	if !s.isDefaultRepo(image.RegisterId) {
		return nil
	}

	request := &swrmodel.DeleteRepoTagRequest{
		Namespace:  HuaweiNamespace,
		Repository: image.Name,
		Tag:        delTag.Name,
	}

	_, err = SwrClient.DeleteRepoTag(request)
	if err != nil {
		klog.Errorf("删除远程镜像 %s tag(%s) 失败 %v", image.Name, delTag.Name, err)
	}

	return nil
}
