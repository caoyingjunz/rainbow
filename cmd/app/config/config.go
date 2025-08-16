package config

const (
	DefaultNormalRateLimitMaxRequests  = 100
	DefaultSpecialRateLimitMaxRequests = 50
	DefaultSpecialRateLimitedPath      = "/rainbow/search"
	DefaultUserRateLimitCap            = 1000
	DefaultUserRateLimitQuantum        = 10
	DefaultUserRateLimitCapacity       = 100

	defaultRainbowdTemplateDir = "/data/template"
)

// SetDefaults 设置配置的默认值
func (c *Config) SetDefaults() {
	if c.RateLimit.NormalRateLimit.MaxRequests == 0 {
		c.RateLimit.NormalRateLimit.MaxRequests = DefaultNormalRateLimitMaxRequests
	}
	if c.RateLimit.SpecialRateLimit.MaxRequests == 0 {
		c.RateLimit.SpecialRateLimit.MaxRequests = DefaultSpecialRateLimitMaxRequests
	}
	if c.RateLimit.UserRateLimit.Cap == 0 {
		c.RateLimit.UserRateLimit.Cap = DefaultUserRateLimitCap
	}
	if c.RateLimit.UserRateLimit.Quantum == 0 {
		c.RateLimit.UserRateLimit.Quantum = DefaultUserRateLimitQuantum
	}
	if c.RateLimit.UserRateLimit.Capacity == 0 {
		c.RateLimit.UserRateLimit.Capacity = DefaultUserRateLimitCapacity
	}
	if c.RateLimit.SpecialRateLimit.RateLimitedPath == nil {
		c.RateLimit.SpecialRateLimit.RateLimitedPath = []string{DefaultSpecialRateLimitedPath}
	}
}

type Config struct {
	Default DefaultOption `yaml:"default"`

	Mysql MysqlOptions `yaml:"mysql"`
	Redis RedisOption  `yaml:"redis"`

	Kubernetes KubernetesOption `yaml:"kubernetes"`
	Images     []Image          `yaml:"images"`

	Server   ServerOption   `yaml:"server"`
	Rainbowd RainbowdOption `yaml:"rainbowd"`

	Plugin   PluginOption `yaml:"plugin"`
	Registry Registry     `yaml:"registry"`

	Agent AgentOption `yaml:"agent"`

	RateLimit RateLimitOption `yaml:"rate_limit"`
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

type RainbowdOption struct {
	Name        string `yaml:"name"`
	TemplateDir string `yaml:"template_dir"`
	DataDir     string `yaml:"data_dir"`
	AgentImage  string `yaml:"agent_image"`
}

func (r *RainbowdOption) SetDefault() {
	if len(r.TemplateDir) == 0 {
		r.TemplateDir = defaultRainbowdTemplateDir
	}
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
	Arch       string `yaml:"arch"`
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
	Name        string `yaml:"name"`
	DataDir     string `yaml:"data_dir"`
	RpcServer   string `yaml:"rpc_server"`
	RetainDays  int    `yaml:"retain_days"`
	HealthzPort int    `yaml:"healthz_port"`
}

type RateLimitOption struct {
	NormalRateLimit  NormalRateLimit  `yaml:"normal_rate_limit"`
	SpecialRateLimit SpecialRateLimit `yaml:"special_rate_limit"`
	UserRateLimit    UserRateLimit    `yaml:"user_rate_limit"`
}

type UserRateLimit struct {
	Cap      int `yaml:"cap"`
	Quantum  int `yaml:"quantum"`
	Capacity int `yaml:"capacity"`
}

type NormalRateLimit struct {
	MaxRequests int `yaml:"max_requests"`
}
type SpecialRateLimit struct {
	RateLimitedPath []string `yaml:"rate_limited_path"`
	MaxRequests     int      `yaml:"max_requests"`
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
