package main

import (
	"context"
	"flag"
	"time"

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
	klog.Infof("agent connected to rpcServer(%s)", opts.ComponentConfig.Agent.RpcServer)

	agentConfig := opts.ComponentConfig.Agent
	// 启动协程，接受服务段回调 client 的请求
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				klog.Errorf("Receive error: %v", err)
				return
			}
			klog.Infof("node(%s) received from server: %s", agentConfig.Name, msg.Result)

			// 启动搜索
			if err = opts.Controller.Agent().Search(context.TODO(), msg.Result); err != nil {
				klog.Errorf("failed to search remote repository or tags %v", err)
			}
		}
	}()

	// 向 rpc 服务端进行注册
	if err = stream.Send(&pb.Request{ClientId: agentConfig.Name, Payload: []byte("pong")}); err != nil {
		klog.Fatal("client(%s) 向 rpc 服务注册失败", err)
	}

	// 启动客户端探活API
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		t := time.Now().Format("2006-01-02 15:04:05")
		if err = stream.Send(&pb.Request{
			ClientId: agentConfig.Name,
			Payload:  []byte("pong at " + t),
		}); err != nil {
			klog.Errorf("client(%s) 探活 RPC 服务端失败 at %v %v", agentConfig.Name, t, err)
		}
	}
}
