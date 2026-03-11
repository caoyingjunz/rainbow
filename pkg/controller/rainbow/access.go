package rainbow

import (
	"context"
	"github.com/caoyingjunz/rainbow/pkg/db/model/rainbow"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) CreateAccess(ctx context.Context, req *types.CreateAccessRequest) error {
	ak, err := util.GenerateAK("pixiu")
	if err != nil {
		klog.Errorf("创建ak失败 %v", err)
		return err
	}
	sk, err := util.GenerateSK()
	if err != nil {
		klog.Errorf("创建sk失败 %v", err)
		return err
	}

	obj := &model.Access{
		UserModel: rainbow.UserModel{
			UserId: req.UserId,
		},
		AccessKey: ak,
		SecretKey: sk,
	}
	if len(req.UserName) == 0 {
		userObj, err := s.factory.Task().GetUser(ctx, req.UserId)
		if err == nil {
			obj.UserName = userObj.Name
		}
	}

	if req.ExpireTime != nil {
		expireTime, err := parseTime(*req.ExpireTime)
		if err != nil {
			klog.Errorf("解析 ak/sk 过期时间失败: %v", err)
			return err
		}
		obj.ExpireTime = &expireTime
	}

	if _, err = s.factory.Access().Create(ctx, obj); err != nil {
		klog.Errorf("创建 ak/sk 失败 %v", err)
		return err
	}
	return nil
}

func (s *ServerController) DeleteAccess(ctx context.Context, ak string) error {
	return s.factory.Access().Delete(ctx, ak)
}

func (s *ServerController) ListAccesses(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	listOption.SetDefaultPageOption()

	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}
	opts := []db.Options{
		db.WithUser(listOption.UserId),
		db.WithAccessKeyLike(listOption.NameSelector),
	}

	var err error
	pageResult.Total, err = s.factory.Access().Count(ctx, opts...)
	if err != nil {
		klog.Errorf("获取 ak/sk 总数失败 %v", err)
		pageResult.Message = err.Error()
	}
	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithCreateOrderByASC(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Access().List(ctx, opts...)
	if err != nil {
		klog.Errorf("获取 ak/sk 列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}
