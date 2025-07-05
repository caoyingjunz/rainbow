package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Notification{})
}

type Notification struct {
	rainbow.Model
	rainbow.UserModel

	Type    string `json:"type"` // 支持 webhook, dingding, wecom
	Content string `json:"content"`
}

func (t *Notification) TableName() string {
	return "notifications"
}
