package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
	"time"
)

func init() {
	register(&User{})
}

type User struct {
	rainbow.Model

	UserId     string    `json:"user_id"`
	Name       string    `json:"name"`
	ExpireTime time.Time `json:"expire_time"`
}

func (t *User) TableName() string {
	return "users"
}
