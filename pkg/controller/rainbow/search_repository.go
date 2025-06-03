package rainbow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func (s *ServerController) SearchRepositories(ctx context.Context, req types.RemoteSearchRequest) (interface{}, error) {
	client := GetRpcClient(req.ClientId, RpcClients)
	if client == nil {
		klog.Errorf("client not connected or register")
		return nil, fmt.Errorf("client not connected or register")
	}

	key := uuid.NewString()
	data, err := json.Marshal(types.RemoteMetaRequest{
		Type:                    1,
		Uid:                     key,
		RepositorySearchRequest: req,
	})
	if err != nil {
		return nil, err
	}
	if err = client.Send(&pb.Response{Result: data}); err != nil {
		klog.Errorf("调用 Client(%v)失败 %v", req.ClientId, err)
		return nil, fmt.Errorf("调用 Client(%v) 失败 %v", req.ClientId, err)
	}

	return s.GetResult(ctx, key)
}

func (s *ServerController) SearchRepositoryTags(ctx context.Context, req types.RemoteTagSearchRequest) (interface{}, error) {
	return nil, nil
}

func (s *ServerController) GetResult(ctx context.Context, key string) (string, error) {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == nil {
		return val, nil // key 存在直接返回
	}
	if err != redis.Nil {
		return "", fmt.Errorf("redis error: %w", err) // 非"不存在"错误
	}

	channel := fmt.Sprintf("__keyspace@0__:%s", key) // Redis 通知频道格式
	pubSub := s.redisClient.Subscribe(ctx, channel)
	defer pubSub.Close()
	if _, err = pubSub.Receive(ctx); err != nil {
		return "", fmt.Errorf("subscribe failed: %w", err)
	}

	// 30 秒超时
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
			return "", fmt.Errorf("wait timeout for key: %s", key)
		}
	}
}
