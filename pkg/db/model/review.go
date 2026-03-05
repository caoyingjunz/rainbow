package model

import (
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

	RecordDay string `json:"record_day"`

	Pull  int64 `json:"pull"`
	Task  int64 `json:"task"`
	Image int64 `json:"image"`
	Tags  int64 `json:"tags"`
}

func (s *Metrics) TableName() string {
	return "metrics"
}
