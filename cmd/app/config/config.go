package config

type Config struct {
	Default DefaultOption `yaml:"default"`
	Mysql   MysqlOptions  `yaml:"mysql"`

	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Images     []string         `yaml:"images"`

	Server ServerOption `yaml:"server"`

	Plugin   PluginOption `yaml:"plugin"`
	Registry Registry     `yaml:"registry"`

	Agent AgentOption `yaml:"agent"`
}

type DefaultOption struct {
	Listen int    `yaml:"listen"`
	Mode   string `yaml:"mode"` // debug 和 release 模式

	PushKubernetes bool `yaml:"push_kubernetes"`
	PushImages     bool `yaml:"push_images"`
}

type ServerOption struct {
	Auth Auth `yaml:"auth"`
}

type Auth struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
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

type MysqlOptions struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
}

type AgentOption struct {
	Name    string `yaml:"name"`
	DataDir string `yaml:"data_dir"`
}
