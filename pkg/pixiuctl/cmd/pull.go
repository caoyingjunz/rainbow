package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/docker"
	"github.com/caoyingjunz/rainbow/pkg/util/signatureutil"
)

const (
	baseURL = "http://127.0.0.1:8090"
)

type RepoResult struct {
	Code    int       `json:"code"`
	Result  model.Tag `json:"result,omitempty"`
	Message string    `json:"message,omitempty"`
}

type TaskResult struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type PullOptions struct {
	baseURL   string
	signature string
	cfg       *config.Config

	// flag
	Platform string

	Repos []string
}

func NewPullCommand() *cobra.Command {
	o := &PullOptions{
		baseURL: baseURL,
	}

	cmd := &cobra.Command{
		Use:   "pull [image]",
		Short: "Pull images from remote registry",
		Long:  `Pull images from remote registry to local storage.`,
		//Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVar(&o.Platform, "platform", "linux/amd64", "Platform for the image (e.g. linux/amd64, linux/arm64)")

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
	o.Repos = args

	return nil
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

	return nil
}

func (o *PullOptions) Run() error {
	// 完成客户端证书
	o.signature = signatureutil.GenerateSignature(
		map[string]string{"action": "pullOrCacheRepo", "accessKey": o.cfg.Auth.AccessKey},
		[]byte(o.cfg.Auth.SecretKey))

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

// SearchRepo 搜索镜像是否存在缓存，如果存在，则直接 pull，如果不存在则先构成缓存，然后再pull，最后进行tag
func (o *PullOptions) pullAndCacheOne(repo string) error {
	// TODO
	// 1. 检查是否本地已存在镜像

	// 2. 执行 pull
	existsRepo, err := o.SearchRepo(repo)
	if err != nil {
		if ErrorIsNotFound(err) {
			return o.cacheAndPull(repo)
		}
		return err
	}

	return o.pull(existsRepo)
}

func (o *PullOptions) SearchRepo(repo string) (*model.Tag, error) {
	url := fmt.Sprintf("%s/api/v2/search/repos?nameSelector=%s&arch=%s", o.baseURL, repo, o.Platform)

	var result RepoResult
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodGet).
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{
			"X-ACCESS-KEY":  o.cfg.Auth.AccessKey,
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

	if err := docker.PullImage(sourceImage); err != nil {
		return err
	}
	return docker.TagImage(sourceImage, targetImage)
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

	return o.pull(cache)
}

func (o *PullOptions) buildCache(repo string) error {
	data, err1 := json.Marshal(map[string]interface{}{
		"name":         "PixiuHub-" + repo + "-加速",
		"architecture": o.Platform,
		"images":       []string{repo},
	})
	if err1 != nil {
		return err1
	}

	var result TaskResult
	url := fmt.Sprintf("%s/api/v2/tasks", o.baseURL)
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method(http.MethodPost).
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{"X-ACCESS-KEY": o.cfg.Auth.AccessKey, "Authorization": o.signature}).
		WithBody(bytes.NewBuffer(data)).
		Do(&result); err != nil {
		return err
	}
	if result.Code == 200 {
		return nil
	}
	return fmt.Errorf("%s", result.Message)
}

func (o *PullOptions) waitForCached(repo string) (*model.Tag, error) {

	return nil, nil
}
