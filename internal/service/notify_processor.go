package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/consts"
	"github.com/reusedev/uportal-api/pkg/logs"
	message "github.com/reusedev/uportal-api/pkg/notify"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NotifyProcessor 通知处理器
type NotifyProcessor struct {
	db          *gorm.DB
	mu          sync.RWMutex
	currentTask *model.BackendNotify
	stopping    bool
}

// NewNotifyProcessor 创建通知处理器实例
func NewNotifyProcessor(db *gorm.DB) *NotifyProcessor {
	return &NotifyProcessor{db: db}
}

// ProcessNotifications 处理通知任务
func (p *NotifyProcessor) ProcessNotifications(ctx context.Context) error {
	logs.Business().Info("开始处理通知任务")

	// 查找待发送或发送中的通知
	var notifications []model.BackendNotify
	err := p.db.Where("status IN (?)", []int8{consts.SendReady, consts.Sending}).
		Order("created_at ASC").Find(&notifications).Error
	if err != nil {
		return fmt.Errorf("查询通知记录失败: %w", err)
	}

	if len(notifications) == 0 {
		logs.Business().Info("没有待处理的通知")
		return nil
	}

	logs.Business().Info("找到待处理通知", zap.Int("count", len(notifications)))

	// 处理每个通知
	for _, notify := range notifications {
		if err := p.processNotification(ctx, &notify); err != nil {
			logs.Business().Error("处理通知失败",
				zap.Int64("notify_id", notify.ID),
				zap.String("template_id", notify.TemplateID),
				zap.Error(err))
			continue
		}
	}

	logs.Business().Info("通知处理完成")
	return nil
}

// processNotification 处理单个通知
func (p *NotifyProcessor) processNotification(ctx context.Context, notify *model.BackendNotify) error {
	// 设置当前任务
	p.mu.Lock()
	p.currentTask = notify
	p.mu.Unlock()

	// 更新状态为发送中
	if notify.Status == consts.SendReady {
		notify.Status = consts.Sending
		if err := p.db.Save(notify).Error; err != nil {
			return fmt.Errorf("更新通知状态失败: %w", err)
		}
	}

	// 解析消息数据
	var messageData map[string]message.Kv
	if err := json.Unmarshal([]byte(notify.Data), &messageData); err != nil {
		return fmt.Errorf("解析消息数据失败: %w", err)
	}

	// 循环处理所有用户，直到全部发送完成
	for {
		// 检查是否需要停止
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		p.mu.RLock()
		isStopping := p.stopping
		p.mu.RUnlock()

		if isStopping {
			logs.Business().Info("接收到停止信号，保存当前进度", zap.Int64("notify_id", notify.ID))
			return nil
		}

		// 查询符合条件的用户（批量查询100个）
		users, err := p.findTargetUsers(notify.TemplateID, notify.LastId)
		if err != nil {
			return fmt.Errorf("查询目标用户失败: %w", err)
		}

		if len(users) == 0 {
			// 没有更多用户需要发送，标记为完成
			notify.Status = consts.SendFinish
			if err := p.db.Save(notify).Error; err != nil {
				return fmt.Errorf("更新通知状态为完成失败: %w", err)
			}
			logs.Business().Info("通知发送完成", zap.Int64("notify_id", notify.ID))
			break
		}

		// 发送消息给每个用户
		var lastProcessedUserID string
		for _, user := range users {
			// 再次检查停止信号
			p.mu.RLock()
			isStopping := p.stopping
			p.mu.RUnlock()

			if isStopping {
				logs.Business().Info("接收到停止信号，中断发送循环", zap.String("last_processed_user", lastProcessedUserID))
				break
			}

			if err := p.sendMessageToUser(user, notify.TemplateID, notify.Page, notify.Title, notify.Type, messageData); err != nil {
				logs.Business().Error("发送消息失败",
					zap.String("user_id", user.UserID),
					zap.String("template_id", notify.TemplateID),
					zap.Error(err))
				// 继续处理其他用户，不中断
			}
			lastProcessedUserID = user.UserID

			// 实时更新last_id
			notify.LastId = lastProcessedUserID
			if err := p.db.Save(notify).Error; err != nil {
				logs.Business().Error("更新last_id失败", zap.Error(err))
			}
		}

		// 如果收到停止信号，退出循环
		p.mu.RLock()
		isStopping = p.stopping
		p.mu.RUnlock()

		if isStopping {
			break
		}

		logs.Business().Info("本批次通知发送完成",
			zap.Int64("notify_id", notify.ID),
			zap.Int("user_count", len(users)),
			zap.String("last_id", lastProcessedUserID))
	}

	// 清除当前任务
	p.mu.Lock()
	p.currentTask = nil
	p.mu.Unlock()

	return nil
}

// findTargetUsers 查找目标用户
func (p *NotifyProcessor) findTargetUsers(templateID, lastID string) ([]model.MessageSubscribe, error) {
	query := p.db.Where("template_id = ? AND subscribe_cnt > 0", templateID).
		Order("user_id ASC").
		Limit(100) // 每次处理100个用户

	// 如果有last_id，从last_id之后开始查询
	if lastID != "" {
		query = query.Where("user_id > ?", lastID)
	}

	var users []model.MessageSubscribe
	err := query.Find(&users).Error
	return users, err
}

// sendMessageToUser 发送消息给用户
func (p *NotifyProcessor) sendMessageToUser(user model.MessageSubscribe, templateID, page, title, msgType string, data map[string]message.Kv) error {
	// 获取用户的openid
	var userAuth model.UserAuth
	err := p.db.Where("user_id = ?", user.UserID).First(&userAuth).Error
	if err != nil {
		return fmt.Errorf("获取用户认证信息失败: %w", err)
	}

	// 构造消息数据
	messageData := message.MessageData{
		Touser:     userAuth.ProviderUserID,
		TemplateId: templateID,
		Page:       page,
		Data:       data,
	}

	messageJSON, err := json.Marshal(messageData)
	if err != nil {
		return fmt.Errorf("序列化消息数据失败: %w", err)
	}

	// 发送消息
	token := wechat_token.GetToken()
	err = message.SendMessage(token, string(messageJSON))
	if err != nil {
		logs.Business().Error("发送订阅消息失败",
			zap.String("user_id", user.UserID),
			zap.String("openid", userAuth.ProviderUserID),
			zap.String("template_id", templateID),
			zap.Error(err))

		// 创建失败通知记录
		p.createNotificationRecord(user.UserID, msgType, title, string(messageJSON), 1)
		return err
	}

	// 减少用户的订阅次数
	err = p.db.Model(&model.MessageSubscribe{}).
		Where("user_id = ? AND template_id = ?", user.UserID, templateID).
		Update("subscribe_cnt", gorm.Expr("subscribe_cnt - 1")).Error
	if err != nil {
		logs.Business().Error("更新用户订阅次数失败",
			zap.String("user_id", user.UserID),
			zap.String("template_id", templateID),
			zap.Error(err))
	}

	// 创建成功通知记录
	p.createNotificationRecord(user.UserID, msgType, title, string(messageJSON), 0)

	logs.Business().Info("消息发送成功",
		zap.String("user_id", user.UserID),
		zap.String("template_id", templateID))

	return nil
}

// createNotificationRecord 创建通知记录
func (p *NotifyProcessor) createNotificationRecord(userID, msgType, title, content string, status int8) {
	notification := model.Notification{
		UserID:  userID,
		Type:    msgType,
		Title:   title,
		Content: content,
		Status:  status,
	}

	if err := p.db.Create(&notification).Error; err != nil {
		logs.Business().Error("创建通知记录失败",
			zap.String("user_id", userID),
			zap.Error(err))
	}
}

// Stop 停止处理器并保存当前进度
func (p *NotifyProcessor) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopping = true

	// 如果有正在处理的任务，保存当前进度
	if p.currentTask != nil {
		if err := p.db.Save(p.currentTask).Error; err != nil {
			logs.Business().Error("保存当前任务进度失败",
				zap.Int64("notify_id", p.currentTask.ID),
				zap.String("last_id", p.currentTask.LastId),
				zap.Error(err))
		} else {
			logs.Business().Info("已保存当前任务进度",
				zap.Int64("notify_id", p.currentTask.ID),
				zap.String("last_id", p.currentTask.LastId))
		}
	}
}
