package types

// SystemConfig 系统配置表结构体
type SystemConfig struct {
	ConfigKey   string `gorm:"column:config_key;primaryKey" json:"config_key"` // 配置键，主键
	ConfigValue string `gorm:"column:config_value" json:"config_value"`        // 配置值
	Description string `gorm:"column:description" json:"description"`          // 配置描述
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_config"
}
