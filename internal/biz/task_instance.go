// Package biz 提供业务逻辑层功能
// 包含任务实例管理的核心业务逻辑
package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/workflow-engine/workflow-engine/internal/data/ent"

	"go.uber.org/zap"
)

// TaskInstanceUseCase 任务实例用例，包含任务实例相关的业务逻辑
type TaskInstanceUseCase struct {
	taskInstanceRepo    TaskInstanceRepo
	processInstanceRepo ProcessInstanceRepo
	variableRepo        ProcessVariableRepo
	cache               CacheRepo
	logger              *zap.Logger
}

// NewTaskInstanceUseCase 创建任务实例用例实例
func NewTaskInstanceUseCase(
	taskInstanceRepo TaskInstanceRepo,
	processInstanceRepo ProcessInstanceRepo,
	variableRepo ProcessVariableRepo,
	cache CacheRepo,
	logger *zap.Logger,
) *TaskInstanceUseCase {
	return &TaskInstanceUseCase{
		taskInstanceRepo:    taskInstanceRepo,
		processInstanceRepo: processInstanceRepo,
		variableRepo:        variableRepo,
		cache:               cache,
		logger:              logger,
	}
}

// GetTaskInstance 根据ID获取任务实例
func (uc *TaskInstanceUseCase) GetTaskInstance(ctx context.Context, id string) (*TaskInstanceResponse, error) {
	uc.logger.Debug("获取任务实例", zap.String("id", id))

	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("task_instance:%s", id)
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var task ent.TaskInstance
		if err := json.Unmarshal([]byte(cached), &task); err == nil {
			uc.logger.Debug("从缓存获取任务实例成功", zap.String("id", id))
			variables, _ := uc.getTaskVariables(ctx, task.ID)
			return uc.toTaskInstanceResponse(&task, variables), nil
		}
	}

	// 从数据库获取
	task, err := uc.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取任务实例失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取任务实例失败: %w", err)
	}

	// 获取任务变量
	variables, err := uc.getTaskVariables(ctx, task.ID)
	if err != nil {
		uc.logger.Warn("获取任务变量失败", zap.Error(err))
		variables = make(map[string]interface{})
	}

	// 缓存结果
	if err := uc.cacheTaskInstance(ctx, task); err != nil {
		uc.logger.Warn("缓存任务实例失败", zap.Error(err))
	}

	return uc.toTaskInstanceResponse(task, variables), nil
}

// ListTaskInstances 分页查询任务实例
func (uc *TaskInstanceUseCase) ListTaskInstances(ctx context.Context, req *ListTaskInstancesRequest) (*ListTaskInstancesResponse, error) {
	uc.logger.Debug("分页查询任务实例", zap.Any("request", req))

	// 构建过滤条件
	filter := &TaskInstanceFilter{
		ProcessInstanceID: req.ProcessInstanceID,
		AssigneeID:        req.AssigneeID,
		CreatedFrom:       req.CreatedFrom,
		CreatedTo:         req.CreatedTo,
	}

	// 构建查询选项
	opts := &QueryOptions{
		Page:     req.Page,
		PageSize: req.PageSize,
		OrderBy:  req.OrderBy,
		Order:    req.Order,
		Search:   req.Search,
	}

	// 查询数据
	tasks, pagination, err := uc.taskInstanceRepo.List(ctx, filter, opts)
	if err != nil {
		uc.logger.Error("查询任务实例列表失败", zap.Error(err))
		return nil, fmt.Errorf("查询任务实例列表失败: %w", err)
	}

	// 转换响应
	items := make([]*TaskInstanceResponse, len(tasks))
	for i, task := range tasks {
		// 获取任务变量
		variables, _ := uc.getTaskVariables(ctx, task.ID)
		items[i] = uc.toTaskInstanceResponse(task, variables)
	}

	return &ListTaskInstancesResponse{
		Items:      items,
		Pagination: pagination,
	}, nil
}

// ClaimTask 认领任务
func (uc *TaskInstanceUseCase) ClaimTask(ctx context.Context, id string, req *ClaimTaskRequest) error {
	uc.logger.Info("认领任务", zap.String("id", id), zap.String("assignee_id", req.AssigneeID))

	// 获取任务实例
	task, err := uc.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取任务实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取任务实例失败: %w", err)
	}

	// 检查任务状态
	if uc.isTaskCompleted(task) {
		return fmt.Errorf("任务已完成，无法认领")
	}
	if task.Assignee != "" && task.Assignee != req.AssigneeID {
		return fmt.Errorf("任务已被其他用户认领")
	}

	// 认领任务
	if err := uc.taskInstanceRepo.Claim(ctx, id, req.AssigneeID); err != nil {
		uc.logger.Error("认领任务失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("认领任务失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("task_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除任务实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("任务认领成功", zap.String("id", id), zap.String("assignee_id", req.AssigneeID))
	return nil
}

// CompleteTask 完成任务
func (uc *TaskInstanceUseCase) CompleteTask(ctx context.Context, id string, req *CompleteTaskRequest) error {
	uc.logger.Info("完成任务", zap.String("id", id))

	// 获取任务实例
	task, err := uc.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取任务实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取任务实例失败: %w", err)
	}

	// 检查任务状态
	if uc.isTaskCompleted(task) {
		return fmt.Errorf("任务已完成")
	}

	// 检查任务是否已被认领
	currentUserID := uc.getCurrentUserID(ctx)
	if task.Assignee == "" {
		return fmt.Errorf("任务未被认领，请先认领任务")
	}
	if task.Assignee != currentUserID {
		return fmt.Errorf("只有任务认领人才能完成任务")
	}

	// 保存任务变量
	if req.Variables != nil && len(req.Variables) > 0 {
		if err := uc.saveTaskVariables(ctx, task.ID, req.Variables); err != nil {
			uc.logger.Warn("保存任务变量失败", zap.Error(err))
		}
	}

	// 完成任务
	if err := uc.taskInstanceRepo.Complete(ctx, id, req.Variables); err != nil {
		uc.logger.Error("完成任务失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("完成任务失败: %w", err)
	}

	// TODO: 集成Temporal，推进工作流执行

	// 清除缓存
	cacheKey := fmt.Sprintf("task_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除任务实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("任务完成成功", zap.String("id", id))
	return nil
}

// DelegateTask 委派任务
func (uc *TaskInstanceUseCase) DelegateTask(ctx context.Context, id string, req *DelegateTaskRequest) error {
	uc.logger.Info("委派任务", zap.String("id", id), zap.String("delegate_id", req.DelegateID))

	// 获取任务实例
	task, err := uc.taskInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取任务实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取任务实例失败: %w", err)
	}

	// 检查任务状态
	if uc.isTaskCompleted(task) {
		return fmt.Errorf("任务已完成，无法委派")
	}

	// 检查委派权限
	currentUserID := uc.getCurrentUserID(ctx)
	if task.Assignee != currentUserID && task.Owner != currentUserID {
		return fmt.Errorf("只有任务的认领人或拥有者才能委派任务")
	}

	// 委派任务
	if err := uc.taskInstanceRepo.Delegate(ctx, id, req.DelegateID); err != nil {
		uc.logger.Error("委派任务失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("委派任务失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("task_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除任务实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("任务委派成功", zap.String("id", id), zap.String("delegate_id", req.DelegateID))
	return nil
}

// GetMyTasks 获取当前用户的任务列表
func (uc *TaskInstanceUseCase) GetMyTasks(ctx context.Context, req *ListTaskInstancesRequest) (*ListTaskInstancesResponse, error) {
	uc.logger.Debug("获取当前用户的任务列表")

	// 获取当前用户ID
	currentUserID := uc.getCurrentUserID(ctx)

	// 设置按执行人过滤
	req.AssigneeID = currentUserID

	return uc.ListTaskInstances(ctx, req)
}

// GetAvailableTasks 获取可认领的任务列表
func (uc *TaskInstanceUseCase) GetAvailableTasks(ctx context.Context, req *ListTaskInstancesRequest) (*ListTaskInstancesResponse, error) {
	uc.logger.Debug("获取可认领的任务列表")

	// 清除执行人过滤条件，获取未分配的任务
	req.AssigneeID = ""

	return uc.ListTaskInstances(ctx, req)
}

// saveTaskVariables 保存任务变量
func (uc *TaskInstanceUseCase) saveTaskVariables(ctx context.Context, taskID int64, variables map[string]interface{}) error {
	for name, value := range variables {
		variable := &ent.ProcessVariable{
			TaskID: taskID,
			Name:   name,
			Type:   uc.getVariableType(value),
		}

		// 根据变量类型设置相应的值字段
		switch v := value.(type) {
		case string:
			variable.TextValue = v
		case int, int32, int64:
			if intVal, ok := value.(int64); ok {
				variable.LongValue = intVal
			} else if intVal, ok := value.(int); ok {
				variable.LongValue = int64(intVal)
			} else if intVal, ok := value.(int32); ok {
				variable.LongValue = int64(intVal)
			}
		case float32, float64:
			if floatVal, ok := value.(float64); ok {
				variable.DoubleValue = floatVal
			} else if floatVal, ok := value.(float32); ok {
				variable.DoubleValue = float64(floatVal)
			}
		default:
			// 复杂类型序列化为JSON存储在TextValue中
			valueBytes, err := json.Marshal(value)
			if err != nil {
				uc.logger.Warn("序列化任务变量失败",
					zap.String("name", name),
					zap.Any("value", value),
					zap.Error(err))
				continue
			}
			variable.TextValue = string(valueBytes)
		}

		if _, err := uc.variableRepo.Create(ctx, variable); err != nil {
			uc.logger.Warn("保存任务变量失败",
				zap.String("name", name),
				zap.Error(err))
		}
	}
	return nil
}

// getTaskVariables 获取任务变量
func (uc *TaskInstanceUseCase) getTaskVariables(ctx context.Context, taskID int64) (map[string]interface{}, error) {
	// TODO: 实现获取任务变量的逻辑
	// 这里需要在ProcessVariableRepo中添加根据任务ID查询变量的方法
	return make(map[string]interface{}), nil
}

// getVariableType 获取变量类型
func (uc *TaskInstanceUseCase) getVariableType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "integer"
	case float32, float64:
		return "double"
	case bool:
		return "boolean"
	case time.Time:
		return "date"
	default:
		return "json"
	}
}

// getCurrentUserID 获取当前用户ID (从上下文中获取)
func (uc *TaskInstanceUseCase) getCurrentUserID(ctx context.Context) string {
	// TODO: 从JWT token或上下文中获取当前用户ID
	// 这里先返回一个默认值
	return "system"
}

// isTaskCompleted 检查任务是否已完成
func (uc *TaskInstanceUseCase) isTaskCompleted(task *ent.TaskInstance) bool {
	// TODO: 根据实际的TaskInstance结构检查任务状态
	// 这里需要根据Ent schema的实际字段来实现
	return false // 临时实现
}

// cacheTaskInstance 缓存任务实例
func (uc *TaskInstanceUseCase) cacheTaskInstance(ctx context.Context, task *ent.TaskInstance) error {
	cacheKey := fmt.Sprintf("task_instance:%s", strconv.FormatInt(task.ID, 10))
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化任务实例失败: %w", err)
	}

	// 缓存15分钟
	return uc.cache.Set(ctx, cacheKey, string(data), 15*time.Minute)
}

// toTaskInstanceResponse 转换为响应格式
func (uc *TaskInstanceUseCase) toTaskInstanceResponse(task *ent.TaskInstance, variables map[string]interface{}) *TaskInstanceResponse {
	// TODO: 根据实际的TaskInstance结构实现转换
	// 这里需要根据Ent schema的实际字段来实现
	return &TaskInstanceResponse{
		ID:                  strconv.FormatInt(task.ID, 10),
		ProcessInstanceID:   strconv.FormatInt(task.ProcessInstanceID, 10),
		ProcessDefinitionID: strconv.FormatInt(task.ProcessDefinitionID, 10),
		Name:                task.Name,
		Description:         task.Description,
		TaskDefinitionKey:   task.TaskDefinitionKey,
		Priority:            task.Priority,
		CreateTime:          task.CreateTime,
		ClaimTime:           nil, // TODO: 实现认领时间字段
		DueDate:             task.DueDate,
		Category:            task.Category,
		Owner:               task.Owner,
		Assignee:            task.Assignee,
		Delegation:          task.Delegation,
		FormKey:             task.FormKey,
		IsSuspended:         task.Suspended,
		TenantID:            task.TenantID,
		Variables:           variables,
		CreatedAt:           task.CreatedAt,
		UpdatedAt:           task.UpdatedAt,
	}
}
