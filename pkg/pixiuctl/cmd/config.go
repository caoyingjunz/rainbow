package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

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
	cfgCmd.AddCommand(newConfigShowCommand())

	return cfgCmd
}

func newConfigInitCommand() *cobra.Command {
	o := &ConfigInitOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize pixiuctl configuration file",
		Example: `  # 使用默认配置文件路径（~/.pixiu/config）初始化
  pixiuctl config init --access-key <your-access-key> --secret-key <your-secret-key>

  # 指定超时时间（分钟）
  pixiuctl config init --access-key <your-access-key> --secret-key <your-secret-key> --timeout 15

  # 指定配置文件路径
  pixiuctl --configFile /path/to/config config init --access-key <your-access-key> --secret-key <your-secret-key>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run(cmd)
		},
	}

	cmd.Flags().StringVar(&o.accessKey, "access-key", "", "Access key for PixiuHub")
	cmd.Flags().StringVar(&o.secretKey, "secret-key", "", "Secret key for PixiuHub")
	cmd.Flags().IntVar(&o.timeout, "timeout", 10, "Timeout in minutes for waiting image cache")

	return cmd
}

// newConfigShowCommand shows current pixiuctl configuration.
func newConfigShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current pixiuctl configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 使用 root 命令上的 --configFile 作为配置文件路径（默认为 ~/.pixiu/config）
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

			cfg, err := pixiucfg.LoadConfig(configFile)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "config file: %s\n\n", configFile)

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}

			_, err = out.Write(data)
			return err
		},
	}

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
