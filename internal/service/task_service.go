// Package service 任务实例服务实现
// 处理任务实例相关的API请求并调用业务逻辑
package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/biz"
)

// TaskInstanceService 任务实例服务
// 负责处理任务实例相关的API请求，连接HTTP/gRPC层与业务逻辑层
type TaskInstanceService struct {
	uc     *biz.TaskInstanceUseCase
	logger *zap.Logger
}

// NewTaskInstanceService 创建任务实例服务
func NewTaskInstanceService(
	uc *biz.TaskInstanceUseCase,
	logger *zap.Logger,
) *TaskInstanceService {
	return &TaskInstanceService{
		uc:     uc,
		logger: logger,
	}
}

// GetTaskInstance 获取任务实例
// 根据ID获取任务实例详情
func (s *TaskInstanceService) GetTaskInstance(ctx context.Context, id string) (*biz.TaskInstanceResponse, error) {
	s.logger.Debug("服务层: 获取任务实例", zap.String("id", id))

	if id == "" {
		s.logger.Error("任务实例ID不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "任务实例ID不能为空")
	}

	result, err := s.uc.GetTaskInstance(ctx, id)
	if err != nil {
		s.logger.Error("获取任务实例失败", zap.String("id", id), zap.Error(err))
		return nil, WrapError(err, ErrCodeNotFound, "任务实例不存在")
	}

	return result, nil
}

// ListTaskInstances 查询任务实例列表
// 分页查询任务实例列表
func (s *TaskInstanceService) ListTaskInstances(ctx context.Context, req *biz.ListTaskInstancesRequest) (*biz.ListTaskInstancesResponse, error) {
	s.logger.Debug("服务层: 查询任务实例列表")

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

	result, err := s.uc.ListTaskInstances(ctx, req)
	if err != nil {
		s.logger.Error("查询任务实例列表失败", zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "查询任务实例列表失败")
	}

	s.logger.Info("服务层: 查询任务实例列表成功",
		zap.Int("total", result.Pagination.Total),
		zap.Int("page", req.Page))
	return result, nil
}

// ClaimTask 认领任务
// 用户认领指定的任务
func (s *TaskInstanceService) ClaimTask(ctx context.Context, taskID string, userID string) error {
	s.logger.Info("服务层: 认领任务",
		zap.String("task_id", taskID),
		zap.String("user_id", userID))

	if taskID == "" {
		s.logger.Error("任务ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "任务ID不能为空")
	}
	if userID == "" {
		s.logger.Error("用户ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "用户ID不能为空")
	}

	req := &biz.ClaimTaskRequest{
		AssigneeID: userID,
	}

	err := s.uc.ClaimTask(ctx, taskID, req)
	if err != nil {
		s.logger.Error("认领任务失败",
			zap.String("task_id", taskID),
			zap.String("user_id", userID),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "认领任务失败")
	}

	s.logger.Info("服务层: 认领任务成功",
		zap.String("task_id", taskID),
		zap.String("user_id", userID))
	return nil
}

// CompleteTask 完成任务
// 完成指定的任务
func (s *TaskInstanceService) CompleteTask(ctx context.Context, taskID string, variables map[string]interface{}, comment string) error {
	s.logger.Info("服务层: 完成任务",
		zap.String("task_id", taskID))

	if taskID == "" {
		s.logger.Error("任务ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "任务ID不能为空")
	}

	req := &biz.CompleteTaskRequest{
		Variables: variables,
		Comment:   comment,
	}

	err := s.uc.CompleteTask(ctx, taskID, req)
	if err != nil {
		s.logger.Error("完成任务失败",
			zap.String("task_id", taskID),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "完成任务失败")
	}

	s.logger.Info("服务层: 完成任务成功", zap.String("task_id", taskID))
	return nil
}

// DelegateTask 委派任务
// 将任务委派给其他用户
func (s *TaskInstanceService) DelegateTask(ctx context.Context, taskID string, delegateID string, comment string) error {
	s.logger.Info("服务层: 委派任务",
		zap.String("task_id", taskID),
		zap.String("delegate_id", delegateID))

	if taskID == "" {
		s.logger.Error("任务ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "任务ID不能为空")
	}
	if delegateID == "" {
		s.logger.Error("委派人用户ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "委派人用户ID不能为空")
	}

	req := &biz.DelegateTaskRequest{
		DelegateID: delegateID,
		Comment:    comment,
	}

	err := s.uc.DelegateTask(ctx, taskID, req)
	if err != nil {
		s.logger.Error("委派任务失败",
			zap.String("task_id", taskID),
			zap.String("delegate_id", delegateID),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "委派任务失败")
	}

	s.logger.Info("服务层: 委派任务成功",
		zap.String("task_id", taskID),
		zap.String("delegate_id", delegateID))
	return nil
}

// GetMyTasks 获取我的任务列表
// 获取当前用户的任务列表
func (s *TaskInstanceService) GetMyTasks(ctx context.Context, req *biz.ListTaskInstancesRequest) (*biz.ListTaskInstancesResponse, error) {
	s.logger.Debug("服务层: 获取我的任务列表")

	// 参数验证和默认值设置
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	result, err := s.uc.GetMyTasks(ctx, req)
	if err != nil {
		s.logger.Error("获取我的任务列表失败", zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "获取我的任务列表失败")
	}

	s.logger.Info("服务层: 获取我的任务列表成功",
		zap.Int("total", result.Pagination.Total))
	return result, nil
}

// GetAvailableTasks 获取可认领的任务列表
// 获取当前用户可以认领的任务列表
func (s *TaskInstanceService) GetAvailableTasks(ctx context.Context, req *biz.ListTaskInstancesRequest) (*biz.ListTaskInstancesResponse, error) {
	s.logger.Debug("服务层: 获取可认领的任务列表")

	// 参数验证和默认值设置
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	result, err := s.uc.GetAvailableTasks(ctx, req)
	if err != nil {
		s.logger.Error("获取可认领的任务列表失败", zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "获取可认领的任务列表失败")
	}

	s.logger.Info("服务层: 获取可认领的任务列表成功",
		zap.Int("total", result.Pagination.Total))
	return result, nil
}
