package handler

import "github.com/gin-gonic/gin"

// RegisterGoodsRoutes 商品
func RegisterGoodsRoutes(r *gin.RouterGroup, h *TaskHandler) {
	r.POST("/list", h.ListGoods)      // 获取商品列表
	r.POST("/add", h.CreateGoods)     // 创建商品
	r.POST("/edit", h.UpdateGood)     // 更新商品
	r.POST("/operate", h.OperateGood) // 操作商品
	r.POST("/delete", h.DeleteGood)   // 删除商品
}
