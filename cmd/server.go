package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-openapi/runtime/client"
	harborv1 "github.com/goharbor/go-client/pkg/harbor"
	harbor "github.com/goharbor/go-client/pkg/sdk/v2.0/client"

	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/ping"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/api/server/router"
	"github.com/caoyingjunz/rainbow/cmd/app/options"
)

var (
	serverFilePath = flag.String("configFile", "./config.yaml", "config file")
)

// 自定义 Transport 确保 Accept 头正确
type customTransport struct {
	base http.RoundTripper
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 设置 Accept 头为 application/json
	req.Header.Set("Accept", "application/json")
	// 设置 Content-Type 如果需要的话
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

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

	harborCfg := opts.ComponentConfig.Server.Harbor

	URL, err := url.Parse(harborCfg.URL)
	if err != nil {
		klog.Fatal(err)
	}

	harborClient := harbor.New(harbor.Config{
		URL: URL,
		Transport: &customTransport{
			base: http.DefaultTransport,
		},
		AuthInfo: client.BasicAuth(harborCfg.Username, harborCfg.Password),
	})

	status, err := harborClient.Ping.GetPing(context.TODO(), &ping.GetPingParams{})
	if err != nil {
		klog.Fatal(err)
	}
	fmt.Println("status", status.IsSuccess())

	cs, err := harborv1.NewClientSet(&harborv1.ClientSetConfig{
		URL:      harborCfg.URL,
		Username: harborCfg.Username,
		Password: harborCfg.Password,
	})
	if err != nil {
		klog.Fatal(err)
	}

	users, err := harborClient.User.ListUsers(context.TODO(), &user.ListUsersParams{})
	if err != nil {
		klog.Fatal("dd", err)
	}
	fmt.Println(users)

	projects, err := harborClient.Project.ListProjects(context.TODO(), &project.ListProjectsParams{})
	if err != nil {
		klog.Fatal("xxxxxxx", err)
	}

	fmt.Println("projects", *projects)

	return

	// 安装 http 路由
	router.InstallRouters(opts)

	for _, runner := range []func(context.Context, int) error{opts.Controller.Server().Run} {
		if err = runner(context.TODO(), 5); err != nil {
			klog.Fatal("failed to rainbow agent: ", err)
		}
	}
	defer opts.Controller.Server().Stop(context.TODO())

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
