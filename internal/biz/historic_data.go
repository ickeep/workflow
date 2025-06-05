// Package biz 历史数据管理业务逻辑层
// 提供历史流程实例查询、统计分析和报表功能
package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// HistoricDataUseCase 历史数据用例
// 负责处理历史流程实例、任务的查询、统计和分析业务逻辑
type HistoricDataUseCase struct {
	historicRepo HistoricProcessInstanceRepo
	cache        CacheRepo
	logger       *zap.Logger
}

// NewHistoricDataUseCase 创建历史数据用例
func NewHistoricDataUseCase(
	historicRepo HistoricProcessInstanceRepo,
	cache CacheRepo,
	logger *zap.Logger,
) *HistoricDataUseCase {
	return &HistoricDataUseCase{
		historicRepo: historicRepo,
		cache:        cache,
		logger:       logger,
	}
}

// GetHistoricProcessInstance 获取历史流程实例详情
// 根据实例ID获取历史流程实例的详细信息
func (uc *HistoricDataUseCase) GetHistoricProcessInstance(ctx context.Context, instanceID int64) (*HistoricProcessInstanceResponse, error) {
	uc.logger.Debug("获取历史流程实例", zap.Int64("instanceID", instanceID))

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("historic_process_instance:%d", instanceID)
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var response HistoricProcessInstanceResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			uc.logger.Debug("从缓存获取历史流程实例成功")
			return &response, nil
		}
	}

	// 从数据库获取
	instance, err := uc.historicRepo.GetHistoricProcessInstance(ctx, instanceID)
	if err != nil {
		uc.logger.Error("获取历史流程实例失败", zap.Error(err))
		return nil, fmt.Errorf("获取历史流程实例失败: %w", err)
	}

	// 转换为响应对象
	response := &HistoricProcessInstanceResponse{
		ID:                   strconv.FormatInt(instance.ID, 10),
		ProcessDefinitionID:  strconv.FormatInt(instance.ProcessDefinitionID, 10),
		ProcessDefinitionKey: instance.ProcessDefinitionKey,
		BusinessKey:          instance.BusinessKey,
		StartTime:            instance.StartTime,
		EndTime:              instance.EndTime,
		Duration:             uc.calculateDuration(instance.StartTime, instance.EndTime),
		StartUserID:          instance.StartUserID,
		DeleteReason:         instance.DeleteReason,
		TenantID:             instance.TenantID,
		CreatedAt:            instance.CreatedAt,
		UpdatedAt:            instance.UpdatedAt,
	}

	// 获取流程变量
	if variables, err := uc.historicRepo.GetHistoricVariables(ctx, instanceID); err == nil {
		response.Variables = make(map[string]interface{})
		for _, variable := range variables {
			response.Variables[variable.Name] = uc.parseVariableValue(variable)
		}
	}

	// 缓存结果（历史数据缓存时间较长）
	if data, err := json.Marshal(response); err == nil {
		uc.cache.Set(ctx, cacheKey, string(data), 2*time.Hour)
	}

	uc.logger.Info("获取历史流程实例成功", zap.Int64("instanceID", instanceID))
	return response, nil
}

// ListHistoricProcessInstances 查询历史流程实例列表
// 支持多种过滤条件和分页查询
func (uc *HistoricDataUseCase) ListHistoricProcessInstances(ctx context.Context, req *ListHistoricProcessInstancesRequest) (*ListHistoricProcessInstancesResponse, error) {
	uc.logger.Debug("查询历史流程实例列表", zap.Any("request", req))

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 构建查询过滤器
	filter := &HistoricProcessInstanceFilter{
		ProcessDefinitionID:  req.ProcessDefinitionID,
		ProcessDefinitionKey: req.ProcessDefinitionKey,
		BusinessKey:          req.BusinessKey,
		StartUserID:          req.StartUserID,
		State:                req.State,
		StartTimeAfter:       req.StartTimeAfter,
		StartTimeBefore:      req.StartTimeBefore,
		EndTimeAfter:         req.EndTimeAfter,
		EndTimeBefore:        req.EndTimeBefore,
		TenantID:             req.TenantID,
		Page:                 req.Page,
		PageSize:             req.PageSize,
		OrderBy:              req.OrderBy,
		OrderDirection:       req.OrderDirection,
	}

	// 查询历史流程实例
	instances, total, err := uc.historicRepo.ListHistoricProcessInstances(ctx, filter)
	if err != nil {
		uc.logger.Error("查询历史流程实例列表失败", zap.Error(err))
		return nil, fmt.Errorf("查询历史流程实例列表失败: %w", err)
	}

	// 转换为响应对象
	var items []*HistoricProcessInstanceResponse
	for _, instance := range instances {
		item := &HistoricProcessInstanceResponse{
			ID:                   strconv.FormatInt(instance.ID, 10),
			ProcessDefinitionID:  strconv.FormatInt(instance.ProcessDefinitionID, 10),
			ProcessDefinitionKey: instance.ProcessDefinitionKey,
			BusinessKey:          instance.BusinessKey,
			StartTime:            instance.StartTime,
			EndTime:              instance.EndTime,
			Duration:             uc.calculateDuration(instance.StartTime, instance.EndTime),
			StartUserID:          instance.StartUserID,
			DeleteReason:         instance.DeleteReason,
			TenantID:             instance.TenantID,
			CreatedAt:            instance.CreatedAt,
			UpdatedAt:            instance.UpdatedAt,
		}
		items = append(items, item)
	}

	response := &ListHistoricProcessInstancesResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	uc.logger.Info("查询历史流程实例列表成功", zap.Int("total", total))
	return response, nil
}

// GetProcessStatistics 获取流程统计信息
// 提供流程执行的统计分析数据
func (uc *HistoricDataUseCase) GetProcessStatistics(ctx context.Context, req *ProcessStatisticsRequest) (*ProcessStatisticsResponse, error) {
	uc.logger.Debug("获取流程统计信息", zap.String("processDefinitionKey", req.ProcessDefinitionKey))

	// 参数验证
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().AddDate(0, -1, 0) // 默认最近一个月
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("process_stats:%s:%d:%d", req.ProcessDefinitionKey, req.StartTime.Unix(), req.EndTime.Unix())
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var response ProcessStatisticsResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			uc.logger.Debug("从缓存获取流程统计信息")
			return &response, nil
		}
	}

	// 获取统计数据
	stats, err := uc.historicRepo.GetProcessStatistics(ctx, req.ProcessDefinitionKey, req.StartTime, req.EndTime)
	if err != nil {
		uc.logger.Error("获取流程统计信息失败", zap.Error(err))
		return nil, fmt.Errorf("获取流程统计信息失败: %w", err)
	}

	response := &ProcessStatisticsResponse{
		ProcessDefinitionKey: req.ProcessDefinitionKey,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		TotalInstances:       stats.TotalInstances,
		CompletedInstances:   stats.CompletedInstances,
		ActiveInstances:      stats.ActiveInstances,
		SuspendedInstances:   stats.SuspendedInstances,
		TerminatedInstances:  stats.TerminatedInstances,
		AverageDuration:      stats.AverageDuration,
		MinDuration:          stats.MinDuration,
		MaxDuration:          stats.MaxDuration,
	}

	// 计算完成率
	if stats.TotalInstances > 0 {
		response.CompletionRate = float64(stats.CompletedInstances) / float64(stats.TotalInstances) * 100
	}

	// 缓存结果
	if data, err := json.Marshal(response); err == nil {
		uc.cache.Set(ctx, cacheKey, string(data), 1*time.Hour)
	}

	uc.logger.Info("获取流程统计信息成功", zap.String("processDefinitionKey", req.ProcessDefinitionKey))
	return response, nil
}

// GetProcessTrend 获取流程趋势分析
// 按时间维度分析流程执行趋势
func (uc *HistoricDataUseCase) GetProcessTrend(ctx context.Context, req *ProcessTrendRequest) (*ProcessTrendResponse, error) {
	uc.logger.Debug("获取流程趋势分析", zap.String("processDefinitionKey", req.ProcessDefinitionKey))

	// 参数验证
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().AddDate(0, -1, 0) // 默认最近一个月
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}
	if req.Granularity == "" {
		req.Granularity = "day" // 默认按天统计
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("process_trend:%s:%d:%d:%s", req.ProcessDefinitionKey, req.StartTime.Unix(), req.EndTime.Unix(), req.Granularity)
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var response ProcessTrendResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			uc.logger.Debug("从缓存获取流程趋势分析")
			return &response, nil
		}
	}

	// 获取趋势数据
	trends, err := uc.historicRepo.GetProcessTrend(ctx, req.ProcessDefinitionKey, req.StartTime, req.EndTime, req.Granularity)
	if err != nil {
		uc.logger.Error("获取流程趋势分析失败", zap.Error(err))
		return nil, fmt.Errorf("获取流程趋势分析失败: %w", err)
	}

	response := &ProcessTrendResponse{
		ProcessDefinitionKey: req.ProcessDefinitionKey,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Granularity:          req.Granularity,
		TrendData:            trends,
	}

	// 缓存结果
	if data, err := json.Marshal(response); err == nil {
		uc.cache.Set(ctx, cacheKey, string(data), 30*time.Minute)
	}

	uc.logger.Info("获取流程趋势分析成功", zap.String("processDefinitionKey", req.ProcessDefinitionKey))
	return response, nil
}

// DeleteHistoricProcessInstance 删除历史流程实例
// 物理删除历史流程实例及其相关数据
func (uc *HistoricDataUseCase) DeleteHistoricProcessInstance(ctx context.Context, instanceID int64) error {
	uc.logger.Info("删除历史流程实例", zap.Int64("instanceID", instanceID))

	// 检查实例是否存在
	_, err := uc.historicRepo.GetHistoricProcessInstance(ctx, instanceID)
	if err != nil {
		uc.logger.Error("历史流程实例不存在", zap.Error(err))
		return fmt.Errorf("历史流程实例不存在: %w", err)
	}

	// 删除历史流程实例
	if err := uc.historicRepo.DeleteHistoricProcessInstance(ctx, instanceID); err != nil {
		uc.logger.Error("删除历史流程实例失败", zap.Error(err))
		return fmt.Errorf("删除历史流程实例失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("historic_process_instance:%d", instanceID)
	uc.cache.Delete(ctx, cacheKey)

	uc.logger.Info("删除历史流程实例成功", zap.Int64("instanceID", instanceID))
	return nil
}

// BatchDeleteHistoricProcessInstances 批量删除历史流程实例
// 根据条件批量删除历史流程实例
func (uc *HistoricDataUseCase) BatchDeleteHistoricProcessInstances(ctx context.Context, req *BatchDeleteHistoricProcessInstancesRequest) (*BatchDeleteHistoricProcessInstancesResponse, error) {
	uc.logger.Info("批量删除历史流程实例", zap.String("processDefinitionKey", req.ProcessDefinitionKey))

	// 参数验证
	if req.EndTimeBefore.IsZero() {
		return nil, fmt.Errorf("必须指定结束时间条件")
	}

	// 执行批量删除
	deletedCount, err := uc.historicRepo.BatchDeleteHistoricProcessInstances(ctx, req.ProcessDefinitionKey, req.EndTimeBefore)
	if err != nil {
		uc.logger.Error("批量删除历史流程实例失败", zap.Error(err))
		return nil, fmt.Errorf("批量删除历史流程实例失败: %w", err)
	}

	response := &BatchDeleteHistoricProcessInstancesResponse{
		DeletedCount: deletedCount,
	}

	uc.logger.Info("批量删除历史流程实例成功", zap.Int64("deletedCount", deletedCount))
	return response, nil
}

// calculateDuration 计算流程执行时长
func (uc *HistoricDataUseCase) calculateDuration(startTime time.Time, endTime *time.Time) *int64 {
	if endTime == nil {
		return nil
	}
	duration := endTime.Sub(startTime).Milliseconds()
	return &duration
}

// parseVariableValue 解析变量值
func (uc *HistoricDataUseCase) parseVariableValue(variable *HistoricVariableInstance) interface{} {
	switch variable.Type {
	case "string":
		return variable.TextValue
	case "long":
		return variable.LongValue
	case "double":
		return variable.DoubleValue
	case "boolean":
		return variable.LongValue != nil && *variable.LongValue == 1
	case "date":
		if variable.LongValue != nil {
			return time.Unix(*variable.LongValue/1000, 0)
		}
		return nil
	case "json":
		if variable.TextValue != nil {
			var jsonValue interface{}
			if err := json.Unmarshal([]byte(*variable.TextValue), &jsonValue); err == nil {
				return jsonValue
			}
		}
		return variable.TextValue
	default:
		return variable.TextValue
	}
}
