package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Task{})
}

type Task struct {
	rainbow.Model

	AgentName string `json:"name"`
}

func (a *Task) TableName() string {
	return "tasks"
}
