package config

type Config struct {
	Default DefaultOption `yaml:"default"`

	Mysql MysqlOptions `yaml:"mysql"`
	Redis RedisOption  `yaml:"redis"`

	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Images     []Image          `yaml:"images"`

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

	Time int64 `yaml:"time"`
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
	Callback   string `yaml:"callback"`
	TaskId     int64  `yaml:"task_id"`
	RegistryId int64  `yaml:"registry_id"`
	Synced     bool   `yaml:"synced"`
	Driver     string `yaml:"driver"`
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

type RedisOption struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type AgentOption struct {
	Name      string `yaml:"name"`
	DataDir   string `yaml:"data_dir"`
	RpcServer string `yaml:"rpc_server"`
}

type PluginTemplateConfig struct {
	Default    DefaultOption    `yaml:"default"`
	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Plugin     PluginOption     `yaml:"plugin"`
	Registry   Registry         `yaml:"registry"`
	Images     []Image          `yaml:"images"`
}

type Image struct {
	Name string   `yaml:"name"`
	Id   int64    `yaml:"id"`
	Path string   `yaml:"path"`
	Tags []string `yaml:"tags"`
}

func (i Image) GetMap(repo, ns string) map[string]string {
	m := make(map[string]string)
	for _, tag := range i.Tags {
		m[i.Path+":"+tag] = repo + "/" + ns + "/" + i.Name + ":" + tag
	}
	return m
}

func (i Image) GetId() int64 {
	return i.Id
}

func (i Image) GetPath() string {
	return i.Path
}
