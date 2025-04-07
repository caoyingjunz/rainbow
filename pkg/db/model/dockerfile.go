package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Dockerfile{})
}

type Dockerfile struct {
	rainbow.Model

	Name       string `json:"name"`
	Dockerfile string `json:"dockerfile"`
	UserId     string `json:"user_id"`
}

func (d *Dockerfile) TableName() string {
	return "dockerfiles"
}
