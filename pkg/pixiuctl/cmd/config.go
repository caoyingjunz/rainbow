package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pixiucfg "github.com/caoyingjunz/rainbow/pkg/pixiuctl/config"
)

type ConfigInitOptions struct {
	accessKey string
	secretKey string
	timeout   int
}

func NewConfigCommand() *cobra.Command {
	cfgCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage pixiuctl configuration",
	}

	cfgCmd.AddCommand(newConfigInitCommand())

	return cfgCmd
}

func newConfigInitCommand() *cobra.Command {
	o := &ConfigInitOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize pixiuctl configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(cmd)
		},
	}

	cmd.Flags().StringVar(&o.accessKey, "access-key", "", "Access key for PixiuHub")
	cmd.Flags().StringVar(&o.secretKey, "secret-key", "", "Secret key for PixiuHub")
	cmd.Flags().IntVar(&o.timeout, "timeout", 10, "Timeout in minutes for waiting image cache")

	return cmd
}

func (o *ConfigInitOptions) Run(cmd *cobra.Command) error {
	if len(o.accessKey) == 0 {
		return fmt.Errorf("access-key 必须指定")
	}
	if len(o.secretKey) == 0 {
		return fmt.Errorf("secret-key 必须指定")
	}

	// 默认使用 root 命令上的 --configFile 作为配置文件路径（默认为 ~/.pixiu/config）
	configFile, err := cmd.Root().PersistentFlags().GetString("configFile")
	if err != nil {
		return err
	}
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configFile = filepath.Join(homeDir, ".pixiu", "config")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configFile), 0o755); err != nil {
		return err
	}

	cfg := &pixiucfg.Config{
		Default: &pixiucfg.DefaultConfig{
			Timeout: o.timeout,
		},
		Auth: &pixiucfg.AuthConfig{
			AccessKey: o.accessKey,
			SecretKey: o.secretKey,
		},
	}

	if err := pixiucfg.SaveConfig(configFile, cfg); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "pixiuctl config saved to: %s\n", configFile)
	return nil
}
