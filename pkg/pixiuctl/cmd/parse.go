package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type ParseOptions struct {
	File string
}

func NewParseCommand() *cobra.Command {
	o := &ParseOptions{}
	cmd := &cobra.Command{
		Use:   "parse",
		Short: "Parse images from yaml file",
		Long:  "Parse image references from a YAML file.",
		Example: `  # 从 yaml 中提取镜像
  pixiuctl parse -f ./deployment.yaml`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "YAML file path to parse images from")
	return cmd
}

func (o *ParseOptions) Validate() error {
	if strings.TrimSpace(o.File) == "" {
		return fmt.Errorf("请通过 -f/--file 指定 YAML 文件路径")
	}
	return nil
}

func (o *ParseOptions) Run() error {
	data, err := os.ReadFile(o.File)
	if err != nil {
		return err
	}

	var node yaml.Node
	if err = yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("解析 YAML 失败: %w", err)
	}

	images := map[string]struct{}{}
	collectImages(&node, images)

	list := make([]string, 0, len(images))
	for image := range images {
		list = append(list, image)
	}
	sort.Strings(list)

	printImages(list)
	return nil
}

func collectImages(node *yaml.Node, images map[string]struct{}) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, c := range node.Content {
			collectImages(c, images)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			key := strings.ToLower(strings.TrimSpace(keyNode.Value))

			// Most Kubernetes-style manifests use "image".
			if key == "image" && valNode.Kind == yaml.ScalarNode {
				v := strings.TrimSpace(valNode.Value)
				if v != "" {
					images[v] = struct{}{}
				}
			}

			// Support common custom formats with "images: [a,b]" or list items.
			if key == "images" {
				switch valNode.Kind {
				case yaml.SequenceNode:
					for _, item := range valNode.Content {
						if item.Kind == yaml.ScalarNode {
							v := strings.TrimSpace(item.Value)
							if v != "" {
								images[v] = struct{}{}
							}
						}
					}
				case yaml.ScalarNode:
					v := strings.TrimSpace(valNode.Value)
					if v != "" {
						images[v] = struct{}{}
					}
				}
			}

			collectImages(valNode, images)
		}
	case yaml.SequenceNode:
		for _, c := range node.Content {
			collectImages(c, images)
		}
	}
}

func printImages(images []string) {
	if len(images) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "未在 YAML 中发现镜像字段")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	_, _ = fmt.Fprintln(w, "提取IMAGES:")
	for _, image := range images {
		_, _ = fmt.Fprintf(w, "%s\n", image)
	}

	_, _ = fmt.Fprintf(w, "\n加速执行:\npixiuctl pull \\\n")
	for i, image := range images {
		if i == len(images)-1 {
			_, _ = fmt.Fprintf(w, "    %s\n", image)
		} else {
			_, _ = fmt.Fprintf(w, "    %s \\\n", image)
		}
	}
}
