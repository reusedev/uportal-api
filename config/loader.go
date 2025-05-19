package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	// 如果未指定配置文件路径，使用默认路径
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// 确保配置文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config file error: %v", err)
	}

	// 解析配置文件
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("parse config file error: %v", err)
	}

	// 验证配置
	if err := config.validate(); err != nil {
		return fmt.Errorf("validate config error: %v", err)
	}

	// 设置全局配置
	GlobalConfig = config
	return nil
}

// validate 验证配置
func (c *Config) validate() error {
	// 验证服务器配置
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// 验证数据库配置
	if c.Database.Driver != "mysql" {
		return fmt.Errorf("unsupported database driver: %s", c.Database.Driver)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port <= 0 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// 验证Redis配置
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port <= 0 {
		return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
	}

	// 验证微信小程序配置
	if c.Wechat.MiniProgram.AppID == "" {
		return fmt.Errorf("wechat mini program appid is required")
	}
	if c.Wechat.MiniProgram.AppSecret == "" {
		return fmt.Errorf("wechat mini program app secret is required")
	}

	// 验证微信支付配置
	if c.Wechat.Pay.AppID == "" {
		return fmt.Errorf("wechat pay appid is required")
	}
	if c.Wechat.Pay.MchID == "" {
		return fmt.Errorf("wechat pay merchant id is required")
	}
	if c.Wechat.Pay.MchAPIKey == "" {
		return fmt.Errorf("wechat pay merchant api key is required")
	}
	if c.Wechat.Pay.NotifyURL == "" {
		return fmt.Errorf("wechat pay notify url is required")
	}

	// 验证证书文件
	certDir := filepath.Dir(c.Wechat.Pay.CertFile)
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		return fmt.Errorf("certificate directory not found: %s", certDir)
	}

	// 验证JWT配置
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if c.JWT.ExpireTime <= 0 {
		return fmt.Errorf("invalid jwt expire time: %v", c.JWT.ExpireTime)
	}

	return nil
}
