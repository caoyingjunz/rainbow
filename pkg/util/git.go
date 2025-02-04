package util

import (
	"github.com/caoyingjunz/pixiulib/exec"
)

// Git 封装 git 命令行，以避免依赖，golang 的 git 库需要按照 c 库
type Git struct {
	RepoDir  string
	Branch   string
	Title    string
	executor exec.Interface
}

func NewGit(repoDir string, branch string, title string) *Git {
	return &Git{
		RepoDir:  repoDir,
		Branch:   branch,
		Title:    title,
		executor: exec.New(),
	}
}

func (g *Git) Push() error {
	return nil
}
