package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	// GlobalConfig 全局配置实例
	GlobalConfig *Config
)

// Config 应用配置
type Config struct {
	Server struct {
		Port         int           `yaml:"port"`
		Mode         string        `yaml:"mode"`
		ReadTimeout  time.Duration `yaml:"readTimeout"`
		WriteTimeout time.Duration `yaml:"writeTimeout"`
	} `yaml:"server"`

	Database struct {
		Driver   string `yaml:"driver"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
		Charset  string `yaml:"charset"`
		MaxIdle  int    `yaml:"maxIdle"`
		MaxOpen  int    `yaml:"maxOpen"`
	} `yaml:"database"`

	Logging struct {
		LogDir          string `yaml:"logDir"`
		BusinessLogFile string `yaml:"businessLogFile"`
		DBLogFile       string `yaml:"dbLogFile"`
		Level           string `yaml:"level"`
		Console         bool   `yaml:"console"`
		MaxSize         int    `yaml:"maxSize"`
		MaxBackups      int    `yaml:"maxBackups"`
		MaxAge          int    `yaml:"maxAge"`
		Compress        bool   `yaml:"compress"`
	} `yaml:"logging"`

	JWT struct {
		Secret     string        `yaml:"secret"`
		ExpireTime time.Duration `yaml:"expireTime"`
		Issuer     string        `yaml:"issuer"`
	} `yaml:"jwt"`

	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		PoolSize int    `yaml:"poolSize"`
	} `yaml:"redis"`

	Wechat struct {
		MiniProgram struct {
			AppID     string `yaml:"appId"`
			AppSecret string `yaml:"appSecret"`
		} `yaml:"miniProgram"`
		Pay struct {
			AppID      string `yaml:"appId"`
			MchID      string `yaml:"mchId"`
			MchApiKey  string `yaml:"mchApiKey"`
			NotifyUrl  string `yaml:"notifyUrl"`
			CertFile   string `yaml:"certFile"`
			KeyFile    string `yaml:"keyFile"`
			RootCaFile string `yaml:"rootCaFile"`
		} `yaml:"pay"`
	} `yaml:"wechat"`
}

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
	if c.Server.Mode != "debug" && c.Server.Mode != "release" {
		return fmt.Errorf("invalid server mode: %s", c.Server.Mode)
	}

	// 验证数据库配置
	if c.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
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

	// 验证日志配置
	if c.Logging.LogDir == "" {
		return fmt.Errorf("log directory is required")
	}
	if c.Logging.BusinessLogFile == "" {
		return fmt.Errorf("business log file is required")
	}
	if c.Logging.DBLogFile == "" {
		return fmt.Errorf("database log file is required")
	}
	if c.Logging.Level == "" {
		return fmt.Errorf("log level is required")
	}
	if c.Logging.MaxSize <= 0 {
		return fmt.Errorf("invalid max log size: %d", c.Logging.MaxSize)
	}
	if c.Logging.MaxBackups <= 0 {
		return fmt.Errorf("invalid max backups: %d", c.Logging.MaxBackups)
	}
	if c.Logging.MaxAge <= 0 {
		return fmt.Errorf("invalid max age: %d", c.Logging.MaxAge)
	}

	// 验证JWT配置
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if c.JWT.ExpireTime <= 0 {
		return fmt.Errorf("invalid JWT expire time: %v", c.JWT.ExpireTime)
	}

	// 验证Redis配置
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port <= 0 {
		return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
	}
	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("invalid redis pool size: %d", c.Redis.PoolSize)
	}

	// 验证微信配置
	if c.Wechat.Pay.AppID == "" {
		return fmt.Errorf("wechat pay app id is required")
	}
	if c.Wechat.Pay.MchID == "" {
		return fmt.Errorf("wechat pay merchant id is required")
	}
	if c.Wechat.Pay.MchApiKey == "" {
		return fmt.Errorf("wechat pay merchant api key is required")
	}
	if c.Wechat.Pay.NotifyUrl == "" {
		return fmt.Errorf("wechat pay notify url is required")
	}
	if c.Wechat.Pay.CertFile == "" {
		return fmt.Errorf("wechat pay cert file is required")
	}
	if c.Wechat.Pay.KeyFile == "" {
		return fmt.Errorf("wechat pay key file is required")
	}

	return nil
}
