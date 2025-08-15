package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/api/server/router"
	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	serverFilePath = flag.String("configFile", "./config.yaml", "config file")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	opts, err := options.NewServerOptions(*serverFilePath)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	// 安装 http 路由
	router.InstallRouters(opts)

	for _, runner := range []func(context.Context, int) error{opts.Controller.Server().Run} {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbow agent: ", err)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", opts.ComponentConfig.Default.Listen),
		Handler: opts.HttpEngine,
	}
	go func() {
		err = srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			klog.Fatal("failed to listen rainbow server: ", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	klog.Info("shutting rainbow server down ...")

	// The context is used to inform the server it has 5 seconds to finish the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		klog.Fatalf("rainbow server forced to shutdown: %v", err)
	}
}
