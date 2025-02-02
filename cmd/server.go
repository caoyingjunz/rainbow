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

	"github.com/caoyingjunz/rainbow/api/server/router"
	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	opts, err := options.NewServerOptions(*filePath)
	if err != nil {
		klog.Fatal(err)
	}
	if err = opts.Complete(); err != nil {
		klog.Fatal(err)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8090),
		Handler: opts.HttpEngine,
	}

	// 安装 http 路由
	router.InstallRouters(opts)

	go func() {
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
