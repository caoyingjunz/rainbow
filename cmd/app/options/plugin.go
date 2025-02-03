package options

type Config struct {
	Default DefaultOption `yaml:"default"`
	Mysql   MysqlOptions  `yaml:"mysql"`

	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Images     []string         `yaml:"images"`

	Plugin   PluginOption `yaml:"plugin"`
	Register Register     `yaml:"register"`

	Agent AgentOption `yaml:"agent"`
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

type MysqlOptions struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
}

type AgentOption struct {
	Name string `yaml:"name"`
}
