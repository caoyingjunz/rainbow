package model

import (
	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Image{})
	register(&Tag{})
}

type Image struct {
	rainbow.Model

	GmtDeleted gorm.DeletedAt

	Name       string `json:"name"`
	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	RegisterId int64  `json:"register_id"`

	Logo      string `json:"logo"`
	Path      string `json:"path"`
	Namespace string `json:"namespace"`
	Mirror    string `json:"mirror"`
	Size      int64  `json:"size"`
	Tags      []Tag  `json:"tags" gorm:"foreignKey:ImageId;constraint:OnDelete:CASCADE;"`

	IsPublic      bool `json:"is_public"`
	PublicUpdated bool `json:"public_updated"` // 是否已经同步过远端仓库状态

	Description string `json:"description"`
}

func (t *Image) TableName() string {
	return "images"
}

type Tag struct {
	rainbow.Model

	GmtDeleted gorm.DeletedAt

	ImageId int64  `gorm:"index:idx_image" json:"image_id"`
	Path    string `json:"path"`
	TaskId  int64  `json:"task_id"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Status  string `json:"status"`
	Message string `json:"message"` // 错误信息
}

func (t *Tag) TableName() string {
	return "tags"
}
