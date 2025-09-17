package rainbow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"k8s.io/klog/v2"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

var (
	RpcClients map[string]pb.Tunnel_ConnectServer
)

// Connect 提供 rpc 注册接口
func (s *ServerController) Connect(stream pb.Tunnel_ConnectServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			klog.Errorf("stream.Recv %v", err)
			return err
		}

		s.lock.Lock()
		if RpcClients == nil {
			RpcClients = make(map[string]pb.Tunnel_ConnectServer)
		}
		old, ok := RpcClients[req.ClientId]
		if !ok || old != stream {
			RpcClients[req.ClientId] = stream
			klog.Infof("client(%s) rpc 注册成功", req.ClientId)
		}
		s.lock.Unlock()
		klog.V(2).Infof("Received %s from %s", string(req.Payload), req.ClientId)
	}
}

type HubTagResponse struct {
	Count    int         `json:"count"`
	Next     string      `json:"next"`
	Previous interface{} `json:"previous"` // 可能是 null 或字符串
	Results  []TagResult `json:"results"`
}

type TagResult struct {
	Images              []ImageInfo `json:"images"`
	LastUpdated         time.Time   `json:"last_updated"`
	LastUpdater         int64       `json:"last_updater"`
	LastUpdaterUsername string      `json:"last_updater_username"`
	Name                string      `json:"name"`
	Repository          int64       `json:"repository"`
	FullSize            int64       `json:"full_size"`
	V2                  bool        `json:"v2"`
	TagStatus           string      `json:"tag_status"`
	TagLastPulled       time.Time   `json:"tag_last_pulled"`
	TagLastPushed       time.Time   `json:"tag_last_pushed"`
	MediaType           string      `json:"media_type"`
	ContentType         string      `json:"content_type"`
	Digest              string      `json:"digest"`
}

type HubTagInfoResponse struct {
	Creator int         `json:"creator"`
	Images  []ImageInfo `json:"images"`
}

type ImageInfo struct {
	Features     string    `json:"features"`
	Variant      *string   `json:"variant"` // 可能是 null
	Digest       string    `json:"digest"`
	OS           string    `json:"os"`
	OSFeatures   string    `json:"os_features"`
	OSVersion    *string   `json:"os_version"` // 可能是 null
	Size         int64     `json:"size"`
	Status       string    `json:"status"`
	LastPulled   time.Time `json:"last_pulled"`
	LastPushed   time.Time `json:"last_pushed"`
	Architecture string    `json:"architecture"`
}

func (s *ServerController) preRemoteSearch(ctx context.Context, req types.RemoteSearchRequest) error {
	switch req.Hub {
	case types.ImageHubDocker, types.ImageHubGCR, types.ImageHubQuay:
	default:
		return fmt.Errorf("unsupported image hub type %s", req.Hub)
	}

	return nil
}

func (s *ServerController) SearchRepositories(ctx context.Context, req types.RemoteSearchRequest) (interface{}, error) {
	req.Query = strings.TrimSpace(req.Query)
	if len(req.Query) == 0 {
		return []types.CommonSearchRepositoryResult{}, nil
	}

	// 设置默认仓库类型
	if len(req.Hub) == 0 {
		req.Hub = types.ImageHubDocker
	}

	if err := s.preRemoteSearch(ctx, req); err != nil {
		return nil, err
	}

	key := uuid.NewString()
	data, err := json.Marshal(types.RemoteMetaRequest{
		Type:                    1,
		Uid:                     key,
		RepositorySearchRequest: req,
	})
	if err != nil {
		klog.Errorf("序列化(%v)失败 %v", req, err)
		return nil, err
	}

	val, err := s.doSearch(ctx, req.ClientId, key, data)
	if err != nil {
		return nil, err
	}

	var searchResp []types.CommonSearchRepositoryResult
	if err = json.Unmarshal(val, &searchResp); err != nil {
		klog.Errorf("序列化 HubSearchResponse失败: %v", err)
		return nil, fmt.Errorf("序列化 HubSearchResponse失败: %v", err)
	}

	return searchResp, nil
}

func (s *ServerController) SearchRepositoryTags(ctx context.Context, req types.RemoteTagSearchRequest) (interface{}, error) {
	key := uuid.NewString()
	data, err := json.Marshal(types.RemoteMetaRequest{
		Type:             2,
		Uid:              key,
		TagSearchRequest: req,
	})
	if err != nil {
		klog.Errorf("序列化(%v)失败 %v", req, err)
		return nil, err
	}

	val, err := s.doSearch(ctx, req.ClientId, key, data)
	if err != nil {
		return nil, err
	}

	var tagResp []types.TagResult
	if err = json.Unmarshal(val, &tagResp); err != nil {
		return nil, err
	}

	return tagResp, nil
}

func (s *ServerController) SearchRepositoryTagInfo(ctx context.Context, req types.RemoteTagInfoSearchRequest) (interface{}, error) {
	key := uuid.NewString()
	data, err := json.Marshal(types.RemoteMetaRequest{
		Type:                 3,
		Uid:                  key,
		TagInfoSearchRequest: req,
	})
	if err != nil {
		klog.Errorf("序列化(%v)失败 %v", req, err)
		return nil, err
	}

	val, err := s.doSearch(ctx, req.ClientId, key, data)
	if err != nil {
		return nil, err
	}

	var infoResp HubTagInfoResponse
	if err = json.Unmarshal(val, &infoResp); err != nil {
		return nil, err
	}
	return infoResp, nil
}

func (s *ServerController) doSearch(ctx context.Context, clientId string, key string, data []byte) ([]byte, error) {
	client := GetRpcClient(clientId, RpcClients)
	if client == nil {
		klog.Errorf("未发现可用的 agent，请联系管理员")
		return nil, fmt.Errorf("未发现可用的 agent，请联系管理员")
	}

	if err := client.Send(&pb.Response{Result: data}); err != nil {
		klog.Errorf("调用 Client(%v)失败 %v", clientId, err)
		return nil, fmt.Errorf("调用 Client(%v) 失败 %v", clientId, err)
	}

	val, err := s.GetResult(ctx, key)
	if err != nil {
		return nil, err
	}

	var sr types.SearchResult
	if err = json.Unmarshal([]byte(val), &sr); err != nil {
		klog.Errorf("反序列化（%v）失败 %v", val, err)
		return nil, err
	}
	if sr.StatusCode != 0 {
		klog.Errorf("远程调用失败 %v", err)
		return nil, fmt.Errorf(sr.ErrMessage)
	}

	return sr.Result, nil
}

func (s *ServerController) GetResult(ctx context.Context, key string) (string, error) {
	// 先尝试直接获取
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == nil {
		return val, nil // key 存在直接返回
	}
	if err != redis.Nil {
		return "", fmt.Errorf("redis error: %w", err) // 非"不存在"错误
	}

	// key 不存在，准备订阅通知
	channel := fmt.Sprintf("__keyspace@0__:%s", key) // Redis 通知频道格式
	pubSub := s.redisClient.Subscribe(ctx, channel)
	defer pubSub.Close()

	if _, err = pubSub.Receive(ctx); err != nil {
		return "", fmt.Errorf("subscribe failed: %w", err)
	}

	// 再次检查（避免订阅期间 key 被设置）
	val, err = s.redisClient.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}

	// 60 秒超时
	waitCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	ch := pubSub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg.Payload == "set" { // 只响应 set 操作
				val, err := s.redisClient.Get(ctx, key).Result()
				if err == nil {
					return val, nil
				}
			}
		case <-waitCtx.Done():
			return "", fmt.Errorf("wait timeout for search(%s)", key)
		}
	}
}
