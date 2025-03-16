package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/caoyingjunz/pixiulib/exec"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"k8s.io/klog/v2"

	"github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/util"
)

const (
	Kubeadm   = "kubeadm"
	IgnoreKey = "W0508"
)

type KubeadmVersion struct {
	ClientVersion struct {
		GitVersion string `json:"gitVersion"`
	} `json:"clientVersion"`
}

type KubeadmImage struct {
	Images []string `json:"images"`
}

type PluginController struct {
	KubernetesVersion string
	Callback          string
	RegistryId        int64
	TaskId            int64
	Synced            bool

	httpClient util.HttpInterface
	exec       exec.Interface
	docker     *client.Client

	Cfg      config.Config
	Registry config.Registry
	Images   []string

	Runners []Runner
}

type Runner interface {
	GetName() string
	Run() error
}

type login struct {
	name string
	p    *PluginController
}

func (l *login) GetName() string {
	return l.name
}

func (l *login) Run() error {
	cmd := []string{"docker", "login", "-u", l.p.Registry.Username, "-p", l.p.Registry.Password}
	if l.p.Registry.Repository != "" {
		cmd = append(cmd, l.p.Registry.Repository)
	}
	out, err := l.p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to login in image %v %v", string(out), err)
	}
	return nil
}

type image struct {
	name string
	p    *PluginController
}

func (i *image) GetName() string {
	return i.name
}

func (i *image) Run() error {
	var images []string
	if i.p.Cfg.Default.PushKubernetes {
		kubeImages, err := i.p.getImages()
		if err != nil {
			return fmt.Errorf("获取 k8s 镜像失败: %v", err)
		}

		if err = i.p.CreateImages(kubeImages); err != nil {
			klog.Errorf("回写k8s镜像失败: %v", err)
		}
		images = append(images, kubeImages...)
	}

	if i.p.Cfg.Default.PushImages {
		fileImages, err := i.p.getImagesFromFile()
		if err != nil {
			return fmt.Errorf("")
		}
		images = append(images, fileImages...)
	}

	i.p.Images = images
	return nil
}

func NewPluginController(cfg config.Config) *PluginController {
	return &PluginController{
		Cfg:        cfg,
		Callback:   cfg.Plugin.Callback,
		TaskId:     cfg.Plugin.TaskId,
		RegistryId: cfg.Plugin.RegistryId,
		Synced:     cfg.Plugin.Synced,
		httpClient: util.NewHttpClient(5*time.Second, cfg.Plugin.Callback),
	}
}

func (p *PluginController) Validate() error {
	if p.Cfg.Default.PushKubernetes {
		if len(p.KubernetesVersion) == 0 {
			return fmt.Errorf("failed to find kubernetes version")
		}
		// 检查 kubeadm 的版本是否和 k8s 版本一致
		kubeadmVersion, err := p.getKubeadmVersion()
		if err != nil {
			klog.Error("failed to get kubeadm version: %v", err)
			return fmt.Errorf("failed to get kubeadm version: %v", err)
		}
		if kubeadmVersion != p.KubernetesVersion {
			klog.Errorf("kubeadm version %s not match kubernetes version %s", kubeadmVersion, p.KubernetesVersion)
			return fmt.Errorf("kubeadm version %s not match kubernetes version %s", kubeadmVersion, p.KubernetesVersion)
		}
	}

	// 检查 docker 的客户端是否正常
	if _, err := p.docker.Ping(context.Background()); err != nil {
		klog.Errorf("%v", err)
		return err
	}

	return nil
}

func (p *PluginController) doComplete() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	p.docker = cli
	p.exec = exec.New()
	p.Registry = p.Cfg.Registry

	if p.Cfg.Default.PushKubernetes {
		if len(p.KubernetesVersion) == 0 {
			if len(p.Cfg.Kubernetes.Version) != 0 {
				p.KubernetesVersion = p.Cfg.Kubernetes.Version
			} else {
				p.KubernetesVersion = os.Getenv("KubernetesVersion")
			}
		}
	}

	if p.Cfg.Default.PushKubernetes {
		//cmd := []string{"sudo", "apt-get", "install", "-y", fmt.Sprintf("kubeadm=%s-00", p.Cfg.Kubernetes.Version[1:])}
		cmd := []string{"sudo", "curl", "-LO", fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/amd64/kubeadm", p.Cfg.Kubernetes.Version)}
		klog.Infof("Starting install kubeadm %s", cmd)
		out, err := p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to get kubeadm %v %v", string(out), err)
		}

		cmd2 := []string{"sudo", "install", "-o", "root", "-g", "root", "-m", "0755", "kubeadm", "/usr/local/bin/kubeadm"}
		out, err = p.exec.Command(cmd2[0], cmd2[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install kubeadm %v %v", string(out), err)
		}
	}

	p.Runners = []Runner{
		&login{name: "Registry登陆", p: p},
		&image{name: "解析镜像", p: p},
	}
	return p.Validate()
}

func (p *PluginController) Complete() error {
	if len(p.Cfg.Plugin.Driver) == 0 {
		p.Cfg.Plugin.Driver = "docker"
	}

	status, msg, process := "初始化成功", "初始化环境结束", 1
	var err error
	if err = p.doComplete(); err != nil {
		status, msg, process = "初始化失败", err.Error(), 3
	}
	_ = p.SyncTaskStatus(status, msg, process)
	return err
}

func (p *PluginController) Close() {
	if p.docker != nil {
		_ = p.docker.Close()
	}
}

func (p *PluginController) getKubeadmVersion() (string, error) {
	if _, err := p.exec.LookPath(Kubeadm); err != nil {
		return "", fmt.Errorf("failed to find %s %v", Kubeadm, err)
	}

	cmd := []string{Kubeadm, "version", "-o", "json"}
	out, err := p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to exec kubeadm version %v %v", string(out), err)
	}

	var kubeadmVersion KubeadmVersion
	if err := json.Unmarshal(out, &kubeadmVersion); err != nil {
		return "", fmt.Errorf("failed to unmarshal kubeadm version %v", err)
	}
	klog.V(2).Infof("kubeadmVersion %+v", kubeadmVersion)

	return kubeadmVersion.ClientVersion.GitVersion, nil
}

func (p *PluginController) cleanImages(in []byte) []byte {
	inStr := string(in)
	if !strings.Contains(inStr, IgnoreKey) {
		return in
	}

	klog.V(2).Infof("cleaning images: %+v", inStr)
	parts := strings.Split(inStr, "\n")
	index := 0
	for _, p := range parts {
		if strings.HasPrefix(p, IgnoreKey) {
			index += 1
		}
	}
	newInStr := strings.Join(parts[index:], "\n")
	klog.V(2).Infof("cleaned images: %+v", newInStr)

	return []byte(newInStr)
}

func (p *PluginController) getImages() ([]string, error) {
	cmd := []string{Kubeadm, "config", "images", "list", "--kubernetes-version", p.KubernetesVersion, "-o", "json"}
	out, err := p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to exec kubeadm config images list %v %v", string(out), err)
	}
	out = p.cleanImages(out)
	klog.V(2).Infof("images is %+v", string(out))

	var kubeadmImage KubeadmImage
	if err := json.Unmarshal(out, &kubeadmImage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal kubeadm images %v", err)
	}

	return kubeadmImage.Images, nil
}

func (p *PluginController) parseTargetImage(imageToPush string) (string, error) {
	// real image to push
	parts := strings.Split(imageToPush, "/")

	return p.Registry.Repository + "/" + p.Registry.Namespace + "/" + parts[len(parts)-1], nil
}

func (p *PluginController) doPushImage(imageToPush string) (string, error) {
	targetImage, err := p.parseTargetImage(imageToPush)
	if err != nil {
		return "", fmt.Errorf("failed to parse target image: %v", err)
	}

	klog.Infof("Tagging image from %s to %s", imageToPush, targetImage)

	var cmd []string
	switch p.Cfg.Plugin.Driver {
	case "skopeo":
		cmd = []string{"sudo", "chmod", "+x", "bin/skopeo", "&&", "bin/skopeo", "copy", "docker://" + imageToPush, "docker://" + targetImage}
		klog.Infof("Making skopeo executable and copying image: %s", targetImage)
		out, err := p.exec.Command("sh", "-c", strings.Join(cmd, " ")).CombinedOutput()
		if err != nil {
			klog.Errorf("Failed to execute skopeo commands: %v, output: %s", err, string(out))
			return "", fmt.Errorf("failed to execute skopeo commands: %v", err)
		}
		klog.Infof("Successfully executed skopeo commands: %s", string(out))
		return targetImage, nil

	case "docker":
		klog.Infof("Pulling image: %s", imageToPush)
		reader, err := p.docker.ImagePull(context.TODO(), imageToPush, types.ImagePullOptions{})
		if err != nil {
			klog.Errorf("Failed to pull image %s: %v", imageToPush, err)
			return "", fmt.Errorf("failed to pull image %s: %v", imageToPush, err)
		}
		io.Copy(os.Stdout, reader)

		klog.Infof("Tagging image from %s to %s", imageToPush, targetImage)
		if err := p.docker.ImageTag(context.TODO(), imageToPush, targetImage); err != nil {
			klog.Errorf("Failed to tag image %s to %s: %v", imageToPush, targetImage, err)
			return "", fmt.Errorf("failed to tag image %s to %s: %v", imageToPush, targetImage, err)
		}

		cmd = []string{"docker", "push", targetImage}
	default:
		return "", fmt.Errorf("unsupported driver: %s", p.Cfg.Plugin.Driver)
	}

	klog.Infof("Pushing image: %s", targetImage)
	out, err := p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		klog.Errorf("Failed to push image %s: %v, output: %s", targetImage, err, string(out))
		return "", fmt.Errorf("failed to push image %s: %v", targetImage, err)
	}

	klog.Infof("Successfully pushed image: %s", targetImage)
	return targetImage, nil
}

func (p *PluginController) getImagesFromFile() ([]string, error) {
	var imgs []string
	for _, i := range p.Cfg.Images {
		imageStr := strings.TrimSpace(i)
		if len(imageStr) == 0 {
			continue
		}
		if strings.Contains(imageStr, " ") {
			return nil, fmt.Errorf("error image format: %s", imageStr)
		}

		imgs = append(imgs, imageStr)
	}

	return imgs, nil
}

func (p *PluginController) Run() error {
	for _, runner := range p.Runners {
		name := runner.GetName()
		if err := runner.Run(); err != nil {
			_ = p.SyncTaskStatus(name, name+"失败", 3)
			return err
		} else {
			_ = p.SyncTaskStatus(name, name+"完成", 1)
		}
	}

	_ = p.SyncTaskStatus("开始同步镜像", "", 1)
	diff := len(p.Images)
	errCh := make(chan error, diff)

	var wg sync.WaitGroup
	wg.Add(diff)
	for _, i := range p.Images {
		go func(imageToPush string) {
			defer wg.Done()
			_ = p.SyncImageStatus(imageToPush, "", "同步进行中", "")
			target, err := p.doPushImage(imageToPush)
			if err != nil {
				_ = p.SyncImageStatus(imageToPush, target, "同步异常", err.Error())
				errCh <- err
			} else {
				_ = p.SyncImageStatus(imageToPush, target, "同步完成", "")
			}
		}(i)
	}
	wg.Wait()

	select {
	case err := <-errCh:
		if err != nil {
			_ = p.SyncTaskStatus("镜像同步结束", "存在镜像同步异常", 2)
			return err
		}
	default:
	}

	_ = p.SyncTaskStatus("镜像同步完成", "镜像全部同步完成", 2)
	return nil
}

// SyncTaskStatus
// 0 未开始
// 1 执行中
// 2 执行成功
// 3 执行失败
func (p *PluginController) SyncTaskStatus(status string, msg string, process int) error {
	if !p.Synced {
		return nil
	}
	return p.httpClient.Put(
		fmt.Sprintf("%s/rainbow/tasks/%d/status", p.Callback, p.TaskId),
		nil,
		map[string]interface{}{"status": status, "message": msg, "process": process})
}

func (p *PluginController) SyncImageStatus(name, target, status, msg string) error {
	if !p.Synced {
		return nil
	}
	return p.httpClient.Put(
		fmt.Sprintf("%s/rainbow/images/status", p.Callback),
		nil,
		map[string]interface{}{"status": status, "message": msg, "task_id": p.TaskId, "name": name, "target": target, "registry_id": p.RegistryId})
}

func (p *PluginController) CreateImages(names []string) error {
	if !p.Synced {
		return nil
	}

	return p.httpClient.Post(
		fmt.Sprintf("%s/rainbow/images/batches", p.Callback),
		nil,
		map[string]interface{}{"task_id": p.TaskId, "names": names})
}
