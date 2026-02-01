package rainbow

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

func (s *AgentController) ProcessGithub(ctx context.Context, req *types.CallGithubRequest) ([]byte, error) {
	// op 作为预留字段，后续用于操作更多行为, 默认创建
	// 目前仅支持创建 github 项目
	agent, err := s.factory.Agent().GetByName(ctx, s.name)
	if err != nil {
		klog.Errorf("获取 agent(%s)属性失败 %v", s.name, err)
		return nil, err
	}
	if len(agent.GithubToken) == 0 {
		klog.Infof("agent(%s) 的 token 为空", agent.Name)
		return nil, fmt.Errorf("agent(%s) 的 token 为空", s.name)
	}

	body, err := util.BuildHttpBody(map[string]interface{}{"name": req.Repo, "private": true})
	if err != nil {
		return nil, err
	}
	httpClient := util.HttpClientV2{URL: types.GithubAPIBase + "/user/repos"}
	err1 := httpClient.Method(http.MethodPost).
		WithTimeout(30 * time.Second).
		WithHeader(map[string]string{
			"Content-Type":         "application/json",
			"Accept":               "application/vnd.github+json",
			"Authorization":        fmt.Sprintf("Bearer %s", agent.GithubToken),
			"X-GitHub-Api-Version": "2022-11-28",
		}).
		WithBody(body).
		Do(nil)
	if err1 != nil {
		klog.Errorf("创建 repo(%s) 失败 %v", s.name, err1)
		return nil, err1
	}

	return nil, nil
}
