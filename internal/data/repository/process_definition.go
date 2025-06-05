// Package repository 提供数据仓储层实现
// 包含具体的数据访问逻辑和数据库操作
package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/workflow-engine/workflow-engine/internal/biz"
	"github.com/workflow-engine/workflow-engine/internal/data/ent"
	"github.com/workflow-engine/workflow-engine/internal/data/ent/processdefinition"

	"go.uber.org/zap"
)

// processDefinitionRepo 流程定义仓储实现
type processDefinitionRepo struct {
	data   *ent.Client
	logger *zap.Logger
}

// NewProcessDefinitionRepo 创建流程定义仓储实例
func NewProcessDefinitionRepo(data *ent.Client, logger *zap.Logger) biz.ProcessDefinitionRepo {
	return &processDefinitionRepo{
		data:   data,
		logger: logger,
	}
}

// Create 创建流程定义
func (r *processDefinitionRepo) Create(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error) {
	r.logger.Info("创建流程定义", zap.String("name", pd.Name), zap.String("key", pd.Key))

	result, err := r.data.ProcessDefinition.Create().
		SetName(pd.Name).
		SetKey(pd.Key).
		SetDescription(pd.Description).
		SetCategory(pd.Category).
		SetVersion(pd.Version).
		SetResource(pd.Resource).
		SetSuspended(pd.Suspended).
		SetTenantID(pd.TenantID).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		r.logger.Error("创建流程定义失败", zap.Error(err))
		return nil, fmt.Errorf("创建流程定义失败: %w", err)
	}

	r.logger.Info("流程定义创建成功", zap.String("id", strconv.FormatInt(result.ID, 10)))
	return result, nil
}

// GetByID 根据ID获取流程定义
func (r *processDefinitionRepo) GetByID(ctx context.Context, id string) (*ent.ProcessDefinition, error) {
	r.logger.Debug("根据ID获取流程定义", zap.String("id", id))

	// 将字符串ID转换为int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("无效的流程定义ID: %s", id)
	}

	result, err := r.data.ProcessDefinition.
		Query().
		Where(processdefinition.ID(idInt)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Warn("流程定义不存在", zap.String("id", id))
			return nil, fmt.Errorf("流程定义不存在: %s", id)
		}
		r.logger.Error("获取流程定义失败", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	return result, nil
}

// GetLatestByKey 根据Key获取最新版本的流程定义
func (r *processDefinitionRepo) GetLatestByKey(ctx context.Context, key string) (*ent.ProcessDefinition, error) {
	r.logger.Debug("根据Key获取最新版本流程定义", zap.String("key", key))

	result, err := r.data.ProcessDefinition.
		Query().
		Where(processdefinition.Key(key)).
		Order(ent.Desc(processdefinition.FieldVersion)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Warn("流程定义不存在", zap.String("key", key))
			return nil, fmt.Errorf("流程定义不存在: %s", key)
		}
		r.logger.Error("获取最新流程定义失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取最新流程定义失败: %w", err)
	}

	return result, nil
}

// GetByKeyAndVersion 根据Key和版本获取流程定义
func (r *processDefinitionRepo) GetByKeyAndVersion(ctx context.Context, key string, version int) (*ent.ProcessDefinition, error) {
	r.logger.Debug("根据Key和版本获取流程定义",
		zap.String("key", key),
		zap.Int("version", version))

	result, err := r.data.ProcessDefinition.
		Query().
		Where(
			processdefinition.Key(key),
			processdefinition.Version(int32(version)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Warn("流程定义不存在",
				zap.String("key", key),
				zap.Int("version", version))
			return nil, fmt.Errorf("流程定义不存在: %s v%d", key, version)
		}
		r.logger.Error("获取流程定义失败",
			zap.String("key", key),
			zap.Int("version", version),
			zap.Error(err))
		return nil, fmt.Errorf("获取流程定义失败: %w", err)
	}

	return result, nil
}

// Update 更新流程定义
func (r *processDefinitionRepo) Update(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error) {
	r.logger.Info("更新流程定义", zap.String("id", strconv.FormatInt(pd.ID, 10)))

	result, err := r.data.ProcessDefinition.
		UpdateOneID(pd.ID).
		SetName(pd.Name).
		SetDescription(pd.Description).
		SetCategory(pd.Category).
		SetResource(pd.Resource).
		SetSuspended(pd.Suspended).
		SetTenantID(pd.TenantID).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		r.logger.Error("更新流程定义失败", zap.String("id", strconv.FormatInt(pd.ID, 10)), zap.Error(err))
		return nil, fmt.Errorf("更新流程定义失败: %w", err)
	}

	r.logger.Info("流程定义更新成功", zap.String("id", strconv.FormatInt(result.ID, 10)))
	return result, nil
}

// Delete 删除流程定义
func (r *processDefinitionRepo) Delete(ctx context.Context, id string) error {
	r.logger.Info("删除流程定义", zap.String("id", id))

	// 将字符串ID转换为int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的流程定义ID: %s", id)
	}

	err = r.data.ProcessDefinition.
		DeleteOneID(idInt).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Warn("流程定义不存在", zap.String("id", id))
			return fmt.Errorf("流程定义不存在: %s", id)
		}
		r.logger.Error("删除流程定义失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("删除流程定义失败: %w", err)
	}

	r.logger.Info("流程定义删除成功", zap.String("id", id))
	return nil
}

// List 分页查询流程定义
func (r *processDefinitionRepo) List(ctx context.Context, filter *biz.ProcessDefinitionFilter, opts *biz.QueryOptions) ([]*ent.ProcessDefinition, *biz.PaginationResult, error) {
	r.logger.Debug("分页查询流程定义",
		zap.Any("filter", filter),
		zap.Any("options", opts))

	// 构建查询条件
	query := r.data.ProcessDefinition.Query()

	// 应用过滤条件
	if filter != nil {
		if filter.Name != "" {
			query = query.Where(processdefinition.NameContains(filter.Name))
		}
		if filter.Category != "" {
			query = query.Where(processdefinition.Category(filter.Category))
		}
		if filter.Version > 0 {
			query = query.Where(processdefinition.Version(int32(filter.Version)))
		}
		if filter.Status != "" {
			// 使用 Suspended 字段代替 Status
			suspended := filter.Status == "suspended"
			query = query.Where(processdefinition.Suspended(suspended))
		}
		if filter.CreatedFrom != nil {
			query = query.Where(processdefinition.CreatedAtGTE(*filter.CreatedFrom))
		}
		if filter.CreatedTo != nil {
			query = query.Where(processdefinition.CreatedAtLTE(*filter.CreatedTo))
		}
	}

	// 应用搜索条件
	if opts != nil && opts.Search != "" {
		searchTerm := "%" + strings.ToLower(opts.Search) + "%"
		query = query.Where(
			processdefinition.Or(
				processdefinition.NameContains(searchTerm),
				processdefinition.DescriptionContains(searchTerm),
				processdefinition.CategoryContains(searchTerm),
			),
		)
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("查询流程定义总数失败", zap.Error(err))
		return nil, nil, fmt.Errorf("查询流程定义总数失败: %w", err)
	}

	// 应用排序
	if opts != nil && opts.OrderBy != "" {
		switch opts.OrderBy {
		case "name":
			if opts.Order == "desc" {
				query = query.Order(ent.Desc(processdefinition.FieldName))
			} else {
				query = query.Order(ent.Asc(processdefinition.FieldName))
			}
		case "created_at":
			if opts.Order == "desc" {
				query = query.Order(ent.Desc(processdefinition.FieldCreatedAt))
			} else {
				query = query.Order(ent.Asc(processdefinition.FieldCreatedAt))
			}
		case "version":
			if opts.Order == "desc" {
				query = query.Order(ent.Desc(processdefinition.FieldVersion))
			} else {
				query = query.Order(ent.Asc(processdefinition.FieldVersion))
			}
		default:
			query = query.Order(ent.Desc(processdefinition.FieldCreatedAt))
		}
	} else {
		query = query.Order(ent.Desc(processdefinition.FieldCreatedAt))
	}

	// 应用分页
	if opts != nil && opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	// 执行查询
	results, err := query.All(ctx)
	if err != nil {
		r.logger.Error("查询流程定义失败", zap.Error(err))
		return nil, nil, fmt.Errorf("查询流程定义失败: %w", err)
	}

	// 构建分页结果
	pagination := &biz.PaginationResult{
		Total: total,
	}

	if opts != nil && opts.PageSize > 0 {
		pagination.Page = opts.Page
		pagination.PageSize = opts.PageSize
		pagination.Pages = (total + opts.PageSize - 1) / opts.PageSize
	}

	return results, pagination, nil
}

// Count 计数查询
func (r *processDefinitionRepo) Count(ctx context.Context, filter *biz.ProcessDefinitionFilter) (int, error) {
	r.logger.Debug("计数查询流程定义", zap.Any("filter", filter))

	query := r.data.ProcessDefinition.Query()

	// 应用过滤条件
	if filter != nil {
		if filter.Name != "" {
			query = query.Where(processdefinition.NameContains(filter.Name))
		}
		if filter.Category != "" {
			query = query.Where(processdefinition.Category(filter.Category))
		}
		if filter.Version > 0 {
			query = query.Where(processdefinition.Version(int32(filter.Version)))
		}
		if filter.Status != "" {
			// 使用 Suspended 字段代替 Status
			suspended := filter.Status == "suspended"
			query = query.Where(processdefinition.Suspended(suspended))
		}
		if filter.CreatedFrom != nil {
			query = query.Where(processdefinition.CreatedAtGTE(*filter.CreatedFrom))
		}
		if filter.CreatedTo != nil {
			query = query.Where(processdefinition.CreatedAtLTE(*filter.CreatedTo))
		}
	}

	count, err := query.Count(ctx)
	if err != nil {
		r.logger.Error("计数查询流程定义失败", zap.Error(err))
		return 0, fmt.Errorf("计数查询流程定义失败: %w", err)
	}

	return count, nil
}

// Deploy 部署流程定义（设置为激活状态）
func (r *processDefinitionRepo) Deploy(ctx context.Context, id string) error {
	r.logger.Info("部署流程定义", zap.String("id", id))

	// 将字符串ID转换为int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的流程定义ID: %s", id)
	}

	_, err = r.data.ProcessDefinition.
		UpdateOneID(idInt).
		SetSuspended(false).
		SetDeployTime(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Warn("流程定义不存在", zap.String("id", id))
			return fmt.Errorf("流程定义不存在: %s", id)
		}
		r.logger.Error("部署流程定义失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("部署流程定义失败: %w", err)
	}

	r.logger.Info("流程定义部署成功", zap.String("id", id))
	return nil
}

// Suspend 挂起流程定义
func (r *processDefinitionRepo) Suspend(ctx context.Context, id string) error {
	r.logger.Info("挂起流程定义", zap.String("id", id))

	// 将字符串ID转换为int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的流程定义ID: %s", id)
	}

	_, err = r.data.ProcessDefinition.
		UpdateOneID(idInt).
		SetSuspended(true).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.logger.Warn("流程定义不存在", zap.String("id", id))
			return fmt.Errorf("流程定义不存在: %s", id)
		}
		r.logger.Error("挂起流程定义失败", zap.String("id", id), zap.Error(err))
		return fmt.Errorf("挂起流程定义失败: %w", err)
	}

	r.logger.Info("流程定义挂起成功", zap.String("id", id))
	return nil
}
