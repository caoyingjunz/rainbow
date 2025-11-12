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
	updates["label"] = req.Label
	updates["logo"] = req.Logo
	updates["description"] = req.Description
	updates["is_public"] = req.IsPublic
	updates["is_locked"] = req.IsLocked
	return s.factory.Image().Update(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) TryUpdateRemotePublic(ctx context.Context, req *types.UpdateImageStatusRequest, old *model.Image) error {
	if !s.isDefaultRepo(req.RegistryId) {
		return nil
	}

	if req.Status != types.SyncImageComplete {
		klog.Infof("镜像(%s)状态未推送完成，暂时不需要设置为公开", req.Name)
		return nil
	}

	name := req.Name
	for i := 0; i < 3; i++ {
		if strings.Contains(name, "/") {
			name = strings.ReplaceAll(name, "/", "$")
		}

		klog.Infof("尝试更新镜像 %s 已经为 public", name)
		resp, err := SwrClient.ShowRepository(&swrmodel.ShowRepositoryRequest{Namespace: HuaweiNamespace, Repository: name})
		if err != nil {
			klog.Errorf("获取远端镜像 %s 失败 %v 1s后进行下一次重试", name, err)
			time.Sleep(1 * time.Second)
			continue
		}
		if *resp.IsPublic {
			klog.Infof("镜像%s 已经为 public，无需更新", req.Name)
			return nil
		}

		_, err = SwrClient.UpdateRepo(&swrmodel.UpdateRepoRequest{Namespace: HuaweiNamespace, Repository: name, Body: &swrmodel.UpdateRepoRequestBody{IsPublic: true}})
		if err != nil {
			klog.Errorf("更新远端镜像 %s 失败 %v", name, err)
			time.Sleep(1 * time.Second)
			continue
		} else {
			_ = s.factory.Image().Update(ctx, req.ImageId, old.ResourceVersion, map[string]interface{}{"public_updated": true})
			klog.Infof("镜像 %s 已设置为 public", req.Name)
			return nil
		}
	}

	return fmt.Errorf("更新远端镜像 %s 失败", name)
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
	} else {
		klog.Infof("镜像(%s)已更新过，跳过远程更新", old.Name)
	}

	parts := strings.Split(req.Target, ":")
	tag := parts[1]
	if err = s.factory.Image().UpdateTag(ctx, req.ImageId, tag, map[string]interface{}{"status": req.Status, "message": req.Message}); err != nil {
		klog.Errorf("更新镜像(%d)的版本(%d)状态失败:%v", req.ImageId, tag, err)
		return err
	}

	// 当状态已经变成完成时，更新镜像的修改时间
	if req.Status == types.SyncImageComplete {
		targetName := old.Name
		if strings.Contains(targetName, "/") {
			targetName = strings.ReplaceAll(targetName, "/", "$")
		}
		// 获取版本大小并同步
		newTag, err := SwrClient.ShowRepoTag(&swrmodel.ShowRepoTagRequest{
			Namespace:  HuaweiNamespace,
			Repository: targetName,
			Tag:        tag,
		})
		if err != nil {
			klog.Warningf("获取远端新版本信息失败 %v", err)
			return nil
		}

		updates := make(map[string]interface{})
		updates["size"] = *newTag.Size
		updates["read_size"] = ByteSizeSimple(*newTag.Size)
		updates["digest"] = newTag.Digest
		updates["manifest"] = newTag.Manifest
		if err = s.factory.Image().UpdateTag(ctx, req.ImageId, tag, updates); err != nil {
			klog.Warningf("更新镜像版本失败 %v", err)
			return nil
		}

		// 重新镜像大小
		if err = s.CalculateImageSize(ctx, req.ImageId); err != nil {
			klog.Warningf("计算镜像大小失败 %v", err)
			return nil
		}
	}
	return nil
}

func (s *ServerController) CalculateImageSize(ctx context.Context, imageId int64) error {
	imageInfo, err := s.factory.Image().Get(ctx, imageId, false)
	if err != nil {
		return err
	}

	var total int64
	for _, tag := range imageInfo.Tags {
		total = total + tag.Size
	}
	return s.factory.Image().Update(ctx, imageId, imageInfo.ResourceVersion, map[string]interface{}{
		"size":      total,
		"read_size": ByteSizeSimple(total),
	})
}

// DeleteImage 删除镜像和对应的tags
func (s *ServerController) DeleteImage(ctx context.Context, imageId int64) error {
	// 获取镜像信息，检查是否被锁定
	image, err := s.factory.Image().Get(ctx, imageId, false)
	if err != nil {
		klog.Errorf("获取镜像(%d)失败: %v", imageId, err)
	}
	// 检查 Lock 字段，如果为 true 则不允许删除
	if image.IsLocked {
		return fmt.Errorf("镜像被锁定，不允许删除")
	}

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

	name := delImage.Name
	if strings.Contains(name, "/") {
		name = strings.ReplaceAll(name, "/", "$")
	}
	_, err = SwrClient.DeleteRepo(&swrmodel.DeleteRepoRequest{
		Namespace:  HuaweiNamespace,
		Repository: name,
	})
	if err != nil {
		klog.Warningf("删除远端镜像失败 %v", err)
	}

	return nil
}

func (s *ServerController) ListImages(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	// 初始化分页属性
	listOption.SetDefaultPageOption()

	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}

	opts := []db.Options{ // 先写条件，再写排序，再偏移，再设置每页数量
		db.WithUser(listOption.UserId),
		db.WithNameLike(listOption.NameSelector),
		db.WithNamespace(listOption.Namespace),
	}
	var err error

	// 先获取总数
	pageResult.Total, err = s.factory.Image().Count(ctx, opts...)
	if err != nil {
		klog.Errorf("获取镜像总数失败 %v", err)
		pageResult.Message = err.Error()
	}

	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithModifyOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Image().ListImagesWithTag(ctx, opts...)
	if err != nil {
		klog.Errorf("获取镜像列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}

func (s *ServerController) ListPublicImages(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	return s.factory.Image().List(ctx, db.WithPublic(), db.WithNameLike(listOption.NameSelector), db.WithLimit(listOption.Limit))
}

func (s *ServerController) ListImagesByIds(ctx context.Context, ids []int64) (interface{}, error) {
	return s.factory.Image().List(ctx, db.WithIDIn(ids...))
}

func (s *ServerController) DeleteImagesByIds(ctx context.Context, ids []int64) error {
	return s.factory.Image().DeleteInBatch(ctx, ids)
}

func (s *ServerController) isDefaultRepo(regId int64) bool {
	return regId == *RegistryId
}

func IsDurationExceeded(t time.Time, duration time.Duration) bool {
	now := time.Now()
	return now.Sub(t) > duration
}

func (s *ServerController) GetImage(ctx context.Context, imageId int64) (interface{}, error) {
	object, err := s.factory.Image().Get(ctx, imageId, false)
	if err != nil {
		return nil, err
	}

	// 如果 10 分钟内已更新，则直接返回
	if !IsDurationExceeded(object.LastSyncTime, 10*time.Minute) {
		klog.Infof("镜像(%s/%s) 10分钟内进行过同步，无需重复执行", object.Namespace, object.Name)
		return object, nil
	}
	// 非官方内置仓库，无需更新
	if !s.isDefaultRepo(object.RegisterId) {
		return object, nil
	}

	return s.GetAndUpdateByRemoteImage(ctx, object)
}

func (s *ServerController) UpdateTagFromRemote(ctx context.Context, newTag swrmodel.ShowReposTagResp, oldTag model.Tag) error {
	updates := make(map[string]interface{})
	if oldTag.Size != newTag.Size {
		updates["size"] = newTag.Size
		updates["read_size"] = ByteSizeSimple(newTag.Size)
	}
	if oldTag.Digest != newTag.Digest {
		updates["digest"] = newTag.Digest
	}
	if oldTag.Manifest != newTag.Manifest {
		updates["manifest"] = newTag.Manifest
	}

	if len(updates) == 0 {
		return nil
	}
	return s.factory.Image().UpdateTag(ctx, oldTag.ImageId, oldTag.Name, updates)
}

func (s *ServerController) UpdateImageInfoFromRemote(ctx context.Context, newImage *swrmodel.ShowRepositoryResponse, old *model.Image) error {
	updates := make(map[string]interface{})
	if old.Pull != *newImage.NumDownload {
		updates["pull"] = *newImage.NumDownload
	}
	if old.Size != *newImage.Size {
		updates["size"] = *newImage.Size
		updates["read_size"] = ByteSizeSimple(*newImage.Size)
	}

	if len(updates) == 0 {
		return nil
	}

	updates["last_sync_time"] = time.Now()
	return s.factory.Image().Update(ctx, old.Id, old.ResourceVersion, updates)
}

// GetAndUpdateByRemoteImage 获取远端镜像属性且保存
func (s *ServerController) GetAndUpdateByRemoteImage(ctx context.Context, object *model.Image) (interface{}, error) {
	targetName := object.Name
	if strings.Contains(targetName, "/") {
		targetName = strings.ReplaceAll(targetName, "/", "$")
	}
	oldTagMap := make(map[string]model.Tag)
	for _, t := range object.Tags {
		oldTagMap[t.Name] = t
	}

	// 获取并更新版本
	resp, err := SwrClient.ListRepositoryTags(&swrmodel.ListRepositoryTagsRequest{Namespace: HuaweiNamespace, Repository: targetName})
	if err != nil {
		klog.Errorf("获取远端镜像版本失败 %v", err)
		return object, nil
	}
	tags := *resp.Body
	for _, t := range tags {
		name := t.Tag
		old, ok := oldTagMap[name]
		if !ok {
			klog.Infof("远端镜像(%s)的版本(%s)未收录，忽略", targetName, name)
			continue
		}
		if err = s.UpdateTagFromRemote(ctx, t, old); err != nil {
			klog.Errorf("更新远端镜像版本至本地失败", err)
			return object, err
		}
	}

	// 获取远端镜像属性
	newImage, err := SwrClient.ShowRepository(&swrmodel.ShowRepositoryRequest{Namespace: HuaweiNamespace, Repository: targetName})
	if err != nil {
		klog.Errorf("获取远端镜像详情失败 %v", err)
		return object, nil
	}
	if err = s.UpdateImageInfoFromRemote(ctx, newImage, object); err != nil {
		return object, err
	}

	// 获取最新的属性
	return s.factory.Image().Get(ctx, object.Id, false)
}

func (s *ServerController) CreateImages(ctx context.Context, req *types.CreateImagesRequest) ([]model.Image, error) {
	klog.Infof("CreateImages, req %v", req)
	task, err := s.factory.Task().Get(ctx, req.TaskId)
	if err != nil {
		klog.Errorf("未传任务名，通过任务ID获取任务详情失败 %v", err)
		return nil, err
	}
	taskReq := &types.CreateTaskRequest{
		RegisterId:  task.RegisterId,
		Images:      req.Names,
		UserName:    task.UserName,
		UserId:      task.UserId,
		Namespace:   task.Namespace,
		PublicImage: task.IsPublic,
		IsOfficial:  task.IsOfficial,
		Logo:        task.Logo,
	}
	if err := s.CreateImageWithTag(ctx, req.TaskId, taskReq); err != nil {
		klog.Errorf("创建k8s镜像记录失败 :%v", err)
		return nil, fmt.Errorf("创建k8s镜像记录失败 :%v", err)
	}

	tags, err := s.factory.Image().ListTags(ctx, db.WithTaskLike(req.TaskId))
	if err != nil {
		klog.Errorf("获取k8s镜像tags失败 :%v", err)
		return nil, fmt.Errorf("获取k8s镜像tags失败 :%v", err)
	}
	klog.Infof("已完成k8s镜像创建 %v", tags)

	var imageIds []int64
	for _, tag := range tags {
		imageIds = append(imageIds, tag.ImageId)
	}

	images, err := s.factory.Image().List(ctx, db.WithIDIn(imageIds...))
	if err != nil {
		klog.Errorf("获取已创建的k8s镜像列表失败 :%v", err)
		return nil, fmt.Errorf("获取已创建的k8s镜像列表失败 :%v", err)
	}
	klog.Infof("创建 k8s 镜像成功, 镜像列表为 %v", images)

	return images, nil
}

func (s *ServerController) DeleteImageTag(ctx context.Context, imageId int64, tagId int64) error {
	err := s.factory.Image().DeleteTag(ctx, tagId)
	if err != nil {
		return fmt.Errorf("删除镜像(%d) tag %s 失败:%v", imageId, tagId, err)
	}

	delTag, err := s.factory.Image().GetTag(ctx, tagId, true)
	if err != nil {
		klog.Errorf("获取已删除镜像(%d)的tag(%s) 失败: %v", imageId, tagId, err)
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

	imageName := image.Name
	if strings.Contains(imageName, "/") {
		imageName = strings.ReplaceAll(imageName, "/", "$")
	}
	request := &swrmodel.DeleteRepoTagRequest{
		Namespace:  HuaweiNamespace,
		Repository: imageName,
		Tag:        delTag.Name,
	}
	_, err = SwrClient.DeleteRepoTag(request)
	if err != nil {
		klog.Errorf("删除远程镜像 %s tag(%s) 失败 %v", image.Name, delTag.Name, err)
	}

	return s.CalculateImageSize(ctx, imageId)
}

func (s *ServerController) CreateNamespace(ctx context.Context, req *types.CreateNamespaceRequest) error {
	_, err := s.factory.Image().CreateNamespace(ctx, &model.Namespace{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		klog.Errorf("创建镜像的命名空间失败 %v", err)
	}
	return nil
}

func (s *ServerController) DeleteNamespace(ctx context.Context, objectId int64) error {
	if err := s.factory.Image().DeleteNamespace(ctx, objectId); err != nil {
		return fmt.Errorf("删除命名空间 %d 失败 %v", objectId, err)
	}

	return nil
}

func (s *ServerController) UpdateNamespace(ctx context.Context, req *types.UpdateNamespaceRequest) error {
	updates := make(map[string]interface{})
	updates["description"] = req.Description
	return s.factory.Image().UpdateNamespace(ctx, req.Id, req.ResourceVersion, updates)
}

func (s *ServerController) ListNamespaces(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	return s.factory.Image().ListNamespaces(ctx)
}
