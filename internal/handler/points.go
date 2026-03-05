package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/errors"
	"github.com/reusedev/uportal-api/pkg/response"
	"gorm.io/gorm"
)

type PointsHandler struct {
	db *gorm.DB
}

func NewPointsHandler(db *gorm.DB) *PointsHandler {
	return &PointsHandler{db: db}
}

func RegisterPointsRoutes(group *gin.RouterGroup, h *PointsHandler) {
	group.POST("/get", h.GetBalance)
	group.POST("/records", h.GetRecords)
}

type getBalanceRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

// GetBalance 获取积分余额
func (h *PointsHandler) GetBalance(c *gin.Context) {
	var req getBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	balance, err := model.GetUserBalance(h.db.WithContext(c.Request.Context()), req.UserID)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取余额失败", err))
		return
	}

	response.Success(c, gin.H{"points": balance})
}

type getRecordsRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Prev   string `json:"prev"`
	Limit  int    `json:"limit"`
}

// GetRecords 获取积分明细
func (h *PointsHandler) GetRecords(c *gin.Context) {
	var req getRecordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
		return
	}

	var prev int64
	if req.Prev != "" {
		var err error
		prev, err = strconv.ParseInt(req.Prev, 10, 64)
		if err != nil {
			response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的 prev 参数", err))
			return
		}
	}

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	records, err := model.GetTokenRecords(h.db.WithContext(c.Request.Context()), req.UserID, prev, limit)
	if err != nil {
		response.Error(c, errors.New(errors.ErrCodeInternal, "获取积分记录失败", err))
		return
	}

	response.Success(c, records)
}
