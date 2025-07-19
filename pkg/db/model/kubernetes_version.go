package model

import "github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"

func init() {
	register(&KubernetesVersion{})
}

type KubernetesVersion struct {
	rainbow.Model

	Tag string `json:"tag" gorm:"index:idx"`
}

func (t *KubernetesVersion) TableName() string {
	return "kubernetes_versions"
}
