package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type ImageInterface interface {
	Create(ctx context.Context, object *model.Image) (*model.Image, error)
	Update(ctx context.Context, imageId int64, resourceVersion int64, updates map[string]interface{}) error
	Delete(ctx context.Context, imageId int64) error
	Get(ctx context.Context, imageId int64, del bool) (*model.Image, error)
	List(ctx context.Context, opts ...Options) ([]model.Image, error)

	CreateFlow(ctx context.Context, object *model.Downflow) error

	CreateInBatch(ctx context.Context, objects []model.Image) error
	DeleteInBatch(ctx context.Context, ids []int64) error
	SoftDeleteInBatch(ctx context.Context, taskId int64) error
	ListWithTask(ctx context.Context, taskId int64, opts ...Options) ([]model.Image, error)
	ListWithUser(ctx context.Context, userId string, opts ...Options) ([]model.Image, error)

	Count(ctx context.Context, opts ...Options) (int64, error)

	GetByPath(ctx context.Context, path string, mirror string, opts ...Options) (*model.Image, error)
	ListImagesWithTag(ctx context.Context, opts ...Options) ([]model.Image, error)

	CreateTag(ctx context.Context, object *model.Tag) (*model.Tag, error)
	UpdateTag(ctx context.Context, imageId int64, tag string, updates map[string]interface{}) error
	DeleteTag(ctx context.Context, imageId int64, name string) error
	GetTag(ctx context.Context, imageId int64, name string, del bool) (*model.Tag, error)
	ListTags(ctx context.Context, opts ...Options) ([]model.Tag, error)

	CreateNamespace(ctx context.Context, object *model.Namespace) (*model.Namespace, error)
	UpdateNamespace(ctx context.Context, namespaceId int64, resourceVersion int64, updates map[string]interface{}) error
	DeleteNamespace(ctx context.Context, namespaceId int64) error
	ListNamespaces(ctx context.Context, opts ...Options) ([]model.Namespace, error)
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
	if err := a.db.Clauses(clause.Returning{}).Select("Tags").Where("id = ?", imageId).Delete(&audit).Error; err != nil {
		return err
	}

	return nil
}

func (a *image) DeleteInBatch(ctx context.Context, ids []int64) error {
	for _, id := range ids {
		if err := a.Delete(ctx, id); err != nil {
			return err
		}
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
func (a *image) Get(ctx context.Context, imageId int64, del bool) (*model.Image, error) {
	tx := a.db.WithContext(ctx)
	if del {
		tx = tx.Unscoped()
	}

	var audit model.Image
	if err := tx.Preload("Tags").Where("id = ?", imageId).First(&audit).Error; err != nil {
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
	if err := tx.Find(&audits).Error; err != nil {
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

func (a *image) Count(ctx context.Context, opts ...Options) (int64, error) {
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Image{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (a *image) GetByPath(ctx context.Context, path string, mirror string, opts ...Options) (*model.Image, error) {
	var audit model.Image
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.WithContext(ctx).Where("path = ? and mirror = ?", path, mirror).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *image) ListImagesWithTag(ctx context.Context, opts ...Options) ([]model.Image, error) {
	var audits []model.Image
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Preload("Tags").Find(&audits).Error; err != nil {
		return nil, err
	}
	return audits, nil
}

func (a *image) CreateTag(ctx context.Context, object *model.Tag) (*model.Tag, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now
	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *image) CreateTagsInBatch(ctx context.Context, objects []model.Tag) error {
	for _, object := range objects {
		if _, err := a.CreateTag(ctx, &object); err != nil {
			return err
		}
	}
	return nil
}

func (a *image) DeleteTag(ctx context.Context, imageId int64, name string) error {
	var audit model.Tag
	if err := a.db.Clauses(clause.Returning{}).Where("image_id = ? and name = ?", imageId, name).Delete(&audit).Error; err != nil {
		return err
	}
	return nil
}

func (a *image) ListTags(ctx context.Context, opts ...Options) ([]model.Tag, error) {
	var audits []model.Tag
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}
	return audits, nil
}

func (a *image) GetTag(ctx context.Context, imageId int64, name string, del bool) (*model.Tag, error) {
	tx := a.db.WithContext(ctx)
	if del {
		tx = tx.Unscoped()
	}
	var audit model.Tag
	if err := tx.Where("image_id = ? and name = ?", imageId, name).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *image) UpdateTag(ctx context.Context, imageId int64, tag string, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	f := a.db.WithContext(ctx).Model(&model.Tag{}).Where("image_id = ? and name = ?", imageId, tag).Updates(updates)
	if f.Error != nil {
		return f.Error
	}

	return nil
}

func (a *image) CreateFlow(ctx context.Context, object *model.Downflow) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	return a.db.WithContext(ctx).Create(object).Error
}

func (a *image) CreateNamespace(ctx context.Context, object *model.Namespace) (*model.Namespace, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *image) UpdateNamespace(ctx context.Context, namespaceId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := a.db.WithContext(ctx).Model(&model.Namespace{}).Where("id = ? and resource_version = ?", namespaceId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *image) DeleteNamespace(ctx context.Context, namespaceId int64) error {
	var audit model.Namespace
	return a.db.Clauses(clause.Returning{}).Where("id = ? ", namespaceId).Delete(&audit).Error
}

func (a *image) ListNamespaces(ctx context.Context, opts ...Options) ([]model.Namespace, error) {
	var audits []model.Namespace
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}
	return audits, nil
}
