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

	GithubUser       string  `json:"github_user"`       // github 后端用户名
	GithubRepository string  `json:"github_repository"` // github 仓库地址
	GithubToken      string  `json:"github_token"`      // github token
	GrossAmount      float64 `json:"gross_amount"`      // github 账号开销金额，每个账号上限 16 美金，达到之后自动下线 agent
}

func (a *Agent) TableName() string {
	return "agents"
}
