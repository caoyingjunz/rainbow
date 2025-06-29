package db

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type DockerfileInterface interface {
	Create(ctx context.Context, object *model.Dockerfile) (*model.Dockerfile, error)
	Delete(ctx context.Context, dockerfileId int64) error
	Update(ctx context.Context, DockerfileId int64, resourceVersion int64, updates map[string]interface{}) error
	List(ctx context.Context, opts ...Options) ([]model.Dockerfile, error)
	Get(ctx context.Context, dockerfileId int64) (*model.Dockerfile, error)
}

func newDockerfile(db *gorm.DB) DockerfileInterface {
	return &dockerfile{db}
}

type dockerfile struct {
	db *gorm.DB
}

func (d *dockerfile) Create(ctx context.Context, object *model.Dockerfile) (*model.Dockerfile, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := d.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (d *dockerfile) Get(ctx context.Context, dockerfileId int64) (*model.Dockerfile, error) {
	var audit model.Dockerfile
	if err := d.db.WithContext(ctx).Where("id = ?", dockerfileId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (d *dockerfile) Delete(ctx context.Context, dockerfileId int64) error {
	var audit model.Dockerfile
	if err := d.db.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", dockerfileId).Delete(&audit).Error; err != nil {
		return err
	}

	return nil
}

func (d *dockerfile) Update(ctx context.Context, DockerfileId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := d.db.WithContext(ctx).Model(&model.Dockerfile{}).Where("id = ? and resource_version = ?", DockerfileId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (d *dockerfile) List(ctx context.Context, opts ...Options) ([]model.Dockerfile, error) {
	var audits []model.Dockerfile
	tx := d.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
