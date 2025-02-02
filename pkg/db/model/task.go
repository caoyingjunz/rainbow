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
	Status    string `json:"status"`
	Content   string `json:"content"`
}

func (t *Task) TableName() string {
	return "tasks"
}
