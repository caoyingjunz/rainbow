package builder

import (
	"context"
	"fmt"
	"github.com/caoyingjunz/pixiulib/exec"
	"github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/docker/docker/client"
	"k8s.io/klog/v2"
	"time"
)

type BuilderController struct {
	Callback   string
	RegistryId int64
	BuilderId  int64
	DockerFile string
	Repo       string
	Arch       string

	httpClient util.HttpInterface
	exec       exec.Interface
	docker     *client.Client

	Cfg      config.Config
	Registry config.Registry
	Images   config.Image
}

func (b *BuilderController) Login() error {
	cmd := []string{"docker", "login", "-u", b.Registry.Username, "-p", b.Registry.Password}
	if b.Registry.Repository != "" {
		cmd = append(cmd, b.Registry.Repository)
	}
	out, err := b.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to login in image %v %v", string(out), err)
	}
	klog.Infof("镜像仓库登录完成")
	return nil
}

func NewBuilderController(cfg config.Config) *BuilderController {
	return &BuilderController{
		Cfg:        cfg,
		Callback:   cfg.Builder.Callback,
		BuilderId:  cfg.Builder.BuilderId,
		RegistryId: cfg.Builder.RegistryId,
		Repo:       cfg.Builder.Repo,
		Arch:       cfg.Builder.Arch,
		httpClient: util.NewHttpClient(5*time.Second, cfg.Plugin.Callback),
	}
}

func (b *BuilderController) Complete() error {
	var err error
	if err = b.doComplete(); err != nil {
		klog.Info(err)
	}
	return nil
}

func (b *BuilderController) Close() {
	if b.docker != nil {
		_ = b.docker.Close()
	}
}

func (b *BuilderController) doComplete() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	b.docker = cli
	b.exec = exec.New()
	b.Registry = b.Cfg.Registry

	return b.Validate()
}

func (b *BuilderController) BuildAndPushImage() error {
	imageName := fmt.Sprintf(
		"%s/%s/%s",
		b.Registry.Repository,
		b.Registry.Namespace,
		b.Repo,
	)
	b.SyncBuildStatus("初始化构建环境")
	klog.Info("Init build env")
	buildInitCmd := []string{"docker", "buildx", "create", "--name", "multi-builder", "--driver", "docker-container", "--use"}
	out, err := b.exec.Command(buildInitCmd[0], buildInitCmd[1:]...).CombinedOutput()
	if err != nil {
		b.SyncBuildStatus("初始化构建失败")
		return fmt.Errorf("failed to build image: %w\n%s", err, string(out))

	}
	b.SyncBuildStatus("开始构建上传镜像")
	klog.Infof("Starting build image %s", imageName)
	buildAndPushCmd := []string{"docker", "buildx", "build", "--platform", b.Arch, "-t", imageName, "--push", "."}

	out, err = b.exec.Command(buildAndPushCmd[0], buildAndPushCmd[1:]...).CombinedOutput()
	if err != nil {
		b.SyncBuildStatus("镜像构建上传失败")
		return fmt.Errorf("failed to build image: %w\n%s", err, string(out))
	}
	fmt.Println(string(out))
	b.SyncBuildStatus("镜像构建上传成功")
	klog.Infof("Image built successfully: %s", imageName)

	return nil
}

func (b *BuilderController) Validate() error {
	if _, err := b.docker.Ping(context.Background()); err != nil {
		klog.Errorf("%v", err)
		return err
	}
	klog.Infof("builder validate completed")
	return nil
}

func (b *BuilderController) Run() error {
	err := b.Login()
	if err != nil {
		return err
	}
	err = b.BuildAndPushImage()
	if err != nil {
		return err
	}

	return nil
}

func (b *BuilderController) SyncBuildStatus(msg string) {
	url := fmt.Sprintf("%s/rainbow/builds/%d/status", b.Callback, b.BuilderId)
	err := b.httpClient.Post(url, nil, map[string]interface{}{"Status": msg}, nil)
	if err != nil {
		klog.Errorf("同步 %s 失败 %v", msg, err)
	} else {
		klog.Infof("同步 %s 成功", msg)
	}
}
