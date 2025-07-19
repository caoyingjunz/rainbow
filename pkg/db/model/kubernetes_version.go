package model

import "github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"

func init() {
	register(&KubernetesVersion{})
}

type KubernetesVersion struct {
	rainbow.Model

	Tag string `gorm:"index:idx_tag,unique" json:"tag"` // k8s, db, ai等标识
}

func (t *KubernetesVersion) TableName() string {
	return "kubernetes_versions"
}
