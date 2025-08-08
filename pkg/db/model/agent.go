package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Agent{}, &Account{})
}

const (
	RunAgentType     string = "在线"
	UnRunAgentType   string = "离线"
	UnknownAgentType string = "未知"
	UnStartType      string = "未启动"

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

type Account struct {
	rainbow.Model

	Type            string    `json:"type"` // dockerhub
	UserName        string    `json:"user_name"`
	Password        string    `json:"password"`
	Token           string    `json:"token"`
	TokenExpireTime time.Time `json:"token_expire_time"` // Token 过期时间
	RetainTimes     int       `json:"retain_times"`      // 剩余可查询次数
}

func (a *Account) TableName() string {
	return "accounts"
}
