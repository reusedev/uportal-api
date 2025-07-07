package service

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/consts"
	"github.com/reusedev/uportal-api/pkg/logs"
	message "github.com/reusedev/uportal-api/pkg/notify"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"gorm.io/gorm"
)

// NotifyService 消息通知服务
type NotifyService struct {
	db *gorm.DB
}

// NewNotifyService 创建消息通知服务实例
func NewNotifyService(db *gorm.DB) *NotifyService {
	return &NotifyService{db: db}
}

func newData(openId, templateId, page string, msg map[string]message.Kv) string {
	data := message.MessageData{
		Touser:     openId,
		Page:       page,
		TemplateId: templateId,
		Data:       msg,
	}
	d, _ := json.Marshal(data)
	logs.Business().Info(string(d))
	return string(d)
}

func (notifyService *NotifyService) Notify(ctx context.Context, req *SubscribeReq, userId string) error {
	key := req.Id
	if req.AccessKey == consts.Accept {
		_, err := model.RedisClient.Set(ctx, key, userId, time.Hour*1).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *NotifyService) Send(ctx context.Context, req *SendReq) error {
	key := req.Sign
	_, err := model.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if stdErrors.Is(err, redis.Nil) {
			return nil
		} else {
			return err
		}
	}

	var userAuth model.UserAuth
	err = n.db.Where("user_id = ?", req.UserId).First(&userAuth).Error
	if err != nil {
		return err
	}
	data := newData(userAuth.ProviderUserID, req.TemplateId, req.Page, req.Data)
	t := wechat_token.GetToken()
	for i := 0; i < 3; i++ {
		err = message.SendMessage(t, data)
		if err == nil {
			break
		} else {
			logs.Business().Error(fmt.Sprintf("send message failed %d : err: %s, data: %s, token: %s", i, err, data, t))
			time.Sleep(time.Second * 2 * time.Duration(i+1))
		}
	}
	if err != nil {
		return err
	}
	model.RedisClient.Del(ctx, key)
	notification := model.Notification{
		UserID:    req.UserId,
		Type:      req.Type,
		Title:     req.Title,
		Content:   data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	n.db.Create(&notification)
	return nil
}

type SubscribeReq struct {
	Id        string            `json:"id" binding:"required"`
	Message   map[string]string `json:"message" binding:"required"`
	AccessKey string
}

type Message struct {
	DrawTask string `json:"RqmsNBiXK9bClJ83Z0PkCguj1-wmetG6uGRz1qipf5w" binding:"required"`
}

type SendReq struct {
	UserId     string                `json:"user_id" binding:"required"`
	Sign       string                `json:"sign" binding:"required"` // 订阅标识ID
	Data       map[string]message.Kv `json:"data" binding:"required"`
	Page       string                `json:"page" binding:"required"`
	TemplateId string                `json:"template_id" binding:"required"`
	Type       string                `json:"type" binding:"required"`
	Title      string                `json:"title" binding:"required"`
}
