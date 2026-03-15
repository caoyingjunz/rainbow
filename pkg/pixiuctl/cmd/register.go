package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/signatureutil"
)

const (
	registerBaseURL = "http://peng:8090"
)

type RegistryListResult struct {
	Code    int             `json:"code"`
	Result  []model.Registry `json:"result,omitempty"`
	Message string          `json:"message,omitempty"`
}

type RegistryResult struct {
	Code    int           `json:"code"`
	Result  model.Registry `json:"result,omitempty"`
	Message string        `json:"message,omitempty"`
}

type RegisterOptions struct {
	baseURL string
	cfg     *config.Config

	accessKey string
	signature string
}

func NewRegisterCommand() *cobra.Command {
	o := &RegisterOptions{
		baseURL: registerBaseURL,
	}

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Manage registries",
		Long:  `List and show registry information from PixiuHub.`,
	}

	cmd.AddCommand(NewRegisterListCommand(o))
	cmd.AddCommand(NewRegisterShowCommand(o))

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

func (o *RegisterOptions) initAuth() error {
	if o.cfg.Auth == nil || len(o.cfg.Auth.AccessKey) == 0 || len(o.cfg.Auth.SecretKey) == 0 {
		return fmt.Errorf("配置文件缺少 Auth 或 access_key/secret_key")
	}
	o.accessKey = o.cfg.Auth.AccessKey
	o.signature = signatureutil.GenerateSignature(
		map[string]string{
			"action":    "listRegistries",
			"accessKey": o.accessKey,
		},
		[]byte(o.cfg.Auth.SecretKey))
	return nil
}

// ListRegistries 调用 /api/v2/registries 获取 registry 列表
func (o *RegisterOptions) ListRegistries() ([]model.Registry, error) {
	url := fmt.Sprintf("%s/api/v2/registries", o.baseURL)

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

// GetRegistry 调用 /api/v2/registries/:id 获取单个 registry（若服务端提供该接口则使用，否则可先仅实现 list）
func (o *RegisterOptions) GetRegistry(id int64) (*model.Registry, error) {
	url := fmt.Sprintf("%s/rainbow/registries/%d", o.baseURL, id)

	var result RegistryResult
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
		return &result.Result, nil
	}
	return nil, fmt.Errorf("%s", result.Message)
}

// NewRegisterListCommand 返回 register list 子命令
func NewRegisterListCommand(o *RegisterOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registries",
		Long:  `List registries from PixiuHub via /api/v2/registries.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.initAuth())
			cmdutil.CheckErr(runRegisterList(o))
		},
	}
}

func runRegisterList(o *RegisterOptions) error {
	list, err := o.ListRegistries()
	if err != nil {
		return err
	}
	printRegistryTable(list)
	return nil
}

func printRegistryTable(registries []model.Registry) {
	// 使用 tabwriter 对齐输出
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tCREATED\tID")
	for _, r := range registries {
		created := r.GmtCreate.Format("2006-01-02 15:04:05")
		fmt.Fprintf(w, "%s\t%s\t%d\n", r.Name, created, r.Id)
	}
}

// NewRegisterShowCommand 返回 register show 子命令
func NewRegisterShowCommand(o *RegisterOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "show [registry-id]",
		Short: "Show a registry by id",
		Long:  `Show detailed information of a registry by id.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.initAuth())
			cmdutil.CheckErr(runRegisterShow(o, args[0]))
		},
	}
}

func runRegisterShow(o *RegisterOptions, idStr string) error {
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return fmt.Errorf("invalid registry id: %s", idStr)
	}
	reg, err := o.GetRegistry(id)
	if err != nil {
		return err
	}
	// 简单表格输出
	printRegistryTable([]model.Registry{*reg})
	return nil
}
