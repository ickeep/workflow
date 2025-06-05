// Package service 提供服务层功能
// 流程定义服务实现，处理HTTP/gRPC请求并调用业务逻辑
package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/biz"
)

// ProcessDefinitionService 流程定义服务
// 负责处理流程定义相关的API请求，连接HTTP/gRPC层与业务逻辑层
type ProcessDefinitionService struct {
	uc     *biz.ProcessDefinitionUseCase
	logger *zap.Logger
}

// NewProcessDefinitionService 创建流程定义服务
func NewProcessDefinitionService(
	uc *biz.ProcessDefinitionUseCase,
	logger *zap.Logger,
) *ProcessDefinitionService {
	return &ProcessDefinitionService{
		uc:     uc,
		logger: logger,
	}
}

// CreateProcessDefinition 创建流程定义
// 处理创建流程定义的API请求
func (s *ProcessDefinitionService) CreateProcessDefinition(ctx context.Context, req *biz.CreateProcessDefinitionRequest) (*biz.ProcessDefinitionResponse, error) {
	s.logger.Info("服务层: 创建流程定义",
		zap.String("name", req.Name),
		zap.String("key", req.Key))

	// 参数验证
	if err := s.validateCreateRequest(req); err != nil {
		s.logger.Error("创建流程定义参数验证失败", zap.Error(err))
		return nil, err
	}

	// 调用业务逻辑层
	result, err := s.uc.CreateProcessDefinition(ctx, req)
	if err != nil {
		s.logger.Error("创建流程定义失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("服务层: 创建流程定义成功", zap.String("id", result.ID))
	return result, nil
}

// GetProcessDefinition 获取流程定义
// 根据ID获取流程定义详情
func (s *ProcessDefinitionService) GetProcessDefinition(ctx context.Context, id string) (*biz.ProcessDefinitionResponse, error) {
	s.logger.Debug("服务层: 获取流程定义", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程定义ID不能为空")
		return nil, &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义ID不能为空",
		}
	}

	result, err := s.uc.GetProcessDefinition(ctx, id)
	if err != nil {
		s.logger.Error("获取流程定义失败", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return result, nil
}

// GetLatestProcessDefinition 获取最新版本流程定义
// 根据Key获取最新版本的流程定义
func (s *ProcessDefinitionService) GetLatestProcessDefinition(ctx context.Context, key string) (*biz.ProcessDefinitionResponse, error) {
	s.logger.Debug("服务层: 获取最新版本流程定义", zap.String("key", key))

	if key == "" {
		s.logger.Error("流程定义Key不能为空")
		return nil, &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义Key不能为空",
		}
	}

	result, err := s.uc.GetLatestProcessDefinition(ctx, key)
	if err != nil {
		s.logger.Error("获取最新版本流程定义失败", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	return result, nil
}

// UpdateProcessDefinition 更新流程定义
// 更新指定的流程定义
func (s *ProcessDefinitionService) UpdateProcessDefinition(ctx context.Context, id string, req *biz.UpdateProcessDefinitionRequest) (*biz.ProcessDefinitionResponse, error) {
	s.logger.Info("服务层: 更新流程定义", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程定义ID不能为空")
		return nil, &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义ID不能为空",
		}
	}

	result, err := s.uc.UpdateProcessDefinition(ctx, id, req)
	if err != nil {
		s.logger.Error("更新流程定义失败", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	s.logger.Info("服务层: 更新流程定义成功", zap.String("id", id))
	return result, nil
}

// DeleteProcessDefinition 删除流程定义
// 删除指定的流程定义
func (s *ProcessDefinitionService) DeleteProcessDefinition(ctx context.Context, id string) error {
	s.logger.Info("服务层: 删除流程定义", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程定义ID不能为空")
		return &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义ID不能为空",
		}
	}

	err := s.uc.DeleteProcessDefinition(ctx, id)
	if err != nil {
		s.logger.Error("删除流程定义失败", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("服务层: 删除流程定义成功", zap.String("id", id))
	return nil
}

// ListProcessDefinitions 查询流程定义列表
// 分页查询流程定义列表
func (s *ProcessDefinitionService) ListProcessDefinitions(ctx context.Context, req *biz.ListProcessDefinitionsRequest) (*biz.ListProcessDefinitionsResponse, error) {
	s.logger.Debug("服务层: 查询流程定义列表")

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

	result, err := s.uc.ListProcessDefinitions(ctx, req)
	if err != nil {
		s.logger.Error("查询流程定义列表失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("服务层: 查询流程定义列表成功",
		zap.Int("total", result.Pagination.Total),
		zap.Int("page", req.Page))
	return result, nil
}

// DeployProcessDefinition 部署流程定义
// 将流程定义部署为可用状态
func (s *ProcessDefinitionService) DeployProcessDefinition(ctx context.Context, id string) error {
	s.logger.Info("服务层: 部署流程定义", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程定义ID不能为空")
		return &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义ID不能为空",
		}
	}

	err := s.uc.DeployProcessDefinition(ctx, id)
	if err != nil {
		s.logger.Error("部署流程定义失败", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("服务层: 部署流程定义成功", zap.String("id", id))
	return nil
}

// SuspendProcessDefinition 挂起流程定义
// 将流程定义设置为挂起状态
func (s *ProcessDefinitionService) SuspendProcessDefinition(ctx context.Context, id string) error {
	s.logger.Info("服务层: 挂起流程定义", zap.String("id", id))

	if id == "" {
		s.logger.Error("流程定义ID不能为空")
		return &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义ID不能为空",
		}
	}

	err := s.uc.SuspendProcessDefinition(ctx, id)
	if err != nil {
		s.logger.Error("挂起流程定义失败", zap.String("id", id), zap.Error(err))
		return err
	}

	s.logger.Info("服务层: 挂起流程定义成功", zap.String("id", id))
	return nil
}

// validateCreateRequest 验证创建请求参数
func (s *ProcessDefinitionService) validateCreateRequest(req *biz.CreateProcessDefinitionRequest) error {
	if req.Key == "" {
		return &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义Key不能为空",
		}
	}
	if req.Name == "" {
		return &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义名称不能为空",
		}
	}
	if req.Resource == "" {
		return &ServiceError{
			Code:    ErrCodeBadRequest,
			Message: "流程定义资源不能为空",
		}
	}
	return nil
}
