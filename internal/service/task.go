package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/reusedev/uportal-api/internal/model"
	"github.com/reusedev/uportal-api/pkg/config"
	"github.com/reusedev/uportal-api/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TaskService 任务服务
type TaskService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
	config *config.Config
}

// NewTaskService 创建任务服务
func NewTaskService(db *gorm.DB, redis *redis.Client, logger *zap.Logger, config *config.Config) *TaskService {
	return &TaskService{
		db:     db,
		redis:  redis,
		logger: logger,
		config: config,
	}
}

type ListTaskRequest struct {
	//Page     int    `json:"page" binding:"required"`
	//Limit    int    `json:"limit" binding:"required"`
	Status   *int   `json:"status"`    // 可选，任务状态
	TaskName string `json:"task_name"` // 可选，任务名称模糊查询
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	TaskName        string `json:"task_name" binding:"required"`
	Description     string `json:"task_desc" binding:"required"`
	TokenReward     int    `json:"token_reward" binding:"required"`
	DailyLimit      int    `json:"daily_limit" binding:"required"`
	IntervalSeconds *int   `json:"interval_seconds" binding:"required"`
	ValidFrom       string `json:"valid_from" binding:"required"`
	ValidTo         string `json:"valid_to"`
	Repeatable      *int8  `json:"repeatable" binding:"required"`
	Status          *int8  `json:"status" binding:"required"`
	TaskKey         string `json:"task_key" binding:"required"`
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*model.RewardTask, error) {
	task := &model.RewardTask{
		TaskName:        req.TaskName,
		TaskDesc:        &req.Description,
		TokenReward:     req.TokenReward,
		DailyLimit:      req.DailyLimit,
		IntervalSeconds: *req.IntervalSeconds,
		Repeatable:      *req.Repeatable,
		Status:          *req.Status, // 默认启用
		TaskKey:         req.TaskKey,
	}
	from, _ := time.Parse(time.DateOnly, req.ValidFrom)
	task.ValidFrom = &from
	if req.ValidTo == "" {
		to, err := time.Parse(time.DateOnly, req.ValidTo)
		if err == nil {
			task.ValidTo = &to
		}
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "创建任务失败", err)
	}

	return task, nil
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	TaskId          int    `json:"id" binding:"required"`
	Status          *int8  `json:"status" binding:"required"`
	TaskName        string `json:"task_name" binding:"required"`
	Description     string `json:"task_desc" binding:"required"`
	TokenReward     int    `json:"token_reward" binding:"required"`
	DailyLimit      int    `json:"daily_limit" binding:"required"`
	IntervalSeconds *int   `json:"interval_seconds" binding:"required"`
	ValidFrom       string `json:"valid_from" binding:"required"`
	ValidTo         string `json:"valid_to"`
	Repeatable      *int8  `json:"repeatable" binding:"required"`
	TaskKey         string `json:"task_key" binding:"required"`
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(ctx context.Context, req *UpdateTaskRequest) (*model.RewardTask, error) {
	task, err := s.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, err
	}

	from, _ := time.Parse(time.DateOnly, req.ValidFrom)
	updates := map[string]interface{}{
		"task_name":        req.TaskName,
		"token_reward":     req.TokenReward,
		"status":           *req.Status,
		"task_desc":        req.Description,
		"daily_limit":      req.DailyLimit,
		"interval_seconds": *req.IntervalSeconds,
		"valid_from":       &from,
		"repeatable":       req.Repeatable,
		"task_key":         req.TaskKey,
	}
	if req.ValidTo != "" {
		to, err := time.Parse(time.DateOnly, req.ValidTo)
		if err == nil {
			updates["valid_to"] = &to
		}
	}

	if err := s.db.Model(task).Updates(updates).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "更新任务失败", err)
	}

	return task, nil
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, taskID int) error {
	if err := s.db.Delete(&model.RewardTask{}, taskID).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "删除任务失败", err)
	}
	return nil
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(ctx context.Context, taskID int) (*model.RewardTask, error) {
	var task model.RewardTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(errors.ErrCodeNotFound, "任务不存在", nil)
		}
		return nil, errors.New(errors.ErrCodeInternal, "获取任务失败", err)
	}
	return &task, nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(ctx context.Context, status *int, taskName string) ([]*model.RewardTask, int64, error) {
	var tasks []*model.RewardTask
	var total int64

	query := s.db.Model(&model.RewardTask{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if taskName != "" {
		query = query.Where("task_name LIKE ?", "%"+taskName+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取任务总数失败", err)
	}

	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取任务列表失败", err)
	}

	return tasks, total, nil
}

// ListConsumptionRules 获取代币消耗规则列表
func (s *TaskService) ListConsumptionRules(ctx context.Context) ([]*model.TokenConsumeRule, int64, error) {
	rules, total, err := model.ListTokenConsumptionRules(s.db)
	if err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取代币消耗规则列表失败", err)
	}
	return rules, total, nil
}

// GetConsumptionRules 获取代币消耗规则列表
func (s *TaskService) GetConsumptionRules(ctx context.Context, class string) ([]*model.TokenConsumeRule, error) {
	rules, err := model.GetTokenConsumptionRules(s.db, class)
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取代币消耗规则列表失败", err)
	}
	return rules, nil
}

// UpdateConsumptionRule 更新消费规则
func (s *TaskService) UpdateConsumptionRule(ctx context.Context, id int, req *UpdateConsumptionRuleRequest) error {
	// 查找现有规则
	var rule model.TokenConsumeRule
	if err := s.db.Where("feature_id = ?", id).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrCodeNotFound, "消费规则不存在", err)
		}
		return errors.New(errors.ErrCodeInternal, "查询消费规则失败", err)
	}

	// 更新规则
	updates := make(map[string]interface{})
	if req.FeatureName != "" {
		updates["feature_name"] = req.FeatureName
	}
	if req.FeatureDesc != "" {
		updates["feature_desc"] = req.FeatureDesc
	}
	if req.TokenCost != nil {
		updates["token_cost"] = *req.TokenCost
	}
	if req.FeatureCode != "" {
		updates["feature_code"] = req.FeatureCode
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Class != "" {
		updates["classify"] = req.Class
	}

	if err := s.db.Model(&rule).Updates(updates).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "更新消费规则失败", err)
	}

	return nil
}

type CreateConsumptionRuleRequest struct {
	FeatureName string `json:"feature_name"`
	FeatureDesc string `json:"feature_desc"`
	TokenCost   *int   `json:"token_cost,omitempty"`
	FeatureCode string `json:"feature_code"`
	Status      *int8  `json:"status"`
	Class       string `json:"class"`
}

// CreateConsumptionRule 创建Token消费规则
func (s *TaskService) CreateConsumptionRule(ctx context.Context, req *CreateConsumptionRuleRequest) (*model.TokenConsumeRule, error) {
	rule := &model.TokenConsumeRule{
		FeatureName: req.FeatureName,
		FeatureDesc: &req.FeatureDesc,
		TokenCost:   *req.TokenCost,
		FeatureCode: &req.FeatureCode,
		Status:      *req.Status,
		Class:       req.Class,
	}

	err := model.CreateTokenConsumptionRule(s.db, rule)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// GetAvailableTasks 获取用户可用的任务列表
func (s *TaskService) GetAvailableTasks(ctx context.Context, userID string) ([]*model.RewardTask, error) {
	var tasks []*model.RewardTask
	now := time.Now()

	// 获取所有启用的任务
	if err := s.db.Where("status = 1 AND (valid_from IS NULL OR valid_from <= ?) AND (valid_to IS NULL OR valid_to >= ?)",
		now, now).Find(&tasks).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取可用任务失败", err)
	}

	// 过滤掉已达到每日限制的任务
	var availableTasks []*model.RewardTask
	for _, task := range tasks {
		if task.DailyLimit > 0 {
			count, err := s.getUserTaskCompletionCount(ctx, userID, task.TaskID)
			if err != nil {
				return nil, err
			}
			if count >= int64(task.DailyLimit) {
				continue
			}
		}

		// 检查间隔时间
		if task.IntervalSeconds > 0 {
			lastCompletion, err := s.getLastTaskCompletion(ctx, s.db, userID, task.TaskID)
			if err != nil {
				return nil, err
			}
			if lastCompletion != nil {
				nextAvailableTime := lastCompletion.Add(time.Duration(task.IntervalSeconds) * time.Second)
				if now.Before(nextAvailableTime) {
					continue
				}
			}
		}

		// 检查是否可重复完成
		if task.Repeatable == 0 {
			completed, err := s.hasCompletedTask(ctx, s.db, userID, task.TaskID)
			if err != nil {
				return nil, err
			}
			if completed {
				continue
			}
		}

		availableTasks = append(availableTasks, task)
	}

	return availableTasks, nil
}

// getUserTaskCompletionCount 获取用户任务完成次数
func (s *TaskService) getUserTaskCompletionCount(ctx context.Context, userID string, taskID int) (int64, error) {
	var count int64
	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)

	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Where("user_id = ? AND task_id = ? AND completed_at >= ?",
			userID, taskID, today).
		Count(&count).Error; err != nil {
		return 0, errors.New(errors.ErrCodeInternal, "获取任务完成次数失败", err)
	}

	return count, nil
}

// getUserTotalTaskCompletionCount 获取用户任务总完成次数
func (s *TaskService) getUserTotalTaskCompletionCount(ctx context.Context, userID string, taskID int) (int64, error) {
	var count int64
	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Where("user_id = ? AND task_id = ?",
			userID, taskID).
		Count(&count).Error; err != nil {
		return 0, errors.New(errors.ErrCodeInternal, "获取任务总完成次数失败", err)
	}
	return count, nil
}

// CompleteTaskRequest 完成任务请求
type CompleteTaskRequest struct {
	TaskID    int                    `json:"task_id" binding:"required"`
	ExtraData map[string]interface{} `json:"extra_data"`
}

// TaskCompletionResult 任务完成结果
type TaskCompletionResult struct {
	//TaskID            int        `json:"task_id"`
	//TaskName          string     `json:"task_name"`
	//TokenReward       int        `json:"token_reward"`
	//IsCompleted       bool       `json:"is_completed"`
	//Message           string     `json:"message"`
	//NextAvailableTime *time.Time `json:"next_available_time,omitempty"`
	Balance int64 `json:"balance"`
	Reward  int   `json:"reward"`
}

// CompleteTask 完成任务
func (s *TaskService) CompleteTask(ctx context.Context, userID string, req *CompleteTaskRequest) (*TaskCompletionResult, error) {
	// 获取分布式锁，防止并发完成
	lockKey := fmt.Sprintf("task_completion_lock:%s:%d", userID, req.TaskID)
	acquired, err := s.redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取任务锁失败", err)
	}
	if !acquired {
		return nil, errors.New(errors.ErrCodeTooManyRequests, "任务正在处理中，请稍后重试", nil)
	}
	defer s.redis.Del(ctx, lockKey)

	// 开启事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New(errors.ErrCodeInternal, "开启事务失败", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取任务信息
	task, err := s.GetTask(ctx, req.TaskID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 检查任务状态
	if task.Status != 1 {
		tx.Rollback()
		return nil, errors.New(errors.ErrCodeInvalidParams, "任务未启用", nil)
	}

	// 检查任务有效期
	now := time.Now()
	if task.ValidFrom != nil && now.Before(*task.ValidFrom) {
		tx.Rollback()
		return nil, errors.New(errors.ErrCodeInvalidParams, "任务未开始", nil)
	}
	if task.ValidTo != nil && now.After(*task.ValidTo) {
		tx.Rollback()
		return nil, errors.New(errors.ErrCodeInvalidParams, "任务已结束", nil)
	}

	// 检查任务限制
	if err := s.checkTaskLimits(ctx, tx, userID, task); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 验证任务完成条件
	if err := s.verifyTaskCompletion(ctx, tx, userID, task, req.ExtraData); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 发放奖励
	if err := s.grantTaskReward(ctx, tx, userID, task); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 记录任务完成
	if err := s.recordTaskCompletion(ctx, tx, userID, task); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "提交事务失败", err)
	}

	token, err := s.getUserToken(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	return &TaskCompletionResult{
		Reward:  task.TokenReward,
		Balance: token,
	}, nil
}

// checkTaskLimits 检查任务限制
func (s *TaskService) checkTaskLimits(ctx context.Context, tx *gorm.DB, userID string, task *model.RewardTask) error {
	// 检查每日限制
	if task.DailyLimit > 0 {
		count, err := s.getUserTaskCompletionCount(ctx, userID, task.TaskID)
		if err != nil {
			return err
		}
		if int(count) >= task.DailyLimit {
			return errors.New(errors.ErrCodeInvalidParams, "今日任务完成次数已达上限", nil)
		}
	}

	// 检查间隔时间
	if task.IntervalSeconds > 0 {
		lastCompletion, err := s.getLastTaskCompletion(ctx, tx, userID, task.TaskID)
		if err != nil {
			return err
		}
		if lastCompletion != nil {
			nextAvailableTime := lastCompletion.Add(time.Duration(task.IntervalSeconds) * time.Second)
			if time.Now().Before(nextAvailableTime) {
				return errors.New(errors.ErrCodeInvalidParams,
					fmt.Sprintf("任务冷却中，请在 %s 后重试", nextAvailableTime.Format("2006-01-02 15:04:05")), nil)
			}
		}
	}

	// 检查是否可重复完成
	if task.Repeatable == 0 {
		completed, err := s.hasCompletedTask(ctx, tx, userID, task.TaskID)
		if err != nil {
			return err
		}
		if completed {
			return errors.New(errors.ErrCodeInvalidParams, "该任务已完成且不可重复完成", nil)
		}
	}

	return nil
}

// verifyTaskCompletion 验证任务完成条件
func (s *TaskService) verifyTaskCompletion(ctx context.Context, tx *gorm.DB, userID string, task *model.RewardTask, extraData map[string]interface{}) error {
	// 这里可以根据具体任务类型实现不同的验证逻辑
	// 例如：观看视频任务验证视频是否完整观看，分享任务验证分享是否成功等
	// 目前实现一个简单的示例验证

	// 示例：验证任务参数
	if task.TaskName == "每日签到" {
		// 签到任务不需要额外验证
		return nil
	} else if task.TaskName == "观看视频" {
		// 验证视频ID和观看时长
		if _, ok := extraData["video_id"].(string); !ok {
			return errors.New(errors.ErrCodeInvalidParams, "缺少视频ID", nil)
		}
		if _, ok := extraData["watch_duration"].(float64); !ok {
			return errors.New(errors.ErrCodeInvalidParams, "缺少观看时长", nil)
		}
		// 这里可以添加更多验证逻辑，如检查视频是否存在、观看时长是否足够等
	}

	return nil
}

func (s *TaskService) getUserToken(ctx context.Context, tx *gorm.DB, userID string) (int64, error) {
	return model.GetUserTokenBalance(tx, userID)
}

// grantTaskReward 发放任务奖励
func (s *TaskService) grantTaskReward(ctx context.Context, tx *gorm.DB, userID string, task *model.RewardTask) error {
	// 获取用户信息并加行锁
	var user model.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, userID).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "获取用户信息失败", err)
	}

	// 使用 UpdateUserTokenBalance 更新用户代币余额
	if err := model.UpdateUserTokenBalance(tx, userID, task.TokenReward); err != nil {
		return errors.New(errors.ErrCodeInternal, "更新用户代币余额失败", err)
	}

	// 创建代币记录
	tokenRecord := &model.TokenRecord{
		UserID:       userID,
		ChangeAmount: task.TokenReward,
		BalanceAfter: user.TokenBalance + task.TokenReward,
		ChangeType:   "TASK_REWARD",
		TaskID:       &task.TaskID,
		Remark:       &task.TaskName,
		ChangeTime:   time.Now(),
	}
	if err := model.CreateTokenRecord(tx, tokenRecord); err != nil {
		return errors.New(errors.ErrCodeInternal, "创建代币记录失败", err)
	}

	return nil
}

// recordTaskCompletion 记录任务完成
func (s *TaskService) recordTaskCompletion(ctx context.Context, tx *gorm.DB, userID string, task *model.RewardTask) error {
	// 创建任务完成记录
	record := &model.TaskCompletionRecord{
		UserID:      userID,
		TaskID:      task.TaskID,
		TokenReward: task.TokenReward,
		CompletedAt: time.Now(),
	}

	if err := tx.Create(record).Error; err != nil {
		return errors.New(errors.ErrCodeInternal, "创建任务完成记录失败", err)
	}

	// 发送任务完成通知
	go s.sendTaskCompletionNotification(ctx, userID, task)

	return nil
}

// sendTaskCompletionNotification 发送任务完成通知
func (s *TaskService) sendTaskCompletionNotification(ctx context.Context, userID string, task *model.RewardTask) {
	// 获取用户信息
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return
	}

	// 构建通知内容
	notification := &model.Notification{
		UserID:    userID,
		Type:      "TASK_COMPLETION",
		Title:     "任务完成通知",
		Content:   fmt.Sprintf("恭喜您完成了任务「%s」，获得 %d 代币奖励！", task.TaskName, task.TokenReward),
		Status:    0, // 未读
		CreatedAt: time.Now(),
	}

	// 保存通知
	if err := s.db.Create(notification).Error; err != nil {
		s.logger.Error("创建任务完成通知失败", zap.Error(err))
		return
	}

	// TODO: 如果需要消息推送，可以在这里添加推送逻辑
	// 例如：推送到消息队列，由消息服务处理推送
}

// getLastTaskCompletion 获取最后一次任务完成时间
func (s *TaskService) getLastTaskCompletion(ctx context.Context, tx *gorm.DB, userID string, taskID int) (*time.Time, error) {
	var record model.TaskCompletionRecord
	err := tx.Where("user_id = ? AND task_id = ?",
		userID, taskID).
		Order("completed_at DESC").
		First(&record).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.New(errors.ErrCodeInternal, "获取任务完成记录失败", err)
	}
	return &record.CompletedAt, nil
}

// hasCompletedTask 检查是否已完成任务
func (s *TaskService) hasCompletedTask(ctx context.Context, tx *gorm.DB, userID string, taskID int) (bool, error) {
	var count int64
	err := tx.Model(&model.TaskCompletionRecord{}).
		Where("user_id = ? AND task_id = ?",
			userID, taskID).
		Count(&count).Error
	if err != nil {
		return false, errors.New(errors.ErrCodeInternal, "检查任务完成状态失败", err)
	}
	return count > 0, nil
}

// TaskCompletionRecord 任务完成记录
type TaskCompletionRecord struct {
	RecordID    int64     `json:"record_id"`
	UserID      int64     `json:"user_id"`
	TaskID      int       `json:"task_id"`
	TaskName    string    `json:"task_name"`
	TokenReward int       `json:"token_reward"`
	CompletedAt time.Time `json:"completed_at"`
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TaskID           int    `json:"task_id"`
	TaskName         string `json:"task_name"`
	TotalCompletions int64  `json:"total_completions"`
	TotalRewards     int64  `json:"total_rewards"`
	TodayCompletions int64  `json:"today_completions"`
	TodayRewards     int64  `json:"today_rewards"`
}

// GetUserTaskRecords 获取用户任务完成记录
func (s *TaskService) GetUserTaskRecords(ctx context.Context, userID string, page, pageSize int) ([]*TaskCompletionRecord, int64, error) {
	var records []*TaskCompletionRecord
	var total int64

	// 查询总记录数
	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取任务记录总数失败", err)
	}

	// 查询任务记录
	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Select("task_completion_records.id as record_id, task_completion_records.user_id, task_completion_records.task_id, "+
			"reward_tasks.task_name, task_completion_records.token_reward, "+
			"task_completion_records.completed_at").
		Joins("LEFT JOIN reward_tasks ON task_completion_records.task_id = reward_tasks.task_id").
		Where("task_completion_records.user_id = ?", userID).
		Order("task_completion_records.completed_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&records).Error; err != nil {
		return nil, 0, errors.New(errors.ErrCodeInternal, "获取任务记录失败", err)
	}

	return records, total, nil
}

// GetTaskStatistics 获取任务统计信息
func (s *TaskService) GetTaskStatistics(ctx context.Context, taskID int) (*TaskStatistics, error) {
	var stats TaskStatistics
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 获取任务信息
	task, err := s.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	stats.TaskID = task.TaskID
	stats.TaskName = task.TaskName

	// 查询总完成次数和奖励
	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Where("task_id = ?", taskID).
		Select("COUNT(*) as total_completions, SUM(token_reward) as total_rewards").
		Scan(&stats).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取任务统计信息失败", err)
	}

	// 查询今日完成次数和奖励
	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Where("task_id = ? AND completed_at >= ?", taskID, today).
		Select("COUNT(*) as today_completions, SUM(token_reward) as today_rewards").
		Scan(&stats).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取今日任务统计信息失败", err)
	}

	return &stats, nil
}

// GetUserTaskStatistics 获取用户任务统计信息
func (s *TaskService) GetUserTaskStatistics(ctx context.Context, userID int64) (map[int]*TaskStatistics, error) {
	// 获取用户完成过的所有任务
	var taskIDs []int
	if err := s.db.Model(&model.TaskCompletionRecord{}).
		Where("user_id = ?", userID).
		Distinct("task_id").
		Pluck("task_id", &taskIDs).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取用户任务列表失败", err)
	}

	// 获取每个任务的统计信息
	stats := make(map[int]*TaskStatistics)
	for _, taskID := range taskIDs {
		taskStats, err := s.GetTaskStatistics(ctx, taskID)
		if err != nil {
			return nil, err
		}
		stats[taskID] = taskStats
	}

	return stats, nil
}

// boolToInt8 将 bool 转换为 int8
func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// UserTaskStatus 用户任务状态信息
type UserTaskStatus struct {
	TaskID              int        `json:"task_id"`               // 任务ID
	TaskKey             string     `json:"task_key"`              // 任务唯一标识
	TaskName            string     `json:"task_name"`             // 任务名称
	TaskDesc            *string    `json:"task_desc"`             // 任务描述
	TokenReward         int        `json:"token_reward"`          // 完成一次任务获得的代币数
	DailyLimit          int        `json:"daily_limit"`           // 每日奖励上限
	IntervalSeconds     int        `json:"interval_seconds"`      // 两次完成任务的最小间隔秒数
	Repeatable          int8       `json:"repeatable"`            // 是否可重复完成
	Status              int8       `json:"status"`                // 任务状态
	CanComplete         bool       `json:"can_complete"`          // 当前是否可以完成
	TodayCompletedCount int64      `json:"today_completed_count"` // 今日已完成次数
	TotalCompletedCount int64      `json:"total_completed_count"` // 总完成次数
	NextAvailableTime   *time.Time `json:"next_available_time"`   // 下次可完成时间
	CompletionReason    string     `json:"completion_reason"`     // 不能完成的原因
	ValidFrom           *time.Time `json:"valid_from"`            // 任务生效时间
	ValidTo             *time.Time `json:"valid_to"`              // 任务截止时间
}

// GetUserAvailableTasksWithStatus 获取用户可用任务列表（包含详细状态信息）
func (s *TaskService) GetUserAvailableTasksWithStatus(ctx context.Context, userID string) ([]*UserTaskStatus, error) {
	var tasks []*model.RewardTask
	now := time.Now()

	// 获取所有启用的任务
	if err := s.db.Where("status = 1 AND (valid_from IS NULL OR valid_from <= ?) AND (valid_to IS NULL OR valid_to >= ?)",
		now, now).Find(&tasks).Error; err != nil {
		return nil, errors.New(errors.ErrCodeInternal, "获取可用任务失败", err)
	}

	var result []*UserTaskStatus
	for _, task := range tasks {
		userTaskStatus := &UserTaskStatus{
			TaskID:          task.TaskID,
			TaskKey:         task.TaskKey,
			TaskName:        task.TaskName,
			TaskDesc:        task.TaskDesc,
			TokenReward:     task.TokenReward,
			DailyLimit:      task.DailyLimit,
			IntervalSeconds: task.IntervalSeconds,
			Repeatable:      task.Repeatable,
			Status:          task.Status,
			ValidFrom:       task.ValidFrom,
			ValidTo:         task.ValidTo,
			CanComplete:     true, // 默认可以完成
		}

		// 获取今日完成次数
		todayCount, err := s.getUserTaskCompletionCount(ctx, userID, task.TaskID)
		if err != nil {
			return nil, err
		}
		userTaskStatus.TodayCompletedCount = todayCount

		// 获取总完成次数
		totalCount, err := s.getUserTotalTaskCompletionCount(ctx, userID, task.TaskID)
		if err != nil {
			return nil, err
		}
		userTaskStatus.TotalCompletedCount = totalCount

		// 检查每日限制
		if task.DailyLimit > 0 && todayCount >= int64(task.DailyLimit) {
			userTaskStatus.CanComplete = false
			userTaskStatus.CompletionReason = "今日任务完成次数已达上限"
			// 计算下次可完成时间（明天0点）
			tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			userTaskStatus.NextAvailableTime = &tomorrow
		}

		// 检查间隔时间
		if task.IntervalSeconds > 0 {
			lastCompletion, err := s.getLastTaskCompletion(ctx, s.db, userID, task.TaskID)
			if err != nil {
				return nil, err
			}
			if lastCompletion != nil {
				nextAvailableTime := lastCompletion.Add(time.Duration(task.IntervalSeconds) * time.Second)
				if now.Before(nextAvailableTime) {
					userTaskStatus.CanComplete = false
					userTaskStatus.CompletionReason = "任务冷却中"
					userTaskStatus.NextAvailableTime = &nextAvailableTime
				}
			}
		}

		// 检查是否可重复完成
		if task.Repeatable == 0 {
			completed, err := s.hasCompletedTask(ctx, s.db, userID, task.TaskID)
			if err != nil {
				return nil, err
			}
			if completed {
				userTaskStatus.CanComplete = false
				userTaskStatus.CompletionReason = "该任务已完成且不可重复完成"
			}
		}

		// 如果任务不能完成且没有设置下次可完成时间，设置一个默认值
		if !userTaskStatus.CanComplete && userTaskStatus.NextAvailableTime == nil {
			// 对于每日限制的任务，设置为明天0点
			if task.DailyLimit > 0 && todayCount >= int64(task.DailyLimit) {
				tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
				userTaskStatus.NextAvailableTime = &tomorrow
			}
		}

		result = append(result, userTaskStatus)
	}

	return result, nil
}

// TaskOverview 任务概览信息
type TaskOverview struct {
	TotalTasks          int64             `json:"total_tasks"`            // 总任务数
	AvailableTasks      int64             `json:"available_tasks"`        // 可完成任务数
	CompletedToday      int64             `json:"completed_today"`        // 今日已完成任务数
	TotalRewardsToday   int64             `json:"total_rewards_today"`    // 今日获得奖励总数
	TotalRewardsAllTime int64             `json:"total_rewards_all_time"` // 总获得奖励数
	Tasks               []*UserTaskStatus `json:"tasks"`                  // 任务详细列表
}

// GetUserTaskOverview 获取用户任务概览信息
func (s *TaskService) GetUserTaskOverview(ctx context.Context, userID string) (*TaskOverview, error) {
	// 获取所有任务状态
	tasks, err := s.GetUserAvailableTasksWithStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	overview := &TaskOverview{
		TotalTasks:     int64(len(tasks)),
		AvailableTasks: 0,
		CompletedToday: 0,
		Tasks:          tasks,
	}

	// 计算统计信息
	for _, task := range tasks {
		if task.CanComplete {
			overview.AvailableTasks++
		}
		if task.TodayCompletedCount > 0 {
			overview.CompletedToday++
		}
		overview.TotalRewardsToday += task.TodayCompletedCount * int64(task.TokenReward)
		overview.TotalRewardsAllTime += task.TotalCompletedCount * int64(task.TokenReward)
	}

	return overview, nil
}
