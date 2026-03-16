package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
)

type TaskOptions struct {
	baseURL string
	cfg     *config.Config
	client  *PixiuHubClient

	// flag
	Platform string
	Register int
	Name     string
	Private  bool
	Images   []string
}

func NewTaskCommand() *cobra.Command {
	o := &TaskOptions{
		baseURL: baseURL,
	}

	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage image synchronization tasks",
		Long:  `Manage image synchronization tasks on PixiuHub.`,
	}

	createCmd := &cobra.Command{
		Use:   "create [image-repo]",
		Short: "Create a new image synchronization task",
		Long:  `Create a new image synchronization task.`,
		Example: `  # Create a task for a single repo
  pixiuctl task create nginx:latest

  # Create a task with a specific name and specific register
  pixiuctl task create nginx:latest --name my-nginx-task --register <register_id>

  # Create a task for a specific platform
  pixiuctl task create nginx:latest --name <sync_task_name> --platform linux/arm64 --register <register_id>`,
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

	createCmd.Flags().StringVar(&o.Name, "name", "", "name of the image sync task, default is empty")
	createCmd.Flags().StringVar(&o.Platform, "platform", "linux/amd64", "Platform for the image (e.g. linux/amd64, linux/arm64)")
	createCmd.Flags().IntVar(&o.Register, "register", 0, "Image register the task image sync to, default is pixiuhub")
	createCmd.Flags().BoolVar(&o.Private, "private", false, "whether the sync images is private (default false)")

	cmd.AddCommand(createCmd)

	return cmd
}

func (o *TaskOptions) Complete(cmd *cobra.Command, args []string) error {
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

	o.Images = args
	return nil
}

func (o *TaskOptions) Validate(cmd *cobra.Command, args []string) error {
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

func (o *TaskOptions) Run() error {
	pc, err := NewPixiuHubClient(o.baseURL, o.cfg.Auth.AccessKey, o.cfg.Auth.SecretKey)
	if err != nil {
		return err
	}

	return pc.CreateTask(context.TODO(), CreateTaskOption{
		Name:     o.Name,
		Platform: o.Platform,
		Register: o.Register,
		Private:  o.Private,
		Images:   o.Images,
	})
}
