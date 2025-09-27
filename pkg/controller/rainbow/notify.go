package rainbow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

const (
	DingTalkType = "dingtalk"
	WeComType    = "wecom"
)

type PushMessage struct {
	Text    map[string]string `json:"text"`
	Msgtype string            `json:"msgtype"`
}

func (s *ServerController) CreateNotify(ctx context.Context, req *types.CreateNotificationRequest) error {
	var pushCfgJSON []byte
	var err error

	// 根据类型序列化对应的配置结构体
	switch req.PushCfg.Type {
	case DingTalkType:
		if req.PushCfg.DingTalkPushCfg == nil {
			return fmt.Errorf("钉钉推送配置不能为空")
		}
		pushCfgJSON, err = json.Marshal(req.PushCfg.DingTalkPushCfg)
		if err != nil {
			klog.Errorf("序列化钉钉推送配置失败: %v", err)
			return err
		}

	case WeComType:
		if req.PushCfg.WeComPushCfg == nil {
			return fmt.Errorf("企微推送配置不能为空")
		}
		pushCfgJSON, err = json.Marshal(req.PushCfg.WeComPushCfg)
		if err != nil {
			klog.Errorf("序列化企业微信推送配置失败: %v", err)
			return err
		}

	default:
		return fmt.Errorf("不支持的推送类型: %s", req.PushCfg.Type)
	}

	_, err = s.factory.Notify().Create(ctx, &model.Notification{
		Name:   req.Name,
		Role:   req.Role,
		Enable: req.Enable,
		Type:   req.PushCfg.Type,
		UserModel: rainbow.UserModel{
			UserId:   req.UserId,
			UserName: req.UserName,
		},
		PushCfg:   string(pushCfgJSON),
		ShortDesc: req.ShortDesc,
	})
	if err != nil {
		klog.Error("创建推送(%s)记录失败: %v", req.Name, err)
	}

	return err
}

func (s *ServerController) SendNotify(ctx context.Context, req *types.SendNotificationRequest) error {
	switch req.Role {
	case 1:
		return s.SendRegisterNotify(ctx, req)
	case 0:
		return s.sendMessageNotify(ctx, req)
	default:
		return fmt.Errorf("invalid role: %d", req.Role)
	}
}

func (s *ServerController) SendRegisterNotify(ctx context.Context, req *types.SendNotificationRequest) error {
	notifies, err := s.factory.Notify().List(ctx, db.WithRole(req.Role), db.WithEnable(1))
	if err != nil {
		return fmt.Errorf("failed to query notification configs: %w", err)
	}

	msg := fmt.Sprintf("注册通知\n用户名: %s\n时间: %v\nEmail: %s",
		req.UserName, time.Now().Format("2006-01-02 15:04:05"), req.Email)

	for _, notify := range notifies {
		if err := s.sendNotification(&notify, msg); err != nil {
			klog.Errorf("notify(%s) 推送失败: %v", notify.Name, err)
			continue
		}
		klog.Infof("notify(%s) 推送成功", notify.Name)
	}

	return nil
}

func (s *ServerController) sendNotification(notify *model.Notification, msg string) error {
	type Config struct {
		Url string `json:"url"`
	}

	var cfg Config
	if err := json.Unmarshal([]byte(notify.PushCfg), &cfg); err != nil {
		return fmt.Errorf("failed to parse config (ID: %d): %w", notify.Id, err)
	}

	httpClient := util.NewHttpClient(5*time.Second, cfg.Url)
	payload := PushMessage{
		Text:    map[string]string{"content": msg},
		Msgtype: "text",
	}

	if err := httpClient.Post(
		cfg.Url,
		nil,
		payload,
		map[string]string{"Content-Type": "application/json"},
	); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *ServerController) sendMessageNotify(ctx context.Context, req *types.SendNotificationRequest) error {
	task, err := s.factory.Task().Get(ctx, req.TaskId)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	list, err := s.factory.Notify().List(ctx, db.WithUser(task.UserId), db.WithEnable(1))

	for _, n := range list {
		if err := s.sendNotification(&n, req.Content); err != nil {
			klog.Errorf("failed to send notification via config (ID: %d): %v", n.Id, err)
		}
	}
	return nil
}
