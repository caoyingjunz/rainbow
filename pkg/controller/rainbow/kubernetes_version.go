package rainbow

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

func (s *ServerController) ListKubernetesVersions(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	// 初始化分页属性
	listOption.SetDefaultPageOption()
	pageResult := types.PageResult{
		PageRequest: types.PageRequest{
			Page:  listOption.Page,
			Limit: listOption.Limit,
		},
	}

	opts := []db.Options{
		db.WithTagLike(listOption.NameSelector),
	} // 先写条件，再写排序，再偏移，再设置每页数量

	var err error
	// 先获取总数
	pageResult.Total, err = s.factory.Task().GetKubernetesVersionCount(ctx, opts...)
	if err != nil {
		klog.Errorf("获取镜像总数失败 %v", err)
		pageResult.Message = err.Error()
	}

	offset := (listOption.Page - 1) * listOption.Limit
	opts = append(opts, []db.Options{
		db.WithTagOrderByDESC(),
		db.WithOffset(offset),
		db.WithLimit(listOption.Limit),
	}...)
	pageResult.Items, err = s.factory.Task().ListKubernetesVersions(ctx, opts...)
	if err != nil {
		klog.Errorf("获取镜像列表失败 %v", err)
		pageResult.Message = err.Error()
		return pageResult, err
	}

	return pageResult, nil
}

func (s *ServerController) SyncKubernetesTags(ctx context.Context, req *types.CallKubernetesTagRequest) (interface{}, error) {
	key := uuid.NewString()

	data, err := json.Marshal(types.CallMetaRequest{
		Type:                     types.CallKubernetesTagType,
		Uid:                      key,
		CallKubernetesTagRequest: req,
	})
	if err != nil {
		return nil, err
	}

	val, err := s.Call(ctx, req.ClientId, key, data)
	if err != nil {
		return nil, err
	}
	var Tags []Tag
	if err = json.Unmarshal(val, &Tags); err != nil {
		klog.Errorf("序列号 k8s tag 失败 %v", err)
		return nil, err
	}

	oldTags, err := s.factory.Task().ListKubernetesVersions(ctx)
	if err != nil {
		klog.Errorf("获取历史版本失败 %v", err)
		return nil, err
	}
	oldMap := make(map[string]bool)
	for _, oldTag := range oldTags {
		oldMap[oldTag.Tag] = true
	}

	addVersions := make([]string, 0)
	for _, tag := range Tags {
		if oldMap[tag.Name] {
			continue
		}
		err = s.factory.Task().CreateKubernetesVersion(ctx, &model.KubernetesVersion{
			Tag: tag.Name,
		})
		// 新增成功
		if err == nil {
			addVersions = append(addVersions, tag.Name)
		} else {
			klog.Errorf("同步k8s 版本(%s) 失败 %v", tag.Name, err)
		}
	}

	klog.Infof("新增k8s同步版本(%v)", addVersions)
	return addVersions, nil
}

type Tag struct {
	Name string `json:"name"`
}
