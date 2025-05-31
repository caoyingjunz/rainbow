package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/api/agent/router"
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

	// 安装 agent  路由
	router.InstallAgentRouter(opts)

	runers := []func(context.Context, int) error{opts.Controller.Agent().Run}
	for _, runner := range runers {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbow agent: ", err)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8091),
		Handler: opts.HttpEngine,
	}
	go func() {
		klog.Infof("启动 rainbow agent 服务")
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			klog.Fatal("failed to listen rainbow agent: ", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	klog.Info("shutting rainbow agent down ...")

	// The context is used to inform the server it has 5 seconds to finish the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		klog.Fatalf("rainbow agent forced to shutdown: %v", err)
	}
}
