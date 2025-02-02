package options

type Config struct {
	Default DefaultOption `yaml:"default"`
	Mysql   MysqlOptions  `yaml:"mysql"`

	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Register   Repository       `yaml:"register"`
	Images     []string         `yaml:"images"`

	Agent AgentOption `yaml:"agent"`
}

type DefaultOption struct {
	PushKubernetes bool `yaml:"push_kubernetes"`
	PushImages     bool `yaml:"push_images"`
}

type KubernetesOption struct {
	Version string `yaml:"version"`
}

type Repository struct {
	Target RepositoryOption `yaml:"target"`
	Source RepositoryOption `yaml:"source"`
}

type RepositoryOption struct {
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Registry  string `yaml:"registry"`
	Namespace string `yaml:"namespace"`
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
