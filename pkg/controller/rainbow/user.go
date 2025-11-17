package rainbow

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
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

func (s *ServerController) isUserExist(ctx context.Context, userId string) (bool, error) {
	_, err := s.factory.Task().GetUser(ctx, userId)
	if err == nil {
		return true, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

func (s *ServerController) CreateOrUpdateUser(ctx context.Context, user *model.User) error {
	return nil
}

func (s *ServerController) CreateOrUpdateUsers(ctx context.Context, req *types.CreateUsersRequest) error {
	for _, user := range req.Users {
		if err := s.CreateOrUpdateUser(ctx, &model.User{
			Name:       user.Name,
			UserId:     user.UserId,
			UserType:   user.UserType,
			ExpireTime: user.ExpireTime,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServerController) CreateUser(ctx context.Context, req *types.CreateUserRequest) error {
	if err := s.factory.Task().CreateUser(ctx, &model.User{
		Name:       req.Name,
		UserId:     req.UserId,
		UserType:   req.UserType,
		ExpireTime: req.ExpireTime,
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
	return s.factory.Task().UpdateUser(ctx, req.UserId, req.ResourceVersion, map[string]interface{}{"name": req.Name, "user_type": req.UserType, "expire_time": req.ExpireTime})
}

func (s *ServerController) DeleteUser(ctx context.Context, userId string) error {
	return s.factory.Task().DeleteUser(ctx, userId)
}
