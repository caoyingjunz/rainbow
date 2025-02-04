package template

import "github.com/caoyingjunz/rainbow/cmd/app/options"

type PluginTemplateConfig struct {
	Default    options.DefaultOption    `yaml:"default"`
	Kubernetes options.KubernetesOption `yaml:"kubernetes"`
	Plugin     options.PluginOption     `yaml:"plugin"`
	Register   options.Register         `yaml:"registry"`
	Images     []string                 `yaml:"images"`
}
