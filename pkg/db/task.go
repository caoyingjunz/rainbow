package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
)

type TaskInterface interface {
	Create(ctx context.Context, object *model.Task) (*model.Task, error)
	Update(ctx context.Context, taskId int64, resourceVersion int64, updates map[string]interface{}) error
	Delete(ctx context.Context, taskId int64) error
	Get(ctx context.Context, taskId int64) (*model.Task, error)
	List(ctx context.Context, opts ...Options) ([]model.Task, error)

	DeleteInBatch(ctx context.Context, taskIds []int64) error
	UpdateDirectly(ctx context.Context, taskId int64, updates map[string]interface{}) error

	DeleteBySubscribe(ctx context.Context, subId int64) error

	GetOne(ctx context.Context, taskId int64, resourceVersion int64) (*model.Task, error)
	AssignToAgent(ctx context.Context, taskId int64, agentName string) error
	ListWithAgent(ctx context.Context, agentName string, process int, opts ...Options) ([]model.Task, error)
	ListWithNoAgent(ctx context.Context, process int, opts ...Options) ([]model.Task, error)
	ListWithUser(ctx context.Context, userId string, opts ...Options) ([]model.Task, error)
	GetOneForSchedule(ctx context.Context, opts ...Options) (*model.Task, error)
	GetRunningTask(ctx context.Context, opts ...Options) ([]model.Task, error)

	Count(ctx context.Context, opts ...Options) (int64, error)
	CountSubscribe(ctx context.Context, opts ...Options) (int64, error)

	ListReview(ctx context.Context) ([]model.Review, error)
	AddDailyReview(ctx context.Context, object *model.Daily) error
	CountDailyReview(ctx context.Context) (int64, error)

	CreateTaskMessage(ctx context.Context, object *model.TaskMessage) error
	DeleteTaskMessages(ctx context.Context, taskId int64) error
	ListTaskMessages(ctx context.Context, opts ...Options) ([]model.TaskMessage, error)

	CreateUser(ctx context.Context, object *model.User) error
	ListUsers(ctx context.Context, opts ...Options) ([]model.User, error)
	GetUser(ctx context.Context, userId string) (*model.User, error)
	DeleteUser(ctx context.Context, userId string) error
	UpdateUser(ctx context.Context, userId string, resourceVersion int64, updates map[string]interface{}) error

	ListKubernetesVersions(ctx context.Context, opts ...Options) ([]model.KubernetesVersion, error)
	GetKubernetesVersionCount(ctx context.Context, opts ...Options) (int64, error)
	GetKubernetesVersion(ctx context.Context, name string) (*model.KubernetesVersion, error)
	CreateKubernetesVersion(ctx context.Context, object *model.KubernetesVersion) error

	CreateSubscribe(ctx context.Context, object *model.Subscribe) error
	UpdateSubscribe(ctx context.Context, subId int64, resourceVersion int64, updates map[string]interface{}) error
	DeleteSubscribe(ctx context.Context, subId int64) error
	GetSubscribe(ctx context.Context, subId int64) (*model.Subscribe, error)
	ListSubscribes(ctx context.Context, opts ...Options) ([]model.Subscribe, error)

	DeleteSubscribeAllMessage(ctx context.Context, subId int64) error
	UpdateSubscribeDirectly(ctx context.Context, subId int64, updates map[string]interface{}) error

	CreateSubscribeMessage(ctx context.Context, object *model.SubscribeMessage) error
	DeleteSubscribeMessage(ctx context.Context, subId int64) error
	ListSubscribeMessages(ctx context.Context, opts ...Options) ([]model.SubscribeMessage, error)
}

func newTask(db *gorm.DB) TaskInterface {
	return &task{db}
}

type task struct {
	db *gorm.DB
}

func (a *task) Create(ctx context.Context, object *model.Task) (*model.Task, error) {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return nil, err
	}
	return object, nil
}

func (a *task) Update(ctx context.Context, taskId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := a.db.WithContext(ctx).Model(&model.Task{}).Where("id = ? and resource_version = ?", taskId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *task) UpdateDirectly(ctx context.Context, taskId int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	f := a.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", taskId).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *task) Delete(ctx context.Context, taskId int64) error {
	return a.db.WithContext(ctx).Where("id = ?", taskId).Delete(&model.Task{}).Error
}

func (a *task) DeleteBySubscribe(ctx context.Context, subId int64) error {
	return a.db.WithContext(ctx).Where("subscribe_id = ?", subId).Delete(&model.Task{}).Error
}

func (a *task) DeleteInBatch(ctx context.Context, taskIds []int64) error {
	return a.db.WithContext(ctx).Where("id in ?", taskIds).Delete(&model.Task{}).Error
}

func (a *task) Get(ctx context.Context, agentId int64) (*model.Task, error) {
	var audit model.Task
	if err := a.db.WithContext(ctx).Where("id = ?", agentId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *task) GetOne(ctx context.Context, taskId int64, resourceVersion int64) (*model.Task, error) {
	updates := make(map[string]interface{})
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1
	updates["process"] = 1

	f := a.db.WithContext(ctx).Model(&model.Task{}).Where("id = ? and resource_version = ?", taskId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return nil, f.Error
	}
	if f.RowsAffected == 0 {
		return nil, errors.ErrRecordNotUpdate
	}

	return a.Get(ctx, taskId)
}

func (a *task) List(ctx context.Context, opts ...Options) ([]model.Task, error) {
	var audits []model.Task
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) ListWithAgent(ctx context.Context, agentName string, process int, opts ...Options) ([]model.Task, error) {
	var audits []model.Task
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("agent_name = ? and process = ?", agentName, process).Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) ListWithNoAgent(ctx context.Context, process int, opts ...Options) ([]model.Task, error) {
	var audits []model.Task
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("agent_name = ? and process = ?", "", process).Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) GetOneForSchedule(ctx context.Context, opts ...Options) (*model.Task, error) {
	var audits []model.Task
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("agent_name = ? and process = ? and mode = ?", "", 0, 0).Find(&audits).Error; err != nil {
		return nil, err
	}
	if len(audits) == 0 {
		return nil, nil
	}

	one := audits[0]
	return &one, nil
}

func (a *task) GetRunningTask(ctx context.Context, opts ...Options) ([]model.Task, error) {
	var audits []model.Task
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	if err := tx.Where("process = ?", 1).Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) ListWithUser(ctx context.Context, userId string, opts ...Options) ([]model.Task, error) {
	var audits []model.Task
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Where("user_id = ?", userId).Order("gmt_create DESC").Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) AssignToAgent(ctx context.Context, taskId int64, agentName string) error {
	f := a.db.WithContext(ctx).Model(&model.Task{}).Where("id = ?", taskId).Updates(map[string]interface{}{
		"gmt_modified": time.Now(),
		"agent_name":   agentName,
	})

	return f.Error
}

func (a *task) Count(ctx context.Context, opts ...Options) (int64, error) {
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Task{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (a *task) CountSubscribe(ctx context.Context, opts ...Options) (int64, error) {
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.Subscribe{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (a *task) ListReview(ctx context.Context) ([]model.Review, error) {
	var audits []model.Review
	tx := a.db.WithContext(ctx)

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) AddDailyReview(ctx context.Context, object *model.Daily) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	err := a.db.WithContext(ctx).Create(object).Error
	return err
}

func (a *task) CountDailyReview(ctx context.Context) (int64, error) {
	var total int64
	if err := a.db.WithContext(ctx).Model(&model.Daily{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (a *task) CreateTaskMessage(ctx context.Context, object *model.TaskMessage) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	err := a.db.WithContext(ctx).Create(object).Error
	return err
}

func (a *task) DeleteTaskMessages(ctx context.Context, taskId int64) error {
	return a.db.WithContext(ctx).Where("task_id = ?", taskId).Delete(&model.TaskMessage{}).Error
}

func (a *task) GetSubscribe(ctx context.Context, subId int64) (*model.Subscribe, error) {
	var audit model.Subscribe
	if err := a.db.WithContext(ctx).Where("id = ?", subId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *task) ListTaskMessages(ctx context.Context, opts ...Options) ([]model.TaskMessage, error) {
	var audits []model.TaskMessage
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}
	return audits, nil
}

func (a *task) CreateUser(ctx context.Context, object *model.User) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	err := a.db.WithContext(ctx).Create(object).Error
	return err
}

func (a *task) UpdateUser(ctx context.Context, userId string, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := a.db.WithContext(ctx).Model(&model.User{}).Where("user_id = ? and resource_version = ?", userId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *task) GetUser(ctx context.Context, userId string) (*model.User, error) {
	var audit model.User
	if err := a.db.WithContext(ctx).Where("user_id = ?", userId).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *task) ListUsers(ctx context.Context, opts ...Options) ([]model.User, error) {
	var audits []model.User
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) DeleteUser(ctx context.Context, userId string) error {
	return a.db.WithContext(ctx).Where("user_id = ?", userId).Delete(&model.User{}).Error
}

func (a *task) ListKubernetesVersions(ctx context.Context, opts ...Options) ([]model.KubernetesVersion, error) {
	var audits []model.KubernetesVersion
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) GetKubernetesVersionCount(ctx context.Context, opts ...Options) (int64, error) {
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	var total int64
	if err := tx.Model(&model.KubernetesVersion{}).Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

func (a *task) GetKubernetesVersion(ctx context.Context, name string) (*model.KubernetesVersion, error) {
	var audit model.KubernetesVersion
	if err := a.db.WithContext(ctx).Where("name = ?", name).First(&audit).Error; err != nil {
		return nil, err
	}
	return &audit, nil
}

func (a *task) CreateKubernetesVersion(ctx context.Context, object *model.KubernetesVersion) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return err
	}
	return nil
}

func (a *task) CreateSubscribe(ctx context.Context, object *model.Subscribe) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return err
	}
	return nil
}
func (a *task) ListSubscribes(ctx context.Context, opts ...Options) ([]model.Subscribe, error) {
	var audits []model.Subscribe
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}

func (a *task) UpdateSubscribe(ctx context.Context, subId int64, resourceVersion int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	updates["resource_version"] = resourceVersion + 1

	f := a.db.WithContext(ctx).Model(&model.Subscribe{}).Where("id = ? and resource_version = ?", subId, resourceVersion).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *task) UpdateSubscribeDirectly(ctx context.Context, subId int64, updates map[string]interface{}) error {
	updates["gmt_modified"] = time.Now()
	f := a.db.WithContext(ctx).Model(&model.Subscribe{}).Where("id = ?", subId).Updates(updates)
	if f.Error != nil {
		return f.Error
	}
	if f.RowsAffected == 0 {
		return fmt.Errorf("record not updated")
	}

	return nil
}

func (a *task) DeleteSubscribe(ctx context.Context, subId int64) error {
	return a.db.WithContext(ctx).Where("id = ?", subId).Delete(&model.Subscribe{}).Error
}

func (a *task) CreateSubscribeMessage(ctx context.Context, object *model.SubscribeMessage) error {
	now := time.Now()
	object.GmtCreate = now
	object.GmtModified = now

	if err := a.db.WithContext(ctx).Create(object).Error; err != nil {
		return err
	}
	return nil
}

func (a *task) DeleteSubscribeAllMessage(ctx context.Context, subId int64) error {
	return a.db.WithContext(ctx).Where("subscribe_id = ?", subId).Delete(&model.SubscribeMessage{}).Error
}

func (a *task) DeleteSubscribeMessage(ctx context.Context, subId int64) error {
	result := a.db.Exec(`
	DELETE FROM subscribe_messages
	WHERE subscribe_id = ?
	AND id NOT IN (
		SELECT id FROM (
		SELECT id
	FROM subscribe_messages
	WHERE subscribe_id = ?
	ORDER BY id DESC
	LIMIT 5
	) AS temp
	)`, subId, subId)
	return result.Error
}

func (a *task) ListSubscribeMessages(ctx context.Context, opts ...Options) ([]model.SubscribeMessage, error) {
	var audits []model.SubscribeMessage
	tx := a.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}

	if err := tx.Find(&audits).Error; err != nil {
		return nil, err
	}

	return audits, nil
}
