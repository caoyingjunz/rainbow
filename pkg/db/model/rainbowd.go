package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Rainbowd{}, &RainbowdEvent{})
}

type Rainbowd struct {
	rainbow.Model

	Name               string    `json:"name"`
	Host               string    `json:"host"`
	Status             string    `gorm:"column:status;" json:"status"`
	LastTransitionTime time.Time `gorm:"column:last_transition_time;type:datetime;default:current_timestamp;not null" json:"last_transition_time"`
}

func (t *Rainbowd) TableName() string {
	return "rainbowds"
}

type RainbowdEvent struct {
	rainbow.Model

	RainbowdId int64  `json:"rainbowd_id" gorm:"index:idx"`
	Message    string `json:"message"`
}

func (t *RainbowdEvent) TableName() string {
	return "rainbowd_events"
}
