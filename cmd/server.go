package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/member"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/user"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	harborv1 "github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
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

	harborCfg := opts.ComponentConfig.Server.Harbor

	cs, err := harborv1.NewClientSet(&harborv1.ClientSetConfig{
		URL:      harborCfg.URL,
		Username: harborCfg.Username,
		Password: harborCfg.Password,
	})
	if err != nil {
		klog.Fatal(err)
	}
	harborClient := cs.V2()

	name := "test7"
	// 创建项目
	_, err = harborClient.Project.CreateProject(context.TODO(), &project.CreateProjectParams{
		Project: &models.ProjectReq{
			Metadata: &models.ProjectMetadata{
				Public: "true",
			},
			ProjectName: name,
		},
	})
	if err != nil {
		klog.Fatal("create", err.Error())
	}

	// 创建用户
	_, err = harborClient.User.CreateUser(context.TODO(), &user.CreateUserParams{
		UserReq: &models.UserCreationReq{
			Username: name,
			Password: "Test123456!",
			Comment:  "pixiuhub",
			Email:    fmt.Sprintf("%s@qq.com", name),
			Realname: name,
		},
	})

	if err != nil {
		if apiErr, ok := err.(*runtime.APIError); ok {
			fmt.Printf("API Error: Code=%d, Response=%s\n", apiErr.Code, apiErr.Response)
		} else {
			fmt.Printf("Other Error: %v\n", err)
		}
	}

	// 关联用户到项目
	_, err = harborClient.Member.CreateProjectMember(context.TODO(), &member.CreateProjectMemberParams{
		ProjectNameOrID: name,
		ProjectMember: &models.ProjectMember{
			RoleID: 4,
			MemberUser: &models.UserEntity{
				Username: name,
			},
		},
	})
	if err != nil {
		if apiErr, ok := err.(*runtime.APIError); ok {
			fmt.Printf("API Error: Code=%d, Response=%s\n", apiErr.Code, apiErr.Response)
		} else {
			fmt.Printf("Other Error: %v\n", err)
		}
	}

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
