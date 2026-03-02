package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type MetricsInterface interface {
	Create(ctx context.Context, object *model.Metrics) (*model.Metrics, error)
	Delete(ctx context.Context, opts ...Options) error
	List(ctx context.Context, opts ...Options) ([]model.Metrics, error)
}

type metrics struct {
	db *gorm.DB
}

func newMetrics(db *gorm.DB) MetricsInterface {
	return &metrics{db: db}
}

func (c *metrics) Create(ctx context.Context, object *model.Metrics) (*model.Metrics, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := c.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (c *metrics) Delete(ctx context.Context, opts ...Options) error {
	return nil
}

func (c *metrics) List(ctx context.Context, opts ...Options) ([]model.Metrics, error) {
	var audits []model.Metrics
	tx := c.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
