package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type NotifyInterface interface {
	Create(ctx context.Context, object *model.Notification) (*model.Notification, error)
	Update(ctx context.Context, id int64, resourceVersion int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*model.Notification, error)
	List(ctx context.Context, opts ...Options) ([]model.Notification, error)
}

type notify struct {
	db *gorm.DB
}

func newNotify(db *gorm.DB) NotifyInterface {
	return &notify{db}
}

func (d *notify) Create(ctx context.Context, object *model.Notification) (*model.Notification, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := d.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (d *notify) Get(ctx context.Context, id int64) (*model.Notification, error) {
	var audit model.Notification
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (d *notify) Delete(ctx context.Context, id int64) error {
	var audit model.Notification
	if err := d.db.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", id).Delete(&audit).Error; err != nil {
		return err
	}

	return nil
}

func (d *notify) Update(ctx context.Context, id int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := d.db.WithContext(ctx).Model(&model.Notification{}).Where("id = ? and resource_version = ?", id, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (d *notify) List(ctx context.Context, opts ...Options) ([]model.Notification, error) {
	var audits []model.Notification
	tx := d.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}
	return audits, nil
}
