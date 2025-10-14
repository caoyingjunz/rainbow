package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	rainbowdFile = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rlog.SetLogLevel("warn")

	opts, err := options.NewOptions(*rainbowdFile)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	for _, runner := range []func(context.Context, int) error{opts.Controller.Rainbowd().Run} {
		if err = runner(context.TODO(), 2); err != nil {
			klog.Fatal("failed to run rainbowd: %v", err)
		}
	}

	rocketmqCfg := opts.ComponentConfig.Rocketmq
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(rocketmqCfg.NameServers), // NameServer地址
		consumer.WithCredentials(primitive.Credentials{AccessKey: rocketmqCfg.Credential.AccessKey, SecretKey: rocketmqCfg.Credential.SecretKey}),
		consumer.WithGroupName(rocketmqCfg.GroupName),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	)
	if err != nil {
		klog.Fatalf("new rocketmq consumer error: %s", err.Error())
	}
	err = c.Subscribe(rocketmqCfg.Topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: fmt.Sprintf("rainbowd-%s", opts.ComponentConfig.Rainbowd.Name)},
		opts.Controller.Rainbowd().Subscribe)
	if err != nil {
		klog.Fatalf("订阅主题失败: %v", err)
	}

	err = c.Start()
	if err != nil {
		klog.Fatalf("启动消费者失败: %v", err)
	}
	defer func() {
		_ = c.Shutdown()
	}()

	klog.Infof("rainbowd 已启动")
	select {}
}
