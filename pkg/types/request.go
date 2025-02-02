package types

type (
	CreateTaskRequest struct {
		Name        string `json:"name" binding:"omitempty"`        // optional
		AliasName   string `json:"alias_name" binding:"omitempty"`  // optional
		KubeConfig  string `json:"kube_config" binding:"required"`  // required
		Description string `json:"description" binding:"omitempty"` // optional
		Protected   bool   `json:"protected" binding:"omitempty"`   // optional
	}
)
