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
	updates["task_id"] = req.TaskId
	updates["task_name"] = req.TaskName
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
	old, err := s.factory.Image().Get(ctx, req.ImageId)
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

// SoftDeleteImage 软删除
func (s *ServerController) SoftDeleteImage(ctx context.Context, imageId int64) error {
	old, err := s.factory.Image().Get(ctx, imageId)
	if err != nil {
		return err
	}
	if old.IsDeleted {
		return nil
	}

	return s.factory.Image().Update(ctx, imageId, old.ResourceVersion, map[string]interface{}{
		"gmt_deleted": time.Now(),
		"is_deleted":  true,
	})
}

func (s *ServerController) ListImages(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	if listOption.Limits == 0 {
		return s.factory.Image().ListImagesWithTag(ctx, db.WithTask(listOption.TaskId), db.WithUser(listOption.UserId), db.WithNameLike(listOption.NameSelector))
	}

	// TODO: 临时实现，后续再优化
	return s.factory.Image().ListImagesWithTag(ctx, db.WithStatus("同步完成"), db.WithLimit(listOption.Limits))
}

func (s *ServerController) GetImage(ctx context.Context, imageId int64) (interface{}, error) {
	return s.factory.Image().Get(ctx, imageId)
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
