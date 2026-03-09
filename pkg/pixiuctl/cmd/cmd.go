package cmd

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func NewDefaultPixiuCtlCommand() *cobra.Command {
	return NewPixiuCtlCommand(os.Stdin, os.Stdout, os.Stderr)
}

func NewPixiuCtlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pixiuctl",
		Short: "pixiuctl controls the PixiuHub cluster manager",
		Long: `pixiuctl controls the PixiuHub cluster manager.

Find more information at: https://github.com/caoyingjunz/rainbow`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(NewPullCommand())

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
