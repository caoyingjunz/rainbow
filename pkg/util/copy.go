package util

import (
	"fmt"

	"github.com/caoyingjunz/pixiulib/exec"
)

func Copy(src, dest string) error {
	executor := exec.New()
	out, err := executor.Command("cp", "-r", src, dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(out))
	}

	return nil
}

func Move(src, dest string) error {
	executor := exec.New()
	out, err := executor.Command("mv", src, dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(out))
	}

	return nil
}
