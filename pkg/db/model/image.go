package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Image{})
	register(&Tag{})
}

type Image struct {
	rainbow.Model

	Name     string `json:"name"`
	TaskId   int64  `gorm:"index:idx_task" json:"task_id"`
	TaskName string `json:"task_name"`

	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	RegisterId int64  `json:"register_id"`
	Status     string `json:"status"`

	GmtDeleted time.Time `gorm:"column:gmt_deleted;type:datetime" json:"gmt_deleted"`
	IsDeleted  bool      `json:"is_deleted"`

	Logo      string `json:"logo"`
	Path      string `json:"path"`
	Namespace string `json:"namespace"`
	Mirror    string `json:"mirror"`
	Size      int64  `json:"size"`
	Tags      []Tag  `json:"tags" gorm:"foreignKey:ImageId"`

	IsPublic      bool `json:"is_public"`
	PublicUpdated bool `json:"public_updated"` // 是否已经同步过远端仓库状态

	Description string `json:"description"`
}

func (t *Image) TableName() string {
	return "images"
}

type Tag struct {
	rainbow.Model

	ImageId int64  `gorm:"index:idx_image" json:"image_id"`
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Status  string `json:"status"`
}

func (t *Tag) TableName() string {
	return "tags"
}
