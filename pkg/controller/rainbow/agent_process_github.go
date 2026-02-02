package rainbow

import (
	"bytes"
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

func (s *AgentController) ProcessKubernetesTags(ctx context.Context, req *types.CallKubernetesTagRequest) ([]byte, error) {
	if !req.SyncAll {
		url := fmt.Sprintf("https://api.github.com/repos/kubernetes/kubernetes/tags?per_page=10")
		return DoHttpRequest(url)
	}

	var allData []byte
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/repos/kubernetes/kubernetes/tags?per_page=100&page=%d", page)
		data, err := DoHttpRequest(url)
		if err != nil {
			return nil, err
		}
		if len(data) == 0 || bytes.Equal(data, []byte("[]")) {
			break
		}
		allData = appendData(allData, data)

		page++
		time.Sleep(10 * time.Millisecond) // 避免请求过快
	}

	return allData, nil
}

func appendData(allData, newData []byte) []byte {
	// 情况1: 原始二进制直接拼接
	// return append(allData, newData...)

	// 情况2: JSON数组合并（更常见）
	if len(allData) == 0 {
		return newData
	}

	// 移除allData最后的']'和newData开头的'['
	if allData[len(allData)-1] == ']' && newData[0] == '[' {
		allData = allData[:len(allData)-1]
		newData = newData[1:]
		return append(append(allData, ','), newData...)
	}

	return append(allData, newData...)
}
