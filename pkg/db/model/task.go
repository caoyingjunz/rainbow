package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
	"time"
)

func init() {
	register(&Task{}, &TaskMessage{}, &Subscribe{}, &SubscribeMessage{})
}

type Task struct {
	rainbow.Model

	Name              string `json:"name"`
	UserId            string `json:"user_id"`
	UserName          string `json:"user_name"`
	RegisterId        int64  `json:"register_id"`
	AgentName         string `json:"agent_name"`
	Process           int    `json:"process"`
	Mode              int64  `json:"mode"`
	Status            string `json:"status"`
	Message           string `json:"message"`
	Type              int    `json:"type"` // 0：直接指定镜像列表 1: 指定 kubernetes 版本
	KubernetesVersion string `json:"kubernetes_version"`
	Driver            string `json:"driver"` // docker or skopeo
	Namespace         string `json:"namespace"`
	IsPublic          bool   `json:"is_public"`
	IsOfficial        bool   `json:"is_official"`
	Logo              string `json:"logo"`
}

func (t *Task) TableName() string {
	return "tasks"
}

type TaskMessage struct {
	rainbow.Model

	TaskId  int64  `json:"task_id" gorm:"index:idx"`
	Message string `json:"message"`
}

func (t *TaskMessage) TableName() string {
	return "task_messages"
}

type Subscribe struct { // 同步远端镜像更新状态
	rainbow.Model
	rainbow.UserModel

	Path           string        `json:"path"` // 默认会自动填充命名空间，比如 nginx 表示 library/nginx， jenkins/jenkins 则直接使用
	RawPath        string        `json:"raw_path"`
	Enable         bool          `json:"enable"` // 启动或者关闭
	Status         string        `json:"status"`
	Limit          int           `json:"limit"` // 同步最新多少个版本
	RegisterId     int64         `json:"register_id"`
	Namespace      string        `json:"namespace"`
	LastNotifyTime time.Time     `json:"last_notify_time" gorm:"column:last_notify_time;type:datetime;default:current_timestamp;not null"` // 上次触发时间
	Interval       time.Duration `json:"interval"`                                                                                         // 间隔多久同步一次
	FailTimes      int           `json:"fail_times"`                                                                                       // 失败次数
}

func (t *Subscribe) TableName() string {
	return "subscribes"
}

type SubscribeMessage struct {
	rainbow.Model

	SubscribeId int64  `json:"subscribe_id" gorm:"index:idx"`
	Message     string `json:"message"`
}

func (t *SubscribeMessage) TableName() string {
	return "subscribe_messages"
}
