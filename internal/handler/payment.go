package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentHandler struct {
	db *gorm.DB
}

func NewPaymentHandler(db *gorm.DB) *PaymentHandler {
	return &PaymentHandler{db: db}
}

func RegisterPaymentRoutes(group *gin.RouterGroup, h *PaymentHandler) {
	group.POST("/plan_list", h.PlanList)
	group.POST("/report", h.Report)
}

// PlanList 获取充值方案列表
func (h *PaymentHandler) PlanList(c *gin.Context) {
	plans, err := model.GetRechargePlans(h.db.WithContext(c.Request.Context()))
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取充值方案失败", err))
		return
	}
	response.Success(c, plans)
}

type reportRequest struct {
	UserID          int64  `json:"user_id" binding:"required"`
	OrderID         string `json:"order_id" binding:"required"`
	PackageID       string `json:"package_id" binding:"required"`
	Points          int    `json:"points" binding:"required"`
	Stars           int    `json:"stars" binding:"required"`
	PaidAt          string `json:"paid_at" binding:"required"`
	PaymentChargeID string `json:"payment_charge_id" binding:"required"`
}

// Report 充值上报：在单个事务中创建订单 + 给用户加积分（幂等）
func (h *PaymentHandler) Report(c *gin.Context) {
	var req reportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	db := h.db.WithContext(c.Request.Context())
	err := db.Transaction(func(tx *gorm.DB) error {
		// 尝试创建订单，主键冲突则忽略（幂等）
		order := &model.RechargeOrder{
			OrderID:         req.OrderID,
			UserID:          req.UserID,
			PlanID:          req.PackageID,
			Points:          req.Points,
			Stars:           req.Stars,
			PaymentChargeID: req.PaymentChargeID,
			PaidAt:          req.PaidAt,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(order)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		title := fmt.Sprintf("充值 %d 积分", req.Points)
		desc := fmt.Sprintf("购买 %d Stars 套餐", req.Stars)
		return model.AddPoints(tx, req.UserID, req.Points, "recharge", title, desc, req.OrderID)
	})
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "充值上报失败", err))
		return
	}

	response.Success(c, gin.H{"message": "success"})
}
