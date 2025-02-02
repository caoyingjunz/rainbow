package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Registry{})
}

type Registry struct {
	rainbow.Model

	Repository string `json:"repository"`
	Namespace  string `json:"namespace"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func (t *Registry) TableName() string {
	return "registries"
}
