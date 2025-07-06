package docker

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getDockerCommand() string {
	// 首先尝试系统PATH中的docker命令
	if path, err := exec.LookPath("docker"); err == nil {
		return path
	}
	var dockerPath string
	// 根据架构选择项目中的docker二进制
	arch := runtime.GOARCH
	if arch == "amd64" {
		dockerPath = fmt.Sprintf("./docker-bin/docker-%s", arch)
	} else {
		dockerPath = "./docker-bin/docker"
	}

	if _, err := os.Stat(dockerPath); err == nil {
		return dockerPath
	}

	// 默认返回docker命令，让exec.Command报错
	return "docker"
}

func LoginDocker(registry, username, password string) error {
	if registry == "" || username == "" || password == "" {
		return fmt.Errorf("missing required environment variables")
	}

	dockerCmd := getDockerCommand()
	cmd := exec.Command(dockerCmd, "login", registry, "-u", username, "--password-stdin")
	cmd.Stdin = strings.NewReader(password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker login failed (using %s): %v, output: %s", dockerCmd, err, string(output))
	}

	return nil
}
