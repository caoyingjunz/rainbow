package rpc

import (
	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	"github.com/caoyingjunz/rainbow/cmd/app/options"
	"google.golang.org/grpc"
	"log"
	"net"
)

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
