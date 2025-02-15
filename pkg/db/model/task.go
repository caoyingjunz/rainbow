package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Task{})
}

type Task struct {
	rainbow.Model

	Name       string `json:"name"`
	UserId     string `json:"user_id"`
	RegisterId int64  `json:"register_id"`
	AgentName  string `json:"agent_name"`
	Process    int    `json:"process"`
	Mode       int64  `json:"mode"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

func (t *Task) TableName() string {
	return "tasks"
}
