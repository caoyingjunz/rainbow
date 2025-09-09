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

	Name      string `json:"name"`
	Role      int    `json:"role"`
	Enable    bool   `json:"enable"`
	Type      string `json:"type"`     // 支持 webhook, dingtalk, wecom, feishu, email, sms, lark
	PushCfg   string `json:"push_cfg"` // json 格式
	ShortDesc string `json:"short_desc"`
}

func (t *Notification) TableName() string {
	return "notifications"
}
