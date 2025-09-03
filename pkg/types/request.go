package types

import "time"

type (
	UserMetaRequest struct {
		UserId   string `json:"user_id"`
		UserName string `json:"user_name"`
	}

	CreateDockerfileRequest struct {
		Name       string `json:"name"`
		Dockerfile string `json:"dockerfile"`
	}

	UpdateDockerfileRequest struct {
		Id              int64  `json:"id"`
		ResourceVersion int64  `json:"resource_version"`
		Dockerfile      string `json:"dockerfile"`
	}

	CreateLabelRequest struct {
		Name string `json:"name" binding:"required"`
	}

	CreateLogoRequest struct {
		Name string `json:"name" binding:"required"`
		Logo string `json:"logo"`
	}

	UpdateLogoRequest struct {
		Id              int64  `json:"id"`
		ResourceVersion int64  `json:"resource_version"`
		Name            string `json:"name" binding:"required"`
		Logo            string `json:"logo"`
	}

	UpdateLabelRequest struct {
		Id              int64 `json:"id"`
		ResourceVersion int64 `json:"resource_version"`

		Name string `json:"name" binding:"required"`
	}

	CreateTaskRequest struct {
		Name              string   `json:"name"`
		UserId            string   `json:"user_id"`
		UserName          string   `json:"user_name"`
		RegisterId        int64    `json:"register_id"`
		Type              int      `json:"type"` // 0：直接指定镜像列表 1: 指定 kubernetes 版本
		KubernetesVersion string   `json:"kubernetes_version"`
		Images            []string `json:"images"` // 镜像列表中含镜像和架构，格式类似 nginx:1.0.1/amd64，直接指定架构优先级高于任务 arch
		AgentName         string   `json:"agent_name"`
		Mode              int64    `json:"mode"`
		PublicImage       bool     `json:"public_image"`
		Driver            string   `json:"driver"`
		Logo              string   `json:"logo"`
		Namespace         string   `json:"namespace"`
		IsOfficial        bool     `json:"is_official"`
		Architecture      string   `json:"architecture"`
		OwnerRef          int      `json:"owner_ref"` // 任务所属，直接创建 0，订阅创建 1
	}

	UpdateTaskRequest struct {
		Id                int64    `json:"id"`
		ResourceVersion   int64    `json:"resource_version"`
		Name              string   `json:"name"`
		UserId            string   `json:"user_id"`
		RegisterId        int64    `json:"register_id"`
		Type              int      `json:"type"` // 0：直接指定镜像列表 1: 指定 kubernetes 版本
		KubernetesVersion string   `json:"kubernetes_version"`
		AgentName         string   `json:"agent_name"`
		Status            string   `json:"status"`
		Images            []string `json:"images"`
		Mode              int64    `json:"mode"`
		PublicImage       bool     `json:"public_image"`
		OnlyPushError     bool     `json:"only_push_error"` // 仅同步推送异常
	}

	UpdateTaskStatusRequest struct {
		TaskId  int64  `json:"task_id"`
		Status  string `json:"status"`
		Message string `json:"message"`
		Process int    `json:"process"`
	}

	CreateRegistryRequest struct {
		Name       string `json:"name"`
		UserId     string `json:"user_id"`
		Repository string `json:"repository"`
		Namespace  string `json:"namespace"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Role       int    `json:"role"`
	}

	UpdateRegistryRequest struct {
		Id              int64  `json:"id"`
		ResourceVersion int64  `json:"resource_version"`
		Name            string `json:"name"`
		UserId          string `json:"user_id"`
		Repository      string `json:"repository"`
		Namespace       string `json:"namespace"`
		Username        string `json:"username"`
		Password        string `json:"password"`
	}

	CreateImageRequest struct {
		TaskId     int64  `json:"task_id"`
		TaskName   string `json:"task_name"`
		UserId     string `json:"user_id"`
		RegisterId int64  `json:"register_id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		IsPublic   bool   `json:"is_public"`
		IsLocked   bool   `json:"is_locked"`
		Arch       string `json:"arch"`
	}

	CreateImagesRequest struct {
		TaskId   int64    `json:"task_id"`
		TaskName string   `json:"task_name"`
		Names    []string `json:"names"`
	}

	UpdateImageRequest struct {
		Id              int64  `json:"id"`
		ResourceVersion int64  `json:"resource_version"`
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		Label           string `json:"label"`
		IsPublic        bool   `json:"is_public"`
		Logo            string `json:"logo"`
		Description     string `json:"description"`
		IsLocked        bool   `json:"is_locked"`
	}

	UpdateImageStatusRequest struct {
		Name       string `json:"name"` // 对应仓里的镜像名称 比如，nginx
		ImageId    int64  `json:"image_id"`
		TaskId     int64  `json:"task_id"`
		RegistryId int64  `json:"registry_id"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		Target     string `json:"target"`
	}

	CreateNamespaceRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	UpdateNamespaceRequest struct {
		Id              int64  `json:"id"`
		ResourceVersion int64  `json:"resource_version"`
		Name            string `json:"name"`
		Description     string `json:"description"`
	}

	CreateUserRequest struct {
		Name       string `json:"name"`
		UserId     string `json:"user_id"`
		UserType   string `json:"user_type"` // 个人版，专有版
		ExpireTime string `json:"expire_time"`
	}

	UpdateUserRequest struct {
		ResourceVersion int64 `json:"resource_version"`

		CreateUserRequest `json:",inline"`
	}

	CreateAgentRequest struct {
		AgentName        string `json:"agent_name"`
		Type             string `json:"type"`
		GithubUser       string `json:"github_user"`       // github 后端用户名
		GithubRepository string `json:"github_repository"` // github 仓库地址
		GithubToken      string `json:"github_token"`      // github token
		GithubEmail      string `json:"github_email"`
		HealthzPort      int    `json:"healthz_port"`
		RainbowdName     string `json:"rainbowd_name"`
	}

	UpdateAgentRequest struct {
		AgentName string `json:"agent_name"`

		GithubUser       string `json:"github_user"`       // github 后端用户名
		GithubRepository string `json:"github_repository"` // github 仓库地址
		GithubToken      string `json:"github_token"`      // github token
		GithubEmail      string `json:"github_email"`
		HealthzPort      int    `json:"healthz_port"`
		RainbowdName     string `json:"rainbowd_name"`
	}

	UpdateAgentStatusRequest struct {
		AgentName string `json:"agent_name"`
		Status    string `json:"status"`
	}

	CreateNotificationRequest struct {
		UserMetaRequest `json:",inline"`

		Name      string `json:"name"`
		Role      int    `json:"role"` // 1 管理员 0 普通用户
		Enable    bool   `json:"enable"`
		Type      string `json:"type"` // 支持 webhook, dingtalk, wecom
		Url       string `json:"url"`
		Content   string `json:"content"`
		ShortDesc string `json:"short_desc"`
	}
	SendNotificationRequest struct {
		CreateNotificationRequest `json:",inline"`

		Email string `json:"email"`
	}

	// PageRequest 分页配置
	PageRequest struct {
		Page  int `form:"page" json:"page"`   // 页数，表示第几页
		Limit int `form:"limit" json:"limit"` // 每页数量，表示每页几个对象
	}
	// QueryOption 搜索配置
	QueryOption struct {
		LabelSelector string `form:"labelSelector" json:"labelSelector"` // 标签搜索
		NameSelector  string `form:"nameSelector" json:"nameSelector"`   // 名称搜索
	}

	CustomMeta struct {
		Status    int    `form:"status"`
		Namespace string `form:"namespace"`
		Agent     string `form:"agent"`
		OwnerRef  string `form:"ownerRef"`
	}

	RemoteSearchRequest struct {
		Hub      string `json:"hub" form:"hub"`
		ClientId string `json:"client_id" form:"client_id"` // 指定后端执行 clientId
		Query    string `json:"query" form:"query"`
		Page     string `json:"page" form:"page"`
		PageSize string `json:"page_size" form:"page_size"`
	}

	RemoteTagSearchRequest struct {
		Hub        string `json:"hub" form:"hub"`
		ClientId   string `json:"client_id" form:"client_id"`
		Namespace  string `json:"namespace" form:"namespace"`
		Repository string `json:"repository" form:"repository"`
		Query      string `json:"query" form:"query"`
		Page       string `json:"page" form:"page"`
		PageSize   string `json:"page_size" form:"page_size"`
	}

	RemoteTagInfoSearchRequest struct {
		Hub        string `json:"hub" form:"hub"`
		ClientId   string `json:"client_id" form:"client_id"`
		Namespace  string `json:"namespace" form:"namespace"`
		Repository string `json:"repository" form:"repository"`
		Tag        string `json:"tag" form:"tag"`
		Arch       string `json:"arch" form:"arch"`

		Query    string `json:"query" form:"query"`
		Page     string `json:"page" form:"page"`
		PageSize string `json:"page_size" form:"page_size"`
	}

	KubernetesTagRequest struct {
		ClientId string `json:"client_id" form:"client_id"`
		SyncAll  bool   `json:"sync_all"`
	}

	RemoteMetaRequest struct {
		Type                    int
		Uid                     string `json:"uid"`
		RepositorySearchRequest RemoteSearchRequest
		TagSearchRequest        RemoteTagSearchRequest
		TagInfoSearchRequest    RemoteTagInfoSearchRequest
		KubernetesTagRequest    KubernetesTagRequest
	}

	CreateTaskMessageRequest struct {
		Id      int64  `json:"id"`
		Message string `json:"message"`
	}

	CreateSubscribeRequest struct {
		UserMetaRequest `json:",inline"`

		Path       string        `json:"path"`   // 默认会自动填充命名空间，比如 nginx 表示 library/nginx， jenkins/jenkins 则直接使用
		Enable     bool          `json:"enable"` // 启动或者关闭
		Size       int           `json:"size"`   // 同步最新多少个版本
		RegisterId int64         `json:"register_id"`
		Namespace  string        `json:"namespace"`
		Interval   time.Duration `json:"interval"` // 间隔多久同步一次
	}

	UpdateSubscribeRequest struct {
		Id              int64 `json:"id"`
		ResourceVersion int64 `json:"resource_version"`

		Enable   bool          `json:"enable"`   // 启动或者关闭
		Size     int           `json:"size"`     // 同步最新多少个版本
		Interval time.Duration `json:"interval"` // 间隔多久同步一次
	}
)

// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
	CustomMeta `json:",inline"`

	UserMeta `json:",inline"`
	TaskMeta `json:",inline"`

	PageRequest `json:",inline"` // 分页请求属性
	QueryOption `json:",inline"` // 搜索内容
}

func (o *ListOptions) SetDefaultPageOption() {
	// 初始化分页属性
	if o.Page <= 0 {
		o.Page = 1
	}
	if o.Limit <= 0 || o.Limit > 100 {
		o.Limit = 10
	}
}

type PageResult struct {
	PageRequest `json:",inline"`

	Total   int64       `json:"total"`   // 总记录数
	Items   interface{} `json:"items"`   // 数据列表
	Message string      `json:"message"` // 正常或异常信息
}
