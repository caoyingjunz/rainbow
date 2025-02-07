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
	"github.com/caoyingjunz/rainbow/pkg/db/model"
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

	httpClient util.HttpInterface
	exec       exec.Interface
	docker     *client.Client

	Cfg      config.Config
	Registry config.Registry
}

func NewPluginController(cfg config.Config) *PluginController {
	return &PluginController{
		Callback:   cfg.Plugin.Callback,
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
			return fmt.Errorf("failed to get kubeadm version: %v", err)
		}
		if kubeadmVersion != p.KubernetesVersion {
			return fmt.Errorf("kubeadm version %s not match kubernetes version %s", kubeadmVersion, p.KubernetesVersion)
		}
	}

	// 检查 docker 的客户端是否正常
	if _, err := p.docker.Ping(context.Background()); err != nil {
		return err
	}

	return nil
}

func (p *PluginController) Complete() error {
	p.ReportImage()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	p.docker = cli

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
		klog.Infof("Starting install kubeadm", cmd)
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

	p.exec = exec.New()
	p.Registry = p.Cfg.Registry
	return nil
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

func (p *PluginController) doPushImage(imageToPush string) error {
	targetImage, err := p.parseTargetImage(imageToPush)
	if err != nil {
		return err
	}

	klog.Infof("starting pull image %s", imageToPush)
	// start pull
	reader, err := p.docker.ImagePull(context.TODO(), imageToPush, types.ImagePullOptions{})
	if err != nil {
		klog.Errorf("failed to pull %s: %v", imageToPush, err)
		return err
	}
	io.Copy(os.Stdout, reader)

	klog.Infof("tag %s to %s", imageToPush, targetImage)
	if err := p.docker.ImageTag(context.TODO(), imageToPush, targetImage); err != nil {
		klog.Errorf("failed to tag %s to %s: %v", imageToPush, targetImage, err)
		return err
	}

	klog.Infof("starting push image %s", targetImage)

	cmd := []string{"docker", "push", targetImage}
	out, err := p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push image %s %v %v", targetImage, string(out), err)
	}

	klog.Infof("complete push image %s", imageToPush)
	return nil
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
	var images []string

	if p.Cfg.Default.PushKubernetes {
		kubeImages, err := p.getImages()
		if err != nil {
			return fmt.Errorf("获取 k8s 镜像失败: %v", err)
		}
		images = append(images, kubeImages...)
	}

	if p.Cfg.Default.PushImages {
		fileImages, err := p.getImagesFromFile()
		if err != nil {
			return fmt.Errorf("")
		}
		images = append(images, fileImages...)
	}

	klog.V(2).Infof("get images: %v", images)
	diff := len(images)
	errCh := make(chan error, diff)

	// 登陆
	cmd := []string{"docker", "login", "-u", p.Registry.Username, "-p", p.Registry.Password}
	if p.Registry.Repository != "" {
		cmd = append(cmd, p.Registry.Repository)
	}
	out, err := p.exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to login in image %v %v", string(out), err)
	}

	var wg sync.WaitGroup
	wg.Add(diff)
	for _, i := range images {
		go func(imageToPush string) {
			defer wg.Done()
			if err := p.doPushImage(imageToPush); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	default:
	}

	return nil
}

func (p *PluginController) ReportImage() error {
	url := p.Callback + "/rainbow/agents"

	var result []model.Agent
	if err := p.httpClient.Get(url, &result); err != nil {
		fmt.Println("err", err)
		return err
	}

	return nil
}

func (p *PluginController) ReportTask() error {
	return nil
}
