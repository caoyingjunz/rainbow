package types

type (
	CreateTaskRequest struct {
	}

	CreateRegistryRequest struct {
		UserId     int64  `json:"user_id"`
		Repository string `json:"repository"`
		Namespace  string `json:"namespace"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}
)
