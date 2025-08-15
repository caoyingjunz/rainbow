package docker

import (
	"fmt"
	"os/exec"
	"strings"
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
