package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Registry{})
}

type Registry struct {
	rainbow.Model

	// 所属用户
	UserId     int64  `gorm:"index:idx_user" json:"user_id"`
	Repository string `json:"repository"`
	Namespace  string `json:"namespace"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func (t *Registry) TableName() string {
	return "registries"
}
