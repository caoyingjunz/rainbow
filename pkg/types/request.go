package types

type (
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
		Images            []string `json:"images"`
		AgentName         string   `json:"agent_name"`
		Mode              int64    `json:"mode"`
		PublicImage       bool     `json:"public_image"`
		Driver            string   `json:"driver"`
		Logo              string   `json:"logo"`
		Namespace         string   `json:"namespace"`
		IsOfficial        bool     `json:"is_official"`
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
		Name string `json:"name"`
	}

	UpdateAgentStatusRequest struct {
		AgentName string `json:"agent_name"`
		Status    string `json:"status"`
	}

	// PageRequest 分页配置
	PageRequest struct {
		Page  int `form:"page" json:"page"`   // 页数，表示第几页
		Limit int `form:"limit" json:"limit"` // 每页数量
	}
	// QueryOption 搜索配置
	QueryOption struct {
		LabelSelector string `form:"labelSelector" json:"labelSelector"` // 标签搜索
		NameSelector  string `form:"nameSelector" json:"nameSelector"`   // 名称搜索
	}

	CustomMeta struct {
		Status int `form:"status"`
		Limits int `form:"limits"`
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

	RemoteMetaRequest struct {
		Type                    int
		Uid                     string `json:"uid"`
		RepositorySearchRequest RemoteSearchRequest
		TagSearchRequest        RemoteTagSearchRequest
		TagInfoSearchRequest    RemoteTagInfoSearchRequest
	}

	CreateTaskMessageRequest struct {
		Id      int64  `json:"id"`
		Message string `json:"message"`
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
