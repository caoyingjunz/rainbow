package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type RegistryInterface interface {
	Create(ctx context.Context, object *model.Registry) (*model.Registry, error)
	Delete(ctx context.Context, registryId int64) error
	Get(ctx context.Context, registryId int64) (*model.Registry, error)
	List(ctx context.Context, opts ...Options) ([]model.Registry, error)
	ListWithUser(ctx context.Context, userId int64, opts ...Options) ([]model.Registry, error)
}

func newRegistry(db *gorm.DB) RegistryInterface {
	return &registry{db}
}

type registry struct {
	db *gorm.DB
}

func (a *registry) Create(ctx context.Context, object *model.Registry) (*model.Registry, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *registry) Delete(ctx context.Context, registryId int64) error {
	return nil
}

func (a *registry) Get(ctx context.Context, registryId int64) (*model.Registry, error) {
	var audit model.Registry
	if err := a.db.WithContext(ctx).Where("id = ?", registryId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *registry) List(ctx context.Context, opts ...Options) ([]model.Registry, error) {
	var audits []model.Registry
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *registry) ListWithUser(ctx context.Context, userId int64, opts ...Options) ([]model.Registry, error) {
	var audits []model.Registry
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("user_id = ?", userId).Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
