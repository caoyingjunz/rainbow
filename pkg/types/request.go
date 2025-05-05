package types

type (
	CreateLabelRequest struct {
		Name string `json:"name" binding:"required"`
	}

	CreateLogoRequest struct {
		Name string `json:"name" binding:"required"`
		Logo string `json:"logo"`
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
)

// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
	CustomMeta `json:",inline"`

	UserMeta `json:",inline"`
	TaskMeta `json:",inline"`

	PageRequest `json:",inline"` // 分页请求属性
	QueryOption `json:",inline"` // 搜索内容
}
