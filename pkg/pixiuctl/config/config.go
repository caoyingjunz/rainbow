package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Default *DefaultConfig `yaml:"default"`
	Auth    *AuthConfig    `yaml:"auth"`
}

type AuthConfig struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type DefaultConfig struct {
	URL     string `yaml:"url,omitempty"`
	Timeout int    `yaml:"timeout"` // 单位是分钟
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// 确保文件存在时被覆盖
	if err := ioutil.WriteFile(path, data, 0o644); err != nil {
		return err
	}

	return nil
}
