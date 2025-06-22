package rainbow

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	"github.com/caoyingjunz/rainbow/pkg/util/huaweicloud"
)

type ServerGetter interface {
	Server() ServerInterface
}

type ServerInterface interface {
	CreateRegistry(ctx context.Context, req *types.CreateRegistryRequest) error
	UpdateRegistry(ctx context.Context, req *types.UpdateRegistryRequest) error
	DeleteRegistry(ctx context.Context, registryId int64) error
	GetRegistry(ctx context.Context, registryId int64) (interface{}, error)
	ListRegistries(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	CreateTask(ctx context.Context, req *types.CreateTaskRequest) error
	UpdateTask(ctx context.Context, req *types.UpdateTaskRequest) error
	ListTasks(ctx context.Context, listOption types.ListOptions) (interface{}, error)
	DeleteTask(ctx context.Context, taskId int64) error
	GetTask(ctx context.Context, taskId int64) (interface{}, error)
	UpdateTaskStatus(ctx context.Context, req *types.UpdateTaskStatusRequest) error

	ListTaskImages(ctx context.Context, taskId int64, listOption types.ListOptions) (interface{}, error)
	ReRunTask(ctx context.Context, req *types.UpdateTaskRequest) error

	GetAgent(ctx context.Context, agentId int64) (interface{}, error)
	ListAgents(ctx context.Context) (interface{}, error)
	UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error

	CreateImage(ctx context.Context, req *types.CreateImageRequest) error
	UpdateImage(ctx context.Context, req *types.UpdateImageRequest) error
	DeleteImage(ctx context.Context, imageId int64) error
	GetImage(ctx context.Context, imageId int64) (interface{}, error)
	ListImages(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	ListPublicImages(ctx context.Context, listOption types.ListOptions) (interface{}, error)

	UpdateImageStatus(ctx context.Context, req *types.UpdateImageStatusRequest) error
	CreateImages(ctx context.Context, req *types.CreateImagesRequest) ([]model.Image, error)
	DeleteImageTag(ctx context.Context, imageId int64, name string) error

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

	CreateUser(ctx context.Context, req *types.CreateUserRequest) error

	Run(ctx context.Context, workers int) error
}

var (
	SwrClient  *swr.SwrClient
	RegistryId *int64
)

type ServerController struct {
	factory     db.ShareDaoFactory
	cfg         rainbowconfig.Config
	redisClient *redis.Client

	// rpcServer
	pb.UnimplementedTunnelServer
	lock sync.RWMutex
}

func NewServer(f db.ShareDaoFactory, cfg rainbowconfig.Config, redisClient *redis.Client) *ServerController {
	sc := &ServerController{
		factory:     f,
		cfg:         cfg,
		redisClient: redisClient,
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

func (s *ServerController) UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error {
	return s.factory.Agent().UpdateByName(ctx, req.AgentName, map[string]interface{}{"status": req.Status, "message": fmt.Sprintf("Agent has been set to %s", req.Status)})
}

func (s *ServerController) ListAgents(ctx context.Context) (interface{}, error) {
	return s.factory.Agent().List(ctx)
}

func (s *ServerController) Run(ctx context.Context, workers int) error {
	go s.schedule(ctx)
	go s.sync(ctx)
	go s.startSyncDailyPulls(ctx)
	go s.startRpcServer(ctx)
	go s.startAgentHeartbeat(ctx)

	return nil
}

func (s *ServerController) startSyncDailyPulls(ctx context.Context) {
	location, _ := time.LoadLocation("Asia/Shanghai") // 设置时区
	c := cron.New(cron.WithLocation(location))
	_, err := c.AddFunc("* * * * *", func() {
		klog.Infof("执行每日 0 点任务...")
		s.syncPulls(ctx)
	})
	if err != nil {
		klog.Fatal("定时任务配置错误:", err)
	}
	c.Start()
	klog.Infof("定时任务已启动")

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
		return err
	}

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
			klog.Errorf("获取远程镜像列表失败", err)
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
			klog.Errorf("查询本地镜像列表失败", err)
			continue
		}
		for _, targetImage := range targetImages {
			pull := imageMap[targetImage.Name]
			if targetImage.Pull == pull {
				klog.Infof("镜像(%s)下载量未发生变量，无需更新", targetImage.Name)
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
			klog.Error("failed to get agents %v", err)
			continue
		}

		for _, agent := range agents {
			diff := time.Now().Sub(agent.LastTransitionTime)
			if diff > time.Minute*5 {
				if agent.Status == model.UnknownAgentType {
					continue
				}
				err = s.factory.Agent().UpdateByName(ctx, agent.Name, map[string]interface{}{"status": model.UnknownAgentType, "message": "Agent stopped posting status"})
				if err != nil {
					klog.Error("failed to sync agent %s status %v", agent.Name, err)
				}
			}
		}
	}
}
