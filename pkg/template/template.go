package template

type PluginTemplateConfig struct {
	Default    DefaultOption    `yaml:"default"`
	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Plugin     PluginOption     `yaml:"plugin"`
	Register   Register         `yaml:"registry"`
	Images     []string         `yaml:"images"`
}

type DefaultOption struct {
	PushKubernetes bool `yaml:"push_kubernetes"`
	PushImages     bool `yaml:"push_images"`
}

type KubernetesOption struct {
	Version string `yaml:"version"`
}

type PluginOption struct {
	Callback string `yaml:"callback"`
}

type Register struct {
	Repository string `yaml:"repository"`
	Namespace  string `yaml:"namespace"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}
