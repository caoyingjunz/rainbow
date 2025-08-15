package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type AgentInterface interface {
	Create(ctx context.Context, object *model.Agent) (*model.Agent, error)
	Update(ctx context.Context, agentId int64, resourceVersion int64, updates map[string]interface{}) error
	Delete(ctx context.Context, agentId int64) error
	Get(ctx context.Context, agentId int64) (*model.Agent, error)
	GetByName(ctx context.Context, agentName string) (*model.Agent, error)
	UpdateByName(ctx context.Context, agentName string, updates map[string]interface{}) error
	List(ctx context.Context, opts ...Options) ([]model.Agent, error)
	ListForSchedule(ctx context.Context, opts ...Options) ([]model.Agent, error)
}

func newAgent(db *gorm.DB) AgentInterface {
	return &agent{db}
}

type agent struct {
	db *gorm.DB
}

func (a *agent) Create(ctx context.Context, object *model.Agent) (*model.Agent, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *agent) Update(ctx context.Context, agentId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := a.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ? and resource_version = ?", agentId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *agent) UpdateByName(ctx context.Context, agentName string, updates map[string]interface{}) error {
	f := a.db.WithContext(ctx).Model(&model.Agent{}).Where("name = ?", agentName).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *agent) Delete(ctx context.Context, agentId int64) error {
	if err := a.db.WithContext(ctx).Where("id = ?", agentId).Delete(&model.Agent{}).Error; err != nil {
		return err
	}
	return nil
}

func (a *agent) Get(ctx context.Context, agentId int64) (*model.Agent, error) {
	var audit model.Agent
	if err := a.db.WithContext(ctx).Where("id = ?", agentId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *agent) GetByName(ctx context.Context, agentName string) (*model.Agent, error) {
	var audit model.Agent
	if err := a.db.WithContext(ctx).Where("name = ?", agentName).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *agent) List(ctx context.Context, opts ...Options) ([]model.Agent, error) {
	var audits []model.Agent
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *agent) ListForSchedule(ctx context.Context, opts ...Options) ([]model.Agent, error) {
	var audits []model.Agent
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("status = ?", "在线").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
