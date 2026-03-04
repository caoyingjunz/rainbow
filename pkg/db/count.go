package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type CountInterface interface {
	Create(ctx context.Context, object *model.Count) (*model.Count, error)
	Delete(ctx context.Context, opts ...Options) error
	List(ctx context.Context, opts ...Options) ([]model.Count, error)
}

type count struct {
	db *gorm.DB
}

func newCount(db *gorm.DB) CountInterface {
	return &count{db: db}
}

func (c *count) Create(ctx context.Context, object *model.Count) (*model.Count, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := c.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (c *count) Delete(ctx context.Context, opts ...Options) error {
	return nil
}

func (c *count) List(ctx context.Context, opts ...Options) ([]model.Count, error) {
	var audits []model.Count
	tx := c.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
