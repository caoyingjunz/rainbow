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
		return fmt.Errorf("订阅未开启")
	}

	if sub.Rewrite {
		return s.runSubscribeFull(ctx, sub)
	}
	return s.runSubscribeIncrement(ctx, sub)
}

// 全量同步
func (s *ServerController) runSubscribeFull(ctx context.Context, sub *model.Subscribe) error {
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
	if err = s.CreateTask(ctx, &types.CreateTaskRequest{
		Name:         uuid.NewRandName(fmt.Sprintf("订阅-%s-", sub.Path), 8),
		UserId:       sub.UserId,
		UserName:     sub.UserName,
		RegisterId:   sub.RegisterId,
		Namespace:    sub.Namespace,
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
	exists, err := s.factory.Image().ListImagesWithTag(ctx, db.WithUser(sub.UserId), db.WithName(sub.SrcPath))
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

	klog.Infof("即将增量推送订阅镜像(%v)", imagesNeedSync)
	if err = s.CreateTask(ctx, &types.CreateTaskRequest{
		Name:         uuid.NewRandName(fmt.Sprintf("订阅-%s-", sub.Path), 8),
		UserId:       sub.UserId,
		UserName:     sub.UserName,
		RegisterId:   sub.RegisterId,
		Namespace:    sub.Namespace,
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

	tagResp, err := s.SearchRepositoryTags(ctx, types.CallSearchRequest{
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
	if err = s.reRunSubErrTags(ctx, errTags); err != nil {
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

	remotes, err := s.SearchRepositoryTags(ctx, types.CallSearchRequest{
		Namespace:  ns,
		Repository: repo,
		//Config: &types.SearchConfig{
		//	ImageFrom: sub.ImageFrom,
		//	Page:      1, // 从第一页开始搜索
		//	Size:      size,
		//	Policy:    sub.Policy,
		//	Arch:      sub.Arch,
		//},
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
			Name:         uuid.NewRandName(fmt.Sprintf("订阅-%s-", sub.Path), 8) + "-" + arch,
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

	parts2 := strings.Split(req.Path, "/")
	srcPath := parts2[len(parts2)-1]
	if len(srcPath) == 0 {
		return fmt.Errorf("不合规镜像名称 %s", req.Path)
	}

	ns := WrapNamespace(req.Namespace, req.UserName)
	if len(ns) != 0 {
		srcPath = ns + "/" + srcPath
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
		Namespace: ns,
		Path:      req.Path,
		RawPath:   rawPath,
		SrcPath:   srcPath,
		Enable:    req.Enable,   // 是否启动订阅
		Size:      req.Size,     // 最多同步多少个版本
		Interval:  req.Interval, // 多久执行一次
		ImageFrom: req.ImageFrom,
		Policy:    strings.TrimSpace(req.Policy),
		Arch:      req.Arch,
		Rewrite:   false, // ToDo： 目前仅支持增量
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
		"namespace":  req.Namespace,
	}

	enable := req.Enable
	old, err := s.factory.Task().GetSubscribe(ctx, req.Id)
	if err == nil {
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

			// 同步更新（订正）命名空间
			newNS := WrapNamespace(req.Namespace, old.UserName)
			if newNS != old.Namespace {
				update["namespace"] = newNS
			}
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
