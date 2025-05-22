package router

import (
	"uportal-api/internal/handler"
	"uportal-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(taskHandler *handler.TaskHandler) *gin.Engine {
	r := gin.Default()

	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 任务相关路由
	taskGroup := r.Group("/api/v1/tasks")
	{
		// 管理员接口
		adminTaskGroup := taskGroup.Group("/admin")
		adminTaskGroup.Use(middleware.AdminAuth())
		{
			adminTaskGroup.POST("", taskHandler.CreateTask)                           // 创建任务
			adminTaskGroup.PUT("/:task_id", taskHandler.UpdateTask)                   // 更新任务
			adminTaskGroup.DELETE("/:task_id", taskHandler.DeleteTask)                // 删除任务
			adminTaskGroup.GET("/:task_id", taskHandler.GetTask)                      // 获取任务详情
			adminTaskGroup.GET("", taskHandler.ListTasks)                             // 获取任务列表
			adminTaskGroup.GET("/statistics/:task_id", taskHandler.GetTaskStatistics) // 获取任务统计信息
		}

		// 用户接口
		userTaskGroup := taskGroup.Group("")
		userTaskGroup.Use(middleware.Auth())
		{
			userTaskGroup.GET("/available", taskHandler.GetAvailableTasks)      // 获取可用任务列表
			userTaskGroup.POST("/complete", taskHandler.CompleteTask)           // 完成任务
			userTaskGroup.GET("/records", taskHandler.GetUserTaskRecords)       // 获取任务完成记录
			userTaskGroup.GET("/statistics", taskHandler.GetUserTaskStatistics) // 获取用户任务统计
		}
	}

	return r
}
