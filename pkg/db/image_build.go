package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type BuildInterface interface {
	Create(ctx context.Context, object *model.Build) (*model.Build, error)
	Delete(ctx context.Context, dockerfileId int64) error
	Update(ctx context.Context, DockerfileId int64, resourceVersion int64, updates map[string]interface{}) error
	List(ctx context.Context, opts ...Options) ([]model.Build, error)
	Get(ctx context.Context, dockerfileId int64) (*model.Build, error)
	UpdateBy(ctx context.Context, updates map[string]interface{}, opts ...Options) error
	Count(ctx context.Context, opts ...Options) (int64, error)
}

func newBuild(db *gorm.DB) BuildInterface {
	return &build{db}
}

type build struct {
	db *gorm.DB
}

func (d *build) Create(ctx context.Context, object *model.Build) (*model.Build, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := d.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (d *build) Get(ctx context.Context, dockerfileId int64) (*model.Build, error) {
	var audit model.Build
	if err := d.db.WithContext(ctx).Where("id = ?", dockerfileId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (d *build) Delete(ctx context.Context, dockerfileId int64) error {
	var audit model.Build
	if err := d.db.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", dockerfileId).Delete(&audit).Error; err != nil {
		return err
	}

	return nil
}

func (d *build) Update(ctx context.Context, DockerfileId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := d.db.WithContext(ctx).Model(&model.Build{}).Where("id = ? and resource_version = ?", DockerfileId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (d *build) List(ctx context.Context, opts ...Options) ([]model.Build, error) {
	var audits []model.Build
	tx := d.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (d *build) UpdateBy(ctx context.Context, updates map[string]interface{}, opts ...Options) error {
	updates["gmt_modified"] = time.Now()

	tx := d.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	f := tx.Model(&model.Build{}).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (d *build) Count(ctx context.Context, opts ...Options) (int64, error) {
	tx := d.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Build{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
