package main

import (
	"context"
	"flag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"

	pb "github.com/caoyingjunz/rainbow/api/rpc/proto"
	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	filePath = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	opts, err := options.NewOptions(*filePath)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	for _, runner := range []func(context.Context, int) error{opts.Controller.Agent().Run} {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbow agent: ", err)
		}
	}

	conn, err := grpc.Dial(opts.ComponentConfig.Agent.RpcServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		klog.Fatalf("failed to connect rpc server %v", err)
	}
	defer conn.Close()

	c := pb.NewTunnelClient(conn)
	stream, err := c.Connect(context.Background())
	if err != nil {
		klog.Fatalf("%v", err)
	}
	klog.Infof("agent connected to rpcServer")

	agentConfig := opts.ComponentConfig.Agent
	// 启动协程，接受服务段回调 client 的请求
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				klog.Errorf("Receive error: %v", err)
				return
			}
			// TODO
			klog.Infof("node(%s) received from server: %s", agentConfig.Name, msg.Result)
		}
	}()

	select {}
}
