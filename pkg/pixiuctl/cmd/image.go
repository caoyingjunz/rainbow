package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
)

type ImageOptions struct {
	baseURL string
	cfg     *config.Config

	// flags
	Page     int
	PageSize int
}

func NewImageCommand() *cobra.Command {
	o := &ImageOptions{
		baseURL: baseURL,
	}

	cmd := &cobra.Command{
		Use:   "image",
		Short: "Manage PixiuHub images",
		Long:  `List and show image information from PixiuHub.`,
	}

	cmd.AddCommand(NewImageListCommand(o))

	return cmd
}

func (o *ImageOptions) Complete(cmd *cobra.Command, args []string) error {
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

func (o *ImageOptions) Validate(cmd *cobra.Command, args []string) error {
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

func (o *ImageOptions) RunImageList() error {
	pc, err := NewPixiuHubClient(o.baseURL, o.cfg.Auth.AccessKey, o.cfg.Auth.SecretKey)
	if err != nil {
		return err
	}
	result, err := pc.ListImages(o.Page, o.PageSize)
	if err != nil {
		return err
	}

	PrintImagesTable(result)
	return nil
}

// NewImageListCommand 返回 image list 子命令
func NewImageListCommand(o *ImageOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List images",
		Long:  `List images from PixiuHub.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(cmd, args))
			cmdutil.CheckErr(o.RunImageList())
		},
	}

	cmd.Flags().IntVar(&o.Page, "page", 1, "Page number")
	cmd.Flags().IntVar(&o.PageSize, "page-size", 10, "Page size")

	return cmd
}
