package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&User{})
}

type User struct {
	rainbow.Model

	UserId     string    `gorm:"index:idx_user" json:"user_id"` // 所属用户
	Name       string    `json:"name"`
	UserType   string    `json:"user_type"` // 个人版，专有版
	ExpireTime time.Time `json:"expire_time"`
}

func (t *User) TableName() string {
	return "users"
}
