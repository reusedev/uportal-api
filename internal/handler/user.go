package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func RegisterUserRoutes(group *gin.RouterGroup, h *UserHandler) {
	group.POST("/message_report", h.MessageReport)
}

type messageReportRequest struct {
	ID                    int64  `json:"id" binding:"required"`
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	Username              string `json:"username"`
	LanguageCode          string `json:"language_code"`
	IsBot                 bool   `json:"is_bot"`
	IsPremium             bool   `json:"is_premium"`
	AllowsWriteToPm       bool   `json:"allows_write_to_pm"`
	AddedToAttachmentMenu bool   `json:"added_to_attachment_menu"`
	PhotoURL              string `json:"photo_url"`
}

// MessageReport 用户信息上报
func (h *UserHandler) MessageReport(c *gin.Context) {
	var req messageReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	user := &model.User{
		UserID:                req.ID,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Username:              req.Username,
		LanguageCode:          req.LanguageCode,
		IsBot:                 req.IsBot,
		IsPremium:             req.IsPremium,
		AllowsWriteToPm:       req.AllowsWriteToPm,
		AddedToAttachmentMenu: req.AddedToAttachmentMenu,
		PhotoURL:              req.PhotoURL,
	}
	if err := model.UpsertUser(h.db.WithContext(c.Request.Context()), user); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "用户信息上报失败", err))
		return
	}

	response.Success(c, gin.H{"message": "success"})
}
