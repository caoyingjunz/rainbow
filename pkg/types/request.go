package types

type (
	CreateTaskRequest struct {
		Name       string   `json:"name"`
		UserId     string   `json:"user_id"`
		RegisterId int64    `json:"register_id"`
		Images     []string `json:"images"`
		AgentName  string   `json:"agent_name"`
		Mode       int64    `json:"mode"`
	}

	UpdateTaskRequest struct {
		Id              int64    `json:"id"`
		ResourceVersion int64    `json:"resource_version"`
		Name            string   `json:"name"`
		UserId          string   `json:"user_id"`
		RegisterId      int64    `json:"register_id"`
		AgentName       string   `json:"agent_name"`
		Status          string   `json:"status"`
		Images          []string `json:"images"`
		Mode            int64    `json:"mode"`
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
		TaskId   int64  `json:"task_id"`
		TaskName string `json:"task_name"`
		Name     string `json:"name"`
		Status   string `json:"status"`
		Message  string `json:"message"`
	}

	UpdateImageRequest struct {
		Id              int64  `json:"id"`
		TaskId          int64  `json:"task_id"`
		TaskName        string `json:"task_name"`
		ResourceVersion int64  `json:"resource_version"`
		Name            string `json:"name"`
		Status          string `json:"status"`
		Message         string `json:"message"`
	}

	UpdateImageStatusRequest struct {
		Name     string `json:"name"`
		TaskId   int64  `json:"task_id"`
		TaskName string `json:"task_name"`
		Status   string `json:"status"`
		Message  string `json:"message"`
		Target   string `json:"target"`
	}

	UpdateAgentStatusRequest struct {
		AgentName string `json:"agent_name"`
		Status    string `json:"status"`
	}
)
