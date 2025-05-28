package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Task{})
}

type Task struct {
	rainbow.Model

	Name              string `json:"name"`
	UserId            string `json:"user_id"`
	UserName          string `json:"user_name"`
	RegisterId        int64  `json:"register_id"`
	AgentName         string `json:"agent_name"`
	Process           int    `json:"process"`
	Mode              int64  `json:"mode"`
	Status            string `json:"status"`
	Message           string `json:"message"`
	Type              int    `json:"type"` // 0：直接指定镜像列表 1: 指定 kubernetes 版本
	KubernetesVersion string `json:"kubernetes_version"`
	Driver            string `json:"driver"` // docker or skopeo
	Namespace         string `json:"namespace"`
	IsPublic          bool   `json:"is_public"`
}

func (t *Task) TableName() string {
	return "tasks"
}
