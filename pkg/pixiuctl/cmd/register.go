package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
)

type RegisterOptions struct {
	baseURL string
	cfg     *config.Config
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

func (o *RegisterOptions) Validate(cmd *cobra.Command, args []string) error {
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

func (o *RegisterOptions) RunRegisterList() error {
	pc, err := NewPixiuHubClient(o.baseURL, o.cfg.Auth.AccessKey, o.cfg.Auth.SecretKey)
	if err != nil {
		return err
	}
	registries, err := pc.ListRegistries()
	if err != nil {
		return err
	}

	PrintTable(registries)
	return nil
}

// NewRegisterListCommand 返回 register list 子命令
func NewRegisterListCommand(o *RegisterOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registries",
		Long:  `List registries from PixiuHub.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.RunRegisterList())
		},
	}
}
