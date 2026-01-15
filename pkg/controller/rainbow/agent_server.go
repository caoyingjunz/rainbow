package rainbow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"

	rainbowconfig "github.com/caoyingjunz/rainbow/cmd/app/config"
	"github.com/caoyingjunz/rainbow/pkg/db"
	"github.com/caoyingjunz/rainbow/pkg/db/model"
	"github.com/caoyingjunz/rainbow/pkg/types"
	"github.com/caoyingjunz/rainbow/pkg/util"
	"github.com/caoyingjunz/rainbow/pkg/util/sshutil"
)

const GitConfig = `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = {{ .URL }}
	fetch = +refs/heads/*:refs/remotes/origin/*
`

func (s *ServerController) GetAgent(ctx context.Context, agentId int64) (interface{}, error) {
	return s.factory.Agent().Get(ctx, agentId)
}

func (s *ServerController) ReconcileAgent(ctx context.Context, sshConfig *sshutil.SSHConfig, agent *model.Agent) error {
	// TODO 测试，后续删除
	agent.Name = "harbor-jobservice"
	agentName := agent.Name

	old, err := s.GetAgentContainer(sshConfig, agentName)
	if err != nil {
		klog.Errorf("获取 agent 容器失败 %v", err)
		return err
	}
	if old == nil {
		klog.Errorf("agent 容器(%s)不存在", agentName)
	}

	var (
		needUpdated bool
	)

	switch agent.Status {
	case model.RestartingAgentType:
		klog.Infof("agent(%s)重启中", agentName)
		err = s.RestartAgentContainer(sshConfig, agent)
		needUpdated = true
	case model.UpgradeAgentType:
		klog.Infof("agent(%s)升级中", agentName)
		if old != nil {
			// 先卸载原有容器，然后刷新配置，重新启动
			if err1 := s.UninstallAgentContainer(sshConfig, agent); err1 != nil {
				return fmt.Errorf("卸载 agent(%s) 失败 %v", agentName, err1)
			}
		}
		if err = s.ResetAgentMetadata(sshConfig, agent); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	if needUpdated {
		fmt.Println("需要更新状态")
	}
	return nil
}

func (s *ServerController) IsAgentRunningStatus(status string) bool {
	return false
}

func (s *ServerController) ResetAgentMetadata(sshConfig *sshutil.SSHConfig, agent *model.Agent) error {
	sshClient, err := sshutil.NewSSHClient(sshConfig)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	containerName := agent.Name
	destDir := filepath.Join(s.cfg.Rainbowd.DataDir, containerName)

	result, err := sshClient.RunCommand(fmt.Sprintf("mkdir -p %s", destDir))
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("远程命令执行失败: %s", result.Stderr)
	}

	// 配置文件 config.yaml
	data, err := util.ReadFromFile(s.cfg.Rainbowd.TemplateDir + "/config.yaml")
	if err != nil {
		return err
	}
	var cfg rainbowconfig.Config
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	// 追加差异化配置
	cfg.Agent.Name = containerName
	cfgData, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err = util.WriteIntoFile(string(cfgData), s.cfg.Rainbowd.TemplateDir+fmt.Sprintf("/%s-config.yaml", containerName)); err != nil {
		return err
	}
	// 拷贝配置文件
	if err = sshClient.UploadFile(s.cfg.Rainbowd.TemplateDir+fmt.Sprintf("/%s-config.yaml", containerName), destDir+"/config.yaml", "0644"); err != nil {
		klog.Errorf("传输 yaml 配置文件失败 %v", err)
		return err
	}

	// 拷贝 agent 每次都重置最新
	if err = sshClient.UploadFile(s.cfg.Rainbowd.TemplateDir+"/agent", destDir+"/agent", "0755"); err != nil {
		klog.Errorf("传输 agent 二进制文件失败 %v", err)
		return err
	}

	// 拷贝 plugin 项目
	if err = sshClient.UploadDir(s.cfg.Rainbowd.TemplateDir+"/plugin", destDir+"/plugin"); err != nil {
		klog.Errorf("传输 plugin 文件失败 %v", err)
		return err
	}

	gc := struct{ URL string }{URL: fmt.Sprintf("https://%s:%s@github.com/%s/plugin.git", agent.GithubUser, agent.GithubToken, agent.GithubUser)}
	tpl := template.New(containerName)
	t := template.Must(tpl.Parse(GitConfig))
	var buf bytes.Buffer
	if err = t.Execute(&buf, gc); err != nil {
		return err
	}
	if err = ioutil.WriteFile(s.cfg.Rainbowd.TemplateDir+fmt.Sprintf("/%s-git-config", containerName), buf.Bytes(), 0644); err != nil {
		klog.Errorf("生成 git config 文件失败 %v", err)
		return err
	}
	if err = sshClient.UploadFile(s.cfg.Rainbowd.TemplateDir+fmt.Sprintf("/%s-git-config", containerName), destDir+"/plugin/.git/config", "0755"); err != nil {
		klog.Errorf("传输 /plugin/.git/config 失败 %v", err)
		return err
	}

	return nil
}

func (s *ServerController) UpdateAgentStatus(ctx context.Context, req *types.UpdateAgentStatusRequest) error {
	if req.Status == "强制在线" || req.Status == "强制离线" {
		realStatus := strings.Replace(req.Status, "强制", "", -1)
		return s.factory.Agent().UpdateByName(ctx, req.AgentName, map[string]interface{}{"status": realStatus, "message": fmt.Sprintf("Agent has been set to %s", realStatus)})
	}

	old, err := s.factory.Agent().GetByName(ctx, req.AgentName)
	if err != nil {
		return err
	}
	// 如果是进行中状态，先返回等待
	if s.IsAgentRunningStatus(old.Status) {
		return fmt.Errorf("agent(%s)状态为(%s), 请稍后再试", req.AgentName, old.Status)
	}

	sshConfig, ok := s.sshConfigMap[old.RainbowdName]
	if !ok {
		return fmt.Errorf("未加载 rainbow node(%s)", old.RainbowdName)
	}

	if err := s.factory.Agent().UpdateByName(ctx, req.AgentName, map[string]interface{}{"status": req.Status, "message": fmt.Sprintf("Agent has been set to %s", req.Status)}); err != nil {
		return err
	}
	go func() {
		if err = s.ReconcileAgent(ctx, &sshConfig, old); err != nil {
			klog.Errorf("远程更新agent失败 %v", err)
		}
	}()
	return nil
}

func (s *ServerController) UpdateAgent(ctx context.Context, req *types.UpdateAgentRequest) error {
	repo := req.GithubRepository
	if len(repo) == 0 {
		repo = fmt.Sprintf("https://github.com/%s/plugin.git", req.GithubUser)
	}

	updates := make(map[string]interface{})
	updates["github_user"] = req.GithubUser
	updates["github_repository"] = repo
	updates["github_token"] = req.GithubToken
	updates["github_email"] = req.GithubEmail
	updates["rainbowd_name"] = req.RainbowdName
	return s.factory.Agent().UpdateByName(ctx, req.AgentName, updates)
}

func (s *ServerController) ListAgents(ctx context.Context, listOption types.ListOptions) (interface{}, error) {
	return s.factory.Agent().List(ctx, db.WithNameLike(listOption.NameSelector))
}

func (s *ServerController) GetAgentContainer(sshConfig *sshutil.SSHConfig, containerName string) (*ContainerInfo, error) {
	sshClient, err := sshutil.NewSSHClient(sshConfig)
	if err != nil {
		return nil, err
	}
	defer sshClient.Close()

	result, err := sshClient.RunCommand(fmt.Sprintf("docker ps -a --filter name=%s --format json", containerName))
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("docker 命令执行失败: %s", result.Stderr)
	}

	containers, err := s.parseDockerJSONOutput(result.Stdout)
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.Names == containerName {
			return &container, nil
		}
	}

	return nil, nil
}

func (s *ServerController) InstallAgentContainer() error {
	return nil
}

func (s *ServerController) RemoveAgentContainer(sshConfig *sshutil.SSHConfig, containerName string) error {
	return nil
}

func (s *ServerController) UninstallAgentContainer(sshConfig *sshutil.SSHConfig, agent *model.Agent) error {
	sshClient, err := sshutil.NewSSHClient(sshConfig)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	containerName := agent.Name
	destDir := filepath.Join(s.cfg.Rainbowd.DataDir, containerName)
	klog.V(1).Infof("agent(%s) 工作目录(%s) 正在被回收", containerName, destDir)
	cmds := []string{
		fmt.Sprintf("docker rm %s -f", containerName),
		fmt.Sprintf("rm -rf %s", destDir),
	}
	results, err := sshClient.RunCommands(cmds)
	if err != nil {
		return err
	}
	for _, result := range results {
		if result.ExitCode != 0 {
			return fmt.Errorf("docker 命令执行失败: %s", result.Stderr)
		}
	}

	return nil
}

func (s *ServerController) UpgradeAgentContainer() error {
	return nil
}

func (s *ServerController) RestartAgentContainer(sshConfig *sshutil.SSHConfig, agent *model.Agent) error {
	sshClient, err := sshutil.NewSSHClient(sshConfig)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	containerName := agent.Name
	result, err := sshClient.RunCommand(fmt.Sprintf("docker restart %s", containerName))
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("docker 命令执行失败: %s", result.Stderr)
	}

	return nil
}

type ContainerInfo struct {
	ID           string `json:"ID"`
	Names        string `json:"Names"`
	Image        string `json:"Image"`
	Command      string `json:"Command"`
	CreatedAt    string `json:"CreatedAt"`
	Status       string `json:"Status"`
	Ports        string `json:"Ports"`
	Size         string `json:"Size"`
	Labels       string `json:"Labels"`
	LocalVolumes string `json:"LocalVolumes"`
	Mounts       string `json:"Mounts"`
	Networks     string `json:"Networks"`
}

// parseDockerJSONOutput 解析Docker JSON格式输出
func (s *ServerController) parseDockerJSONOutput(output string) ([]ContainerInfo, error) {
	if strings.TrimSpace(output) == "" {
		return []ContainerInfo{}, nil
	}

	// Docker输出的JSON是每行一个JSON对象，不是JSON数组
	var containers []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var container ContainerInfo
		if err := json.Unmarshal([]byte(line), &container); err != nil {
			return nil, fmt.Errorf("解析JSON失败: %w, 原始数据: %s", err, line)
		}

		containers = append(containers, container)
	}

	return containers, nil
}
