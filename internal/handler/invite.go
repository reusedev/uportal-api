package handler

import (
	basicErr "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/internal/service"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/consts"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/qrcode"
	"github.com/reusedev/uportal-api/pkg/response"
	"github.com/reusedev/uportal-api/pkg/wechat_token"
	"gorm.io/gorm"
	"strconv"
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

type QrcodeInviteRequest struct {
	Scene     string `json:"scene" binding:"required"`
	IsHyaline bool   `json:"is_hyaline"`
}

// ReportPointsRewardRequest 代币奖励上报请求
type ReportPointsRewardRequest struct {
	Type string `json:"type" binding:"required"` // 奖励类型
}

func (h *InviteHandler) QrcodeInvite(c *gin.Context) {
	var req QrcodeInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	userID := c.GetString(consts.UserId)
	//检查当前用户是否有专属二维码了
	var currentUser model.User
	if err := h.inviteSvc.GetDB().Where("id = ?", userID).First(&currentUser).Error; err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取用户信息失败", err))
		return
	}
	if currentUser.Qrcode != "" {
		imageUrl := model.GetUrlById(currentUser.Qrcode, config.GlobalConfig.DrawApi.UploadFileUrl, "input")
		if imageUrl != "" {
			response.Success(c, imageUrl)
			return
		}
	}
	savePath := fmt.Sprintf("tmp/%s.png", userID)
	t := wechat_token.GetToken()
	err := qrcode.GetQrcode(t, req.Scene, "", savePath, req.IsHyaline)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取专属二维码错误", err))
		return
	}
	//上传
	uploadFile, err := model.UploadFile(savePath, config.GlobalConfig.DrawApi.UploadFileUrl)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "上传文件失败", err))
		return
	}
	currentUser.Qrcode = strconv.Itoa(uploadFile.Data.Id)
	h.inviteSvc.GetDB().Save(&currentUser)
	response.Success(c, uploadFile.Data.Url)
}

func (h *InviteHandler) QrcodeWork(c *gin.Context) {
	var data []byte
	var err error
	if data, err = c.GetRawData(); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	userID := c.GetString(consts.UserId)
	savePath := fmt.Sprintf("tmp/work_%s.png", userID)
	t := wechat_token.GetToken()
	err = qrcode.GetWorksQrcode(t, savePath, data)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取专属二维码错误", err))
		return
	}
	//上传
	uploadFile, err := model.UploadFile(savePath, config.GlobalConfig.DrawApi.UploadFileUrl)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "上传文件失败", err))
		return
	}
	response.Success(c, uploadFile.Data.Url)
}

type QrcodeReq struct {
	WorkId string `json:"work_id" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

func (h *InviteHandler) Qrcode(c *gin.Context) {
	var req QrcodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}
	savePath := fmt.Sprintf("tmp/work_%s.png", req.UserID)
	t := wechat_token.GetToken()
	data := map[string]interface{}{
		"page":       "pages/creative/creative",
		"scene":      fmt.Sprintf("id=%s!invite_by=%s", req.WorkId, req.UserID),
		"is_hyaline": true,
	}
	err := qrcode.GetWorksQrcode(t, savePath, data)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取专属二维码错误", err))
		return
	}
	//上传
	uploadFile, err := model.UploadFile(savePath, config.GlobalConfig.DrawApi.UploadFileUrl)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "上传文件失败", err))
		return
	}
	response.Success(c, uploadFile.Data.Url)
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
		response.Success(c, nil)
		return
	}

	// 检查当前用户是否已经被邀请
	var currentUser model.User
	if err := h.inviteSvc.GetDB().Where("id = ?", userID).First(&currentUser).Error; err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取用户信息失败", err))
		return
	}
	if currentUser.InviterID != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "您已经被邀请", nil))
		return
	}

	// 检查邀请人状态
	var inviter model.User
	if err := h.inviteSvc.GetDB().Where("id = ?", req.InviteBy).First(&inviter).Error; err != nil {
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
	tokenReward := 1000 // 临时使用固定值，后续从配置获取
	if inviteBy == "5uiw4fCEidU" {
		tokenReward = 3000
	}

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
	r.POST("/qrcode", h.QrcodeInvite)
}

// RegisterWorkRoutes 注册相关路由
func RegisterWorkRoutes(r *gin.RouterGroup, h *InviteHandler) {
	r.POST("/qrcode", h.QrcodeWork)
}
