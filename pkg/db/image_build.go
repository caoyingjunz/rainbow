package db

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type BuildInterface interface {
	Create(ctx context.Context, object *model.Build) (*model.Build, error)
	Delete(ctx context.Context, dockerfileId int64) error
	Update(ctx context.Context, DockerfileId int64, resourceVersion int64, updates map[string]interface{}) error
	UpdateDirectly(ctx context.Context, buildId int64, updates map[string]interface{}) error
	GetOne(ctx context.Context, buildId int64, resourceVersion int64) (*model.Build, error)
	List(ctx context.Context, opts ...Options) ([]model.Build, error)
	ListWithAgent(ctx context.Context, agentName string, opts ...Options) ([]model.Build, error)
	Get(ctx context.Context, dockerfileId int64) (*model.Build, error)
	UpdateBy(ctx context.Context, updates map[string]interface{}, opts ...Options) error
	Count(ctx context.Context, opts ...Options) (int64, error)

	CreateBuildMessage(ctx context.Context, object *model.BuildMessage) error
	ListBuildMessages(ctx context.Context, opts ...Options) ([]model.BuildMessage, error)
}

func newBuild(db *gorm.DB) BuildInterface {
	return &build{db}
}

type build struct {
	db *gorm.DB
}

func (b *build) Create(ctx context.Context, object *model.Build) (*model.Build, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := b.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (b *build) Get(ctx context.Context, dockerfileId int64) (*model.Build, error) {
	var audit model.Build
	if err := b.db.WithContext(ctx).Where("id = ?", dockerfileId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (b *build) Delete(ctx context.Context, dockerfileId int64) error {
	var audit model.Build
	if err := b.db.WithContext(ctx).Clauses(clause.Returning{}).Where("id = ?", dockerfileId).Delete(&audit).Error; err != nil {
		return err
	}

	return nil
}

func (b *build) Update(ctx context.Context, DockerfileId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := b.db.WithContext(ctx).Model(&model.Build{}).Where("id = ? and resource_version = ?", DockerfileId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (b *build) UpdateDirectly(ctx context.Context, buildId int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	f := b.db.WithContext(ctx).Model(&model.Build{}).Where("id = ?", buildId).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (b *build) List(ctx context.Context, opts ...Options) ([]model.Build, error) {
	var audits []model.Build
	tx := b.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
func (b *build) ListWithAgent(ctx context.Context, agentName string, opts ...Options) ([]model.Build, error) {
	var audits []model.Build
	tx := b.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("agent_name = ? and status = ?", agentName, "调度中").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (b *build) UpdateBy(ctx context.Context, updates map[string]interface{}, opts ...Options) error {
	updates["gmt_modified"] = time.Now()

	tx := b.db.WithContext(ctx)
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

func (b *build) Count(ctx context.Context, opts ...Options) (int64, error) {
	tx := b.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Build{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (b *build) CreateBuildMessage(ctx context.Context, object *model.BuildMessage) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	err := b.db.WithContext(ctx).Create(object).Error
	return err
}

func (b *build) ListBuildMessages(ctx context.Context, opts ...Options) ([]model.BuildMessage, error) {
	var audits []model.BuildMessage
	tx := b.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}
	return audits, nil
}

func (b *build) GetOne(ctx context.Context, buildId int64, resourceVersion int64) (*model.Build, error) {
	updates := make(map[string]interface{})
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := b.db.WithContext(ctx).Model(&model.Build{}).Where("id = ? and resource_version = ?", buildId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return nil, f.Error
	}
	if f.RowsAffected == 0 {
		return nil, errors.ErrRecordNotUpdate
	}

	return b.Get(ctx, buildId)
}
