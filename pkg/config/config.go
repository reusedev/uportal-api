package config

import (
	"sync"

	"github.com/spf13/viper"
)

var (
	config *Config
	once   sync.Once
)

// Config 应用配置
type Config struct {
	Server struct {
		Port         int    `mapstructure:"port"`
		Mode         string `mapstructure:"mode"`
		ReadTimeout  string `mapstructure:"readTimeout"`
		WriteTimeout string `mapstructure:"writeTimeout"`
	} `mapstructure:"server"`

	Database struct {
		Driver   string `mapstructure:"driver"`
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
		Charset  string `mapstructure:"charset"`
		MaxIdle  int    `mapstructure:"maxIdle"`
		MaxOpen  int    `mapstructure:"maxOpen"`
	} `mapstructure:"database"`

	JWT struct {
		Secret     string `mapstructure:"secret"`
		ExpireTime string `mapstructure:"expireTime"`
		Issuer     string `mapstructure:"issuer"`
	} `mapstructure:"jwt"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
		PoolSize int    `mapstructure:"poolSize"`
	} `mapstructure:"redis"`

	Wechat struct {
		MiniProgram struct {
			AppID     string `mapstructure:"appId"`
			AppSecret string `mapstructure:"appSecret"`
		} `mapstructure:"miniProgram"`
		Pay struct {
			AppID      string `mapstructure:"appId"`
			MchID      string `mapstructure:"mchId"`
			MchApiKey  string `mapstructure:"mchApiKey"`
			NotifyUrl  string `mapstructure:"notifyUrl"`
			CertFile   string `mapstructure:"certFile"`
			KeyFile    string `mapstructure:"keyFile"`
			RootCaFile string `mapstructure:"rootCaFile"`
		} `mapstructure:"pay"`
	} `mapstructure:"wechat"`
}

// Get 获取配置实例
func Get() *Config {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}

		config = &Config{}
		if err := viper.Unmarshal(config); err != nil {
			panic(err)
		}
	})
	return config
}
