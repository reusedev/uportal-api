package config

// Config 应用配置
type Config struct {
	// 通知配置
	Notification NotificationConfig
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Enabled bool `json:"enabled"` // 是否启用通知
}
