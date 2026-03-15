package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/signatureutil"
)

type RegistryListResult struct {
	Code    int              `json:"code"`
	Result  []model.Registry `json:"result,omitempty"`
	Message string           `json:"message,omitempty"`
}

type RegisterOptions struct {
	baseURL string
	cfg     *config.Config

	accessKey string
	signature string

	user *model.User
}

func NewRegisterCommand() *cobra.Command {
	o := &RegisterOptions{
		baseURL: baseURL,
	}

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Manage PixiuHub registries",
		Long:  `List and show registry information from PixiuHub.`,
	}

	cmd.AddCommand(NewRegisterListCommand(o))

	return cmd
}

func (o *RegisterOptions) Complete(cmd *cobra.Command, args []string) error {
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
	return nil
}

func (o *RegisterOptions) preRun() error {
	if o.cfg.Auth == nil || len(o.cfg.Auth.AccessKey) == 0 || len(o.cfg.Auth.SecretKey) == 0 {
		return fmt.Errorf("配置文件缺少 Auth 或 access_key/secret_key")
	}
	o.accessKey = o.cfg.Auth.AccessKey
	o.signature = signatureutil.GenerateSignature(
		map[string]string{
			"action":    "pullOrCacheRepo",
			"accessKey": o.accessKey,
		},
		[]byte(o.cfg.Auth.SecretKey))

	var err error
	o.user, err = GetUserInfoByAccessKey(o.baseURL, o.accessKey, o.signature)
	if err != nil {
		return err
	}

	return nil
}

// ListRegistries 调用 /api/v2/registries 获取 registry 列表
func (o *RegisterOptions) ListRegistries() ([]model.Registry, error) {
	url := fmt.Sprintf("%s/api/v2/registries?user_id=%s", o.baseURL, o.user.UserId)

	var result RegistryListResult
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method("GET").
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{
			"X-ACCESS-KEY":  o.accessKey,
			"Authorization": o.signature,
		}).
		Do(&result); err != nil {
		return nil, err
	}
	if result.Code == 200 {
		return result.Result, nil
	}
	return nil, fmt.Errorf("%s", result.Message)
}

// NewRegisterListCommand 返回 register list 子命令
func NewRegisterListCommand(o *RegisterOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registries",
		Long:  `List registries from PixiuHub.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.preRun())
			cmdutil.CheckErr(runRegisterList(o))
		},
	}
}

func runRegisterList(o *RegisterOptions) error {
	list, err := o.ListRegistries()
	if err != nil {
		return err
	}
	PrintTable(list)
	return nil
}
