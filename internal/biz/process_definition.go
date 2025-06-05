// Package biz 提供业务逻辑层功能
// 包含流程定义管理的核心业务逻辑
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

// ProcessDefinitionUseCase 流程定义用例，包含流程定义相关的业务逻辑
type ProcessDefinitionUseCase struct {
	repo   ProcessDefinitionRepo
	cache  CacheRepo
	logger *zap.Logger
}

// NewProcessDefinitionUseCase 创建流程定义用例实例
func NewProcessDefinitionUseCase(
	repo ProcessDefinitionRepo,
	cache CacheRepo,
	logger *zap.Logger,
) *ProcessDefinitionUseCase {
	return &ProcessDefinitionUseCase{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// CreateProcessDefinition 创建流程定义
func (uc *ProcessDefinitionUseCase) CreateProcessDefinition(ctx context.Context, req *CreateProcessDefinitionRequest) (*ProcessDefinitionResponse, error) {
	uc.logger.Info("开始创建流程定义",
		zap.String("name", req.Name),
		zap.String("key", req.Key))

	// 验证输入参数
	if err := uc.validateCreateRequest(req); err != nil {
		uc.logger.Error("创建流程定义参数验证失败", zap.Error(err))
		return nil, fmt.Errorf("参数验证失败: %w", err)
	}

	// 检查流程键是否已存在
	existing, err := uc.repo.GetLatestByKey(ctx, req.Key)
	if err == nil && existing != nil {
		// 如果存在，创建新版本
		req.Version = existing.Version + 1
		uc.logger.Info("流程定义已存在，创建新版本",
			zap.String("key", req.Key),
			zap.Int32("new_version", req.Version))
	} else {
		// 如果不存在，创建第一个版本
		req.Version = 1
		uc.logger.Info("创建流程定义第一个版本", zap.String("key", req.Key))
	}

	// 验证流程定义内容
	if err := uc.validateProcessDefinition(req.Resource); err != nil {
		uc.logger.Error("流程定义内容验证失败", zap.Error(err))
		return nil, fmt.Errorf("流程定义内容验证失败: %w", err)
	}

	// 构建流程定义实体
	pd := &ent.ProcessDefinition{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Version:     req.Version,
		Resource:    req.Resource,
		Suspended:   false, // 新创建的流程定义默认为激活状态
		TenantID:    req.TenantID,
	}

	// 保存到数据库
	result, err := uc.repo.Create(ctx, pd)
	if err != nil {
		uc.logger.Error("保存流程定义失败", zap.Error(err))
		return nil, fmt.Errorf("保存流程定义失败: %w", err)
	}

	// 缓存流程定义
	if err := uc.cacheProcessDefinition(ctx, result); err != nil {
		uc.logger.Warn("缓存流程定义失败", zap.Error(err))
		// 缓存失败不影响主要流程
	}

	uc.logger.Info("流程定义创建成功",
		zap.String("id", strconv.FormatInt(result.ID, 10)),
		zap.String("key", result.Key),
		zap.Int32("version", result.Version))

	return uc.toProcessDefinitionResponse(result), nil
}

// GetProcessDefinition 根据ID获取流程定义
func (uc *ProcessDefinitionUseCase) GetProcessDefinition(ctx context.Context, id string) (*ProcessDefinitionResponse, error) {
	uc.logger.Debug("获取流程定义", zap.String("id", id))

	// 先尝试从缓存获取
	cacheKey := fmt.Sprintf("process_definition:%s", id)
	if cached, err := uc.cache.Get(ctx, cacheKey); err == nil {
		var pd ent.ProcessDefinition
		if err := json.Unmarshal([]byte(cached), &pd); err == nil {
			uc.logger.Debug("从缓存获取流程定义成功", zap.String("id", id))
			return uc.toProcessDefinitionResponse(&pd), nil
		}
	}

	// 从数据库获取
	pd, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取流程定义失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	// 缓存结果
	if err := uc.cacheProcessDefinition(ctx, pd); err != nil {
		uc.logger.Warn("缓存流程定义失败", zap.Error(err))
	}

	return uc.toProcessDefinitionResponse(pd), nil
}

// GetLatestProcessDefinition 根据Key获取最新版本的流程定义
func (uc *ProcessDefinitionUseCase) GetLatestProcessDefinition(ctx context.Context, key string) (*ProcessDefinitionResponse, error) {
	uc.logger.Debug("获取最新版本流程定义", zap.String("key", key))

	pd, err := uc.repo.GetLatestByKey(ctx, key)
	if err != nil {
		uc.logger.Error("获取最新版本流程定义失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取最新版本流程定义失败: %w", err)
	}

	return uc.toProcessDefinitionResponse(pd), nil
}

// UpdateProcessDefinition 更新流程定义
func (uc *ProcessDefinitionUseCase) UpdateProcessDefinition(ctx context.Context, id string, req *UpdateProcessDefinitionRequest) (*ProcessDefinitionResponse, error) {
	uc.logger.Info("更新流程定义", zap.String("id", id))

	// 获取现有流程定义
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("获取待更新的流程定义失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取待更新的流程定义失败: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Resource != "" {
		// 验证新的流程定义内容
		if err := uc.validateProcessDefinition(req.Resource); err != nil {
			uc.logger.Error("流程定义内容验证失败", zap.Error(err))
			return nil, fmt.Errorf("流程定义内容验证失败: %w", err)
		}
		existing.Resource = req.Resource
	}

	// 保存更新
	result, err := uc.repo.Update(ctx, existing)
	if err != nil {
		uc.logger.Error("更新流程定义失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("更新流程定义失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("process_definition:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程定义缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程定义更新成功", zap.String("id", id))
	return uc.toProcessDefinitionResponse(result), nil
}

// DeleteProcessDefinition 删除流程定义
func (uc *ProcessDefinitionUseCase) DeleteProcessDefinition(ctx context.Context, id string) error {
	uc.logger.Info("删除流程定义", zap.String("id", id))

	// 检查是否有正在运行的流程实例
	// TODO: 实现流程实例检查逻辑

	// 删除流程定义
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.logger.Error("删除流程定义失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("删除流程定义失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("process_definition:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程定义缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程定义删除成功", zap.String("id", id))
	return nil
}

// ListProcessDefinitions 分页查询流程定义
func (uc *ProcessDefinitionUseCase) ListProcessDefinitions(ctx context.Context, req *ListProcessDefinitionsRequest) (*ListProcessDefinitionsResponse, error) {
	uc.logger.Debug("分页查询流程定义", zap.Any("request", req))

	// 构建过滤条件
	filter := &ProcessDefinitionFilter{
		Name:        req.Name,
		Category:    req.Category,
		Status:      req.Status,
		CreatedFrom: req.CreatedFrom,
		CreatedTo:   req.CreatedTo,
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
	definitions, pagination, err := uc.repo.List(ctx, filter, opts)
	if err != nil {
		uc.logger.Error("查询流程定义列表失败", zap.Error(err))
		return nil, fmt.Errorf("查询流程定义列表失败: %w", err)
	}

	// 转换响应
	items := make([]*ProcessDefinitionResponse, len(definitions))
	for i, pd := range definitions {
		items[i] = uc.toProcessDefinitionResponse(pd)
	}

	return &ListProcessDefinitionsResponse{
		Items:      items,
		Pagination: pagination,
	}, nil
}

// DeployProcessDefinition 部署流程定义
func (uc *ProcessDefinitionUseCase) DeployProcessDefinition(ctx context.Context, id string) error {
	uc.logger.Info("部署流程定义", zap.String("id", id))

	// 部署流程定义
	if err := uc.repo.Deploy(ctx, id); err != nil {
		uc.logger.Error("部署流程定义失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("部署流程定义失败: %w", err)
	}

	// 清除缓存，强制重新加载
	cacheKey := fmt.Sprintf("process_definition:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程定义缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程定义部署成功", zap.String("id", id))
	return nil
}

// SuspendProcessDefinition 挂起流程定义
func (uc *ProcessDefinitionUseCase) SuspendProcessDefinition(ctx context.Context, id string) error {
	uc.logger.Info("挂起流程定义", zap.String("id", id))

	// 挂起流程定义
	if err := uc.repo.Suspend(ctx, id); err != nil {
		uc.logger.Error("挂起流程定义失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("挂起流程定义失败: %w", err)
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("process_definition:%s", id)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.Warn("清除流程定义缓存失败", zap.Error(err))
	}

	uc.logger.Info("流程定义挂起成功", zap.String("id", id))
	return nil
}

// validateCreateRequest 验证创建请求参数
func (uc *ProcessDefinitionUseCase) validateCreateRequest(req *CreateProcessDefinitionRequest) error {
	if req.Key == "" {
		return fmt.Errorf("流程键不能为空")
	}
	if req.Name == "" {
		return fmt.Errorf("流程名称不能为空")
	}
	if req.Resource == "" {
		return fmt.Errorf("流程资源不能为空")
	}
	return nil
}

// validateProcessDefinition 验证流程定义内容
func (uc *ProcessDefinitionUseCase) validateProcessDefinition(resource string) error {
	// 简单的JSON格式验证
	var definition map[string]interface{}
	if err := json.Unmarshal([]byte(resource), &definition); err != nil {
		return fmt.Errorf("流程定义不是有效的JSON格式: %w", err)
	}

	// 检查必要字段
	if _, ok := definition["id"]; !ok {
		return fmt.Errorf("流程定义缺少id字段")
	}
	if _, ok := definition["name"]; !ok {
		return fmt.Errorf("流程定义缺少name字段")
	}
	if _, ok := definition["elements"]; !ok {
		return fmt.Errorf("流程定义缺少elements字段")
	}

	return nil
}

// cacheProcessDefinition 缓存流程定义
func (uc *ProcessDefinitionUseCase) cacheProcessDefinition(ctx context.Context, pd *ent.ProcessDefinition) error {
	cacheKey := fmt.Sprintf("process_definition:%s", strconv.FormatInt(pd.ID, 10))
	data, err := json.Marshal(pd)
	if err != nil {
		return fmt.Errorf("序列化流程定义失败: %w", err)
	}

	// 缓存1小时
	return uc.cache.Set(ctx, cacheKey, string(data), time.Hour)
}

// toProcessDefinitionResponse 转换为响应格式
func (uc *ProcessDefinitionUseCase) toProcessDefinitionResponse(pd *ent.ProcessDefinition) *ProcessDefinitionResponse {
	return &ProcessDefinitionResponse{
		ID:          strconv.FormatInt(pd.ID, 10),
		Key:         pd.Key,
		Name:        pd.Name,
		Description: pd.Description,
		Category:    pd.Category,
		Version:     pd.Version,
		Resource:    pd.Resource,
		Suspended:   pd.Suspended,
		TenantID:    pd.TenantID,
		DeployTime:  pd.DeployTime,
		CreatedAt:   pd.CreatedAt,
		UpdatedAt:   pd.UpdatedAt,
	}
}
