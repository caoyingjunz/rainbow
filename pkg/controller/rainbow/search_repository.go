package rainbow

import (
	"context"
	"fmt"
	"io"

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

		// TODO 目前是DEMO
		//klog.Infof("Received from %s %s", req.ClientId, string(req.Payload))
	}
}

func (s *ServerController) Callback(clientId string, data []byte) ([]byte, error) {
	stream, ok := RpcClients[clientId]
	if !ok {
		klog.Errorf("client not connected")
		return nil, fmt.Errorf("client not connected")
	}

	// 发送调用请求
	err := stream.Send(&pb.Response{Result: []byte(clientId + " server callback")})
	if err != nil {
		klog.Errorf("回调客户端(%s)失败: %v", clientId, err)
		return nil, err
	}

	return nil, err
}

func (s *ServerController) SearchRepositories(ctx context.Context, req types.RemoteSearchRequest) (interface{}, error) {
	fmt.Println("req2", RpcClients)
	return nil, nil
}

func (s *ServerController) SearchRepositoryTags(ctx context.Context, req types.RemoteTagSearchRequest) (interface{}, error) {
	return nil, nil
}
