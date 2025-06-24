package db

import (
	"context"
	"fmt"

	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type LabelInterface interface {
	Create(ctx context.Context, object *model.Label) (*model.Label, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, labelId int64, resourceVersion int64, updates map[string]interface{}) error
	List(ctx context.Context, opts ...Options) ([]model.Label, error)

	CreateLogo(ctx context.Context, object *model.Logo) (*model.Logo, error)
	UpdateLogo(ctx context.Context, labelId int64, resourceVersion int64, updates map[string]interface{}) error
	DeleteLogo(ctx context.Context, logoId int64) error
	ListLogos(ctx context.Context, opts ...Options) ([]model.Logo, error)
}

func newLabel(db *gorm.DB) LabelInterface {
	return &label{db}
}

type label struct {
	db *gorm.DB
}

func (l *label) Create(ctx context.Context, object *model.Label) (*model.Label, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := l.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (l *label) Update(ctx context.Context, labelId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := l.db.WithContext(ctx).Model(&model.Label{}).Where("id = ? and resource_version = ?", labelId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (l *label) Delete(ctx context.Context, labelId int64) error {
	return l.db.WithContext(ctx).Where("id = ?", labelId).Delete(&model.Label{}).Error
}

func (l *label) List(ctx context.Context, opts ...Options) ([]model.Label, error) {
	var audits []model.Label
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (l *label) CreateLogo(ctx context.Context, object *model.Logo) (*model.Logo, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := l.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (l *label) UpdateLogo(ctx context.Context, labelId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := l.db.WithContext(ctx).Model(&model.Logo{}).Where("id = ? and resource_version = ?", labelId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (l *label) DeleteLogo(ctx context.Context, logoId int64) error {
	return l.db.WithContext(ctx).Where("id = ?", logoId).Delete(&model.Logo{}).Error
}

func (l *label) ListLogos(ctx context.Context, opts ...Options) ([]model.Logo, error) {
	var audits []model.Logo
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
