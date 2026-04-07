package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/signatureutil"
)

type PixiuHubClient struct {
	baseURL  string
	userInfo UserOption

	// auth
	accessKey string
	signature string
}

type UserOption struct {
	Name   string
	UserId string
}

type CreateTaskOption struct {
	Name     string
	Platform string
	Register int
	Private  bool
	Images   []string
}

type CreateResult struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type ListResult struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type RegistryListResult struct {
	ListResult `json:",inline"`

	Result []model.Registry `json:"result,omitempty"`
}

type ImageListResult struct {
	ListResult `json:",inline"`

	Result []model.Image `json:"result,omitempty"`
}

func NewPixiuHubClient(url, accessKey, secretKey string) (*PixiuHubClient, error) {
	pc := &PixiuHubClient{
		baseURL:   url,
		accessKey: accessKey,
	}

	pc.signature = signatureutil.GenerateSignature(
		map[string]string{"action": "pullOrCacheRepo", "accessKey": accessKey},
		[]byte(secretKey))

	user, err := GetUserInfoByAccessKey(pc.baseURL, pc.accessKey, pc.signature)
	if err != nil {
		return nil, err
	}
	pc.userInfo = UserOption{
		Name:   user.Name,
		UserId: user.UserId,
	}

	return pc, nil
}

func (pc *PixiuHubClient) CreateTask(ctx context.Context, o CreateTaskOption) error {
	data, err1 := json.Marshal(map[string]interface{}{
		"name":         o.Name,
		"architecture": o.Platform,
		"images":       o.Images,
		"user_id":      pc.userInfo.UserId,
		"user_name":    pc.userInfo.Name,
		"public_image": !o.Private,
		"register_id":  o.Register,
	})
	if err1 != nil {
		return err1
	}

	var result CreateResult
	httpClient := util.HttpClientV2{URL: fmt.Sprintf("%s/api/v2/tasks", pc.baseURL)}
	if err := httpClient.Method(http.MethodPost).
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{"X-ACCESS-KEY": pc.accessKey, "Authorization": pc.signature}).
		WithBody(bytes.NewBuffer(data)).
		Do(&result); err != nil {
		return err
	}

	if result.Code == 200 {
		klog.Infof("Image sync task created successfully")
		return nil
	}
	return fmt.Errorf("image sync task created failed %s", result.Message)
}

func (pc *PixiuHubClient) ListTasks() error {
	return nil
}

func (pc *PixiuHubClient) ListRegistries() ([]model.Registry, error) {
	var result RegistryListResult
	httpClient := util.HttpClientV2{URL: fmt.Sprintf("%s/api/v2/registries?user_id=%s", pc.baseURL, pc.userInfo.UserId)}
	if err := httpClient.Method("GET").
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{"X-ACCESS-KEY": pc.accessKey, "Authorization": pc.signature}).
		Do(&result); err != nil {
		return nil, err
	}
	if result.Code == 200 {
		return result.Result, nil
	}
	return nil, fmt.Errorf("%s", result.Message)
}

func (pc *PixiuHubClient) ListImages(page, pageSize int) ([]model.Image, error) {
	var result ImageListResult
	url := fmt.Sprintf("%s/api/v2/images?user_id=%s&page=%d&limit=%d", pc.baseURL, pc.userInfo.UserId, page, pageSize)
	httpClient := util.HttpClientV2{URL: url}
	if err := httpClient.Method("GET").
		WithTimeout(5 * time.Second).
		WithHeader(map[string]string{"X-ACCESS-KEY": pc.accessKey, "Authorization": pc.signature}).
		Do(&result); err != nil {
		return nil, err
	}
	if result.Code == 200 {
		return result.Result, nil
	}
	return nil, fmt.Errorf("%s", result.Message)
}
