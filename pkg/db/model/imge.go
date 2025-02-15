package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Image{})
}

type Image struct {
	rainbow.Model

	Name     string `json:"name"`
	TaskId   int64  `gorm:"index:idx_task" json:"task_id"`
	TaskName string `json:"task_name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

func (t *Image) TableName() string {
	return "images"
}
