package handler

import (
	basicErr "errors"
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/consts"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"gorm.io/gorm"
)

// InviteHandler 邀请处理器
type InviteHandler struct {
	inviteSvc *service.InviteService
}

// NewInviteHandler 创建邀请处理器
func NewInviteHandler(inviteSvc *service.InviteService) *InviteHandler {
	return &InviteHandler{
		inviteSvc: inviteSvc,
	}
}

// ReportInviteRequest 邀请上报请求
type ReportInviteRequest struct {
	InviteBy string `json:"invite_by" binding:"required"` // 邀请人ID
}

// ReportPointsRewardRequest 代币奖励上报请求
type ReportPointsRewardRequest struct {
	Type string `json:"type" binding:"required"` // 奖励类型
}

func (h *InviteHandler) ReportInvite(c *gin.Context) {
	// 从上下文获取当前用户ID
	userID := c.GetString(consts.UserId)
	if userID == "" {
		response.Error(c, errors.New(errors.ErrCodeUnauthorized, "未登录", nil))
		return
	}

	// 绑定请求参数
	var req ReportInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	inviteBy := req.InviteBy

	// 检查邀请人ID是否有效
	if inviteBy == "" {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的邀请人ID", nil))
		return
	}

	// 不能邀请自己
	if inviteBy == userID {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "不能邀请自己", nil))
		return
	}

	// 检查当前用户是否已经被邀请
	var currentUser model.User
	if err := h.inviteSvc.GetDB().First(&currentUser, userID).Error; err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取用户信息失败", err))
		return
	}
	if currentUser.InviterID != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "您已经被邀请", nil))
		return
	}

	// 检查邀请人状态
	var inviter model.User
	if err := h.inviteSvc.GetDB().First(&inviter, req.InviteBy).Error; err != nil {
		if basicErr.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, errors.New(errors.ErrCodeUserNotFound, "邀请人不存在", nil))
			return
		}
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取邀请人信息失败", err))
		return
	}
	if inviter.Status != 1 {
		response.Error(c, errors.New(errors.ErrCodeUserDisabled, "邀请人账号已被禁用", nil))
		return
	}

	// 从系统配置获取邀请奖励代币数
	// TODO: 从系统配置中获取实际的奖励代币数
	tokenReward := 100 // 临时使用固定值，后续从配置获取

	// 开启事务处理邀请记录和奖励
	tx := h.inviteSvc.GetDB().Begin()
	if tx.Error != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "开启事务失败", tx.Error))
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新当前用户的邀请人ID
	if err := tx.Model(&currentUser).Update("inviter_id", req.InviteBy).Error; err != nil {
		tx.Rollback()
		response.Error(c, errors.New(errors.ErrCodeInternal, "更新邀请关系失败", err))
		return
	}
	// 创建邀请记录
	if err := h.inviteSvc.CreateInviteRecordWithTx(c.Request.Context(), tx, inviteBy, userID, tokenReward); err != nil {
		tx.Rollback()
		response.Error(c, err)
		return
	}

	// 立即处理邀请奖励
	if err := h.inviteSvc.ProcessInviteRewardWithTx(c.Request.Context(), tx, userID); err != nil {
		tx.Rollback()
		response.Error(c, err)
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "提交事务失败", err))
		return
	}

	response.Success(c, nil)
}

// RegisterInviteRoutes 注册邀请相关路由
func RegisterInviteRoutes(r *gin.RouterGroup, h *InviteHandler) {
	r.POST("/report", h.ReportInvite)
}
