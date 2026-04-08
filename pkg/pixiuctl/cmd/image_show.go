package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type ImageShowOptions struct {
	*ImageOptions

	Size  int
	Limit int
	Query string
}

func NewImageShowCommand(base *ImageOptions) *cobra.Command {
	o := &ImageShowOptions{
		ImageOptions: base,
		Size:         1,
		Limit:        10,
	}

	cmd := &cobra.Command{
		Use:   "show <imageId>",
		Short: "Show image tags by image ID",
		Long:  "Show paged tags of one image from PixiuHub by image ID.",
		Example: `  pixiuctl image show 12
  pixiuctl image show 12 --size 1 --limit 20
  pixiuctl image show 12 --query v1 -q latest`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.ValidateShow(args))
			cmdutil.CheckErr(o.RunShow(args[0]))
		},
	}

	cmd.Flags().IntVar(&o.Size, "size", 1, "Page number")
	cmd.Flags().IntVar(&o.Limit, "limit", 10, "Maximum number of tags per page")
	cmd.Flags().StringVarP(&o.Query, "query", "q", "", "query tag name selector")

	return cmd
}

func (o *ImageShowOptions) ValidateShow(args []string) error {
	if err := o.ImageOptions.Validate(nil, nil); err != nil {
		return err
	}
	if _, err := strconv.ParseInt(strings.TrimSpace(args[0]), 10, 64); err != nil {
		return fmt.Errorf("imageId 必须是整数，当前为 %q", args[0])
	}
	if o.Size <= 0 {
		return fmt.Errorf("--size must be at least 1")
	}
	if o.Limit <= 0 {
		return fmt.Errorf("--limit must be at least 1")
	}
	return nil
}

func (o *ImageShowOptions) RunShow(imageIDRaw string) error {
	imageID, _ := strconv.ParseInt(strings.TrimSpace(imageIDRaw), 10, 64)

	pc, err := NewPixiuHubClient(o.baseURL, o.cfg.Auth.AccessKey, o.cfg.Auth.SecretKey)
	if err != nil {
		return err
	}

	items, err := pc.ListImageTags(imageID, o.Size, o.Limit, o.Query)
	if err != nil {
		return err
	}
	PrintImageTagsTable(items)
	return nil
}
