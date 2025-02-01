package db

import "gorm.io/gorm"

type AgentInterface interface{}

func newAgent(db *gorm.DB) AgentInterface {
	return &agent{db}
}

type agent struct {
	db *gorm.DB
}
