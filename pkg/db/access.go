package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type AccessInterface interface {
	Create(ctx context.Context, object *model.Access) (*model.Access, error)
	Delete(ctx context.Context, ak string) error
	Get(ctx context.Context, ak string) (*model.Access, error)
	List(ctx context.Context, opts ...Options) ([]model.Access, error)
	Count(ctx context.Context, opts ...Options) (int64, error)
}

type access struct {
	db *gorm.DB
}

func newAccess(db *gorm.DB) AccessInterface {
	return &access{db: db}
}

func (a *access) Create(ctx context.Context, object *model.Access) (*model.Access, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *access) Get(ctx context.Context, ak string) (*model.Access, error) {
	var audit model.Access
	if err := a.db.WithContext(ctx).Where("access_key = ?", ak).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *access) Delete(ctx context.Context, ak string) error {
	return a.db.WithContext(ctx).Where("access_key = ?", ak).Delete(&model.Access{}).Error
}

func (a *access) List(ctx context.Context, opts ...Options) ([]model.Access, error) {
	var list []model.Access
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (a *access) Count(ctx context.Context, opts ...Options) (int64, error) {
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Access{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
