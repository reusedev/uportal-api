package types

// TokenConsumeRule 代币消耗功能表结构体
type TokenConsumeRule struct {
	FeatureID   int    `gorm:"column:feature_id;primaryKey;autoIncrement" json:"feature_id"` // 功能ID，主键，自增
	FeatureName string `gorm:"column:feature_name" json:"feature_name"`                      // 功能名称
	FeatureDesc string `gorm:"column:feature_desc" json:"feature_desc"`                      // 功能描述
	TokenCost   int    `gorm:"column:token_cost" json:"token_cost"`                          // 使用一次该功能消耗的代币数
	FeatureCode string `gorm:"column:feature_code" json:"feature_code"`                      // 功能代码
	Status      int8   `gorm:"column:status;default:1" json:"status"`                        // 功能状态：1=启用，0=停用
}

// TableName 指定表名
func (TokenConsumeRule) TableName() string {
	return "token_consume_rules"
}
