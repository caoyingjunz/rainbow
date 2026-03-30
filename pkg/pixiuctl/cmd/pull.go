package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/ctrutil"
	"github.com/caoyingjunz/rainbow/pkg/util/docker"
	"github.com/caoyingjunz/rainbow/pkg/util/signatureutil"
)

const (
	driverAuto   = "auto"
	driverDocker = "docker"
	driverCtr    = "ctr"

	defaultCtrNamespace = "k8s.io"
)

type RepoResult struct {
	Code    int       `json:"code"`
	Result  model.Tag `json:"result,omitempty"`
	Message string    `json:"message,omitempty"`
}

type UserResult struct {
	Code    int        `json:"code"`
	Result  model.User `json:"result"`
	Message string     `json:"message,omitempty"`
}

type PullOptions struct {
	baseURL     string
	waitTimeout time.Duration
	cfg         *config.Config

	accessKey string
	signature string

	// flag
	Platform  string
	Rewrite   bool
	Driver    string
	Namespace string

	resolvedDriver string

	Repos []string

	user *model.User
}

func NewPullCommand() *cobra.Command {
	o := &PullOptions{
		baseURL:     baseURL,
		waitTimeout: 10 * time.Minute,
		Driver:      driverAuto,
		Namespace:   defaultCtrNamespace,
	}

	cmd := &cobra.Command{
		Use:   "pull [image]",
		Short: "Pull and cache images from PixiuHub(https://hub.pixiuio.com)",
		Long: `Pull and cache images from PixiuHub(https://hub.pixiuio.com) to local storage.

Use --driver to choose the local runtime: auto (default) picks docker if available, else ctr;
--driver docker uses docker pull; --driver ctr uses ctr images pull with --namespace (default k8s.io).`,
		Example: `  # 拉取单个镜像
  pixiuctl pull nginx:latest

  # 拉取多个镜像
  pixiuctl pull nginx:latest redis:alpine

  # 指定架构拉取
  pixiuctl pull nginx --platform linux/arm64

  # 强制重建缓存后拉取
  pixiuctl pull nginx --rewrite

  # 使用 containerd ctr 并指定命名空间
  pixiuctl pull nginx:latest --driver ctr --namespace k8s.io`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVar(&o.Platform, "platform", "linux/amd64", "Platform for the image (e.g. linux/amd64, linux/arm64)")
	cmd.Flags().BoolVar(&o.Rewrite, "rewrite", false, "Rewrite the repo even if it exists")
	cmd.Flags().StringVarP(&o.Driver, "driver", "d", driverAuto, "Local pull driver: auto (prefer docker, then ctr), docker, or ctr")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", defaultCtrNamespace, "containerd namespace for ctr (only used when --driver is ctr or resolves to ctr)")

	return cmd
}

func (o *PullOptions) Complete(cmd *cobra.Command, args []string) error {
	// Load config file from root flag.
	configFile, err := cmd.Root().PersistentFlags().GetString("configFile")
	if err != nil {
		return err
	}
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return err
	}
	o.cfg = cfg
	if o.cfg.Default != nil && len(o.cfg.Default.URL) != 0 {
		o.baseURL = o.cfg.Default.URL
	}
	if o.cfg.Default != nil && o.cfg.Default.Timeout > 0 {
		o.waitTimeout = time.Duration(o.cfg.Default.Timeout) * time.Minute
	}

	o.Repos = args

	if err = o.resolveDriver(); err != nil {
		return err
	}

	return nil
}

func (o *PullOptions) resolveDriver() error {
	d := strings.ToLower(strings.TrimSpace(o.Driver))
	switch d {
	case driverAuto:
		if _, err := exec.LookPath(driverDocker); err == nil {
			o.resolvedDriver = driverDocker
			return nil
		}
		if _, err := exec.LookPath(driverCtr); err == nil {
			o.resolvedDriver = driverCtr
			return nil
		}
		return fmt.Errorf("未在 PATH 中找到 docker 或 ctr，请安装其一或显式指定 --driver")
	case driverDocker:
		if _, err := exec.LookPath(driverDocker); err != nil {
			return fmt.Errorf("指定了 --driver docker 但未找到 docker 可执行文件: %w", err)
		}
		o.resolvedDriver = driverDocker
		return nil
	case driverCtr:
		if _, err := exec.LookPath(driverCtr); err != nil {
			return fmt.Errorf("指定了 --driver ctr 但未找到 ctr 可执行文件: %w", err)
		}
		o.resolvedDriver = driverCtr
		return nil
	default:
		return fmt.Errorf("--driver 必须是 auto、docker 或 ctr，当前为 %q", o.Driver)
	}
}

// Validate makes sure that provided values for command-line options are valid
func (o *PullOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(o.Repos) == 0 {
		return fmt.Errorf("未指定任何待同步镜像名称")
	}

	if o.cfg.Auth == nil {
		return fmt.Errorf("配置文件缺少 Auth")
	}
	if len(o.cfg.Auth.AccessKey) == 0 {
		return fmt.Errorf("配置文件缺少 auth.access_key")
	}
	if len(o.cfg.Auth.SecretKey) == 0 {
		return fmt.Errorf("配置文件缺少 auth.secret_key")
	}

	if o.resolvedDriver != driverCtr && o.Namespace != defaultCtrNamespace {
		return fmt.Errorf("--namespace 仅在与 containerd ctr 一起使用时生效（--driver ctr 或 auto 解析为 ctr），当前驱动为 %s", o.resolvedDriver)
	}

	return nil
}

func (o *PullOptions) Run() error {
	// 运行前初始化必要属性
	if err := o.preRun(); err != nil {
		return err
	}

	// 执行
	diff := len(o.Repos)
	errCh := make(chan error, diff)

	var wg sync.WaitGroup
	wg.Add(diff)
	for _, repo := range o.Repos {
		go func(i string) {
			defer wg.Done()
			// 等待执行
			errCh <- o.pullAndCacheOne(i)
		}(repo)
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for pullErr := range errCh {
		if pullErr != nil {
			errs = append(errs, pullErr)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func (o *PullOptions) preRun() error {
	o.accessKey = o.cfg.Auth.AccessKey
	// 完成客户端证书
	o.signature = signatureutil.GenerateSignature(
		map[string]string{
			"action":    "pullOrCacheRepo",
			"accessKey": o.accessKey},
		[]byte(o.cfg.Auth.SecretKey))

	// 初始化用户信息
	var err error
	o.user, err = GetUserInfoByAccessKey(o.baseURL, o.accessKey, o.signature)
	if err != nil {
		return err
	}

	return nil
}

// SearchRepo 搜索镜像是否存在缓存，如果存在，则直接 pull，如果不存在则先构成缓存，然后再pull，最后进行tag
func (o *PullOptions) pullAndCacheOne(repo string) error {
	if o.Rewrite {
		klog.Infof("rewirting cache")
		return o.cacheAndPull(repo)
	}

	// 1. 执行 pull
	existsRepo, err := o.SearchRepo(repo)
	if err != nil {
		if ErrorIsNotFound(err) {
			return o.cacheAndPull(repo)
		}
		return err
	}

	// 2. 首次状态异常，则重新触发构建
	if existsRepo.Status == types.SyncImageError {
		klog.Infof("going rebuild cache")
		return o.cacheAndPull(repo)
	}
	if existsRepo.Status == types.SyncImageComplete {
		return o.pull(existsRepo)
	}

	return o.waitAndPull(repo)
}

func (o *PullOptions) SearchRepo(repo string) (*model.Tag, error) {
	url := fmt.Sprintf("%s/api/v2/search/repos?nameSelector=%s&arch=%s&user_id=%s", o.baseURL, repo, o.Platform, o.user.UserId)

	var result RepoResult
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodGet).
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{
			"X-ACCESS-KEY":  o.accessKey,
			"Authorization": o.signature,
		}).
		Do(&result); err != nil {
		return nil, err
	}
	if result.Code == 200 {
		return &result.Result, nil
	}

	return nil, fmt.Errorf("%s", result.Message)
}

// 下载镜像
// 重命名镜像，删除 mirror 镜像
func (o *PullOptions) pull(tag *model.Tag) error {
	sourceImage := tag.Mirror + ":" + tag.Name
	targetImage := tag.Path + ":" + tag.Name

	switch o.resolvedDriver {
	case driverCtr:
		if err := ctrutil.PullImage(o.Namespace, sourceImage); err != nil {
			return err
		}
		return ctrutil.TagAndRemoveSource(o.Namespace, sourceImage, targetImage)
	default:
		if err := docker.PullImage(sourceImage); err != nil {
			return err
		}
		return docker.TagImage(sourceImage, targetImage)
	}
}

// 构造并等待缓存完成
func (o *PullOptions) cacheAndPull(repo string) error {
	if err := o.buildCache(repo); err != nil {
		return err
	}
	cache, err := o.waitForCached(repo)
	if err != nil {
		return err
	}

	klog.Infof("image cache completed")
	return o.pull(cache)
}

func (o *PullOptions) waitAndPull(repo string) error {
	klog.Infof("waiting for cache completed")
	cache, err := o.waitForCached(repo)
	if err != nil {
		return err
	}

	klog.Infof("image cache completed")
	return o.pull(cache)
}

func (o *PullOptions) buildCache(repo string) error {
	data, err1 := json.Marshal(map[string]interface{}{
		"name":         "PixiuHub-" + repo + "-加速",
		"architecture": o.Platform,
		"images":       []string{repo},
		"user_id":      o.user.UserId,
		"user_name":    o.user.Name,
		"public_image": true,
	})
	if err1 != nil {
		return err1
	}

	var result CreateResult
	url := fmt.Sprintf("%s/api/v2/tasks", o.baseURL)
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodPost).
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{"X-ACCESS-KEY": o.accessKey, "Authorization": o.signature}).
		WithBody(bytes.NewBuffer(data)).
		Do(&result); err != nil {
		return err
	}
	if result.Code == 200 {
		klog.Infof("building cache, please wait...")
		return nil
	}
	return fmt.Errorf("%s", result.Message)
}

func (o *PullOptions) waitForCached(repo string) (*model.Tag, error) {
	// 创建一个计时器用于超时控制
	timeoutTimer := time.NewTimer(o.waitTimeout)
	defer timeoutTimer.Stop()

	// 创建一个 ticker 用于定期轮询
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			// 超时退出
			return nil, fmt.Errorf("构建(%s)缓存已超时，请稍后再试或调整超时时间再试", repo)
		case <-ticker.C:
			// 执行轮询：获取镜像当前状态
			cacheTag, err := o.SearchRepo(repo)
			if err != nil {
				klog.V(1).Infof("获取构建失败(%v)，等待下一次查询", err)
				continue
			}

			if cacheTag.Status == types.SyncImageComplete {
				return cacheTag, nil
			}
			if cacheTag.Status == types.SyncImageError {
				msg := cacheTag.Message
				parts := strings.Split(msg, "msg=")
				if len(parts) == 2 {
					msg = parts[1]
				}
				return nil, fmt.Errorf("%s\n更多信息参考 https://hub.pixiuio.com/", msg)
			}
		}
	}
}
