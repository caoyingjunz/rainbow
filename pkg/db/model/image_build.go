package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Build{})
}

type Build struct {
	rainbow.Model
	rainbow.UserModel

	Name       string `json:"name"`
	Status     string `json:"status"`
	Arch       string `json:"arch"`        // 架构
	Dockerfile string `json:"dockerfile"`  // 镜像构建 dockerfile
	RegistryId int64  `json:"registry_id"` // 推送镜像仓库
	AgentName  string `json:"agent_name"`  // 执行代理
}

func (b *Build) TableName() string {
	return "builds"
}
