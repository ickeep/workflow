// Package biz 提供业务逻辑层功能
// 包含流程实例管理的核心业务逻辑
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

// ProcessInstanceUseCase 流程实例用例，包含流程实例相关的业务逻辑
type ProcessInstanceUseCase struct {
	processInstanceRepo ProcessInstanceRepo
	processDefRepo      ProcessDefinitionRepo
	variableRepo        ProcessVariableRepo
	cache               CacheRepo
	logger              *zap.Logger
}

// NewProcessInstanceUseCase 创建流程实例用例实例
func NewProcessInstanceUseCase(
	processInstanceRepo ProcessInstanceRepo,
	processDefRepo ProcessDefinitionRepo,
	variableRepo ProcessVariableRepo,
	cache CacheRepo,
	logger *zap.Logger,
) *ProcessInstanceUseCase {
	return &ProcessInstanceUseCase{
		processInstanceRepo: processInstanceRepo,
		processDefRepo:      processDefRepo,
		variableRepo:        variableRepo,
		cache:               cache,
		logger:              logger,
	}
}

// StartProcessInstance 启动流程实例
func (uc *ProcessInstanceUseCase) StartProcessInstance(ctx context.Context, req *StartProcessInstanceRequest) (*ProcessInstanceResponse, error) {
	uc.logger.Info("开始启动流程实例",
		zap.String("process_definition_id", req.ProcessDefinitionID),
		zap.String("process_definition_key", req.ProcessDefinitionKey),
		zap.String("business_key", req.BusinessKey))

	// 验证请求参数
	if err := uc.validateStartRequest(req); err != nil {
		uc.logger.Error("启动流程实例参数验证失败", zap.Error(err))
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 获取流程定义
	var processDef *ent.ProcessDefinition
	var err error

	if req.ProcessDefinitionID != "" {
		processDef, err = uc.processDefRepo.GetByID(ctx, req.ProcessDefinitionID)
	} else if req.ProcessDefinitionKey != "" {
		processDef, err = uc.processDefRepo.GetLatestByKey(ctx, req.ProcessDefinitionKey)
	}

	if err != nil {
		uc.logger.Error("获取流程定义失败", zap.Error(err))
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	// 检查流程定义是否被挂起
	if processDef.Suspended {
		uc.logger.Error("流程定义已被挂起，无法启动实例",
			zap.String("process_definition_id", strconv.FormatInt(processDef.ID, 10)))
		return nil, fmt.Errorf("流程定义已被挂起，无法启动实例")
	}

	// 构建流程实例
	instance := &ent.ProcessInstance{
		ProcessDefinitionID:      processDef.ID,
		ProcessDefinitionKey:     processDef.Key,
		ProcessDefinitionName:    processDef.Name,
		ProcessDefinitionVersion: processDef.Version,
		BusinessKey:              req.BusinessKey,
		StartUserID:              uc.getCurrentUserID(ctx),
		StartTime:                time.Now(),
		Name:                     req.Name,
		Description:              req.Description,
		Suspended:                false,
		TenantID:                 req.TenantID,
	}

	// 保存流程实例
	result, err := uc.processInstanceRepo.Create(ctx, instance)
	if err != nil {
		uc.logger.Error("保存流程实例失败", zap.Error(err))
		return nil, fmt.Errorf("保存流程实例失败: %w", err)
	}

	// 保存流程变量
	if req.Variables != nil && len(req.Variables) > 0 {
		if err := uc.saveProcessVariables(ctx, result.ID, req.Variables); err != nil {
			uc.logger.Warn("保存流程变量失败", zap.Error(err))
			// 变量保存失败不影响主要流程
		}
	}

	// TODO: 集成Temporal，启动工作流执行

	// 缓存流程实例
	if err := uc.cacheProcessInstance(ctx, result); err != nil {
		uc.logger.Warn("缓存流程实例失败", zap.Error(err))
	}

	uc.logger.Info("流程实例启动成功",
		zap.String("instance_id", strconv.FormatInt(result.ID, 10)),
		zap.String("process_definition_id", strconv.FormatInt(processDef.ID, 10)))

	return uc.toProcessInstanceResponse(result, req.Variables), nil
}

// GetProcessInstance 根据ID获取流程实例
func (uc *ProcessInstanceUseCase) GetProcessInstance(ctx context.Context, id string) (*ProcessInstanceResponse, error) {
	uc.logger.Debug("获取流程实例", zap.String("id", id))

	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("process_instance:%s", id)
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var instance ent.ProcessInstance
		if err := json.Unmarshal([]byte(cached), &instance); err == nil {
			uc.logger.Debug("从缓存获取流程实例成功", zap.String("id", id))
			variables, _ := uc.getProcessVariables(ctx, instance.ID)
			return uc.toProcessInstanceResponse(&instance, variables), nil
		}
	}

	// 从数据库获取
	instance, err := uc.processInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取流程实例失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取流程实例失败: %w", err)
	}

	// 获取流程变量
	variables, err := uc.getProcessVariables(ctx, instance.ID)
	if err != nil {
		uc.logger.Warn("获取流程变量失败", zap.Error(err))
		variables = make(map[string]interface{})
	}

	// 缓存结果
	if err := uc.cacheProcessInstance(ctx, instance); err != nil {
		uc.logger.Warn("缓存流程实例失败", zap.Error(err))
	}

	return uc.toProcessInstanceResponse(instance, variables), nil
}

// ListProcessInstances 分页查询流程实例
func (uc *ProcessInstanceUseCase) ListProcessInstances(ctx context.Context, req *ListProcessInstancesRequest) (*ListProcessInstancesResponse, error) {
	uc.logger.Debug("分页查询流程实例", zap.Any("request", req))

	// 构建过滤条件
	filter := &ProcessInstanceFilter{
		ProcessDefinitionID: req.ProcessDefinitionID,
		StartedFrom:         req.StartedFrom,
		StartedTo:           req.StartedTo,
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
	instances, pagination, err := uc.processInstanceRepo.List(ctx, filter, opts)
	if err != nil {
		uc.logger.Error("查询流程实例列表失败", zap.Error(err))
		return nil, fmt.Errorf("查询流程实例列表失败: %w", err)
	}

	// 转换响应
	items := make([]*ProcessInstanceResponse, len(instances))
	for i, instance := range instances {
		// 获取流程变量
		variables, _ := uc.getProcessVariables(ctx, instance.ID)
		items[i] = uc.toProcessInstanceResponse(instance, variables)
	}

	return &ListProcessInstancesResponse{
		Items:      items,
		Pagination: pagination,
	}, nil
}

// SuspendProcessInstance 挂起流程实例
func (uc *ProcessInstanceUseCase) SuspendProcessInstance(ctx context.Context, id string) error {
	uc.logger.Info("挂起流程实例", zap.String("id", id))

	// 获取流程实例
	instance, err := uc.processInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取流程实例失败: %w", err)
	}

	// 检查实例状态
	if instance.EndTime != nil {
		return fmt.Errorf("已结束的流程实例无法挂起")
	}
	if instance.Suspended {
		return fmt.Errorf("流程实例已处于挂起状态")
	}

	// 更新实例状态
	instance.Suspended = true
	_, err = uc.processInstanceRepo.Update(ctx, instance)
	if err != nil {
		uc.logger.Error("挂起流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("挂起流程实例失败: %w", err)
	}

	// TODO: 集成Temporal，暂停工作流执行

	// 清除缓存
	cacheKey := fmt.Sprintf("process_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程实例挂起成功", zap.String("id", id))
	return nil
}

// ActivateProcessInstance 激活流程实例
func (uc *ProcessInstanceUseCase) ActivateProcessInstance(ctx context.Context, id string) error {
	uc.logger.Info("激活流程实例", zap.String("id", id))

	// 获取流程实例
	instance, err := uc.processInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取流程实例失败: %w", err)
	}

	// 检查实例状态
	if instance.EndTime != nil {
		return fmt.Errorf("已结束的流程实例无法激活")
	}
	if !instance.Suspended {
		return fmt.Errorf("流程实例已处于激活状态")
	}

	// 更新实例状态
	instance.Suspended = false
	_, err = uc.processInstanceRepo.Update(ctx, instance)
	if err != nil {
		uc.logger.Error("激活流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("激活流程实例失败: %w", err)
	}

	// TODO: 集成Temporal，恢复工作流执行

	// 清除缓存
	cacheKey := fmt.Sprintf("process_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程实例激活成功", zap.String("id", id))
	return nil
}

// TerminateProcessInstance 终止流程实例
func (uc *ProcessInstanceUseCase) TerminateProcessInstance(ctx context.Context, id, reason string) error {
	uc.logger.Info("终止流程实例", zap.String("id", id), zap.String("reason", reason))

	// 获取流程实例
	instance, err := uc.processInstanceRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("获取流程实例失败: %w", err)
	}

	// 检查实例状态
	if instance.EndTime != nil {
		return fmt.Errorf("流程实例已经结束")
	}

	// 更新实例状态
	now := time.Now()
	instance.EndTime = &now
	instance.DeleteReason = reason
	if instance.StartTime.Before(now) {
		duration := now.Sub(instance.StartTime).Milliseconds()
		instance.Duration = duration
	}

	_, err = uc.processInstanceRepo.Update(ctx, instance)
	if err != nil {
		uc.logger.Error("终止流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("终止流程实例失败: %w", err)
	}

	// TODO: 集成Temporal，终止工作流执行

	// 清除缓存
	cacheKey := fmt.Sprintf("process_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程实例终止成功", zap.String("id", id))
	return nil
}

// DeleteProcessInstance 删除流程实例
func (uc *ProcessInstanceUseCase) DeleteProcessInstance(ctx context.Context, id, reason string) error {
	uc.logger.Info("删除流程实例", zap.String("id", id), zap.String("reason", reason))

	// 先终止流程实例（如果还在运行）
	if err := uc.TerminateProcessInstance(ctx, id, reason); err != nil {
		// 如果终止失败，检查是否是因为实例已经结束
		instance, getErr := uc.processInstanceRepo.GetByID(ctx, id)
		if getErr != nil || instance.EndTime == nil {
			return fmt.Errorf("终止流程实例失败: %w", err)
		}
	}

	// 删除流程实例
	if err := uc.processInstanceRepo.Delete(ctx, id); err != nil {
		uc.logger.Error("删除流程实例失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("删除流程实例失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("process_instance:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程实例缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程实例删除成功", zap.String("id", id))
	return nil
}

// validateStartRequest 验证启动请求参数
func (uc *ProcessInstanceUseCase) validateStartRequest(req *StartProcessInstanceRequest) error {
	if req.ProcessDefinitionID == "" && req.ProcessDefinitionKey == "" {
		return fmt.Errorf("流程定义ID或流程定义键必须提供其中之一")
	}
	if req.ProcessDefinitionID != "" && req.ProcessDefinitionKey != "" {
		return fmt.Errorf("流程定义ID和流程定义键不能同时提供")
	}
	return nil
}

// saveProcessVariables 保存流程变量
func (uc *ProcessInstanceUseCase) saveProcessVariables(ctx context.Context, instanceID int64, variables map[string]interface{}) error {
	for name, value := range variables {
		variable := &ent.ProcessVariable{
			ProcessInstanceID: instanceID,
			Name:              name,
			Type:              uc.getVariableType(value),
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
				uc.logger.Warn("序列化流程变量失败",
					zap.String("name", name),
					zap.Any("value", value),
					zap.Error(err))
				continue
			}
			variable.TextValue = string(valueBytes)
		}

		if _, err := uc.variableRepo.Create(ctx, variable); err != nil {
			uc.logger.Warn("保存流程变量失败",
				zap.String("name", name),
				zap.Error(err))
		}
	}
	return nil
}

// GetProcessVariables 获取流程实例的所有变量 (公共方法)
func (uc *ProcessInstanceUseCase) GetProcessVariables(ctx context.Context, instanceID string) (map[string]interface{}, error) {
	id, err := strconv.ParseInt(instanceID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的流程实例ID: %s", instanceID)
	}
	return uc.getProcessVariables(ctx, id)
}

// SetProcessVariables 批量设置流程变量 (公共方法)
func (uc *ProcessInstanceUseCase) SetProcessVariables(ctx context.Context, instanceID string, variables map[string]interface{}) error {
	id, err := strconv.ParseInt(instanceID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的流程实例ID: %s", instanceID)
	}
	return uc.saveProcessVariables(ctx, id, variables)
}

// GetProcessVariable 获取单个流程变量 (公共方法)
func (uc *ProcessInstanceUseCase) GetProcessVariable(ctx context.Context, instanceID string, variableName string) (interface{}, error) {
	variables, err := uc.GetProcessVariables(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	value, exists := variables[variableName]
	if !exists {
		return nil, fmt.Errorf("变量 %s 不存在", variableName)
	}

	return value, nil
}

// SetProcessVariable 设置单个流程变量 (公共方法)
func (uc *ProcessInstanceUseCase) SetProcessVariable(ctx context.Context, instanceID string, variableName string, value interface{}) error {
	variables := map[string]interface{}{
		variableName: value,
	}
	return uc.SetProcessVariables(ctx, instanceID, variables)
}

// getProcessVariables 获取流程变量 (私有方法)
func (uc *ProcessInstanceUseCase) getProcessVariables(ctx context.Context, instanceID int64) (map[string]interface{}, error) {
	variables, err := uc.variableRepo.ListByProcessInstanceID(ctx, strconv.FormatInt(instanceID, 10))
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, variable := range variables {
		var value interface{}

		switch variable.Type {
		case "string":
			value = variable.TextValue
		case "integer":
			value = variable.LongValue
		case "double":
			value = variable.DoubleValue
		case "boolean":
			value = variable.TextValue == "true"
		case "json":
			if variable.TextValue != "" {
				if err := json.Unmarshal([]byte(variable.TextValue), &value); err != nil {
					uc.logger.Warn("反序列化流程变量失败",
						zap.String("name", variable.Name),
						zap.String("value", variable.TextValue),
						zap.Error(err))
					continue
				}
			}
		default:
			value = variable.TextValue
		}

		result[variable.Name] = value
	}

	return result, nil
}

// getVariableType 获取变量类型
func (uc *ProcessInstanceUseCase) getVariableType(value interface{}) string {
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
func (uc *ProcessInstanceUseCase) getCurrentUserID(ctx context.Context) string {
	// TODO: 从JWT token或上下文中获取当前用户ID
	// 这里先返回一个默认值
	return "system"
}

// cacheProcessInstance 缓存流程实例
func (uc *ProcessInstanceUseCase) cacheProcessInstance(ctx context.Context, instance *ent.ProcessInstance) error {
	cacheKey := fmt.Sprintf("process_instance:%s", strconv.FormatInt(instance.ID, 10))
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("序列化流程实例失败: %w", err)
	}

	// 缓存30分钟
	return uc.cache.Set(ctx, cacheKey, string(data), 30*time.Minute)
}

// toProcessInstanceResponse 转换为响应格式
func (uc *ProcessInstanceUseCase) toProcessInstanceResponse(instance *ent.ProcessInstance, variables map[string]interface{}) *ProcessInstanceResponse {
	// 计算状态
	isActive := instance.EndTime == nil && !instance.Suspended
	isEnded := instance.EndTime != nil
	isSuspended := instance.Suspended

	// 计算持续时间
	var duration *int64
	if instance.Duration != 0 {
		duration = &instance.Duration
	}

	return &ProcessInstanceResponse{
		ID:                  strconv.FormatInt(instance.ID, 10),
		ProcessDefinitionID: strconv.FormatInt(instance.ProcessDefinitionID, 10),
		BusinessKey:         instance.BusinessKey,
		StartUserID:         instance.StartUserID,
		StartTime:           instance.StartTime,
		EndTime:             instance.EndTime,
		Duration:            duration,
		DeleteReason:        instance.DeleteReason,
		ActivityID:          "", // TODO: 从执行状态中获取
		Name:                instance.Name,
		Description:         instance.Description,
		IsActive:            isActive,
		IsEnded:             isEnded,
		IsSuspended:         isSuspended,
		TenantID:            instance.TenantID,
		Variables:           variables,
		CreatedAt:           instance.CreatedAt,
		UpdatedAt:           instance.UpdatedAt,
	}
}
