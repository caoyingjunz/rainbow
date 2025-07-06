package rainbow

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

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
		Content:   req.Content,
		ShortDesc: req.ShortDesc,
	})
	if err != nil {
		klog.Error("创建推送(%s)记录失败: %v", req.Name, err)
	}

	return err
}
func (s *ServerController) SendNotify(ctx context.Context, req *types.SendNotificationRequest) error {
	return nil
}
