package db

import (
	"context"
	"fmt"
	"gorm.io/gorm/clause"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type ImageInterface interface {
	Create(ctx context.Context, object *model.Image) (*model.Image, error)
	Update(ctx context.Context, imageId int64, resourceVersion int64, updates map[string]interface{}) error
	Delete(ctx context.Context, imageId int64) error
	Get(ctx context.Context, imageId int64) (*model.Image, error)
	List(ctx context.Context, opts ...Options) ([]model.Image, error)

	UpdateDirectly(ctx context.Context, name string, taskId int64, updates map[string]interface{}) error

	CreateInBatch(ctx context.Context, objects []model.Image) error
	SoftDeleteInBatch(ctx context.Context, taskId int64) error
	ListWithTask(ctx context.Context, taskId int64, opts ...Options) ([]model.Image, error)
	ListWithUser(ctx context.Context, userId string, opts ...Options) ([]model.Image, error)
}

func newImage(db *gorm.DB) ImageInterface {
	return &image{db}
}

type image struct {
	db *gorm.DB
}

func (a *image) Create(ctx context.Context, object *model.Image) (*model.Image, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now
	object.GmtDeleted = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *image) Update(ctx context.Context, imageId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := a.db.WithContext(ctx).Model(&model.Image{}).Where("id = ? and resource_version = ?", imageId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *image) UpdateDirectly(ctx context.Context, name string, taskId int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	f := a.db.WithContext(ctx).Model(&model.Image{}).Where("name = ? and task_id = ?", name, taskId).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *image) CreateInBatch(ctx context.Context, objects []model.Image) error {
	for _, object := range objects {
		if _, err := a.Create(ctx, &object); err != nil {
			return err
		}
	}
	return nil
}

func (a *image) Delete(ctx context.Context, imageId int64) error {
	var audit model.Image
	if err := a.db.Clauses(clause.Returning{}).Where("id = ?", imageId).Delete(&audit).Error; err != nil {
		return err
	}

	return nil
}

func (a *image) SoftDeleteInBatch(ctx context.Context, taskId int64) error {
	updates := make(map[string]interface{})
	updates["gmt_deleted"] = time.Now()
	updates["is_deleted"] = true
	f := a.db.WithContext(ctx).Model(&model.Image{}).Where("task_id = ?", taskId).Updates(updates)
	if f.Error != nil {
		return f.Error
	}

	return nil
}

func (a *image) Get(ctx context.Context, imageId int64) (*model.Image, error) {
	var audit model.Image
	if err := a.db.WithContext(ctx).Where("id = ?", imageId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *image) List(ctx context.Context, opts ...Options) ([]model.Image, error) {
	var audits []model.Image
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("is_deleted = 0").Order("gmt_create DESC").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *image) ListWithTask(ctx context.Context, taskId int64, opts ...Options) ([]model.Image, error) {
	var audits []model.Image
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("task_id = ? and is_deleted = 0", taskId).Order("gmt_create DESC").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *image) ListWithUser(ctx context.Context, userId string, opts ...Options) ([]model.Image, error) {
	var audits []model.Image
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("user_id = ? and is_deleted = 0", userId).Order("gmt_create DESC").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
