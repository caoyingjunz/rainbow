package rainbow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
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

func (s *ServerController) RunSubscribe(ctx context.Context, req *types.RunSubscribeRequest) error {
	sub, err := s.factory.Task().GetSubscribe(ctx, req.SubscribeId)
	if err != nil {
		return err
	}
	if !sub.Enable {
		klog.Warningf("订阅已被关闭")
		return fmt.Errorf("订阅已关闭，需启用后执行")
	}

	s.CreateSubscribeMessageWithLog(ctx, *sub, "手动执行镜像订阅")
	if sub.Rewrite {
		return s.runSubscribeFull(ctx, sub)
	}
	return s.runSubscribeIncrement(ctx, sub)
}

// 全量同步
func (s *ServerController) runSubscribeFull(ctx context.Context, sub *model.Subscribe) error {
	klog.Infof("开始执行%s全量同步,策略 %s", sub.Path, sub.Policy)

	var ns, repo string
	parts := strings.Split(sub.RawPath, "/")
	if len(parts) == 2 {
		ns, repo = parts[0], parts[1]
	}
	size := sub.Size
	if size > 100 {
		size = 100 // 最大并发是 100
	}

	klog.Infof("开始搜索远端最新镜像")
	newImages, err := s.SearchRepositoryTags(ctx, types.CallSearchRequest{
		Namespace:    ns,
		Repository:   repo,
		CustomConfig: &types.SearchCustomConfig{Policy: sub.Policy, Arch: sub.Arch},
	})
	if err != nil {
		return s.HandlerSearchRepositoryTags(ctx, sub, err)
	}

	tagResult, ok := newImages.(types.CommonSearchTagResult)
	if !ok {
		klog.Errorf("转换tag类型失败")
		return fmt.Errorf("转换tag类型失败")
	}

	var imagesNeedSync []string
	for _, tag := range tagResult.TagResult {
		imagesNeedSync = append(imagesNeedSync, sub.Path+":"+tag.Name)
	}
	if len(imagesNeedSync) == 0 {
		klog.Infof("未发现有新增订阅镜像(%s)，本次忽略", sub.Path)
		return nil
	}

	klog.Infof("即将增量推送订阅镜像(%v)", imagesNeedSync)

	// 订阅的默认空间，在同步任务时做一次转换
	taskNamespace := sub.Namespace
	if len(taskNamespace) == 0 {
		taskNamespace = defaultNamespace
	}

	if err = s.CreateTask(ctx, &types.CreateTaskRequest{
		Name:         uuid.NewRandName(fmt.Sprintf("订阅-%s-", sub.Path), 8),
		UserId:       sub.UserId,
		UserName:     sub.UserName,
		RegisterId:   sub.RegisterId,
		Namespace:    taskNamespace,
		Images:       imagesNeedSync,
		OwnerRef:     1,
		SubscribeId:  sub.Id,
		Driver:       types.SkopeoDriver,
		PublicImage:  true,
		Architecture: sub.Arch,
	}); err != nil {
		klog.Errorf("创建订阅镜像任务失败 %v", err)
		return err
	}

	return s.afterRunSubscribe(ctx, sub)
}

// 增量同步
func (s *ServerController) runSubscribeIncrement(ctx context.Context, sub *model.Subscribe) error {
	exists, err := s.factory.Image().ListImagesWithTag(ctx, db.WithUser(sub.UserId), db.WithName(sub.DestPath))
	if err != nil {
		klog.Errorf("获取定义镜像(%s)失败 %v", sub.Path, err)
		return err
	}

	// 常规情况下 exists 只含有一个镜像
	if len(exists) > 1 {
		klog.Warningf("查询到镜像(%s)存在多个记录，不太正常，取第一个订阅", sub.Path)
	}
	if len(exists) == 0 {
		// 不存在，则首次全量触发
		klog.Infof("未发现镜像%s存在,订阅首次执行，触发全量同步", sub.Path)
		return s.runSubscribeFull(ctx, sub)
	}

	// 非首次，做增量推送
	subImage := exists[0]

	existImageTagMap := make(map[string]bool) // 已存在的推送成功镜像
	errTags := make([]model.Tag, 0)           // 之前推送失败的镜像列表
	for _, tag := range subImage.Tags {
		if tag.Status == types.SyncImageError {
			klog.Infof("镜像(%s)版本(%s)状态异常，即将重新同步", sub.Path, tag.Name)
			errTags = append(errTags, tag)
			continue
		}
		existImageTagMap[tag.Name] = true
	}

	// 重新触发之前推送失败的tag
	if err = s.reRunSubErrTags(ctx, errTags); err != nil {
		klog.Errorf("重新触发异常tag失败: %v", err)
	}

	var ns, repo string
	parts := strings.Split(sub.RawPath, "/")
	if len(parts) == 2 {
		ns, repo = parts[0], parts[1]
	}
	size := sub.Size
	if size > 100 {
		size = 100 // 最大并发是 100
	}

	klog.Infof("开始搜索远端最新镜像")
	newImages, err := s.SearchRepositoryTags(ctx, types.CallSearchRequest{
		Namespace:    ns,
		Repository:   repo,
		CustomConfig: &types.SearchCustomConfig{Policy: sub.Policy, Arch: sub.Arch},
	})
	if err != nil {
		return s.HandlerSearchRepositoryTags(ctx, sub, err)
	}

	tagResult, ok := newImages.(types.CommonSearchTagResult)
	if !ok {
		klog.Errorf("转换tag类型失败")
		return fmt.Errorf("转换tag类型失败")
	}

	var imagesNeedSync []string
	// TODO: 忽略架构差异,目前仅支持当架构订阅
	for _, tag := range tagResult.TagResult {
		if existImageTagMap[tag.Name] {
			continue
		}
		imagesNeedSync = append(imagesNeedSync, sub.Path+":"+tag.Name)
	}
	if len(imagesNeedSync) == 0 {
		klog.Infof("未发现有新增订阅镜像(%s)，本次忽略", sub.Path)
		return nil
	}

	// 订阅的默认空间，在同步任务时做一次转换
	taskNamespace := sub.Namespace
	if len(taskNamespace) == 0 {
		taskNamespace = defaultNamespace
	}
	klog.Infof("即将增量推送订阅镜像(%v)", imagesNeedSync)
	if err = s.CreateTask(ctx, &types.CreateTaskRequest{
		Name:         uuid.NewRandName(fmt.Sprintf("订阅-%s-", sub.Path), 8),
		UserId:       sub.UserId,
		UserName:     sub.UserName,
		RegisterId:   sub.RegisterId,
		Namespace:    taskNamespace,
		Images:       imagesNeedSync,
		OwnerRef:     1,
		SubscribeId:  sub.Id,
		Driver:       types.SkopeoDriver,
		PublicImage:  true,
		Architecture: sub.Arch,
	}); err != nil {
		klog.Errorf("创建订阅镜像任务失败 %v", err)
		return err
	}

	return s.afterRunSubscribe(ctx, sub)
}

func (s *ServerController) afterRunSubscribe(ctx context.Context, sub *model.Subscribe) error {
	updates := make(map[string]interface{})
	updates["last_notify_time"] = time.Now()
	if err := s.factory.Task().UpdateSubscribe(ctx, sub.Id, sub.ResourceVersion, updates); err != nil {
		klog.Infof("订阅 (%s) 更新失败 %v", sub.Path, err)
	}
	return nil
}

func (s *ServerController) HandlerSearchRepositoryTags(ctx context.Context, sub *model.Subscribe, err error) error {
	klog.Errorf("获取 dockerhub 镜像(%s)最新镜像版本失败 %v", sub.Path, err)
	// 如果返回报错是 404 Not Found 则说明远端进行不存在，终止订阅
	if strings.Contains(err.Error(), "404 Not Found") {
		klog.Infof("订阅镜像(%s)不存在，关闭订阅", sub.Path)
		if err = s.factory.Task().UpdateSubscribe(ctx, sub.Id, sub.ResourceVersion, map[string]interface{}{"status": "镜像不存在", "enable": false}); err != nil {
			klog.Infof("镜像(%s)不存在关闭订阅失败 %v", sub.Path, err)
		}
		if err2 := s.factory.Task().CreateSubscribeMessage(ctx, &model.SubscribeMessage{
			SubscribeId: sub.Id, Message: fmt.Sprintf("订阅镜像(%s)不存在，已自动关闭 %v", sub.Path, err.Error()),
		}); err2 != nil {
			klog.Errorf("创建订阅限制事件失败 %v", err)
		}
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

func (s *ServerController) reRunSubErrTags(ctx context.Context, errTags []model.Tag) error {
	if len(errTags) == 0 {
		return nil
	}

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

func (s *ServerController) preCreateSubscribe(ctx context.Context, req *types.CreateSubscribeRequest) error {
	// 同一个镜像，禁止重复订阅
	_, err := s.factory.Task().GetSubscribeBy(ctx, db.WithPath(req.Path))
	if err == nil {
		return fmt.Errorf("镜像(%s)的订阅已存在", req.Path)
	}
	if !errors.IsNotFound(err) {
		return err
	}

	if err := s.ValidateSubscribeSize(ctx, req.Size, req.UserId); err != nil {
		return err
	}
	if err := ValidateArch(req.Arch); err != nil {
		return err
	}

	return nil
}

func (s *ServerController) ValidateSubscribeSize(ctx context.Context, size int, userId string) error {
	if size < 0 {
		return fmt.Errorf("订阅镜像版本数不得定于 0")
	}

	user, err := s.factory.Task().GetUser(ctx, userId)
	if err != nil {
		klog.Warningf("获取失败，不做精细化校验")
		if size > 100 {
			return fmt.Errorf("订阅镜像版本数超过阈值 100")
		}
	} else {
		switch user.UserType {
		case types.FreeUserType: // 个人版本
			if size > 10 {
				return fmt.Errorf("个人版订阅镜像版本数超过阈值 10")
			}
		case types.PayUserType:
			if size > 100 {
				return fmt.Errorf("专业版订阅镜像版本数超过阈值 100")
			}
		}
	}

	return nil
}

func (s *ServerController) CreateSubscribe(ctx context.Context, req *types.CreateSubscribeRequest) error {
	if err := s.preCreateSubscribe(ctx, req); err != nil {
		return err
	}

	// 初始化仓库
	if req.RegisterId == 0 {
		req.RegisterId = *RegistryId
	}

	ns := WrapNamespace(req.Namespace, req.UserName)
	destPath, err := s.parseImageNameFromPath(ctx, req.Path, req.RegisterId, ns)
	if err != nil {
		return err
	}

	// 初始化镜像来源
	if len(req.ImageFrom) == 0 {
		req.ImageFrom = types.ImageHubDocker
	}

	rawPath := req.Path
	// 默认dockerhub不指定 ns时，会添加默认值
	if req.ImageFrom == types.ImageHubDocker {
		parts := strings.Split(rawPath, "/")
		if len(parts) == 1 {
			rawPath = types.DefaultDockerhubNamespace + "/" + rawPath
		}
	}

	return s.factory.Task().CreateSubscribe(ctx, &model.Subscribe{
		UserModel: rainbow.UserModel{
			UserId:   req.UserId,
			UserName: req.UserName,
		},
		Namespace:  ns,
		Path:       req.Path,
		RawPath:    rawPath,
		DestPath:   destPath,
		RegisterId: req.RegisterId,
		Enable:     req.Enable,   // 是否启动订阅
		Size:       req.Size,     // 最多同步多少个版本
		Interval:   req.Interval, // 多久执行一次
		ImageFrom:  req.ImageFrom,
		Policy:     strings.TrimSpace(req.Policy),
		Arch:       req.Arch,
		Rewrite:    req.Rewrite,
	})
}

func (s *ServerController) preUpdateSubscribe(ctx context.Context, req *types.UpdateSubscribeRequest) error {
	if err := s.ValidateSubscribeSize(ctx, req.Size, req.UserId); err != nil {
		return err
	}
	if err := ValidateArch(req.Arch); err != nil {
		return err
	}
	return nil
}

func (s *ServerController) UpdateSubscribe(ctx context.Context, req *types.UpdateSubscribeRequest) error {
	if err := s.preUpdateSubscribe(ctx, req); err != nil {
		return err
	}

	update := map[string]interface{}{
		"size":       req.Size,
		"interval":   req.Interval,
		"image_from": req.ImageFrom,
		"policy":     req.Policy,
		"arch":       req.Arch,
		"rewrite":    req.Rewrite,
	}

	enable := req.Enable
	old, err := s.factory.Task().GetSubscribe(ctx, req.Id)
	if err == nil {
		// 设置开关属性和消息
		if enable != old.Enable {
			update["enable"] = enable

			msg := ""
			// 原先是关闭，最新开启，则刷新 sub message 为开启
			if !old.Enable && enable {
				msg = fmt.Sprintf("手动启动制品订阅")
			}
			if old.Enable && !enable {
				msg = fmt.Sprintf("手动关闭制品订阅")
			}
			s.CreateSubscribeMessageWithLog(ctx, *old, msg)
		}

		// 同步更新（订正）命名空间
		newNS := WrapNamespace(req.Namespace, old.UserName)
		if newNS != old.Namespace {
			update["namespace"] = newNS
			// 如果命名空间发生变化，则目标path也需同步改变
			destPath, err := s.parseImageNameFromPath(ctx, old.Path, old.RegisterId, newNS)
			if err != nil {
				return err
			}

			update["dest_path"] = destPath
		}
	}

	if err = s.factory.Task().UpdateSubscribe(ctx, req.Id, req.ResourceVersion, update); err != nil {
		return err
	}
	return nil
}

// DeleteSubscribe 删除订阅
// 1. 删除订阅
// 2. 删除订阅记录
// 3. 删除订阅关联的同步任务
func (s *ServerController) DeleteSubscribe(ctx context.Context, subId int64) error {
	if err := s.factory.Task().DeleteSubscribe(ctx, subId); err != nil {
		return err
	}
	if err := s.factory.Task().DeleteSubscribeAllMessage(ctx, subId); err != nil {
		return err
	}
	if err := s.factory.Task().DeleteBySubscribe(ctx, subId); err != nil {
		return err
	}

	return nil
}
