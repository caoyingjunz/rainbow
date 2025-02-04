package types

type (
	CreateTaskRequest struct {
		UserId     int64    `json:"user_id"`
		RegisterId int64    `json:"register_id"`
		Images     []string `json:"images"`
		AgentName  string   `json:"agent_name"`
	}

	CreateRegistryRequest struct {
		UserId     int64  `json:"user_id"`
		Repository string `json:"repository"`
		Namespace  string `json:"namespace"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}
)
