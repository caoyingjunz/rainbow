package model

import "github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"

func init() {
	register(&Label{})
	register(&Logo{})

}

type Label struct {
	rainbow.Model

	Name string `gorm:"index:idx_name,unique" json:"name"` // k8s, db, ai等标识
}

func (l *Label) TableName() string {
	return "labels"
}

type Logo struct {
	rainbow.Model

	Name string `json:"name"`
	Logo string `json:"logo"`
}

func (l *Logo) TableName() string {
	return "logos"
}
