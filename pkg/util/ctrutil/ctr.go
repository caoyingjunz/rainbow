package ctrutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PullImage runs: ctr -n <namespace> images pull <ref>
func PullImage(namespace, ref string) error {
	cmd := exec.Command("ctr", "-n", namespace, "images", "pull", ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// TagAndRemoveSource runs ctr images tag then ctr images rm on source, matching docker.TagImage behavior.
func TagAndRemoveSource(namespace, source, target string) error {
	tagCmd := exec.Command("ctr", "-n", namespace, "images", "tag", source, target)
	if out, err := tagCmd.CombinedOutput(); err != nil {
		// If target tag already exists, keep behavior idempotent and continue.
		if !strings.Contains(string(out), "already exists") && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("ctr images tag: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}
	rmCmd := exec.Command("ctr", "-n", namespace, "images", "rm", source)
	rmCmd.Stdout = os.Stdout
	rmCmd.Stderr = os.Stderr
	if err := rmCmd.Run(); err != nil {
		return fmt.Errorf("ctr images rm: %w", err)
	}
	return nil
}
