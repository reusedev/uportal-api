package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/logs"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentService 支付服务
type PaymentService struct {
	db            *gorm.DB
	redis         *redis.Client
	orderSvc      *OrderService
	wxPayClient   *core.Client
	notifyHandler *notify.Handler
	config        *config.Config
}

// NewPaymentService 创建支付服务
func NewPaymentService(db *gorm.DB, redis *redis.Client, orderSvc *OrderService, cfg *config.Config) (*PaymentService, error) {
	// 加载商户证书
	mchPrivateKey, err := utils.LoadPrivateKey(cfg.Wechat.Pay.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load merchant private key error: %v", err)
	}

	// 创建微信支付客户端
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(cfg.Wechat.Pay.MchID, cfg.Wechat.Pay.MchApiKey, mchPrivateKey, cfg.Wechat.Pay.CertFile),
	}
	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("new wechat pay client error: %v", err)
	}

	// 创建回调处理器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.Wechat.Pay.MchID)
	handler := notify.NewNotifyHandler(cfg.Wechat.Pay.MchApiKey, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))

	return &PaymentService{
		db:            db,
		redis:         redis,
		orderSvc:      orderSvc,
		wxPayClient:   client,
		notifyHandler: handler,
		config:        cfg,
	}, nil
}

// acquireLock 获取分布式锁
func (s *PaymentService) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return s.redis.SetNX(ctx, key, "1", ttl).Result()
}

// releaseLock 释放分布式锁
func (s *PaymentService) releaseLock(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}

// CreateWxPayOrder 创建微信支付订单
func (s *PaymentService) CreateWxPayOrder(ctx context.Context, orderID int64, description string, amount float64) (*jsapi.PrepayWithRequestPaymentResponse, error) {
	// 获取订单信息
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		return nil, errors.New(errors.ErrCodeInvalidParams, "订单状态不正确", nil)
	}

	// 检查支付金额
	if order.Amount != amount {
		return nil, errors.New(errors.ErrCodeInvalidParams, "支付金额不正确", nil)
	}

	// 获取用户的微信OpenID
	var userAuth model.UserAuth
	err = s.db.Where("user_id = ? AND provider = ?", order.UserID, "wechat").First(&userAuth).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errors.ErrCodeInvalidParams, "用户未绑定微信账号", nil)
		}
		return nil, errors.New(errors.ErrCodeInternal, "获取用户微信信息失败", err)
	}

	// 创建支付订单
	svc := jsapi.JsapiApiService{Client: s.wxPayClient}
	resp, _, err := svc.PrepayWithRequestPayment(ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(s.config.Wechat.Pay.AppID),
			Mchid:       core.String(s.config.Wechat.Pay.MchID),
			Description: core.String(description),
			OutTradeNo:  core.String(order.OrderNo),
			NotifyUrl:   core.String(s.config.Wechat.Pay.NotifyUrl),
			Amount: &jsapi.Amount{
				Total:    core.Int64(int64(amount * 100)), // 转换为分
				Currency: core.String("CNY"),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(userAuth.ProviderUserID),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create wx pay order error: %v", err)
	}

	// 更新订单状态为支付中
	err = s.orderSvc.UpdateOrderStatus(ctx, orderID, model.OrderStatusPending, map[string]interface{}{
		"payment_method": "wechat",
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// HandleWxPayNotify 处理微信支付回调
func (s *PaymentService) HandleWxPayNotify(ctx context.Context, requestBody []byte, headers map[string]string) error {
	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", "", nil)
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 解析回调通知
	var transaction payments.Transaction
	notifyReq, err := s.notifyHandler.ParseNotifyRequest(ctx, req, &transaction)
	if err != nil {
		return fmt.Errorf("parse notify request error: %v", err)
	}

	// 验证回调通知
	if notifyReq.EventType != "TRANSACTION.SUCCESS" {
		return fmt.Errorf("unexpected event type: %s", notifyReq.EventType)
	}

	// 获取订单号
	orderNo := *transaction.OutTradeNo

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("start transaction error: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建通知记录
	notifyRecord := &model.PaymentNotifyRecord{
		OrderID:       0, // 稍后更新
		TransactionID: *transaction.TransactionId,
		NotifyType:    notifyReq.EventType,
		NotifyTime:    time.Now(),
		ProcessStatus: model.NotifyStatusPending,
	}

	// 检查是否已处理过该通知
	existingRecord, err := model.GetNotifyRecord(tx, 0, *transaction.TransactionId)
	if err == nil {
		// 如果已处理成功，直接返回成功
		if existingRecord.ProcessStatus == model.NotifyStatusSuccess {
			log.Printf("Notification for transaction %s already processed successfully", *transaction.TransactionId)
			if err := tx.Commit().Error; err != nil {
				return fmt.Errorf("commit transaction error: %v", err)
			}
			return nil
		}
		// 如果处理失败且未超过重试次数，更新重试次数
		if existingRecord.ProcessStatus == model.NotifyStatusFailed && existingRecord.RetryCount < model.MaxRetryCount {
			notifyRecord = existingRecord
			notifyRecord.RetryCount++
		} else {
			tx.Rollback()
			return fmt.Errorf("notification processing failed and exceeded retry limit")
		}
	} else if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("check notify record error: %v", err)
	}

	// 获取订单信息
	order, err := model.GetOrderByOrderNo(tx, orderNo)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("get order error: %v", err)
	}

	// 更新通知记录的订单ID
	notifyRecord.OrderID = order.OrderID

	// 如果通知记录不存在，创建新的通知记录
	if existingRecord == nil {
		if err := model.CreateNotifyRecord(tx, notifyRecord); err != nil {
			tx.Rollback()
			return fmt.Errorf("create notify record error: %v", err)
		}
	}

	// 获取分布式锁
	lockKey := fmt.Sprintf("payment_notify_lock:%d:%s", order.OrderID, *transaction.TransactionId)
	acquired, err := s.acquireLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("acquire lock error: %v", err)
	}
	if !acquired {
		tx.Rollback()
		return fmt.Errorf("failed to acquire lock, another process is handling this notification")
	}
	defer s.releaseLock(ctx, lockKey)

	// 幂等性检查：如果订单已经支付成功，直接返回成功
	if order.Status == model.OrderStatusPaid {
		log.Printf("Order %s already paid, skip processing", orderNo)
		// 更新通知记录为成功
		now := time.Now()
		notifyRecord.ProcessStatus = model.NotifyStatusSuccess
		notifyRecord.ProcessTime = &now
		if err := model.UpdateNotifyRecord(tx, notifyRecord.RecordID, map[string]interface{}{
			"process_status": notifyRecord.ProcessStatus,
			"process_time":   notifyRecord.ProcessTime,
		}); err != nil {
			tx.Rollback()
			return fmt.Errorf("update notify record error: %v", err)
		}
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit transaction error: %v", err)
		}
		return nil
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		tx.Rollback()
		return fmt.Errorf("invalid order status: %s", order.Status)
	}

	// 检查支付金额
	paidAmount := float64(*transaction.Amount.Total) / 100.0
	if paidAmount != order.Amount {
		tx.Rollback()
		return fmt.Errorf("payment amount mismatch: expected %.2f, got %.2f", order.Amount, paidAmount)
	}

	// 更新订单状态
	err = s.orderSvc.UpdateOrderStatus(ctx, order.OrderID, model.OrderStatusPaid, map[string]interface{}{
		"transaction_id": *transaction.TransactionId,
		"payment_time":   time.Now(),
	})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("update order status error: %v", err)
	}

	// 更新通知记录
	now := time.Now()
	notifyRecord.ProcessStatus = model.NotifyStatusSuccess
	notifyRecord.ProcessTime = &now
	if err := model.UpdateNotifyRecord(tx, notifyRecord.RecordID, map[string]interface{}{
		"process_status": notifyRecord.ProcessStatus,
		"process_time":   notifyRecord.ProcessTime,
	}); err != nil {
		tx.Rollback()
		return fmt.Errorf("update notify record error: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction error: %v", err)
	}

	log.Printf("Successfully processed payment notification for order %s, transaction %s", orderNo, *transaction.TransactionId)
	return nil
}

// RetryFailedNotifications 重试失败的通知
func (s *PaymentService) RetryFailedNotifications(ctx context.Context) error {
	// 获取待处理的通知记录
	records, err := model.ListPendingNotifyRecords(s.db, 100) // 每次处理100条
	if err != nil {
		return fmt.Errorf("list pending notify records error: %v", err)
	}

	for _, record := range records {
		// 获取分布式锁
		lockKey := fmt.Sprintf("payment_notify_retry_lock:%d:%s", record.OrderID, record.TransactionID)
		acquired, err := s.acquireLock(ctx, lockKey, 30*time.Second)
		if err != nil {
			log.Printf("Failed to acquire lock for retry: %v", err)
			continue
		}
		if !acquired {
			log.Printf("Failed to acquire lock for retry, skipping record %d", record.RecordID)
			continue
		}

		// 重新处理通知
		headers := map[string]string{
			"Wechatpay-Signature": record.TransactionID, // 这里需要保存原始签名
			"Wechatpay-Timestamp": record.NotifyTime.Format(time.RFC3339),
			"Wechatpay-Nonce":     strconv.FormatInt(record.RecordID, 10),
			"Wechatpay-Serial":    "1", // 这里需要保存原始证书序列号
		}

		err = s.HandleWxPayNotify(ctx, nil, headers)
		if err != nil {
			log.Printf("Retry failed for record %d: %v", record.RecordID, err)
		}

		// 释放锁
		if err := s.releaseLock(ctx, lockKey); err != nil {
			log.Printf("Failed to release lock for retry: %v", err)
		}
	}

	return nil
}

// QueryWxPayOrder 查询微信支付订单
func (s *PaymentService) QueryWxPayOrder(ctx context.Context, orderID int64) (*payments.Transaction, error) {
	// 获取订单信息
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 查询微信支付订单
	svc := jsapi.JsapiApiService{Client: s.wxPayClient}
	resp, _, err := svc.QueryOrderByOutTradeNo(ctx, jsapi.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(order.OrderNo),
		Mchid:      core.String(s.config.Wechat.Pay.MchID),
	})
	if err != nil {
		return nil, fmt.Errorf("query wx pay order error: %v", err)
	}

	return resp, nil
}

// CloseWxPayOrder 关闭微信支付订单
func (s *PaymentService) CloseWxPayOrder(ctx context.Context, orderID int64) error {
	// 获取订单信息
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		return errors.New(errors.ErrCodeInvalidParams, "订单状态不正确", nil)
	}

	// 关闭微信支付订单
	svc := jsapi.JsapiApiService{Client: s.wxPayClient}
	_, err = svc.CloseOrder(ctx, jsapi.CloseOrderRequest{
		OutTradeNo: core.String(order.OrderNo),
		Mchid:      core.String(s.config.Wechat.Pay.MchID),
	})
	if err != nil {
		return fmt.Errorf("close wx pay order error: %v", err)
	}

	// 更新订单状态为已取消
	err = s.orderSvc.UpdateOrderStatus(ctx, orderID, model.OrderStatusCancelled, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetOrder 获取订单信息
func (s *PaymentService) GetOrder(ctx context.Context, orderID int64) (*model.Order, error) {
	order, err := model.GetOrderByID(s.db, orderID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
		}
		return nil, errors.New(errors.ErrCodeInternal, "查询订单失败", err)
	}
	return order, nil
}

// CreateOrder 创建订单
func (s *PaymentService) CreateOrder(ctx context.Context, userID string, amount float64, productID string, productName string) (*model.Order, error) {
	// 创建订单
	order := &model.Order{
		UserID:      userID,
		OrderNo:     generateOrderNo(),
		Amount:      amount,
		ProductID:   productID,
		ProductName: productName,
		Status:      model.OrderStatusPending,
	}

	err := model.CreateOrder(s.db, order)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建订单失败", err)
	}

	logs.Business().Info("订单创建成功",
		zap.Int64("order_id", order.OrderID),
		zap.String("user_id", userID),
		zap.String("order_no", order.OrderNo),
		zap.Float64("amount", amount),
	)

	return order, nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *PaymentService) GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	order, err := model.GetOrderByOrderNo(s.db, orderNo)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
		}
		return nil, errors.New(errors.ErrCodeInternal, "查询订单失败", err)
	}
	return order, nil
}

// UpdateOrderStatus 更新订单状态
func (s *PaymentService) UpdateOrderStatus(ctx context.Context, orderID int64, status model.OrderStatus, paymentInfo map[string]interface{}) error {
	// 获取订单信息
	order, err := model.GetOrderByID(s.db, orderID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.ErrCodeNotFound, "订单不存在", nil)
		}
		return errors.New(errors.ErrCodeInternal, "查询订单失败", err)
	}

	// 检查订单状态是否可以更新
	if !canUpdateOrderStatus(order.Status, status) {
		return errors.New(errors.ErrCodeInvalidParams, "订单状态不允许更新", nil)
	}

	// 更新订单状态
	updates := map[string]interface{}{
		"status": status,
	}

	// 如果支付成功，记录支付信息
	if status == model.OrderStatusPaid {
		paymentInfoJSON, err := json.Marshal(paymentInfo)
		if err != nil {
			return errors.New(errors.ErrCodeInternal, "序列化支付信息失败", err)
		}
		updates["payment_info"] = string(paymentInfoJSON)
		updates["paid_at"] = time.Now()
	}

	err = model.UpdateOrder(s.db, orderID, updates)
	if err != nil {
		return errors.New(errors.ErrCodeInternal, "更新订单状态失败", err)
	}

	logs.Business().Info("订单状态更新成功",
		zap.Int64("order_id", orderID),
		zap.String("order_no", order.OrderNo),
		zap.String("old_status", string(order.Status)),
		zap.String("new_status", string(status)),
	)

	return nil
}

// GetUserOrders 获取用户订单列表
func (s *PaymentService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*model.Order, int64, error) {
	orders, total, err := model.GetUserOrders(s.db, userID, page, pageSize)
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "查询用户订单失败", err)
	}
	return orders, total, nil
}

// 生成订单号
func generateOrderNo() string {
	return fmt.Sprintf("%d%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// 检查订单状态是否可以更新
func canUpdateOrderStatus(oldStatus, newStatus model.OrderStatus) bool {
	switch oldStatus {
	case model.OrderStatusPending:
		return newStatus == model.OrderStatusPaid || newStatus == model.OrderStatusCancelled
	case model.OrderStatusPaid:
		return newStatus == model.OrderStatusCompleted || newStatus == model.OrderStatusRefunded
	case model.OrderStatusRefunded:
		return newStatus == model.OrderStatusCompleted
	default:
		return false
	}
}
