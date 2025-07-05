package rainbow

import (
	"strconv"
	"time"
)

type Model struct {
	Id              int64     `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	GmtCreate       time.Time `gorm:"column:gmt_create;type:datetime;default:current_timestamp;not null" json:"gmt_create"`
	GmtModified     time.Time `gorm:"column:gmt_modified;type:datetime;default:current_timestamp;not null" json:"gmt_modified"`
	ResourceVersion int64     `gorm:"column:resource_version;default:0;not null" json:"resource_version"`
}

func (m Model) GetSID() string {
	return strconv.FormatInt(m.Id, 10)
}

type UserModel struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
}
