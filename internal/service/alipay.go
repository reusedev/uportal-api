package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	apperrors "github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/logs"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlipayService 支付宝支付服务
type AlipayService struct {
	db       *gorm.DB
	client   *alipay.Client
	config   *config.Config
	orderSvc *OrderService
}

// NewAlipayService 创建支付宝支付服务
func NewAlipayService(db *gorm.DB, orderSvc *OrderService, cfg *config.Config) (*AlipayService, error) {
	// 创建支付宝客户端
	client, err := alipay.New(cfg.Alipay.AppID, cfg.Alipay.PrivateKey, cfg.Alipay.IsProd)
	if err != nil {
		return nil, fmt.Errorf("create alipay client error: %v", err)
	}

	// 加载支付宝公钥
	if err = client.LoadAliPayPublicKey(cfg.Alipay.PublicKey); err != nil {
		return nil, fmt.Errorf("load alipay public key error: %v", err)
	}

	return &AlipayService{
		db:       db,
		client:   client,
		config:   cfg,
		orderSvc: orderSvc,
	}, nil
}

// CreateAlipayOrder 创建支付宝支付订单
func (s *AlipayService) CreateAlipayOrder(ctx context.Context, orderID int64, description string, amount float64) (string, error) {
	// 获取订单信息
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return "", err
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		return "", errors.New(errors.ErrCodeInvalidParams, "订单状态不正确", nil)
	}

	// 检查支付金额
	if order.Amount != amount {
		return "", errors.New(errors.ErrCodeInvalidParams, "支付金额不正确", nil)
	}

	// 创建支付宝支付请求
	p := alipay.TradePagePay{}
	p.NotifyURL = s.config.Alipay.NotifyUrl
	p.ReturnURL = s.config.Alipay.ReturnUrl
	p.Subject = description
	p.OutTradeNo = order.OrderNo
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	// 生成支付链接
	url, err := s.client.TradePagePay(p)
	if err != nil {
		return "", fmt.Errorf("create alipay order error: %v", err)
	}

	// 更新订单状态为支付中
	err = s.orderSvc.UpdateOrderStatus(ctx, orderID, model.OrderStatusPending, map[string]interface{}{
		"payment_method": "alipay",
	})
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// HandleAlipayNotify 处理支付宝支付回调
func (s *AlipayService) HandleAlipayNotify(ctx context.Context, notifyData map[string]string) error {
	// 将 map[string]string 转换为 url.Values
	values := make(url.Values)
	for k, v := range notifyData {
		values.Set(k, v)
	}

	// 验证签名
	err := s.client.VerifySign(values)
	if err != nil {
		return apperrors.New(apperrors.ErrCodeInvalidParams, "签名验证失败", err)
	}

	// 获取订单号
	orderNo := notifyData["out_trade_no"]
	tradeNo := notifyData["trade_no"]

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return apperrors.New(apperrors.ErrCodeInternal, "开启事务失败", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建通知记录
	notifyRecord := &model.PaymentNotifyRecord{
		OrderID:       0, // 稍后更新
		TransactionID: tradeNo,
		NotifyType:    "alipay_trade_success",
		NotifyTime:    time.Now(),
		ProcessStatus: model.NotifyStatusPending,
	}

	// 检查是否已处理过该通知
	existingRecord, err := model.GetNotifyRecord(tx, 0, tradeNo)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，继续处理
		} else {
			tx.Rollback()
			return apperrors.New(apperrors.ErrCodeInternal, "查询通知记录失败", err)
		}
	} else {
		// 如果已处理成功，直接返回成功
		if existingRecord.ProcessStatus == model.NotifyStatusSuccess {
			logs.Business().Info("支付宝支付通知已处理",
				zap.String("trade_no", tradeNo),
			)
			if err := tx.Commit().Error; err != nil {
				return apperrors.New(apperrors.ErrCodeInternal, "提交事务失败", err)
			}
			return nil
		}
		// 如果处理失败且未超过重试次数，更新重试次数
		if existingRecord.ProcessStatus == model.NotifyStatusFailed && existingRecord.RetryCount < model.MaxRetryCount {
			notifyRecord = existingRecord
			notifyRecord.RetryCount++
		} else {
			tx.Rollback()
			return apperrors.New(apperrors.ErrCodeInternal, "通知处理失败且超过重试次数限制", nil)
		}
	}

	// 获取订单信息
	order, err := model.GetOrderByOrderNo(tx, orderNo)
	if err != nil {
		tx.Rollback()
		return apperrors.New(apperrors.ErrCodeInternal, "查询订单失败", err)
	}

	// 更新通知记录的订单ID
	notifyRecord.OrderID = order.OrderID

	// 如果通知记录不存在，创建新的通知记录
	if existingRecord == nil {
		if err := model.CreateNotifyRecord(tx, notifyRecord); err != nil {
			tx.Rollback()
			return apperrors.New(apperrors.ErrCodeInternal, "创建通知记录失败", err)
		}
	}

	// 幂等性检查：如果订单已经支付成功，直接返回成功
	if order.Status == model.OrderStatusPaid {
		logs.Business().Info("订单已支付，跳过处理",
			zap.String("order_no", orderNo),
		)
		// 更新通知记录为成功
		now := time.Now()
		notifyRecord.ProcessStatus = model.NotifyStatusSuccess
		notifyRecord.ProcessTime = &now
		if err := model.UpdateNotifyRecord(tx, notifyRecord.RecordID, map[string]interface{}{
			"process_status": notifyRecord.ProcessStatus,
			"process_time":   notifyRecord.ProcessTime,
		}); err != nil {
			tx.Rollback()
			return apperrors.New(apperrors.ErrCodeInternal, "更新通知记录失败", err)
		}
		if err := tx.Commit().Error; err != nil {
			return apperrors.New(apperrors.ErrCodeInternal, "提交事务失败", err)
		}
		return nil
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		tx.Rollback()
		return apperrors.New(apperrors.ErrCodeInvalidParams, "订单状态不正确", nil)
	}

	// 检查支付金额
	paidAmount, _ := strconv.ParseFloat(notifyData["total_amount"], 64)
	if paidAmount != order.Amount {
		tx.Rollback()
		return apperrors.New(apperrors.ErrCodeInvalidParams, "支付金额不匹配", nil)
	}

	// 更新订单状态
	paymentInfo := map[string]interface{}{
		"transaction_id": tradeNo,
		"payment_time":   time.Now(),
		"payment_method": "alipay",
		"trade_status":   notifyData["trade_status"],
	}
	paymentInfoJSON, _ := json.Marshal(paymentInfo)

	err = s.orderSvc.UpdateOrderStatus(ctx, order.OrderID, model.OrderStatusPaid, map[string]interface{}{
		"payment_info": string(paymentInfoJSON),
		"paid_at":      time.Now(),
	})
	if err != nil {
		tx.Rollback()
		return apperrors.New(apperrors.ErrCodeInternal, "更新订单状态失败", err)
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
		return apperrors.New(apperrors.ErrCodeInternal, "更新通知记录失败", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return apperrors.New(apperrors.ErrCodeInternal, "提交事务失败", err)
	}

	logs.Business().Info("支付宝支付通知处理成功",
		zap.String("order_no", orderNo),
		zap.String("trade_no", tradeNo),
	)
	return nil
}

// QueryAlipayOrder 查询支付宝支付订单
func (s *AlipayService) QueryAlipayOrder(ctx context.Context, orderID int64) (*alipay.TradeQueryRsp, error) {
	// 获取订单信息
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 查询支付宝订单
	p := alipay.TradeQuery{}
	p.OutTradeNo = order.OrderNo

	result, err := s.client.TradeQuery(ctx, p)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrCodeInternal, "查询支付宝订单失败", err)
	}

	return result, nil
}

// CloseAlipayOrder 关闭支付宝支付订单
func (s *AlipayService) CloseAlipayOrder(ctx context.Context, orderID int64) error {
	// 获取订单信息
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	// 检查订单状态
	if order.Status != model.OrderStatusPending {
		return apperrors.New(apperrors.ErrCodeInvalidParams, "订单状态不正确", nil)
	}

	// 关闭支付宝订单
	p := alipay.TradeClose{}
	p.OutTradeNo = order.OrderNo

	_, err = s.client.TradeClose(ctx, p)
	if err != nil {
		return apperrors.New(apperrors.ErrCodeInternal, "关闭支付宝订单失败", err)
	}

	// 更新订单状态为已取消
	err = s.orderSvc.UpdateOrderStatus(ctx, orderID, model.OrderStatusCancelled, nil)
	if err != nil {
		return apperrors.New(apperrors.ErrCodeInternal, "更新订单状态失败", err)
	}

	return nil
}
