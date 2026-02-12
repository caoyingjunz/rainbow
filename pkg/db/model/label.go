package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Label{}, &Logo{}, &ImageLabel{})
}

type Label struct {
	rainbow.Model

	Name   string  `gorm:"index:idx_name,unique" json:"name"` // k8s, db, ai等标识
	Images []Image `gorm:"many2many:image_label;constraint:OnDelete:CASCADE"`
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

// ImageLabel 显式定义关联表，便于扩展字段（如操作人、时间等）
type ImageLabel struct {
	rainbow.Model

	ImageID int64 `gorm:"primaryKey"`
	LabelD  int64 `gorm:"primaryKey"`
}

func (l *ImageLabel) TableName() string {
	return "image_labels"
}
