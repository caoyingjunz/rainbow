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
		return fmt.Errorf("docker login failed (using %s): %v, output: %s", "docker", err, string(output))
	}

	return nil
}
