package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	filePath = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rlog.SetLogLevel("warn")

	opts, err := options.NewOptions(*filePath)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	for _, runner := range []func(context.Context, int) error{opts.Controller.Agent().Run} {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbow agent: %v", err)
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
		Expression: fmt.Sprintf("%s || all", opts.ComponentConfig.Agent.Name), // 只订阅指定自身或者未指定的agent
	},
		opts.Controller.Agent().Search)
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

	// 启动 RPC 探活 API
	r := gin.Default()
	healthz := r.Group("/healthz")
	{
		healthz.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		})
	}

	r.Run(fmt.Sprintf(":%d", opts.ComponentConfig.Agent.HealthzPort))
}
