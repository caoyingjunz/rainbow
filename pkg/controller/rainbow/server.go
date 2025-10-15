package rainbow

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"

	"github.com/go-redis/redis/v8"
	swr "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2"
	swrmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/model"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
	"github.com/caoyingjunz/rainbow/pkg/util/huaweicloud"
	"github.com/caoyingjunz/rainbow/pkg/util/uuid"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	CreateDockerfile(ctx context.Context, req *types.CreateDockerfileRequest) error
	DeleteDockerfile(ctx context.Context, dockerfileId int64) error
	UpdateDockerfile(ctx context.Context, req *types.UpdateDockerfileRequest) error
	ListDockerfile(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	GetDockerfile(ctx context.Context, dockerfileId int64) (interface{}, error)

	CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error
	UpdateRegistry(ctx context.Context, req *types.UpdateRegistryRequest) error
	DeleteRegistry(ctx context.Context, registryId int64) error
	GetRegistry(ctx context.Context, registryId int64) (interface{}, error)
	ListRegistries(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	LoginRegistry(ctx context.Context, req *types.CreateRegistryRequest) error

	CreateTask(ctx context.Context, req *types.CreateTaskRequest) error
	UpdateTask(ctx context.Context, req *types.UpdateTaskRequest) error
	ListTasks(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	DeleteTask(ctx context.Context, taskId int64) error
	GetTask(ctx context.Context, taskId int64) (interface{}, error)
	UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error

	CreateSubscribe(ctx context.Context, req *types.CreateSubscribeRequest) error
	ListSubscribes(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	UpdateSubscribe(ctx context.Context, req *types.UpdateSubscribeRequest) error
	DeleteSubscribe(ctx context.Context, subId int64) error
	GetSubscribe(ctx context.Context, subId int64) (interface{}, error)

	ListSubscribeMessages(ctx context.Context, subId int64) (interface{}, error)
	RunSubscribeImmediately(ctx context.Context, req *types.UpdateSubscribeRequest) error

	ListTaskImages(ctx context.Context, taskId int64, listOption types.ListOptions) (interface{}, error)
	ReRunTask(ctx context.Context, req *types.UpdateTaskRequest) error

	ListTasksByIds(ctx context.Context, ids []int64) (interface{}, error)
	DeleteTasksByIds(ctx context.Context, ids []int64) error

	CreateAgent(ctx context.Context, req *types.CreateAgentRequest) error
	UpdateAgent(ctx context.Context, req *types.UpdateAgentRequest) error
	DeleteAgent(ctx context.Context, agentId int64) error
	GetAgent(ctx context.Context, agentId int64) (interface{}, error)
	ListAgents(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error

	CreateImage(ctx context.Context, req *types.CreateImageRequest) error
	UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error
	DeleteImage(ctx context.Context, imageId int64) error
	GetImage(ctx context.Context, imageId int64) (interface{}, error)
	ListImages(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	ListImagesByIds(ctx context.Context, ids []int64) (interface{}, error)
	DeleteImagesByIds(ctx context.Context, ids []int64) error

	ListPublicImages(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error
	CreateImages(ctx context.Context, req *types.CreateImagesRequest) ([]model.Image, error)
	DeleteImageTag(ctx context.Context, imageId int64, TagId int64) error

	GetCollection(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	AddDailyReview(ctx context.Context, page string) error

	CreateLabel(ctx context.Context, req *types.CreateLabelRequest) error
	DeleteLabel(ctx context.Context, labelId int64) error
	UpdateLabel(ctx context.Context, req *types.UpdateLabelRequest) error
	ListLabels(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	CreateLogo(ctx context.Context, req *types.CreateLogoRequest) error
	UpdateLogo(ctx context.Context, req *types.UpdateLogoRequest) error
	DeleteLogo(ctx context.Context, logoId int64) error
	ListLogos(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	CreateNamespace(ctx context.Context, req *types.CreateNamespaceRequest) error
	UpdateNamespace(ctx context.Context, req *types.UpdateNamespaceRequest) error
	DeleteNamespace(ctx context.Context, objectId int64) error
	ListNamespaces(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	Overview(ctx context.Context) (interface{}, error)
	Downflow(ctx context.Context) (interface{}, error)
	Store(ctx context.Context) (interface{}, error)
	ImageDownflow(ctx context.Context, downflowMeta types.DownflowMeta) (interface{}, error)

	SearchRepositories(ctx context.Context, req types.RemoteSearchRequest) (interface{}, error)
	SearchRepositoryTags(ctx context.Context, req types.RemoteTagSearchRequest) (interface{}, error)
	SearchRepositoryTagInfo(ctx context.Context, req types.RemoteTagInfoSearchRequest) (interface{}, error)

	CreateTaskMessage(ctx context.Context, req types.CreateTaskMessageRequest) error
	ListTaskMessages(ctx context.Context, taskId int64) (interface{}, error)

	ListArchitectures(ctx context.Context, listOption types.ListOptions) ([]string, error)

	CreateUser(ctx context.Context, req *types.CreateUserRequest) error
	UpdateUser(ctx context.Context, req *types.UpdateUserRequest) error
	ListUsers(ctx context.Context, listOption types.ListOptions) ([]model.User, error)
	GetUser(ctx context.Context, userId string) (*model.User, error)
	DeleteUser(ctx context.Context, userId string) error

	CreateNotify(ctx context.Context, req *types.CreateNotificationRequest) error
	SendNotify(ctx context.Context, req *types.SendNotificationRequest) error

	ListKubernetesVersions(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	SyncKubernetesVersions(ctx context.Context, req *types.KubernetesTagRequest) (interface{}, error)

	ListRainbowds(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	Run(ctx context.Context, workers int) error
	Stop(ctx context.Context)
}

var (
	SwrClient  *swr.SwrClient
	RegistryId *int64
)

type ServerController struct {
	factory     db.ShareDaoFactory
	cfg         rainbowconfig.Config
	redisClient *redis.Client
	Producer    rocketmq.Producer

	// rpcServer
	pb.UnimplementedTunnelServer
	lock sync.RWMutex
}

func NewServer(f db.ShareDaoFactory, cfg rainbowconfig.Config, redisClient *redis.Client, p rocketmq.Producer) *ServerController {
	sc := &ServerController{
		factory:     f,
		cfg:         cfg,
		redisClient: redisClient,
		Producer:    p,
	}

	if SwrClient == nil || RegistryId == nil {
		reg, err := f.Registry().GetDefaultRegistry(context.TODO())
		if err == nil {
			if len(reg.Ak) == 0 || len(reg.Sk) == 0 || len(reg.RegionId) == 0 {
				klog.Errorf("默认华为仓库未设置必要配置, ak(%s) sk(%s) regionId(%s)", reg.Ak, reg.Sk, reg.RegionId)
			} else {
				client, err := huaweicloud.NewHuaweiCloudClient(huaweicloud.HuaweiCloudConfig{
					AK:       reg.Ak,
					SK:       reg.Sk,
					RegionId: reg.RegionId,
				})
				if err == nil {
					SwrClient = client
					RegistryId = &reg.Id
					klog.Infof("创建华为仓库客户端成功，仓库名称: %s(%d) ", reg.Name, *RegistryId)
				} else {
					klog.Errorf("创建为仓库客户端失败 %v", err)
				}
			}
		} else {
			klog.Errorf("获取默认华为仓库失败: %v", err)
		}
	}

	return sc
}

func (s *ServerController) GetAgent(ctx context.Context, agentId int64) (interface{}, error) {
	return s.factory.Agent().Get(ctx, agentId)
}

func (s *ServerController) sendMessageForRainbowd(ctx context.Context, rainbowName string, data []byte) error {
	msg := &primitive.Message{
		Topic: s.cfg.Rocketmq.Topic,
		Body:  data,
	}

	msg.WithTag(fmt.Sprintf("rainbowd-%s", rainbowName))
	msg.WithKeys([]string{"Rainbowd"})
	res, err := s.Producer.SendSync(ctx, msg)
	if err != nil {
		klog.Errorf("send message to rainbowd error: %v", err)
		return err
	}

	klog.V(0).Infof("send message to rainbowd success: result=%s", res.String())
	return nil
}

func (s *ServerController) UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error {
	old, err := s.factory.Agent().GetByName(ctx, req.AgentName)
	if err != nil {
		return err
	}
	if err := s.factory.Agent().UpdateByName(ctx, req.AgentName, map[string]interface{}{"status": req.Status, "message": fmt.Sprintf("Agent has been set to %s", req.Status)}); err != nil {
		return err
	}

	return s.sendMessageForRainbowd(ctx, old.RainbowdName, []byte(fmt.Sprintf("%d/%d", old.Id, old.ResourceVersion)))
}

func (s *ServerController) UpdateAgent(ctx context.Context, req *types.UpdateAgentRequest) error {
	repo := req.GithubRepository
	if len(repo) == 0 {
		repo = fmt.Sprintf("https://github.com/%s/plugin.git", req.GithubUser)
	}

	updates := make(map[string]interface{})
	updates["github_user"] = req.GithubUser
	updates["github_repository"] = repo
	updates["github_token"] = req.GithubToken
	updates["github_email"] = req.GithubEmail
	updates["healthz_port"] = req.HealthzPort
	updates["rainbowd_name"] = req.RainbowdName
	return s.factory.Agent().UpdateByName(ctx, req.AgentName, updates)
}

func (s *ServerController) ListAgents(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	return s.factory.Agent().List(ctx, db.WithNameLike(listOption.NameSelector))
}

func (s *ServerController) Run(ctx context.Context, workers int) error {
	go s.schedule(ctx)
	go s.sync(ctx)
	go s.startSyncDailyPulls(ctx)
	//go s.startRpcServer(ctx)
	go s.startAgentHeartbeat(ctx)
	go s.startSyncKubernetesVersion(ctx)
	go s.startSubscribeController(ctx)

	klog.Infof("starting rocketmq producer")
	if err := s.Producer.Start(); err != nil {
		return err
	}

	return nil
}

func (s *ServerController) Stop(ctx context.Context) {
	klog.Infof("停止服务!!!")
	_ = s.Producer.Shutdown()
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

func (s *ServerController) startSubscribeController(ctx context.Context) {
	klog.Infof("starting subscribe controller")

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		subscribes, err := s.factory.Task().ListSubscribes(ctx, db.WithEnable(1), db.WithFailTimes(6))
		if err != nil {
			klog.Errorf("获取全部订阅失败 %v 15分钟后重新执行订阅", err)
			continue
		}

		for _, sub := range subscribes {
			if sub.FailTimes > 5 {
				klog.Warningf("订阅 (%s) 失败超过限制，已终止订阅", sub.Path)
				s.DisableSubscribeWithMessage(ctx, sub, fmt.Sprintf("订阅(%s)失败已超过限制，终止订阅", sub.Path))
				continue
			}
			now := time.Now()
			if now.Sub(sub.LastNotifyTime) < sub.Interval*time.Second {
				klog.Infof("订阅 (%s) 时间间隔 %v 暂时无需执行", sub.Path, sub.Interval*time.Second)
				continue
			}

			changed, err := s.subscribe(ctx, sub)
			if err == nil {
				// 订阅触发成功
				if changed {
					s.CreateSubscribeMessageWithLog(ctx, sub, fmt.Sprintf("%s 在 %v 订阅触发成功", sub.Path, time.Now().Format("2006-01-02 15:04:05")))
				}
			} else {
				klog.Error("failed to do Subscribe(%s) %v", sub.Path, err)
				s.CreateSubscribeMessageAndFailTimesAdd(ctx, sub, err.Error())
			}

			// 仅保留最新的 n 个事件
			_ = s.cleanSubscribeMessages(ctx, sub.Id, 5)
		}
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
	if size > 100 {
		size = 100 // 最大并发是 100
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

func (s *ServerController) startSyncKubernetesVersion(ctx context.Context) {
	klog.Infof("starting kubernetes version syncer")
	ticker := time.NewTicker(3600 * time.Second)
	defer ticker.Stop()

	opt := types.KubernetesTagRequest{SyncAll: false}
	for range ticker.C {
		if _, err := s.SyncKubernetesVersions(ctx, &opt); err != nil {
			klog.Error("failed kubernetes version syncer %v", err)
		}
	}
}

func (s *ServerController) startSyncDailyPulls(ctx context.Context) {
	c := cron.New()
	_, err := c.AddFunc("0 1 * * *", func() {
		klog.Infof("执行每天凌晨 1 点任务...")
		s.syncPulls(ctx)
	})
	if err != nil {
		klog.Fatal("定时任务配置错误:", err)
	}
	c.Start()
	klog.Infof("starting cronjob controller")

	// 优雅关闭（可选）
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	c.Stop()
	klog.Infof("定时任务已停止")
}

func (s *ServerController) startRpcServer(ctx context.Context) {
	listener, err := net.Listen("tcp", ":8091")
	if err != nil {
		klog.Fatalf("failed to listen %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterTunnelServer(gs, s)

	klog.Infof("starting rpc server (listening at %v)", listener.Addr())
	if err = gs.Serve(listener); err != nil {
		klog.Fatalf("failed to start rpc serve %v", err)
	}
}

func (s *ServerController) syncPulls(ctx context.Context) {
	_, err := s.factory.Image().List(ctx)
	if err != nil {
		klog.Errorf("获取镜像列表失败 %v", err)
		return
	}
}

func (s *ServerController) schedule(ctx context.Context) {
	klog.Infof("starting scheduler controller")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.doSchedule(ctx); err != nil {
			klog.Error("failed to do schedule %v", err)
		}
	}
}

func (s *ServerController) doSchedule(ctx context.Context) error {
	item, err := s.factory.Task().GetOneForSchedule(ctx)
	if err != nil {
		return err
	}
	if item == nil {
		return nil
	}
	klog.Infof("获取待处理任务 %v", item)

	targetAgent, err := s.assignAgent(ctx)
	if err != nil {
		return err
	}
	if targetAgent == "" {
		return nil
	}
	if err = s.factory.Task().Update(ctx, item.Id, item.ResourceVersion, map[string]interface{}{
		"agent_name": targetAgent,
	}); err != nil {
		return err
	}
	klog.Infof("任务 %s 已被分配给 agent %s，等待处理中", item.Name, targetAgent)

	return nil
}

func (s *ServerController) sync(ctx context.Context) {
	if SwrClient == nil {
		klog.Infof("未设置默认远程仓库，无需镜像同步")
		return
	}

	klog.Infof("starting remote image sync controller")
	ticker := time.NewTicker(300 * time.Second)
	defer ticker.Stop()

	defaultNamespace := HuaweiNamespace
	for range ticker.C {
		//overview, err := SwrClient.ShowDomainOverview(&swrmodel.ShowDomainOverviewRequest{})
		//if err != nil {
		//	klog.Errorf("获取远程仓库概览失败", err)
		//	continue
		//}
		//klog.Infof("获取远程仓库概览成功 %v", overview)

		// TODO: 后续分页查询
		resp, err := SwrClient.ListReposDetails(&swrmodel.ListReposDetailsRequest{Namespace: &defaultNamespace})
		if err != nil {
			klog.Errorf("获取远程镜像列表失败 %v", err)
			continue
		}
		if resp.Body == nil || len(*resp.Body) == 0 {
			klog.Infof("获取远程镜像为空")
			return
		}

		var imageNames []string
		imageMap := make(map[string]int64)
		for _, reRepo := range *resp.Body {
			imageNames = append(imageNames, reRepo.Name)
			imageMap[reRepo.Name] = reRepo.NumDownload
		}

		targetImages, err := s.factory.Image().List(ctx, db.WithNameIn(imageNames...))
		if err != nil {
			klog.Errorf("查询本地镜像列表失败 %v", err)
			continue
		}
		for _, targetImage := range targetImages {
			pull := imageMap[targetImage.Name]
			if targetImage.Pull == pull {
				klog.V(1).Infof("镜像(%s)下载量未发生变量，无需更新", targetImage.Name)
				continue
			}

			klog.Infof("镜像(%s)下载量已发生变量，延迟更新", targetImage.Name)
			err = s.factory.Image().Update(ctx, targetImage.Id, targetImage.ResourceVersion, map[string]interface{}{"pull": pull})
			if err != nil {
				klog.Errorf("更新镜像(%s)的下载量失败 %v", targetImage.Name, err)
			}
		}
	}
}

func (s *ServerController) assignAgent(ctx context.Context) (string, error) {
	agents, err := s.factory.Agent().ListForSchedule(ctx)
	if err != nil {
		return "", err
	}
	if len(agents) == 0 {
		klog.Warningf("不存在可用工作节点，等待下一次调度")
		return "", nil
	}

	var agentNames []string
	agentMap := make(map[string]int)
	for _, agent := range agents {
		agentNames = append(agentNames, agent.Name)
		agentMap[agent.Name] = 0
	}
	agentSet := sets.NewString(agentNames...)

	runningTasks, err := s.factory.Task().GetRunningTask(ctx)
	if err != nil {
		return "", err
	}

	if len(runningTasks) == 0 {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(len(agentNames))
		agent := agentNames[x]
		klog.Infof("当前节点均空闲，工作节点 %s 被随机选中", agent)
		return agent, nil
	} else {
		for _, t := range runningTasks {
			if !agentSet.Has(t.AgentName) {
				continue
			}
			old, ok := agentMap[t.AgentName]
			if ok {
				agentMap[t.AgentName] = old + 1
			} else {
				continue
			}
		}

		min := len(runningTasks)
		agent := ""
		for k, v := range agentMap {
			if min >= v {
				min = v
				agent = k
			}
		}
		// 一个 agent 最大并发为 10
		if min > 10 {
			klog.Warningf("工作节点已满负载，等待下一次调度")
			return "", nil
		}
		if agent == "" {
			klog.Warningf("未选中工作节点，等待下一次调度")
			return "", nil
		}

		klog.Infof("工作节点 %s 已选中", agent)
		return agent, nil
	}
}

func (s *ServerController) startAgentHeartbeat(ctx context.Context) {
	klog.Infof("starting agent heartbeat")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		agents, err := s.factory.Agent().List(ctx)
		if err != nil {
			klog.Error("获取 agents 列表失败，等待下一次重试 %v", err)
			continue
		}

		for _, agent := range agents {
			if agent.Status != model.RunAgentType {
				klog.V(1).Infof("agent(%s)非在线状态，忽略", agent.Name)
				continue
			}

			diff := time.Now().Sub(agent.LastTransitionTime)
			if diff > time.Minute*5 {
				if agent.Status == model.UnknownAgentType {
					continue
				}
				err = s.factory.Agent().UpdateByName(ctx, agent.Name, map[string]interface{}{"status": model.UnknownAgentType, "message": "Agent stopped posting status"})
				if err != nil {
					klog.Error("failed to sync agent %s status %v", agent.Name, err)
				} else {
					klog.Infof("agent(%s)被设置成未知", agent.Name)
				}
			}
		}
	}
}

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

func (s *ServerController) ListArchitectures(ctx context.Context, listOption types.ListOptions) ([]string, error) {
	return []string{
		"linux/amd64",
		"linux/arm64",
		"windows/amd64",
	}, nil
}
