package model

import "github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"

func init() {
	register(&Label{})
}

type Label struct {
	rainbow.Model

	Name string `gorm:"index:idx_name,unique" json:"name"`
	Logo string `json:"logo"`
}

func (l *Label) TableName() string {
	return "labels"
}
