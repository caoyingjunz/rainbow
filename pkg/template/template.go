package template

type PluginTemplateConfig struct {
	Default    DefaultOption    `yaml:"default"`
	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Plugin     PluginOption     `yaml:"plugin"`
	Registry   Registry         `yaml:"registry"`
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
	TaskId   int64  `yaml:"task_id"`
	Synced   bool   `yaml:"synced"`
}

type Registry struct {
	Repository string `yaml:"repository"`
	Namespace  string `yaml:"namespace"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}
