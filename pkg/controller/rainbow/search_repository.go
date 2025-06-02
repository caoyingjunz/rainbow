package rainbow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

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

	data, err := json.Marshal(types.RemoteMetaRequest{
		Type:                    1,
		Uid:                     uuid.NewString(),
		RepositorySearchRequest: req,
	})
	if err != nil {
		return nil, err
	}
	if err = client.Send(&pb.Response{Result: data}); err != nil {
		klog.Errorf("调用 Client(%v)失败 %v", req.ClientId, err)
		return nil, fmt.Errorf("调用 Client(%v) 失败 %v", req.ClientId, err)
	}

	// TODO: 回调客户端不提供返回值，通过redis缓存临时规避
	return nil, nil
}

func (s *ServerController) SearchRepositoryTags(ctx context.Context, req types.RemoteTagSearchRequest) (interface{}, error) {
	return nil, nil
}
