package sshutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host       string        // 主机地址
	Port       int           // 端口
	Username   string        // 用户名
	Password   string        // 密码
	PrivateKey string        // 私钥路径（如果使用密钥认证）
	Timeout    time.Duration // 连接超时时间
}

type SSHClient struct {
	config *SSHConfig
	client *ssh.Client
}

type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

func NewSSHClient(config *SSHConfig) (*SSHClient, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Port == 0 {
		config.Port = 22
	}

	client := &SSHClient{config: config}
	err := client.connect()
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}

	return client, nil
}

func (s *SSHClient) connect() error {
	var authMethods []ssh.AuthMethod
	if s.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.config.Password))
	}
	if len(authMethods) == 0 {
		return fmt.Errorf("必须提供密码或私钥")
	}

	sshConfig := &ssh.ClientConfig{
		User:            s.config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         s.config.Timeout,
	}
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("连接SSH服务器失败: %w", err)
	}

	s.client = client
	return nil
}

// RunCommand 执行单个命令
func (s *SSHClient) RunCommand(cmd string) (*CommandResult, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SSH连接未建立")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	err = session.Run(cmd)
	result := &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		exitErr, ok := err.(*ssh.ExitError)
		if ok {
			result.ExitCode = exitErr.ExitStatus()
		} else {
			result.Error = err
			return result, fmt.Errorf("执行命令失败: %w", err)
		}
	}

	return result, nil
}

func (s *SSHClient) RunCommands(commands []string) ([]*CommandResult, error) {
	var results []*CommandResult

	for _, cmd := range commands {
		result, err := s.RunCommand(cmd)
		if err != nil {
			return results, fmt.Errorf("执行命令'%s'失败: %w", cmd, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// UploadFile 上传文件到远程服务器
func (s *SSHClient) UploadFile(localPath, remotePath string, mode string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}
	// 使用scp命令接收文件
	cmd := fmt.Sprintf("scp -t %s", remotePath)
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("启动SCP失败: %w", err)
	}

	// 发送文件
	fmt.Fprintf(w, "C%s %d %s\n", mode, stat.Size(), stat.Name())
	io.Copy(w, file)
	fmt.Fprint(w, "\x00")
	w.Close()

	return session.Wait()
}

// Close 关闭SSH连接
func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Execute 简单执行接口（一步完成连接、执行、关闭）
func Execute(config *SSHConfig, command string) (*CommandResult, error) {
	client, err := NewSSHClient(config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return client.RunCommand(command)
}
