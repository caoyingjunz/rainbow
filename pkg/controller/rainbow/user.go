package rainbow

import (
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util/errors"
)

func parseTime(t string) (time.Time, error) {
	pt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析时间(%s)失败: %v", t, err)
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

func (s *ServerController) preCreateOrUpdateUser(ctx context.Context, user *types.CreateUserRequest) error {
	if user.PaymentType == PackagePaymentType {
		if user.ExpireTime == nil {
			return fmt.Errorf("包年包月类型时，结束时间为必填项")
		}
	}
	return nil
}

func (s *ServerController) CreateOrUpdateUser(ctx context.Context, user *types.CreateUserRequest) error {
	if err := s.preCreateOrUpdateUser(ctx, user); err != nil {
		return err
	}

	old, err := s.factory.Task().GetUser(ctx, user.UserId)
	if err == nil {
		// 同步用户时，忽略已存在用户
		if user.Sync {
			return nil
		}

		// 用户已存在，则更新
		updates := make(map[string]interface{})

		if old.Name != user.Name {
			updates["name"] = user.Name
		}
		if old.UserType != user.UserType {
			updates["user_type"] = user.UserType
		}
		if old.PaymentType != user.PaymentType {
			updates["payment_type"] = user.PaymentType
		}

		if user.PaymentType == PackagePaymentType {
			et, err := parseTime(*user.ExpireTime)
			if err != nil {
				klog.Errorf("%v", err)
				return err
			}
			if old.ExpireTime != et {
				updates["expire_time"] = et
			}
		} else {
			if old.RemainCount != user.RemainCount {
				updates["remain_count"] = user.RemainCount
			}
		}
		// 无变更时，直接退出
		if len(updates) == 0 {
			return nil
		}
		return s.factory.Task().UpdateUser(ctx, user.UserId, old.ResourceVersion, updates)
	} else {
		if !errors.IsNotFound(err) {
			return err
		}
		// 用户不存在，则创建
		if err = s.CreateUser(ctx, user); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServerController) CreateOrUpdateUsers(ctx context.Context, req *types.CreateUsersRequest) error {
	for _, user := range req.Users {
		if err := s.CreateOrUpdateUser(ctx, &user); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServerController) CreateUser(ctx context.Context, req *types.CreateUserRequest) error {
	obj := &model.User{
		Name:        req.Name,
		UserId:      req.UserId,
		UserType:    req.UserType,
		PaymentType: req.PaymentType,
	}

	if req.Sync {
		obj.PaymentType = DemandPaymentType
		obj.RemainCount = defaultRemainCount
	} else {
		if req.PaymentType == PackagePaymentType {
			et, err := parseTime(*req.ExpireTime)
			if err != nil {
				klog.Errorf("%v", err)
				return err
			}
			obj.ExpireTime = et
		} else {
			obj.RemainCount = req.RemainCount
		}
	}

	if err := s.factory.Task().CreateUser(ctx, obj); err != nil {
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
