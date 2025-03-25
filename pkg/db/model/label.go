package model

import "github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"

func init() {
	register(&Label{})
}

type Label struct {
	rainbow.Model

	Name   string   `gorm:"index:idx_name,unique" json:"name"`
	Images []*Image `gorm:"many2many:images_labels;" json:"images,omitempty"` // 多对多关系
}

func (l *Label) TableName() string {
	return "labels"
}
