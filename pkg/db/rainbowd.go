package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
)

type RainbowdInterface interface {
	Create(ctx context.Context, object *model.Rainbowd) (*model.Rainbowd, error)
	Update(ctx context.Context, rainbowdId int64, resourceVersion int64, updates map[string]interface{}) error
	Delete(ctx context.Context, rainbowdId int64) error
	Get(ctx context.Context, rainbowdId int64) (*model.Rainbowd, error)
	List(ctx context.Context, opts ...Options) ([]model.Rainbowd, error)

	GetByName(ctx context.Context, name string) (*model.Rainbowd, error)

	CreateEvent(ctx context.Context, object *model.RainbowdEvent) (*model.RainbowdEvent, error)
	DeleteEvent(ctx context.Context, eid int64) error
	ListEvents(ctx context.Context, opts ...Options) ([]model.RainbowdEvent, error)
}

func newRainbowd(db *gorm.DB) RainbowdInterface {
	return &rainbowd{db}
}

type rainbowd struct {
	db *gorm.DB
}

func (rain *rainbowd) Create(ctx context.Context, object *model.Rainbowd) (*model.Rainbowd, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := rain.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (rain *rainbowd) Update(ctx context.Context, rainbowdId int64, resourceVersion int64, updates map[string]interface{}) error {
	return nil
}

func (rain *rainbowd) Delete(ctx context.Context, rainbowdId int64) error {
	return nil
}

func (rain *rainbowd) Get(ctx context.Context, rainbowdId int64) (*model.Rainbowd, error) {
	var audit model.Rainbowd
	if err := rain.db.WithContext(ctx).Where("id = ?", rainbowdId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (rain *rainbowd) GetByName(ctx context.Context, name string) (*model.Rainbowd, error) {
	var audit model.Rainbowd
	if err := rain.db.WithContext(ctx).Where("name = ?", name).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (rain *rainbowd) List(ctx context.Context, opts ...Options) ([]model.Rainbowd, error) {
	var audits []model.Rainbowd
	tx := rain.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (rain *rainbowd) CreateEvent(ctx context.Context, object *model.RainbowdEvent) (*model.RainbowdEvent, error) {
	return nil, nil
}

func (rain *rainbowd) DeleteEvent(ctx context.Context, eid int64) error {
	return nil
}

func (rain *rainbowd) ListEvents(ctx context.Context, opts ...Options) ([]model.RainbowdEvent, error) {
	return nil, nil
}
