package sshutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/klog/v2"
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
	if s.config.PrivateKey != "" {
		keyAuth, err := s.privateKeyAuthFromFile(s.config.PrivateKey)
		if err == nil {
			authMethods = append(authMethods, keyAuth)
		}
	}

	// 尝试获取私钥
	if len(authMethods) == 0 {
		keyAuth, err := s.tryDefaultPrivateKeys()
		if err == nil {
			authMethods = append(authMethods, keyAuth)
		}
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("必须提供密码或私钥")
	}

	sshConfig := &ssh.ClientConfig{
		User:            "root",
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

	klog.V(1).Infof("执行命令(%s)", cmd)
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
			return result, fmt.Errorf("执行命令失败: %v", err)
		}
	}

	return result, nil
}

func (s *SSHClient) RunCommands(commands []string) ([]*CommandResult, error) {
	var results []*CommandResult

	for _, cmd := range commands {
		result, err := s.RunCommand(cmd)
		if err != nil {
			klog.Errorf("执行命令(%s)失败: %v", cmd, err)
			return results, fmt.Errorf("执行命令(%s)失败: %v", cmd, err)
		}
		if result.ExitCode != 0 {
			return results, fmt.Errorf("执行命令(%s) exitCode %d", cmd, result.ExitCode)
		}

		results = append(results, result)
	}

	return results, nil
}

// UploadDir 上传文件夹到远程服务器
func (s *SSHClient) UploadDir(localDir, remoteDir string, ug string) error {
	var tarBuffer bytes.Buffer

	cmd := exec.Command("tar", "-czf", "-", "-C", localDir, ".")
	cmd.Stdout = &tarBuffer
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("打包目录失败: %w", err)
	}

	// 创建远程会话
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取输入管道失败: %w", err)
	}

	// 在远程解压
	remoteCmd := fmt.Sprintf("mkdir -p %s && tar -xzf - -C %s", remoteDir, remoteDir)
	if len(ug) != 0 {
		remoteCmd = remoteCmd + fmt.Sprintf(" && chown %s -R %s ", ug, remoteDir)
	}
	if err := session.Start(remoteCmd); err != nil {
		return fmt.Errorf("启动远程命令失败: %w", err)
	}
	// 发送tar数据
	if _, err := io.Copy(stdin, &tarBuffer); err != nil {
		return fmt.Errorf("发送压缩数据失败: %w", err)
	}

	stdin.Close()
	return session.Wait()
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

func (s *SSHClient) Ping() error {
	result, err := s.RunCommand("echo pong")
	if err != nil {
		return err
	}
	klog.V(1).Infof("ping 结果 %+v", result)

	return nil
}

// Close 关闭SSH连接
func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSHClient) tryDefaultPrivateKeys() (ssh.AuthMethod, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	defaultKeys := []string{
		filepath.Join(sshDir, "id_rsa"),
		filepath.Join(sshDir, "id_dsa"),
	}
	for _, keyPath := range defaultKeys {
		if _, err := os.Stat(keyPath); err == nil {
			auth, err := s.privateKeyAuthFromFile(keyPath)
			if err == nil {
				return auth, nil
			}
		}
	}

	return nil, fmt.Errorf("没有找到默认私钥文件")
}

func (s *SSHClient) privateKeyAuthFromFile(privateKeyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	return s.privateKeyAuthFromBytes(key)
}

func (s *SSHClient) privateKeyAuthFromBytes(key []byte) (ssh.AuthMethod, error) {
	// 首先尝试解析为普通私钥
	signer, err := ssh.ParsePrivateKey(key)
	if err == nil {
		return ssh.PublicKeys(signer), nil
	}

	// 如果失败，尝试解析为带密码的私钥
	if s.config.Password != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(s.config.Password))
		if err == nil {
			return ssh.PublicKeys(signer), nil
		}
	}

	return nil, fmt.Errorf("解析私钥失败: %w", err)
}
