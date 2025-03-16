package rainbow

import (
	"context"
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
		TaskId:     req.TaskId,
		RegisterId: req.RegisterId,
		TaskName:   req.TaskName,
		Status:     req.Status,
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

func (s *ServerController) UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error {
	if s.registryId != nil {
		reg := s.registryId
		if *reg == req.RegistryId {
			if req.Status == "同步完成" {
				parts := strings.Split(req.Target, "/")
				pps := parts[len(parts)-1]
				p := strings.Split(pps, ":")
				response, err := s.swrClient.UpdateRepo(&swrmodel.UpdateRepoRequest{
					Namespace:  "pixiu-public",
					Repository: p[0],
					Body: &swrmodel.UpdateRepoRequestBody{
						IsPublic: true,
					},
				})
				if err != nil {
					klog.Errorf("尝试设置华为仓库为 public 失败: %v", err)
				} else {
					klog.Infof("设置仓库为公开成功 %v", response)
				}
			}
		}
	}

	return s.factory.Image().UpdateDirectly(ctx, req.Name, req.TaskId, map[string]interface{}{
		"status":      req.Status,
		"message":     req.Message,
		"target":      req.Target,
		"register_id": req.RegistryId,
	})
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
		return s.factory.Image().List(ctx, db.WithTask(listOption.TaskId), db.WithUser(listOption.UserId), db.WithNameLike(listOption.NameSelector))
	}

	// TODO: 临时实现，后续再优化
	return s.factory.Image().List(ctx, db.WithStatus("同步完成"), db.WithLimit(listOption.Limits))
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
			Name:     r,
			TaskId:   req.TaskId,
			TaskName: task.Name,
			UserId:   task.UserId,
			Status:   "镜像同步中",
		})
		if err != nil {
			klog.Errorf("创建镜像失败 %v", err)
			return err
		}
	}

	return nil
}
