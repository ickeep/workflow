// Package service 流程实例服务实现
// 处理流程实例相关的API请求并调用业务逻辑
package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/biz"
)

// ProcessInstanceService 流程实例服务
// 负责处理流程实例相关的API请求，连接HTTP/gRPC层与业务逻辑层
type ProcessInstanceService struct {
	uc     *biz.ProcessInstanceUseCase
	logger *zap.Logger
}

// NewProcessInstanceService 创建流程实例服务
func NewProcessInstanceService(
	uc *biz.ProcessInstanceUseCase,
	logger *zap.Logger,
) *ProcessInstanceService {
	return &ProcessInstanceService{
		uc:     uc,
		logger: logger,
	}
}

// StartProcessInstance 启动流程实例
// 根据流程定义启动新的流程实例
func (s *ProcessInstanceService) StartProcessInstance(ctx context.Context, req *biz.StartProcessInstanceRequest) (*biz.ProcessInstanceResponse, error) {
	s.logger.Info("服务层: 启动流程实例",
		zap.String("process_definition_id", req.ProcessDefinitionID),
		zap.String("process_definition_key", req.ProcessDefinitionKey),
		zap.String("business_key", req.BusinessKey))

	// 参数验证
	if err := s.validateStartRequest(req); err != nil {
		s.logger.Error("启动流程实例参数验证失败", zap.Error(err))
		return nil, err
	}

	// 调用业务逻辑层
	result, err := s.uc.StartProcessInstance(ctx, req)
	if err != nil {
		s.logger.Error("启动流程实例失败", zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "启动流程实例失败")
	}

	s.logger.Info("服务层: 启动流程实例成功", zap.String("instance_id", result.ID))
	return result, nil
}

// GetProcessInstance 获取流程实例
// 根据ID获取流程实例详情
func (s *ProcessInstanceService) GetProcessInstance(ctx context.Context, id string) (*biz.ProcessInstanceResponse, error) {
	s.logger.Debug("服务层: 获取流程实例", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程实例ID不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	result, err := s.uc.GetProcessInstance(ctx, id)
	if err != nil {
		s.logger.Error("获取流程实例失败", zap.String("id", id), zap.Error(err))
		return nil, WrapError(err, ErrCodeNotFound, "流程实例不存在")
	}

	return result, nil
}

// ListProcessInstances 查询流程实例列表
// 分页查询流程实例列表
func (s *ProcessInstanceService) ListProcessInstances(ctx context.Context, req *biz.ListProcessInstancesRequest) (*biz.ListProcessInstancesResponse, error) {
	s.logger.Debug("服务层: 查询流程实例列表")

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

	result, err := s.uc.ListProcessInstances(ctx, req)
	if err != nil {
		s.logger.Error("查询流程实例列表失败", zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "查询流程实例列表失败")
	}

	s.logger.Info("服务层: 查询流程实例列表成功",
		zap.Int("total", result.Pagination.Total),
		zap.Int("page", req.Page))
	return result, nil
}

// SuspendProcessInstance 挂起流程实例
// 将流程实例设置为挂起状态
func (s *ProcessInstanceService) SuspendProcessInstance(ctx context.Context, id string) error {
	s.logger.Info("服务层: 挂起流程实例", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程实例ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	err := s.uc.SuspendProcessInstance(ctx, id)
	if err != nil {
		s.logger.Error("挂起流程实例失败", zap.String("id", id), zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "挂起流程实例失败")
	}

	s.logger.Info("服务层: 挂起流程实例成功", zap.String("id", id))
	return nil
}

// ActivateProcessInstance 激活流程实例
// 将挂起的流程实例重新激活
func (s *ProcessInstanceService) ActivateProcessInstance(ctx context.Context, id string) error {
	s.logger.Info("服务层: 激活流程实例", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程实例ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	err := s.uc.ActivateProcessInstance(ctx, id)
	if err != nil {
		s.logger.Error("激活流程实例失败", zap.String("id", id), zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "激活流程实例失败")
	}

	s.logger.Info("服务层: 激活流程实例成功", zap.String("id", id))
	return nil
}

// TerminateProcessInstance 终止流程实例
// 终止流程实例的执行
func (s *ProcessInstanceService) TerminateProcessInstance(ctx context.Context, id string, reason string) error {
	s.logger.Info("服务层: 终止流程实例",
		zap.String("id", id),
		zap.String("reason", reason))

	if id == "" {
		s.logger.Error("流程实例ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	err := s.uc.TerminateProcessInstance(ctx, id, reason)
	if err != nil {
		s.logger.Error("终止流程实例失败", zap.String("id", id), zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "终止流程实例失败")
	}

	s.logger.Info("服务层: 终止流程实例成功", zap.String("id", id))
	return nil
}

// DeleteProcessInstance 删除流程实例
// 删除指定的流程实例
func (s *ProcessInstanceService) DeleteProcessInstance(ctx context.Context, id string, reason string) error {
	s.logger.Info("服务层: 删除流程实例",
		zap.String("id", id),
		zap.String("reason", reason))

	if id == "" {
		s.logger.Error("流程实例ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	err := s.uc.DeleteProcessInstance(ctx, id, reason)
	if err != nil {
		s.logger.Error("删除流程实例失败", zap.String("id", id), zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "删除流程实例失败")
	}

	s.logger.Info("服务层: 删除流程实例成功", zap.String("id", id))
	return nil
}

// GetProcessVariable 获取流程变量
// 获取流程实例的指定变量
func (s *ProcessInstanceService) GetProcessVariable(ctx context.Context, instanceID string, variableName string) (interface{}, error) {
	s.logger.Debug("服务层: 获取流程变量",
		zap.String("instance_id", instanceID),
		zap.String("variable_name", variableName))

	if instanceID == "" {
		s.logger.Error("流程实例ID不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}
	if variableName == "" {
		s.logger.Error("变量名不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "变量名不能为空")
	}

	result, err := s.uc.GetProcessVariable(ctx, instanceID, variableName)
	if err != nil {
		s.logger.Error("获取流程变量失败",
			zap.String("instance_id", instanceID),
			zap.String("variable_name", variableName),
			zap.Error(err))
		return nil, WrapError(err, ErrCodeNotFound, "流程变量不存在")
	}

	return result, nil
}

// SetProcessVariable 设置流程变量
// 设置流程实例的指定变量
func (s *ProcessInstanceService) SetProcessVariable(ctx context.Context, instanceID string, variableName string, value interface{}) error {
	s.logger.Info("服务层: 设置流程变量",
		zap.String("instance_id", instanceID),
		zap.String("variable_name", variableName))

	if instanceID == "" {
		s.logger.Error("流程实例ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}
	if variableName == "" {
		s.logger.Error("变量名不能为空")
		return NewServiceError(ErrCodeBadRequest, "变量名不能为空")
	}

	err := s.uc.SetProcessVariable(ctx, instanceID, variableName, value)
	if err != nil {
		s.logger.Error("设置流程变量失败",
			zap.String("instance_id", instanceID),
			zap.String("variable_name", variableName),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "设置流程变量失败")
	}

	s.logger.Info("服务层: 设置流程变量成功",
		zap.String("instance_id", instanceID),
		zap.String("variable_name", variableName))
	return nil
}

// GetProcessVariables 获取流程变量列表
// 获取流程实例的所有变量
func (s *ProcessInstanceService) GetProcessVariables(ctx context.Context, instanceID string) (map[string]interface{}, error) {
	s.logger.Debug("服务层: 获取流程变量列表", zap.String("instance_id", instanceID))

	if instanceID == "" {
		s.logger.Error("流程实例ID不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	result, err := s.uc.GetProcessVariables(ctx, instanceID)
	if err != nil {
		s.logger.Error("获取流程变量列表失败",
			zap.String("instance_id", instanceID),
			zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "获取流程变量列表失败")
	}

	s.logger.Debug("服务层: 获取流程变量列表成功",
		zap.String("instance_id", instanceID),
		zap.Int("variable_count", len(result)))
	return result, nil
}

// SetProcessVariables 批量设置流程变量
// 批量设置流程实例的变量
func (s *ProcessInstanceService) SetProcessVariables(ctx context.Context, instanceID string, variables map[string]interface{}) error {
	s.logger.Info("服务层: 批量设置流程变量",
		zap.String("instance_id", instanceID),
		zap.Int("variable_count", len(variables)))

	if instanceID == "" {
		s.logger.Error("流程实例ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}
	if len(variables) == 0 {
		s.logger.Error("变量列表不能为空")
		return NewServiceError(ErrCodeBadRequest, "变量列表不能为空")
	}

	err := s.uc.SetProcessVariables(ctx, instanceID, variables)
	if err != nil {
		s.logger.Error("批量设置流程变量失败",
			zap.String("instance_id", instanceID),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "批量设置流程变量失败")
	}

	s.logger.Info("服务层: 批量设置流程变量成功",
		zap.String("instance_id", instanceID),
		zap.Int("variable_count", len(variables)))
	return nil
}

// validateStartRequest 验证启动流程实例请求参数
func (s *ProcessInstanceService) validateStartRequest(req *biz.StartProcessInstanceRequest) error {
	if req.ProcessDefinitionID == "" && req.ProcessDefinitionKey == "" {
		return NewServiceError(ErrCodeBadRequest, "流程定义ID或Key至少需要指定一个")
	}
	return nil
}
