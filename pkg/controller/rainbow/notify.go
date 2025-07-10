package rainbow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

type PushMessage struct {
	Text    map[string]string `json:"text"`
	Msgtype string            `json:"msgtype"`
}

func (s *ServerController) CreateNotify(ctx context.Context, req *types.CreateNotificationRequest) error {
	_, err := s.factory.Notify().Create(ctx, &model.Notification{
		Name:   req.Name,
		Role:   req.Role,
		Enable: req.Enable,
		Type:   req.Type,
		UserModel: rainbow.UserModel{
			UserId:   req.UserId,
			UserName: req.UserName,
		},
		Url:       req.Url,
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
	default:
		// TODO
	}
	return nil
}

func (s *ServerController) SendRegisterNotify(ctx context.Context, req *types.SendNotificationRequest) error {
	notifies, err := s.factory.Notify().List(ctx, db.WithRole(req.Role), db.WithEnable(1))
	if err != nil {
		klog.Errorf("获取 notify 失败 %v", err)
		return err
	}
	for _, notify := range notifies {
		http := util.NewHttpClient(5*time.Second, notify.Url)
		msg := fmt.Sprintf("注册通知\n用户名: %s\n时间: %v\nEmail: %s", req.UserName, time.Now().Format("2006-01-02 15:04:05"), req.Email)
		if err = http.Post(notify.Url, nil,
			PushMessage{Text: map[string]string{"content": msg}, Msgtype: "text"}, map[string]string{"Content-Type": "application/json"}); err != nil {
			klog.Errorf("notify(%s) 推送失败 %v", notify.Name, err)
			continue
		}
		klog.Errorf("notify(%s) 推送成功", notify.Name)
	}

	return nil
}
