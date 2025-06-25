package service

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"github.com/go-redis/redis/v8"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/consts"
	message "github.com/reusedev/uportal-api/pkg/notify"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"gorm.io/gorm"
	"time"
)

// NotifyService 消息通知服务
type NotifyService struct {
	db *gorm.DB
}

// NewNotifyService 创建消息通知服务实例
func NewNotifyService(db *gorm.DB) *NotifyService {
	return &NotifyService{db: db}
}

func newDrawTask(req *SendReq) map[string]message.Kv {
	remark := "制作已完成，请点击前往查看"
	if req.Remarks != "" {
		remark = req.Remarks
	}
	return map[string]message.Kv{
		"thing1":  {Value: req.Theme},
		"thing2":  {Value: req.Style},
		"phrase3": {Value: req.Result},
		"time4":   {Value: req.CreatedTime},
		"thing5":  {Value: remark},
	}
}

func newData(openId, templateId string, msg map[string]message.Kv) string {
	data := message.MessageData{
		Touser:     openId,
		TemplateId: templateId,
		Data:       msg,
	}
	d, _ := json.Marshal(data)
	return string(d)
}

func (notifyService *NotifyService) Notify(ctx context.Context, req *SubscribeReq, userId string) error {
	key := req.Id
	if req.Message.DrawTask == consts.Accept {
		_, err := model.RedisClient.Set(ctx, key, userId, time.Hour*1).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *NotifyService) Send(ctx context.Context, req *SendReq) error {
	key := req.Id
	userId, err := model.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if stdErrors.Is(err, redis.Nil) {
			return nil
		} else {
			return err
		}
	}
	var userAuth model.UserAuth
	err = n.db.Where("user_id = ?", userId).First(&userAuth).Error
	if err != nil {
		return err
	}
	msg := newDrawTask(req)
	data := newData(userAuth.ProviderUserID, consts.CompleteNotificationTmpId, msg)
	t := wechat_token.GetToken()
	for i := 0; i < 3; i++ {
		err = message.SendMessage(t, data)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	model.RedisClient.Del(ctx, key)
	notification := model.Notification{
		UserID:    userId,
		Type:      "一次性订阅消息",
		Title:     "AI绘画完成通知",
		Content:   data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	n.db.Create(&notification)
	return nil
}

type SubscribeReq struct {
	Id      string  `json:"id" binding:"required"`
	Message Message `json:"message" binding:"required"`
}

type Message struct {
	DrawTask string `json:"RqmsNBiXK9bClJ83Z0PkCguj1-wmetG6uGRz1qipf5w" binding:"required"`
}

type SendReq struct {
	Id          string `json:"id" binding:"required"`
	Theme       string `json:"theme" binding:"required"`
	Style       string `json:"style" binding:"required"`
	Result      string `json:"result" binding:"required"`
	CreatedTime string `json:"created_time" binding:"required"`
	Remarks     string `json:"remarks"`
}
