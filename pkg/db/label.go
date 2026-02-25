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
	Count(ctx context.Context, opts ...Options) (int64, error)
	Get(ctx context.Context, opts ...Options) (*model.Label, error)

	CreateLogo(ctx context.Context, object *model.Logo) (*model.Logo, error)
	UpdateLogo(ctx context.Context, labelId int64, resourceVersion int64, updates map[string]interface{}) error
	DeleteLogo(ctx context.Context, logoId int64) error
	ListLogos(ctx context.Context, opts ...Options) ([]model.Logo, error)
	CountLogos(ctx context.Context, opts ...Options) (int64, error)
	GetLogo(ctx context.Context, opts ...Options) (*model.Logo, error)

	CreateImageLabel(ctx context.Context, object *model.ImageLabel) (*model.ImageLabel, error)
	GetImageLabel(ctx context.Context, opts ...Options) (*model.ImageLabel, error)
	DeleteImageLabel(ctx context.Context, opts ...Options) error
	ListImageLabels(ctx context.Context, opts ...Options) ([]model.ImageLabel, error)

	ListImageLabelsV2(ctx context.Context, imageId int64) ([]model.Label, error)
	ListImageLabelNames(ctx context.Context, imageId int64) ([]string, error)

	ListLabelImages(ctx context.Context, labelIds []int64, page int, limit int) ([]model.Image, int64, error)
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

func (l *label) Count(ctx context.Context, opts ...Options) (int64, error) {
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Label{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (l *label) Get(ctx context.Context, opts ...Options) (*model.Label, error) {
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var audit model.Label
	if err := tx.First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
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

func (l *label) CountLogos(ctx context.Context, opts ...Options) (int64, error) {
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Logo{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (l *label) GetLogo(ctx context.Context, opts ...Options) (*model.Logo, error) {
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var audit model.Logo
	if err := tx.First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

// ImageLabel 增删改查

func (l *label) CreateImageLabel(ctx context.Context, object *model.ImageLabel) (*model.ImageLabel, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := l.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (l *label) GetImageLabel(ctx context.Context, opts ...Options) (*model.ImageLabel, error) {
	tx := l.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var audit model.ImageLabel
	if err := tx.First(&audit).Error; err != nil {
		return nil, err
	}

	return &audit, nil
}

func (l *label) DeleteImageLabel(ctx context.Context, opts ...Options) error {
	tx := l.db.WithContext(ctx).Model(&model.ImageLabel{})
	for _, opt := range opts {
		tx = opt(tx)
	}

	return tx.Delete(&model.ImageLabel{}).Error
}

func (l *label) ListImageLabels(ctx context.Context, opts ...Options) ([]model.ImageLabel, error) {
	var list []model.ImageLabel
	tx := l.db.WithContext(ctx).Model(&model.ImageLabel{})
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (l *label) ListImageLabelsV2(ctx context.Context, imageId int64) ([]model.Label, error) {
	var labels []model.Label
	err := l.db.WithContext(ctx).
		Table("labels").
		Joins("JOIN image_labels ON image_labels.label_id = labels.id").
		Where("image_labels.image_id = ?", imageId).
		Find(&labels).Error
	if err != nil {
		return nil, err
	}

	return labels, nil
}

func (l *label) ListImageLabelNames(ctx context.Context, imageId int64) ([]string, error) {
	var labelNames []string
	err := l.db.WithContext(ctx).
		Table("labels").
		Joins("JOIN image_labels ON image_labels.label_id = labels.id").
		Where("image_labels.image_id = ?", imageId).
		Pluck("labels.name", &labelNames).Error
	if err != nil {
		return nil, err
	}

	return labelNames, nil
}

func (l *label) ListLabelImages(ctx context.Context, labelIds []int64, page int, limit int) ([]model.Image, int64, error) {
	subQuery := l.db.WithContext(ctx).
		Table("image_labels").
		Select("image_id").
		Where("label_id IN ?", labelIds).
		Group("image_id").
		Having("COUNT(DISTINCT label_id) = ?", len(labelIds))

	var total int64
	err := l.db.Where("id IN (?)", subQuery).Model(&model.Image{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var images []model.Image
	offset := (page - 1) * limit
	if err = l.db.Where("id IN (?)", subQuery).
		Offset(offset).
		Limit(limit).
		Order("gmt_modified DESC").
		Find(&images).Error; err != nil {
		return nil, 0, err
	}

	return images, total, nil
}
