package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Agent{})
}

const (
	RunAgentType     string = "在线"
	UnRunAgentType   string = "离线"
	UnknownAgentType string = "未知"

	PublicAgentType  string = "public"
	PrivateAgentType string = "private"
)

type Agent struct {
	rainbow.Model

	Name               string    `gorm:"index:idx_name,unique" json:"name"`
	LastTransitionTime time.Time `gorm:"column:last_transition_time;type:datetime;default:current_timestamp;not null" json:"last_transition_time"`
	Type               string    `json:"type"`
	Status             string    `gorm:"column:status;" json:"status"`
	Message            string    `json:"message"`
}

func (a *Agent) TableName() string {
	return "agents"
}
