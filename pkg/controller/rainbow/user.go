package rainbow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func parseTime(t string) (time.Time, error) {
	pt, err := time.Parse("2006-01-02 15:04:05", t)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析超时时间(%s)失败: %v", t, err)
	}

	return pt, nil
}

func (s *ServerController) preCreateUser(ctx context.Context, req *types.CreateUserRequest) error {
	_, err := s.factory.Task().GetUser(ctx, req.UserId)
	if err == nil {
		return fmt.Errorf("用户(%s)已经存在", req.UserId)
	}

	return nil
}

func (s *ServerController) CreateUser(ctx context.Context, req *types.CreateUserRequest) error {
	if err := s.preCreateUser(ctx, req); err != nil {
		return err
	}

	t, err := parseTime(req.ExpireTime)
	if err != nil {
		klog.Errorf("%v", err)
		return err
	}

	if err = s.factory.Task().CreateUser(ctx, &model.User{
		Name:       req.Name,
		UserId:     req.UserId,
		UserType:   req.UserType,
		ExpireTime: t,
	}); err != nil {
		klog.Errorf("创建用户 %s 失败 %v", req.Name, err)
		return err
	}

	return nil
}

func (s *ServerController) ListUsers(ctx context.Context, listOption types.ListOptions) ([]model.User, error) {
	return s.factory.Task().ListUsers(ctx)
}

func (s *ServerController) GetUser(ctx context.Context, userId string) (*model.User, error) {
	return s.factory.Task().GetUser(ctx, userId)
}

func (s *ServerController) UpdateUser(ctx context.Context, req *types.UpdateUserRequest) error {
	t, err := parseTime(req.ExpireTime)
	if err != nil {
		klog.Errorf("%v", err)
		return err
	}

	return s.factory.Task().UpdateUser(ctx, req.UserId, req.ResourceVersion, map[string]interface{}{"name": req.Name, "user_type": req.UserType, "expire_time": t})
}

func (s *ServerController) DeleteUser(ctx context.Context, userId string) error {
	return s.factory.Task().DeleteUser(ctx, userId)
}
