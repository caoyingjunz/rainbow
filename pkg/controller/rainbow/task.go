package rainbow

import (
	"context"
	"fmt"
	"strings"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
)

const (
	TaskWaitStatus  = "等待执行"
	HuaweiNamespace = "pixiu-public"
)

func (s *ServerController) preCreateTask(ctx context.Context, req *types.CreateTaskRequest) error {
	if req.Type == 1 {
		if !strings.HasPrefix(req.KubernetesVersion, "v1.") {
			return fmt.Errorf("invaild kubernetes version (%s)", req.KubernetesVersion)
		}
	} else {
		var errs []error
		// TODO: 其他不合规检查
		for _, image := range req.Images {
			if strings.Contains(image, "\"") {
				errs = append(errs, fmt.Errorf("invaild image(%s)", image))
			}
		}
		return utilerrors.NewAggregate(errs)
	}

	return nil
}

const (
	defaultNamespace = "emptyNamespace"
)

func (s *ServerController) CreateTask(ctx context.Context, req *types.CreateTaskRequest) error {
	if err := s.preCreateTask(ctx, req); err != nil {
		klog.Errorf("创建任务前置检查未通过 %v", err)
		return err
	}

	if req.RegisterId == 0 {
		req.RegisterId = *RegistryId
	}

	if len(req.Namespace) == 0 {
		req.Namespace = strings.ToLower(req.UserName)
	}
	if req.Namespace == defaultNamespace {
		req.Namespace = ""
	}

	object, err := s.factory.Task().Create(ctx, &model.Task{
		Name:              req.Name,
		UserId:            req.UserId,
		UserName:          req.UserName,
		RegisterId:        req.RegisterId,
		AgentName:         req.AgentName,
		Mode:              req.Mode,
		Status:            TaskWaitStatus,
		Type:              req.Type,
		KubernetesVersion: req.KubernetesVersion,
		Driver:            req.Driver,
		Namespace:         req.Namespace,
		IsPublic:          req.PublicImage,
		Logo:              req.Logo,
		IsOfficial:        req.IsOfficial,
	})
	if err != nil {
		return err
	}
	if req.Type == 1 {
		klog.Infof("创建任务(%s)成功，状态为延迟执行", req.Name)
		return nil
	}

	taskId := object.Id
	if err = s.CreateImageWithTag(ctx, taskId, req); err != nil {
		_ = s.DeleteTaskWithImages(ctx, taskId)
		return fmt.Errorf("failed to create tasks images %v", err)
	}

	return nil
}

func (s *ServerController) CreateImageWithTag(ctx context.Context, taskId int64, req *types.CreateTaskRequest) error {
	klog.Infof("使用镜像仓库(%d)", req.RegisterId)
	reg, err := s.factory.Registry().Get(ctx, req.RegisterId)
	if err != nil {
		return fmt.Errorf("获取仓库(%d)失败 %v", req.RegisterId, err)
	}

	imageMap := make(map[string][]string)
	for _, i := range util.TrimAndFilter(req.Images) {
		path, tag, err := ParseImageItem(i)
		if err != nil {
			return fmt.Errorf("failed to parse image %v", err)
		}

		old, ok := imageMap[path]
		if ok {
			old = append(old, tag)
			imageMap[path] = old
		} else {
			imageMap[path] = []string{tag}
		}
	}

	namespace := req.Namespace
	for path, tags := range imageMap {
		var imageId int64
		parts2 := strings.Split(path, "/")
		name := parts2[len(parts2)-1]
		if len(name) == 0 {
			return fmt.Errorf("不合规镜像名称 %s", path)
		}
		if req.RegisterId == *RegistryId {
			if len(namespace) != 0 {
				name = namespace + "/" + name
			}
		}

		mirror := reg.Repository + "/" + reg.Namespace + "/" + name

		oldImage, err := s.factory.Image().GetByPath(ctx, path, mirror)
		if err != nil {
			// 镜像不存在，则先创建镜像
			if errors.IsNotFound(err) {
				newImage, err := s.factory.Image().Create(ctx, &model.Image{
					UserId:     req.UserId,
					UserName:   req.UserName,
					RegisterId: req.RegisterId,
					Namespace:  namespace,
					Logo:       req.Logo,
					Name:       name,
					Path:       path,
					Mirror:     mirror,
					IsPublic:   req.PublicImage,
					IsOfficial: req.IsOfficial,
				})
				if err != nil {
					return err
				}
				imageId = newImage.Id
			} else {
				return err
			}
		} else {
			imageId = oldImage.Id
		}

		for _, tag := range tags {
			_, err = s.factory.Image().CreateTag(ctx, &model.Tag{
				Path:    path,
				ImageId: imageId,
				TaskId:  taskId,
				Name:    tag,
				Status:  types.SyncImageInitializing,
			})
			if err != nil {
				klog.Errorf("创建镜像(%s)的版本(%s)失败 %v", path, tag, err)
				return err
			}
		}
	}

	return nil
}

func ParseImageItem(image string) (string, string, error) {
	parts := strings.Split(image, ":")
	if len(parts) > 2 || len(parts) == 0 {
		return "", "", fmt.Errorf("不合规镜像名称 %s", image)
	}

	path := parts[0]
	tag := "latest"
	if len(parts) == 2 {
		tag = parts[1]
	}

	// 如果镜像是以 docker.io 开关，则去除 docker.io
	if strings.HasPrefix(path, "docker.io/") {
		path = strings.Replace(path, "docker.io/", "", 1)
	}

	return path, tag, nil
}

func (s *ServerController) UpdateTask(ctx context.Context, req *types.UpdateTaskRequest) error {
	if err := s.factory.Task().Update(ctx, req.Id, req.ResourceVersion, map[string]interface{}{
		"register_id": req.RegisterId,
		"mode":        req.Mode,
	}); err != nil {
		return err
	}

	old, err := s.factory.Image().ListWithTask(ctx, req.Id)
	if err != nil {
		return err
	}
	var oldImages []string
	for _, o := range old {
		oldImages = append(oldImages, o.Name)
	}
	oldImageMap := sets.NewString(oldImages...)

	var addImages []string
	for _, n := range req.Images {
		if oldImageMap.Has(n) {
			continue
		}
		addImages = append(addImages, n)
	}
	var images []model.Image
	trimAddImages := util.TrimAndFilter(req.Images)
	for _, i := range trimAddImages {
		images = append(images, model.Image{
			Name: i,
		})
	}
	if err = s.factory.Image().CreateInBatch(ctx, images); err != nil {
		return fmt.Errorf("failed to create tasks images %v", err)
	}

	return nil
}

func (s *ServerController) ListTasks(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	return s.factory.Task().List(ctx, db.WithUser(listOption.UserId), db.WithNameLike(listOption.NameSelector), db.WithOrderByDesc())
}

func (s *ServerController) UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error {
	return s.factory.Task().UpdateDirectly(ctx, req.TaskId, map[string]interface{}{"status": req.Status, "message": req.Message, "process": req.Process})
}

func (s *ServerController) DeleteTask(ctx context.Context, taskId int64) error {
	return s.factory.Task().Delete(ctx, taskId)
}

func (s *ServerController) DeleteTaskWithImages(ctx context.Context, taskId int64) error {
	_ = s.factory.Task().Delete(ctx, taskId)
	return nil
}

func (s *ServerController) GetTask(ctx context.Context, taskId int64) (interface{}, error) {
	return s.factory.Task().Get(ctx, taskId)
}

func (s *ServerController) ListTaskImages(ctx context.Context, taskId int64) (interface{}, error) {
	return s.factory.Image().ListTags(ctx, db.WithTask(taskId))
}
