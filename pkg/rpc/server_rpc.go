package rpc

import (
	"fmt"

	"io"
	"log"
	"net"
	"sync"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

type Server struct {
	pb.UnimplementedTunnelServer

	clients map[string]pb.Tunnel_ConnectServer
	lock    sync.RWMutex
}

func (s *Server) Connect(stream pb.Tunnel_ConnectServer) error {
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
		_, ok := s.clients[req.ClientId]
		if !ok {
			s.clients[req.ClientId] = stream
		}
		s.lock.Unlock()

		// TODO 目前是DEMO
		klog.Infof("Received from %s %s", req.ClientId, string(req.Payload))
	}
}
func (s *Server) Call(c *gin.Context) {
	_, _ = s.CallClient(c.Query("clientId"), nil)
}

func (s *Server) CallClient(clientId string, data []byte) ([]byte, error) {
	stream, ok := s.clients[clientId]
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

func Install(o *options.ServerOptions) {
	listener, err := net.Listen("tcp", ":8091")
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	cs := &Server{}
	s := grpc.NewServer()
	pb.RegisterTunnelServer(s, cs)

	go func() {
		log.Printf("grpc listening at %v", listener.Addr())
		if err = s.Serve(listener); err != nil {
			log.Fatalf("failed to serve %v", err)
		}
	}()
}
