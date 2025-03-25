package db

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"gorm.io/gorm"
	"time"
)

type LabelInterface interface {
	FindByIds(ctx context.Context, ids []int64, labels *[]*model.Label) error
	FindByNames(ctx context.Context, names []string, labels *[]*model.Label) error
	Create(ctx context.Context, object *model.Label) (*model.Label, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, labelId int64, resourceVersion int64, updates map[string]interface{}) error
	List(ctx context.Context, opts ...Options) ([]model.Label, error)
	ListLabelsInfo(ctx context.Context, opts ...Options) ([]model.Label, error)
}

func newLabel(db *gorm.DB) LabelInterface {
	return &label{db}
}

type label struct {
	db *gorm.DB
}

func (l *label) FindByNames(ctx context.Context, names []string, labels *[]*model.Label) error {
	if err := l.db.WithContext(ctx).Where("name IN ?", names).Find(labels).Error; err != nil {
		return err
	}
	return nil
}

func (l *label) FindByIds(ctx context.Context, ids []int64, labels *[]*model.Label) error {
	if err := l.db.WithContext(ctx).Where("id IN ?", ids).Find(labels).Error; err != nil {
		return err
	}
	return nil
}

func (l *label) FindByID(ctx context.Context, id int64) (*model.Label, error) {
	var label model.Label
	if err := l.db.WithContext(ctx).First(&label, id).Error; err != nil {
		return nil, err
	}
	return &label, nil
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

func (l *label) Delete(ctx context.Context, id int64) error {
	return l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("images_labels").Where("label_id = ?", id).Delete(nil).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.Label{}, id).Error; err != nil {
			return err
		}

		return nil
	})
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

func (l *label) ListLabelsInfo(ctx context.Context, opts ...Options) ([]model.Label, error) {
	var audits []model.Label
	tx := l.db.WithContext(ctx)

	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Preload("Images").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
