package cmd

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	baseURL = "http://14.103.206.75:8090"
)

func NewDefaultPixiuCtlCommand() *cobra.Command {
	return NewPixiuCtlCommand(os.Stdin, os.Stdout, os.Stderr)
}

func NewPixiuCtlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pixiuctl",
		Short: "pixiuctl controls the PixiuHub",
		Long: `pixiuctl controls the PixiuHub.

Find more information at: https://hub.pixiuio.com`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(NewPullCommand())
	cmd.AddCommand(NewTaskCommand())
	cmd.AddCommand(NewImageCommand())
	cmd.AddCommand(NewRegisterCommand())
	cmd.AddCommand(NewConfigCommand())
	cmd.AddCommand(NewVersionCommand())

	// Global config file flag for all subcommands.
	homeDir, err1 := os.UserHomeDir()
	if err1 != nil {
		klog.Fatal(err1)
	}
	defaultConfig := filepath.Join(homeDir, ".pixiu", "config")
	cmd.PersistentFlags().String("configFile", defaultConfig, "Path to the pixiuctl config file")

	// Init logs
	klog.InitFlags(flag.CommandLine)
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
