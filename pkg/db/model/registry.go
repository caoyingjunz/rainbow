package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Registry{})
}

type Registry struct {
	rainbow.Model

	Name       string `json:"name"`
	UserId     string `gorm:"index:idx_user" json:"user_id"` // 所属用户
	Role       int    `json:"role"`                          // 0. 常规权限 1. 超级权限 2. 默认华为仓库
	Repository string `json:"repository"`
	Namespace  string `json:"namespace"`
	Username   string `json:"username"`
	Password   string `json:"password"`

	// 默认华为仓库客户端依赖配置
	RegionId string `json:"region_id"`
	Ak       string `json:"ak"`
	Sk       string `json:"sk"`
}

func (t *Registry) TableName() string {
	return "registries"
}
