package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InviteService 邀请服务
type InviteService struct {
	db *gorm.DB
}

// NewInviteService 创建邀请服务
func NewInviteService(db *gorm.DB) *InviteService {
	return &InviteService{
		db: db,
	}
}

// GenerateInviteCode 生成邀请码
func (s *InviteService) GenerateInviteCode(userID int64) string {
	// 使用用户ID和时间戳生成邀请码
	data := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:8]) // 取前8位作为邀请码
}

// GetInviteLink 获取邀请链接
func (s *InviteService) GetInviteLink(ctx context.Context, userID int64) (string, error) {
	// 检查用户是否存在
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New(errors.ErrCodeUserNotFound, "用户不存在", nil)
		}
		return "", errors.New(errors.ErrCodeInternal, "获取用户信息失败", err)
	}

	// 生成邀请码
	inviteCode := s.GenerateInviteCode(userID)

	// TODO: 从配置中获取域名
	domain := "https://your-domain.com"
	return fmt.Sprintf("%s/register?invite_code=%s", domain, inviteCode), nil
}

// ValidateInviteCode 验证邀请码
func (s *InviteService) ValidateInviteCode(ctx context.Context, inviteCode string) (int64, error) {
	// 从邀请码中提取用户ID（这里需要根据实际生成邀请码的逻辑来实现）
	// 示例实现：假设邀请码的前8位是用户ID的哈希
	userID, err := s.extractUserIDFromInviteCode(inviteCode)
	if err != nil {
		return 0, errors.New(errors.ErrCodeInvalidParams, "无效的邀请码", err)
	}

	// 检查邀请人是否存在且状态正常
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, errors.New(errors.ErrCodeUserNotFound, "邀请人不存在", nil)
		}
		return 0, errors.New(errors.ErrCodeInternal, "验证邀请码失败", err)
	}

	if user.Status != 1 {
		return 0, errors.New(errors.ErrCodeUserDisabled, "邀请人账号已被禁用", nil)
	}

	return userID, nil
}

// extractUserIDFromInviteCode 从邀请码中提取用户ID
func (s *InviteService) extractUserIDFromInviteCode(inviteCode string) (int64, error) {
	// TODO: 实现从邀请码中提取用户ID的逻辑
	// 这里需要根据实际生成邀请码的逻辑来实现
	// 示例实现：假设邀请码的前8位是用户ID的哈希
	return 0, fmt.Errorf("not implemented")
}

// GetDB 获取数据库连接
func (s *InviteService) GetDB() *gorm.DB {
	return s.db
}

// CreateInviteRecord 创建邀请记录
func (s *InviteService) CreateInviteRecord(ctx context.Context, inviterID, inviteeID int64, tokenReward int) error {
	// 检查是否已经存在邀请记录
	var count int64
	if err := s.db.Model(&model.InviteRecord{}).
		Where("invitee_id = ?", inviteeID).
		Count(&count).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "检查邀请记录失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrCodeInvalidParams, "该用户已被邀请", nil)
	}

	// 创建邀请记录
	record := &model.InviteRecord{
		InviterID:   inviterID,
		InviteeID:   inviteeID,
		TokenReward: tokenReward,
		Status:      0, // 待发放
	}

	if err := s.db.Create(record).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "创建邀请记录失败", err)
	}

	return nil
}

// CreateInviteRecordWithTx 在事务中创建邀请记录
func (s *InviteService) CreateInviteRecordWithTx(ctx context.Context, tx *gorm.DB, inviterID, inviteeID int64, tokenReward int) error {
	// 检查是否已经存在邀请记录
	var count int64
	if err := tx.Model(&model.InviteRecord{}).
		Where("invitee_id = ?", inviteeID).
		Count(&count).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "检查邀请记录失败", err)
	}
	if count > 0 {
		return errors.New(errors.ErrCodeInvalidParams, "该用户已被邀请", nil)
	}

	// 创建邀请记录
	record := &model.InviteRecord{
		InviterID:   inviterID,
		InviteeID:   inviteeID,
		TokenReward: tokenReward,
		Status:      0, // 待发放
	}

	if err := tx.Create(record).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "创建邀请记录失败", err)
	}

	return nil
}

// ProcessInviteRewardWithTx - 在事务中处理邀请奖励
func (s *InviteService) ProcessInviteRewardWithTx(ctx context.Context, tx *gorm.DB, inviteeID int64) error {
	// 查找待处理的邀请记录
	var record model.InviteRecord
	if err := tx.Where("invitee_id = ? AND status = 0", inviteeID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return errors.New(errors.ErrCodeInternal, "查询邀请记录失败", err)
	}

	// 获取邀请人信息并加行锁
	var inviter model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&inviter, record.InviterID).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "获取邀请人信息失败", err)
	}

	// 使用 UpdateUserTokenBalance 更新邀请人代币余额
	if err := model.UpdateUserTokenBalance(tx, record.InviterID, record.TokenReward); err != nil {
		return errors.New(errors.ErrCodeInternal, "更新邀请人代币余额失败", err)
	}

	// 创建代币记录
	remark := "邀请奖励"
	tokenRecord := &model.TokenRecord{
		UserID:       record.InviterID,
		ChangeAmount: record.TokenReward,
		BalanceAfter: inviter.TokenBalance + record.TokenReward,
		ChangeType:   "INVITE_REWARD",
		Remark:       &remark,
		ChangeTime:   time.Now(),
	}
	if err := model.CreateTokenRecord(tx, tokenRecord); err != nil {
		return errors.New(errors.ErrCodeInternal, "创建代币记录失败", err)
	}

	// 更新邀请记录状态
	record.Status = 1
	if err := tx.Model(&record).Update("status", 1).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "更新邀请记录状态失败", err)
	}

	return nil
}

// 定义奖励类型常量
const (
	RewardTypeDailyLogin = "DAILY_LOGIN" // 每日登录奖励
	RewardTypeProfile    = "PROFILE"     // 完善资料奖励
	RewardTypeShare      = "SHARE"       // 分享奖励
	RewardTypeFeedback   = "FEEDBACK"    // 反馈奖励
	RewardAd             = "rewardad"    // 激励视频奖励
)

// 定义奖励类型对应的代币数
var rewardAmounts = map[string]int{
	RewardTypeDailyLogin: 10,  // 每日登录奖励10代币
	RewardTypeProfile:    50,  // 完善资料奖励50代币
	RewardTypeShare:      20,  // 分享奖励20代币
	RewardTypeFeedback:   30,  // 反馈奖励30代币
	RewardAd:             100, // 激励视频奖励100代币
}
