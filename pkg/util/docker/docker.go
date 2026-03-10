package docker

import (
	//"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func LoginDocker(registry, username, password string) error {
	if registry == "" || username == "" || password == "" {
		return fmt.Errorf("missing required environment variables")
	}

	cmd := exec.Command("docker", "login", registry, "-u", username, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(output))
	}

	return nil
}

func LogoutDocker(registry string) error {
	cmd := exec.Command("docker", "logout", registry)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(output))
	}

	return nil
}

func PullImage(image string) error {
	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()

	//
	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	return err
	//}
	//// 获取标准错误管道（如果也需要）
	//stderr, err := cmd.StderrPipe()
	//if err != nil {
	//	return err
	//}
	//if err = cmd.Start(); err != nil {
	//	return err
	//}
	//
	//// 使用 scanner 逐行读取 stdout
	//go func() {
	//	scanner := bufio.NewScanner(stdout)
	//	for scanner.Scan() {
	//		fmt.Println(scanner.Text())
	//	}
	//}()
	//// 同样处理 stderr（通常 docker pull 的输出全部在 stdout 中，但为了全面也处理）
	//go func() {
	//	scanner := bufio.NewScanner(stderr)
	//	for scanner.Scan() {
	//		_, _ = fmt.Fprintln(os.Stderr, scanner.Text())
	//	}
	//}()
	//if err = cmd.Wait(); err != nil {
	//	return err
	//}
	//
	//return nil
}

func ImageExist(image string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return false, err
	}
	defer cli.Close()

	_, _, err = cli.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		if client.IsErrNotFound(err) { // 如果是镜像不存在的错误，可以根据 client.IsErrNotFound(err) 判断
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func RemoveImage(image string, force bool) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	options := types.ImageRemoveOptions{Force: force}
	_, err = cli.ImageRemove(context.Background(), image, options)
	return err
}

func TagImage(sourceImage, targetImage string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer cli.Close()

	if err = cli.ImageTag(context.Background(), sourceImage, targetImage); err != nil {
		return err
	}

	options := types.ImageRemoveOptions{Force: false}
	_, err = cli.ImageRemove(context.Background(), sourceImage, options)
	return err
}
