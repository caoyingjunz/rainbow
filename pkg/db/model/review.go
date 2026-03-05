package model

import (
	"time"

	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Review{}, &Metrics{})
}

type Review struct {
	rainbow.Model

	Count int64 `json:"count"`
}

func (t *Review) TableName() string {
	return "reviews"
}

type Daily struct {
	rainbow.Model

	Page string `json:"page"`
}

func (t *Daily) TableName() string {
	return "dailies"
}

type Metrics struct {
	rainbow.Model

	RecordTime time.Time `json:"record_time"`

	Pull      int64  `json:"pull"`
	Task      int64  `json:"task"`
	Image     int64  `json:"image"`
	RecordDay string `json:"record_day"`
}

func (s *Metrics) TableName() string {
	return "metrics"
}
