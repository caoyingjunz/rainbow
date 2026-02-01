package rainbow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
	"github.com/caoyingjunz/rainbow/pkg/util/uuid"
)

func (s *ServerController) ListSubscribeMessages(ctx context.Context, subId int64) (interface{}, error) {
	return s.factory.Task().ListSubscribeMessages(ctx, db.WithSubscribe(subId))
}

func (s *ServerController) GetSubscribe(ctx context.Context, subId int64) (interface{}, error) {
	return s.factory.Task().GetSubscribe(ctx, subId)
}

func (s *ServerController) RunSubscribeImmediately(ctx context.Context, req *types.UpdateSubscribeRequest) error {
	sub, err := s.factory.Task().GetSubscribe(ctx, req.Id)
	if err != nil {
		return err
	}
	if !sub.Enable {
		klog.Warningf("订阅已被关闭")
		return errors.ErrDisableStatus
	}

	changed, err := s.subscribe(ctx, *sub)
	if err != nil {
		klog.Errorf("执行订阅(%d)失败 %v", req.Id, err)
		return err
	}
	if !changed {
		return errors.ErrImageNotFound
	}
	return nil
}

func (s *ServerController) DisableSubscribeWithMessage(ctx context.Context, sub model.Subscribe, msg string) {
	if err := s.factory.Task().UpdateSubscribeDirectly(ctx, sub.Id, map[string]interface{}{
		"enable": false,
	}); err != nil {
		klog.Errorf("自动关闭订阅失败 %v", err)
		return
	}
	if err := s.factory.Task().CreateSubscribeMessage(ctx, &model.SubscribeMessage{
		SubscribeId: sub.Id,
		Message:     msg,
	}); err != nil {
		klog.Errorf("创建订阅限制事件失败 %v", err)
	}
}

func (s *ServerController) CreateSubscribeMessageAndFailTimesAdd(ctx context.Context, sub model.Subscribe, msg string) {
	if err := s.factory.Task().UpdateSubscribeDirectly(ctx, sub.Id, map[string]interface{}{
		"fail_times": sub.FailTimes + 1,
	}); err != nil {
		klog.Errorf("订阅次数+1失败 %v", err)
	}

	if err := s.factory.Task().CreateSubscribeMessage(ctx, &model.SubscribeMessage{
		SubscribeId: sub.Id,
		Message:     msg,
	}); err != nil {
		klog.Errorf("创建订阅限制事件失败 %v", err)
	}
}

func (s *ServerController) CreateSubscribeMessageWithLog(ctx context.Context, sub model.Subscribe, msg string) {
	if err := s.factory.Task().CreateSubscribeMessage(ctx, &model.SubscribeMessage{
		SubscribeId: sub.Id,
		Message:     msg,
	}); err != nil {
		klog.Errorf("创建订阅限制事件失败 %v", err)
	}
}

func (s *ServerController) cleanSubscribeMessages(ctx context.Context, subId int64, retains int) error {
	return s.factory.Task().DeleteSubscribeMessage(ctx, subId)
}

func (s *ServerController) reRunSubscribeTags(ctx context.Context, errTags []model.Tag) error {
	taskIds := make([]string, 0)
	for _, errTag := range errTags {
		parts := strings.Split(errTag.TaskIds, ",")
		for _, p := range parts {
			taskIds = append(taskIds, p)
		}
	}

	tasks, err := s.factory.Task().List(ctx, db.WithIDStrIn(taskIds...))
	if err != nil {
		return err
	}
	for _, t := range tasks {
		klog.Infof("任务(%s)即将触发异常重新推送", t.Name)
		if err = s.ReRunTask(ctx, &types.UpdateTaskRequest{
			Id:              t.Id,
			ResourceVersion: t.ResourceVersion,
			OnlyPushError:   true,
		}); err != nil {
			return err
		}
	}

	return nil
}

// 1. 获取本地已存在的镜像版本
// 2. 获取远端镜像版本列表
// 3. 同步差异镜像版本
func (s *ServerController) subscribe(ctx context.Context, sub model.Subscribe) (bool, error) {
	//if sub.Rewrite {
	//	return s.subscribeAll(ctx, sub)
	//}
	//return s.subscribeDiff(ctx, sub)

	return s.subscribeAll(ctx, sub)
}

func (s *ServerController) subscribeAll(ctx context.Context, sub model.Subscribe) (bool, error) {
	var ns, repo string
	parts := strings.Split(sub.RawPath, "/")
	if len(parts) == 2 || len(parts) == 3 {
		ns, repo = parts[len(parts)-2], parts[len(parts)-1]
	}

	size := sub.Size
	if size > 10 {
		size = 10 // 最大并发是 10，
	}

	tagResp, err := s.SearchRepositoryTags(ctx, types.RemoteTagSearchRequest{
		Hub:        sub.ImageFrom,
		Namespace:  ns,
		Repository: repo,
		Query:      sub.Policy,
		Page:       1,
		PageSize:   size,
	})

	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			klog.Warningf("订阅镜像(%s)不存在，即将关闭订阅", sub.Path)
			s.DisableSubscribeWithMessage(ctx, sub, fmt.Sprintf("订阅镜像(%s)不存在，自动关闭", sub.Path))
			return false, fmt.Errorf("订阅镜像(%s)不存在", sub.Path)
		} else {
			klog.Errorf("获取仓库(%s)的(%s)镜像 %v", sub.ImageFrom, sub.Path, err)
			return false, err
		}
	}

	commonSearchTagResult := tagResp.(types.CommonSearchTagResult)
	if len(commonSearchTagResult.TagResult) == 0 {
		s.DisableSubscribeWithMessage(ctx, sub, fmt.Sprintf("订阅镜像的版本(%s)不存在，即将关闭订阅", sub.Policy))
		return false, fmt.Errorf("订阅镜像(%s)的版本不存在", sub.Path)
	}

	// 构造镜像列表，
	var images []string
	for _, tag := range commonSearchTagResult.TagResult {
		image := fmt.Sprintf("%s:%s", sub.RawPath, tag.Name)
		images = append(images, image)
	}
	// 创建同步任务
	if err = s.CreateTask(ctx, &types.CreateTaskRequest{
		Name:         uuid.NewRandName("", 8),
		UserId:       sub.UserId,
		UserName:     sub.UserName,
		RegisterId:   sub.RegisterId,
		Namespace:    sub.Namespace,
		Images:       images,
		OwnerRef:     1,
		SubscribeId:  sub.Id,
		Driver:       types.SkopeoDriver,
		PublicImage:  true,
		Architecture: sub.Arch,
	}); err != nil {
		klog.Errorf("创建订阅任务失败 %v", err)
		return false, err
	}

	s.CreateSubscribeMessageWithLog(ctx, sub, "订阅执行成功")
	updates := make(map[string]interface{})
	updates["last_notify_time"] = time.Now()
	if err = s.factory.Task().UpdateSubscribe(ctx, sub.Id, sub.ResourceVersion, updates); err != nil {
		klog.Infof("订阅 (%s) 更新失败 %v", sub.Path, err)
	}

	return true, nil
}

func (s *ServerController) subscribeDiff(ctx context.Context, sub model.Subscribe) (bool, error) {
	exists, err := s.factory.Image().ListImagesWithTag(ctx, db.WithUser(sub.UserId), db.WithName(sub.SrcPath))
	if err != nil {
		return false, err
	}
	// 常规情况下 exists 只含有一个镜像
	if len(exists) > 1 {
		klog.Warningf("查询到镜像(%s)存在多个记录，不太正常，取第一个订阅", sub.Path)
	}
	tagMap := make(map[string]bool)
	errTags := make([]model.Tag, 0)
	for _, v := range exists {
		for _, tag := range v.Tags {
			if tag.Status == types.SyncImageError {
				klog.Infof("镜像(%s)版本(%s)状态异常，重新镜像同步", sub.Path, tag.Name)
				errTags = append(errTags, tag)
				continue
			}
			tagMap[tag.Name] = true
		}
		break
	}

	// 重新触发之前推送失败的tag
	if err = s.reRunSubscribeTags(ctx, errTags); err != nil {
		klog.Errorf("重新触发异常tag失败: %v", err)
	}

	var ns, repo string
	parts := strings.Split(sub.RawPath, "/")
	if len(parts) == 2 {
		ns, repo = parts[0], parts[1]
	}

	size := sub.Size
	if size > 10 {
		size = 10 // 最大并发是 100
	}

	remotes, err := s.SearchRepositoryTags(ctx, types.RemoteTagSearchRequest{
		Namespace:  ns,
		Repository: repo,
		Config: &types.SearchConfig{
			ImageFrom: sub.ImageFrom,
			Page:      1, // 从第一页开始搜索
			Size:      size,
			Policy:    sub.Policy,
			Arch:      sub.Arch,
		},
	})
	if err != nil {
		klog.Errorf("获取 dockerhub 镜像(%s)最新镜像版本失败 %v", sub.Path, err)
		// 如果返回报错是 404 Not Found 则说明远端进行不存在，终止订阅
		if strings.Contains(err.Error(), "404 Not Found") {
			klog.Infof("订阅镜像(%s)不存在，关闭订阅", sub.Path)
			if err = s.factory.Task().UpdateSubscribe(ctx, sub.Id, sub.ResourceVersion, map[string]interface{}{
				"status": "镜像不存在",
				"enable": false,
			}); err != nil {
				klog.Infof("镜像(%s)不存在关闭订阅失败 %v", sub.Path, err)
			}
			if err2 := s.factory.Task().CreateSubscribeMessage(ctx, &model.SubscribeMessage{
				SubscribeId: sub.Id, Message: fmt.Sprintf("订阅镜像(%s)不存在，已自动关闭 %v", sub.Path, err.Error()),
			}); err2 != nil {
				klog.Errorf("创建订阅限制事件失败 %v", err)
			}
		}

		return false, err
	}

	tagResults := remotes.([]types.TagResult)

	tagsMap := make(map[string][]string)
	for _, tag := range tagResults {
		for _, img := range tag.Images {
			existImages, ok := tagsMap[img.Architecture]
			if ok {
				existImages = append(existImages, sub.Path+":"+tag.Name)
				tagsMap[img.Architecture] = existImages
			} else {
				tagsMap[img.Architecture] = []string{sub.Path + ":" + tag.Name}
			}
		}
	}

	// TODO: 后续实现增量推送
	// 全部重新推送
	klog.Infof("即将全量推送订阅镜像(%s)", sub.Path)
	for arch, images := range tagsMap {
		if err = s.CreateTask(ctx, &types.CreateTaskRequest{
			Name:         uuid.NewRandName(fmt.Sprintf("sub-%s-", sub.Path), 8) + "-" + arch,
			UserId:       sub.UserId,
			UserName:     sub.UserName,
			RegisterId:   sub.RegisterId,
			Namespace:    sub.Namespace,
			Images:       images,
			OwnerRef:     1,
			SubscribeId:  sub.Id,
			Driver:       types.SkopeoDriver,
			PublicImage:  true,
			Architecture: arch,
		}); err != nil {
			klog.Errorf("创建订阅任务失败 %v", err)
			return false, err
		}
	}

	updates := make(map[string]interface{})
	updates["last_notify_time"] = time.Now()
	if err = s.factory.Task().UpdateSubscribe(ctx, sub.Id, sub.ResourceVersion, updates); err != nil {
		klog.Infof("订阅 (%s) 更新失败 %v", sub.Path, err)
	}
	return true, nil
}

func (s *ServerController) startSyncKubernetesTags(ctx context.Context) {
	klog.Infof("starting kubernetes tags syncer")
	ticker := time.NewTicker(3600 * 6 * time.Second)
	defer ticker.Stop()

	opt := types.CallKubernetesTagRequest{SyncAll: false}
	for range ticker.C {
		if _, err := s.SyncKubernetesTags(ctx, &opt); err != nil {
			klog.Error("failed kubernetes version syncer %v", err)
		}
	}
}

func (s *ServerController) ListSubscribes(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	listOption.SetDefaultPageOption()

	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}
	opts := []db.Options{
		db.WithUser(listOption.UserId),
		db.WithPathLike(listOption.NameSelector),
		db.WithNamespace(listOption.Namespace),
	}
	var err error
	pageResult.Total, err = s.factory.Task().CountSubscribe(ctx, opts...)
	if err != nil {
		klog.Errorf("获取订阅总数失败 %v", err)
		pageResult.Message = err.Error()
	}
	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithModifyOrderByDesc(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Task().ListSubscribes(ctx, opts...)
	if err != nil {
		klog.Errorf("获取订阅列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}
