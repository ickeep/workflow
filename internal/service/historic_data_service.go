// Package service 历史数据服务实现
// 处理历史数据相关的API请求并调用业务逻辑
package service

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/biz"
)

// HistoricDataService 历史数据服务
// 负责处理历史数据相关的API请求，连接HTTP/gRPC层与业务逻辑层
type HistoricDataService struct {
	uc     *biz.HistoricDataUseCase
	logger *zap.Logger
}

// NewHistoricDataService 创建历史数据服务
func NewHistoricDataService(
	uc *biz.HistoricDataUseCase,
	logger *zap.Logger,
) *HistoricDataService {
	return &HistoricDataService{
		uc:     uc,
		logger: logger,
	}
}

// GetHistoricProcessInstance 获取历史流程实例
// 根据ID获取历史流程实例详情
func (s *HistoricDataService) GetHistoricProcessInstance(ctx context.Context, id string) (*biz.HistoricProcessInstanceResponse, error) {
	s.logger.Debug("服务层: 获取历史流程实例", zap.String("id", id))

	if id == "" {
		s.logger.Error("历史流程实例ID不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "历史流程实例ID不能为空")
	}

	// 类型转换
	instanceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		s.logger.Error("无效的历史流程实例ID", zap.String("id", id), zap.Error(err))
		return nil, NewServiceError(ErrCodeBadRequest, "无效的历史流程实例ID")
	}

	result, err := s.uc.GetHistoricProcessInstance(ctx, instanceID)
	if err != nil {
		s.logger.Error("获取历史流程实例失败", zap.String("id", id), zap.Error(err))
		return nil, WrapError(err, ErrCodeNotFound, "历史流程实例不存在")
	}

	return result, nil
}

// ListHistoricProcessInstances 查询历史流程实例列表
// 分页查询历史流程实例列表
func (s *HistoricDataService) ListHistoricProcessInstances(ctx context.Context, req *biz.ListHistoricProcessInstancesRequest) (*biz.ListHistoricProcessInstancesResponse, error) {
	s.logger.Debug("服务层: 查询历史流程实例列表")

	// 参数验证和默认值设置
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页面大小
	}

	result, err := s.uc.ListHistoricProcessInstances(ctx, req)
	if err != nil {
		s.logger.Error("查询历史流程实例列表失败", zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "查询历史流程实例列表失败")
	}

	s.logger.Info("服务层: 查询历史流程实例列表成功",
		zap.Int("total", result.Total),
		zap.Int("page", req.Page))
	return result, nil
}

// GetProcessStatistics 获取流程统计数据
// 获取指定流程的统计信息
func (s *HistoricDataService) GetProcessStatistics(ctx context.Context, req *biz.ProcessStatisticsRequest) (*biz.ProcessStatisticsResponse, error) {
	s.logger.Debug("服务层: 获取流程统计数据",
		zap.String("process_definition_key", req.ProcessDefinitionKey))

	// 参数验证
	if err := s.validateStatisticsRequest(req); err != nil {
		s.logger.Error("流程统计数据参数验证失败", zap.Error(err))
		return nil, err
	}

	result, err := s.uc.GetProcessStatistics(ctx, req)
	if err != nil {
		s.logger.Error("获取流程统计数据失败",
			zap.String("process_definition_key", req.ProcessDefinitionKey),
			zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "获取流程统计数据失败")
	}

	s.logger.Info("服务层: 获取流程统计数据成功",
		zap.String("process_definition_key", req.ProcessDefinitionKey))
	return result, nil
}

// GetProcessTrend 获取流程趋势数据
// 获取指定流程的趋势分析数据
func (s *HistoricDataService) GetProcessTrend(ctx context.Context, req *biz.ProcessTrendRequest) (*biz.ProcessTrendResponse, error) {
	s.logger.Debug("服务层: 获取流程趋势数据",
		zap.String("process_definition_key", req.ProcessDefinitionKey),
		zap.String("granularity", req.Granularity))

	// 参数验证
	if err := s.validateTrendRequest(req); err != nil {
		s.logger.Error("流程趋势数据参数验证失败", zap.Error(err))
		return nil, err
	}

	result, err := s.uc.GetProcessTrend(ctx, req)
	if err != nil {
		s.logger.Error("获取流程趋势数据失败",
			zap.String("process_definition_key", req.ProcessDefinitionKey),
			zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "获取流程趋势数据失败")
	}

	s.logger.Info("服务层: 获取流程趋势数据成功",
		zap.String("process_definition_key", req.ProcessDefinitionKey),
		zap.String("granularity", req.Granularity))
	return result, nil
}

// BatchDeleteHistoricProcessInstances 批量删除历史流程实例
// 批量删除满足条件的历史流程实例
func (s *HistoricDataService) BatchDeleteHistoricProcessInstances(ctx context.Context, req *biz.BatchDeleteHistoricProcessInstancesRequest) (*biz.BatchDeleteHistoricProcessInstancesResponse, error) {
	s.logger.Info("服务层: 批量删除历史流程实例",
		zap.String("process_definition_key", req.ProcessDefinitionKey))

	// 参数验证
	if err := s.validateBatchDeleteRequest(req); err != nil {
		s.logger.Error("批量删除历史流程实例参数验证失败", zap.Error(err))
		return nil, err
	}

	result, err := s.uc.BatchDeleteHistoricProcessInstances(ctx, req)
	if err != nil {
		s.logger.Error("批量删除历史流程实例失败",
			zap.String("process_definition_key", req.ProcessDefinitionKey),
			zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "批量删除历史流程实例失败")
	}

	s.logger.Info("服务层: 批量删除历史流程实例成功",
		zap.String("process_definition_key", req.ProcessDefinitionKey),
		zap.Int64("deleted_count", result.DeletedCount))
	return result, nil
}

// validateStatisticsRequest 验证统计请求参数
func (s *HistoricDataService) validateStatisticsRequest(req *biz.ProcessStatisticsRequest) error {
	if req.ProcessDefinitionKey == "" {
		return NewServiceError(ErrCodeBadRequest, "流程定义Key不能为空")
	}
	if req.StartTime.IsZero() {
		return NewServiceError(ErrCodeBadRequest, "统计开始时间不能为空")
	}
	if req.EndTime.IsZero() {
		return NewServiceError(ErrCodeBadRequest, "统计结束时间不能为空")
	}
	if req.EndTime.Before(req.StartTime) {
		return NewServiceError(ErrCodeBadRequest, "结束时间不能早于开始时间")
	}
	return nil
}

// validateTrendRequest 验证趋势请求参数
func (s *HistoricDataService) validateTrendRequest(req *biz.ProcessTrendRequest) error {
	if req.ProcessDefinitionKey == "" {
		return NewServiceError(ErrCodeBadRequest, "流程定义Key不能为空")
	}
	if req.StartTime.IsZero() {
		return NewServiceError(ErrCodeBadRequest, "开始时间不能为空")
	}
	if req.EndTime.IsZero() {
		return NewServiceError(ErrCodeBadRequest, "结束时间不能为空")
	}
	if req.EndTime.Before(req.StartTime) {
		return NewServiceError(ErrCodeBadRequest, "结束时间不能早于开始时间")
	}

	// 验证时间粒度
	validGranularities := map[string]bool{
		"hour":  true,
		"day":   true,
		"week":  true,
		"month": true,
	}
	if !validGranularities[req.Granularity] {
		return NewServiceError(ErrCodeBadRequest, "无效的时间粒度，支持: hour, day, week, month")
	}

	return nil
}

// validateBatchDeleteRequest 验证批量删除请求参数
func (s *HistoricDataService) validateBatchDeleteRequest(req *biz.BatchDeleteHistoricProcessInstancesRequest) error {
	if req.ProcessDefinitionKey == "" {
		return NewServiceError(ErrCodeBadRequest, "流程定义Key不能为空")
	}
	if req.EndTimeBefore.IsZero() {
		return NewServiceError(ErrCodeBadRequest, "结束时间之前的时间不能为空")
	}
	return nil
}
