package db

import "gorm.io/gorm"

type DriverInterface interface{}

func newDriver(db *gorm.DB) DriverInterface {
	return &driver{db}
}

type driver struct {
	db *gorm.DB
}
