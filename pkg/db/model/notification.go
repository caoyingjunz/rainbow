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

	Name      string `gorm:"index:idx_notifications_name,unique" json:"name"`
	Role      int    `json:"role"`
	Enable    bool   `json:"enable"`
	Type      string `json:"type"` // 支持 webhook, dingtalk, wecom
	Url       string `json:"url"`
	Content   string `json:"content"`
	ShortDesc string `json:"short_desc"`
}

func (t *Notification) TableName() string {
	return "notifications"
}
