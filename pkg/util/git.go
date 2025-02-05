package util

import (
	"fmt"
	"strings"

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

func (g *Git) Checkout() error {
	currentBranch, err := g.CurrentBranch()
	if err != nil {
		return err
	}
	if currentBranch == g.Branch {
		return nil
	}

	localBranches, err := g.LocalBranches()
	if err != nil {
		return err
	}

	var cmd exec.Cmd
	if InSlice(g.Branch, localBranches) {
		cmd = g.executor.Command("git", "checkout", g.Branch)
	} else {
		cmd = g.executor.Command("git", "checkout", "remotes/origin/master", "-b", g.Branch)
	}
	cmd.SetDir(g.RepoDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(out))
	}
	return nil
}

func (g *Git) Push() error {
	if err := g.Add(); err != nil {
		return err
	}
	if err := g.Commit(); err != nil {
		return err
	}

	cmd := g.executor.Command("git", "push", "origin", "HEAD")
	cmd.SetDir(g.RepoDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(out))
	}
	return nil
}

func (g *Git) Add() error {
	cmd := g.executor.Command("git", "add", ".")
	cmd.SetDir(g.RepoDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(out))
	}

	return nil
}

func (g *Git) Commit() error {
	cmd := g.executor.Command("git", "commit", "-m", g.Title)
	cmd.SetDir(g.RepoDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v %s", err, string(out))
	}
	return nil
}

func (g *Git) CurrentBranch() (string, error) {
	cmd := g.executor.Command("git", "branch", "--show-current")
	cmd.SetDir(g.RepoDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *Git) LocalBranches() ([]string, error) {
	cmd := g.executor.Command("git", "branch")
	cmd.SetDir(g.RepoDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%v %s", err, string(out))
	}

	var branches []string
	for _, b := range strings.Split(string(out), "\n") {
		if len(b) == 0 {
			continue
		}
		branches = append(branches, strings.TrimSpace(b))
	}
	return branches, nil
}
