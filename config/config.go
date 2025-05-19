package config

import (
	"time"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Wechat   WechatConfig   `yaml:"wechat"`
	JWT      JWTConfig      `yaml:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	Mode         string        `yaml:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `yaml:"driver"` // mysql
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
	MaxIdle  int    `yaml:"maxIdle"` // 最大空闲连接数
	MaxOpen  int    `yaml:"maxOpen"` // 最大打开连接数
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"` // 连接池大小
}

// WechatConfig 微信配置
type WechatConfig struct {
	MiniProgram struct {
		AppID     string `yaml:"appId"`
		AppSecret string `yaml:"appSecret"`
	} `yaml:"miniProgram"`
	Pay struct {
		AppID      string `yaml:"appId"`
		MchID      string `yaml:"mchId"`
		MchAPIKey  string `yaml:"mchApiKey"`
		NotifyURL  string `yaml:"notifyUrl"`
		CertFile   string `yaml:"certFile"`
		KeyFile    string `yaml:"keyFile"`
		RootCAFile string `yaml:"rootCaFile"`
	} `yaml:"pay"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	ExpireTime time.Duration `yaml:"expireTime"`
	Issuer     string        `yaml:"issuer"`
}
