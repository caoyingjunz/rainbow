package model

import (
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
)

func init() {
	register(&Review{}, &Daily{})
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
