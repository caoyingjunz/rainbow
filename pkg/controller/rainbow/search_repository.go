package rainbow

import (
	"context"
	"fmt"

	"io"
	"sync"

	"k8s.io/klog/v2"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	"github.com/caoyingjunz/rainbow/pkg/types"
)

type RpcServer struct {
	pb.UnimplementedTunnelServer

	clients map[string]pb.Tunnel_ConnectServer
	lock    sync.RWMutex
}

func (rs *RpcServer) Connect(stream pb.Tunnel_ConnectServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			klog.Errorf("stream.Recv %v", err)
			return err
		}

		rs.lock.Lock()
		_, ok := rs.clients[req.ClientId]
		if !ok {
			rs.clients[req.ClientId] = stream
		}
		rs.lock.Unlock()

		// TODO 目前是DEMO
		klog.Infof("Received from %s %s", req.ClientId, string(req.Payload))
	}
}

func (rs *RpcServer) Callback(clientId string, data []byte) ([]byte, error) {
	stream, ok := rs.clients[clientId]
	if !ok {
		return nil, fmt.Errorf("client not connected")
	}

	// 发送调用请求
	err := stream.Send(&pb.Response{
		Result: []byte(clientId + " server callback"),
	})
	if err != nil {
		return nil, err
	}

	return nil, err
}

func (s *ServerController) SearchRepositories(ctx context.Context, req types.RemoteSearchRequest) (interface{}, error) {

	return nil, nil
}
func (s *ServerController) SearchRepositoryTags(ctx context.Context, req types.RemoteTagSearchRequest) (interface{}, error) {
	return nil, nil
}
