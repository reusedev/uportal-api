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
		Port         int           `yaml:"port"`         // 服务器端口
		Mode         string        `yaml:"mode"`         // 运行模式：debug/release
		ReadTimeout  time.Duration `yaml:"readTimeout"`  // 读取超时时间
		WriteTimeout time.Duration `yaml:"writeTimeout"` // 写入超时时间
	} `yaml:"server"`

	Database struct {
		Driver   string `yaml:"driver"`   // 数据库驱动：mysql
		Host     string `yaml:"host"`     // 数据库主机
		Port     int    `yaml:"port"`     // 数据库端口
		Username string `yaml:"username"` // 数据库用户名
		Password string `yaml:"password"` // 数据库密码
		Database string `yaml:"database"` // 数据库名称
		Charset  string `yaml:"charset"`  // 字符集
		MaxIdle  int    `yaml:"maxIdle"`  // 最大空闲连接数
		MaxOpen  int    `yaml:"maxOpen"`  // 最大打开连接数
	} `yaml:"database"`

	Logging struct {
		LogDir          string `yaml:"logDir"`          // 日志目录
		BusinessLogFile string `yaml:"businessLogFile"` // 业务日志文件名
		DBLogFile       string `yaml:"dbLogFile"`       // 数据库日志文件名
		Level           string `yaml:"level"`           // 日志级别：debug/info/warn/error
		Console         bool   `yaml:"console"`         // 是否输出到控制台
		MaxSize         int    `yaml:"maxSize"`         // 单个日志文件最大尺寸(MB)
		MaxBackups      int    `yaml:"maxBackups"`      // 保留的旧日志文件最大数量
		MaxAge          int    `yaml:"maxAge"`          // 保留的旧日志文件最大天数
		Compress        bool   `yaml:"compress"`        // 是否压缩旧日志文件
	} `yaml:"logging"`

	JWT struct {
		Secret     string        `yaml:"secret"`     // JWT密钥
		ExpireTime time.Duration `yaml:"expireTime"` // JWT过期时间
		Issuer     string        `yaml:"issuer"`     // JWT签发者
	} `yaml:"jwt"`

	Redis struct {
		Host     string `yaml:"host"`     // Redis主机
		Port     int    `yaml:"port"`     // Redis端口
		Password string `yaml:"password"` // Redis密码
		DB       int    `yaml:"db"`       // Redis数据库编号
		PoolSize int    `yaml:"poolSize"` // Redis连接池大小
	} `yaml:"redis"`

	Wechat struct {
		MiniProgram struct {
			AppID     string `yaml:"appId"`     // 小程序AppID
			AppSecret string `yaml:"appSecret"` // 小程序AppSecret
		} `yaml:"miniProgram"`
		Pay struct {
			AppID      string `yaml:"appId"`      // 支付AppID
			MchID      string `yaml:"mchId"`      // 商户号
			MchApiKey  string `yaml:"mchApiKey"`  // 商户API密钥
			NotifyUrl  string `yaml:"notifyUrl"`  // 支付回调通知地址
			CertFile   string `yaml:"certFile"`   // 证书文件路径
			KeyFile    string `yaml:"keyFile"`    // 密钥文件路径
			RootCaFile string `yaml:"rootCaFile"` // 根证书文件路径
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

	// 设置默认值
	setDefaults(config)

	// 验证配置
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("validate config error: %v", err)
	}

	// 设置全局配置
	GlobalConfig = config
	return nil
}

// setDefaults 设置配置默认值
func setDefaults(config *Config) {
	// Server 默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}

	// Database 默认值
	if config.Database.Driver == "" {
		config.Database.Driver = "mysql"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 3306
	}
	if config.Database.Charset == "" {
		config.Database.Charset = "utf8mb4"
	}
	if config.Database.MaxIdle == 0 {
		config.Database.MaxIdle = 10
	}
	if config.Database.MaxOpen == 0 {
		config.Database.MaxOpen = 100
	}

	// Logging 默认值
	if config.Logging.LogDir == "" {
		config.Logging.LogDir = "logs"
	}
	if config.Logging.BusinessLogFile == "" {
		config.Logging.BusinessLogFile = "business.log"
	}
	if config.Logging.DBLogFile == "" {
		config.Logging.DBLogFile = "db.log"
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.MaxSize == 0 {
		config.Logging.MaxSize = 100
	}
	if config.Logging.MaxBackups == 0 {
		config.Logging.MaxBackups = 10
	}
	if config.Logging.MaxAge == 0 {
		config.Logging.MaxAge = 30
	}

	// Redis 默认值
	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}
	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}

	// JWT 默认值
	if config.JWT.ExpireTime == 0 {
		config.JWT.ExpireTime = 24 * time.Hour
	}
	if config.JWT.Issuer == "" {
		config.JWT.Issuer = "uportal-api"
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	if config.Server.Mode != "debug" && config.Server.Mode != "release" {
		return fmt.Errorf("invalid server mode: %s", config.Server.Mode)
	}
	if config.Server.ReadTimeout <= 0 {
		return fmt.Errorf("invalid read timeout: %v", config.Server.ReadTimeout)
	}
	if config.Server.WriteTimeout <= 0 {
		return fmt.Errorf("invalid write timeout: %v", config.Server.WriteTimeout)
	}

	// 验证数据库配置
	if config.Database.Driver != "mysql" {
		return fmt.Errorf("unsupported database driver: %s", config.Database.Driver)
	}
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Port <= 0 || config.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Database.Port)
	}
	if config.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if config.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Database.MaxIdle < 0 {
		return fmt.Errorf("invalid max idle connections: %d", config.Database.MaxIdle)
	}
	if config.Database.MaxOpen <= 0 {
		return fmt.Errorf("invalid max open connections: %d", config.Database.MaxOpen)
	}

	// 验证日志配置
	if config.Logging.Level != "debug" && config.Logging.Level != "info" &&
		config.Logging.Level != "warn" && config.Logging.Level != "error" {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}
	if config.Logging.MaxSize <= 0 {
		return fmt.Errorf("invalid max log size: %d", config.Logging.MaxSize)
	}
	if config.Logging.MaxBackups <= 0 {
		return fmt.Errorf("invalid max backups: %d", config.Logging.MaxBackups)
	}
	if config.Logging.MaxAge <= 0 {
		return fmt.Errorf("invalid max age: %d", config.Logging.MaxAge)
	}

	// 验证JWT配置
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if config.JWT.ExpireTime <= 0 {
		return fmt.Errorf("invalid JWT expire time: %v", config.JWT.ExpireTime)
	}

	// 验证Redis配置
	if config.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", config.Redis.Port)
	}
	if config.Redis.PoolSize <= 0 {
		return fmt.Errorf("invalid redis pool size: %d", config.Redis.PoolSize)
	}

	// 验证微信支付配置
	if config.Wechat.Pay.AppID == "" {
		return fmt.Errorf("wechat pay app id is required")
	}
	if config.Wechat.Pay.MchID == "" {
		return fmt.Errorf("wechat pay merchant id is required")
	}
	if config.Wechat.Pay.MchApiKey == "" {
		return fmt.Errorf("wechat pay merchant api key is required")
	}
	if config.Wechat.Pay.NotifyUrl == "" {
		return fmt.Errorf("wechat pay notify url is required")
	}
	if config.Wechat.Pay.CertFile == "" {
		return fmt.Errorf("wechat pay cert file is required")
	}
	if config.Wechat.Pay.KeyFile == "" {
		return fmt.Errorf("wechat pay key file is required")
	}

	return nil
}

// Get 获取配置实例
func Get() *Config {
	if GlobalConfig == nil {
		panic("config not initialized")
	}
	return GlobalConfig
}
