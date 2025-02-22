package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Image{})
}

type Image struct {
	rainbow.Model
	GmtDeleted time.Time `gorm:"column:gmt_deleted;type:datetime" json:"gmt_deleted"`

	Name     string `json:"name"`
	Target   string `json:"target"`
	TaskId   int64  `gorm:"index:idx_task" json:"task_id"`
	TaskName string `json:"task_name"`
	Status   string `json:"status"`
	Message  string `json:"message"`

	IsDeleted bool `json:"is_deleted"`
}

func (t *Image) TableName() string {
	return "images"
}
