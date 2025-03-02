package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

const (
	PublicVisible  = 0
	PrivateVisible = 1
)

func init() {
	register(&Image{})
}

type Image struct {
	rainbow.Model

	Name     string `json:"name"`
	Target   string `json:"target"`
	TaskId   int64  `gorm:"index:idx_task" json:"task_id"`
	UserId   string `json:"user_id"`
	TaskName string `json:"task_name"`
	Status   string `json:"status"`
	Message  string `json:"message"`

	Tags    string `json:"tags"`
	Visible int    `json:"visible"`

	GmtDeleted time.Time `gorm:"column:gmt_deleted;type:datetime" json:"gmt_deleted"`
	IsDeleted  bool      `json:"is_deleted"`
}

func (t *Image) TableName() string {
	return "images"
}
