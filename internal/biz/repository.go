// Package biz 提供业务逻辑层功能
// 定义 Repository 接口，用于数据访问层的抽象
package biz

import (
	"context"
	"time"

	"github.com/workflow-engine/workflow-engine/internal/data/ent"
)

// 通用查询选项
type QueryOptions struct {
	// 分页参数
	Page     int `json:"page"`      // 页码，从1开始
	PageSize int `json:"page_size"` // 每页大小
	// 排序参数
	OrderBy string `json:"order_by"` // 排序字段
	Order   string `json:"order"`    // 排序方向: asc, desc
	// 搜索参数
	Search string `json:"search"` // 搜索关键词
}

// 分页查询结果
type PaginationResult struct {
	Total    int `json:"total"`     // 总记录数
	Page     int `json:"page"`      // 当前页码
	PageSize int `json:"page_size"` // 每页大小
	Pages    int `json:"pages"`     // 总页数
}

// ProcessDefinitionFilter 流程定义过滤条件
type ProcessDefinitionFilter struct {
	Name        string     `json:"name,omitempty"`         // 按名称过滤
	Category    string     `json:"category,omitempty"`     // 按分类过滤
	Version     int        `json:"version,omitempty"`      // 按版本过滤
	Status      string     `json:"status,omitempty"`       // 按状态过滤
	CreatedFrom *time.Time `json:"created_from,omitempty"` // 创建时间起始
	CreatedTo   *time.Time `json:"created_to,omitempty"`   // 创建时间结束
}

// ProcessInstanceFilter 流程实例过滤条件
type ProcessInstanceFilter struct {
	ProcessDefinitionID string     `json:"process_definition_id,omitempty"` // 按流程定义ID过滤
	Status              string     `json:"status,omitempty"`                // 按状态过滤
	CreatedBy           string     `json:"created_by,omitempty"`            // 按创建者过滤
	StartedFrom         *time.Time `json:"started_from,omitempty"`          // 开始时间起始
	StartedTo           *time.Time `json:"started_to,omitempty"`            // 开始时间结束
}

// TaskInstanceFilter 任务实例过滤条件
type TaskInstanceFilter struct {
	ProcessInstanceID string     `json:"process_instance_id,omitempty"` // 按流程实例ID过滤
	AssigneeID        string     `json:"assignee_id,omitempty"`         // 按执行人过滤
	Status            string     `json:"status,omitempty"`              // 按状态过滤
	CreatedFrom       *time.Time `json:"created_from,omitempty"`        // 创建时间起始
	CreatedTo         *time.Time `json:"created_to,omitempty"`          // 创建时间结束
}

// ProcessDefinitionRepo 流程定义仓储接口
type ProcessDefinitionRepo interface {
	// 创建流程定义
	Create(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error)
	// 根据ID获取流程定义
	GetByID(ctx context.Context, id string) (*ent.ProcessDefinition, error)
	// 根据Key获取最新版本的流程定义
	GetLatestByKey(ctx context.Context, key string) (*ent.ProcessDefinition, error)
	// 根据Key和版本获取流程定义
	GetByKeyAndVersion(ctx context.Context, key string, version int) (*ent.ProcessDefinition, error)
	// 更新流程定义
	Update(ctx context.Context, pd *ent.ProcessDefinition) (*ent.ProcessDefinition, error)
	// 删除流程定义
	Delete(ctx context.Context, id string) error
	// 分页查询流程定义
	List(ctx context.Context, filter *ProcessDefinitionFilter, opts *QueryOptions) ([]*ent.ProcessDefinition, *PaginationResult, error)
	// 计数查询
	Count(ctx context.Context, filter *ProcessDefinitionFilter) (int, error)
	// 部署流程定义（设置为激活状态）
	Deploy(ctx context.Context, id string) error
	// 挂起流程定义
	Suspend(ctx context.Context, id string) error
}

// ProcessInstanceRepo 流程实例仓储接口
type ProcessInstanceRepo interface {
	// 创建流程实例
	Create(ctx context.Context, pi *ent.ProcessInstance) (*ent.ProcessInstance, error)
	// 根据ID获取流程实例
	GetByID(ctx context.Context, id string) (*ent.ProcessInstance, error)
	// 更新流程实例
	Update(ctx context.Context, pi *ent.ProcessInstance) (*ent.ProcessInstance, error)
	// 删除流程实例
	Delete(ctx context.Context, id string) error
	// 分页查询流程实例
	List(ctx context.Context, filter *ProcessInstanceFilter, opts *QueryOptions) ([]*ent.ProcessInstance, *PaginationResult, error)
	// 计数查询
	Count(ctx context.Context, filter *ProcessInstanceFilter) (int, error)
	// 根据流程定义ID查询流程实例
	ListByProcessDefinitionID(ctx context.Context, processDefinitionID string, opts *QueryOptions) ([]*ent.ProcessInstance, *PaginationResult, error)
	// 挂起流程实例
	Suspend(ctx context.Context, id string) error
	// 激活流程实例
	Activate(ctx context.Context, id string) error
	// 终止流程实例
	Terminate(ctx context.Context, id string, reason string) error
}

// TaskInstanceRepo 任务实例仓储接口
type TaskInstanceRepo interface {
	// 创建任务实例
	Create(ctx context.Context, ti *ent.TaskInstance) (*ent.TaskInstance, error)
	// 根据ID获取任务实例
	GetByID(ctx context.Context, id string) (*ent.TaskInstance, error)
	// 更新任务实例
	Update(ctx context.Context, ti *ent.TaskInstance) (*ent.TaskInstance, error)
	// 删除任务实例
	Delete(ctx context.Context, id string) error
	// 分页查询任务实例
	List(ctx context.Context, filter *TaskInstanceFilter, opts *QueryOptions) ([]*ent.TaskInstance, *PaginationResult, error)
	// 计数查询
	Count(ctx context.Context, filter *TaskInstanceFilter) (int, error)
	// 根据流程实例ID查询任务实例
	ListByProcessInstanceID(ctx context.Context, processInstanceID string, opts *QueryOptions) ([]*ent.TaskInstance, *PaginationResult, error)
	// 根据执行人查询任务实例
	ListByAssignee(ctx context.Context, assigneeID string, opts *QueryOptions) ([]*ent.TaskInstance, *PaginationResult, error)
	// 认领任务
	Claim(ctx context.Context, id string, assigneeID string) error
	// 完成任务
	Complete(ctx context.Context, id string, variables map[string]interface{}) error
	// 委派任务
	Delegate(ctx context.Context, id string, delegateID string) error
}

// HistoricProcessInstanceRepo 历史流程实例仓储接口
type HistoricProcessInstanceRepo interface {
	// 创建历史流程实例
	Create(ctx context.Context, hpi *ent.HistoricProcessInstance) (*ent.HistoricProcessInstance, error)
	// 根据ID获取历史流程实例
	GetHistoricProcessInstance(ctx context.Context, id int64) (*ent.HistoricProcessInstance, error)
	// 分页查询历史流程实例
	ListHistoricProcessInstances(ctx context.Context, filter *HistoricProcessInstanceFilter) ([]*ent.HistoricProcessInstance, int, error)
	// 计数查询
	Count(ctx context.Context, filter *HistoricProcessInstanceFilter) (int, error)
	// 根据流程定义ID查询历史流程实例
	ListByProcessDefinitionID(ctx context.Context, processDefinitionID string, opts *QueryOptions) ([]*ent.HistoricProcessInstance, *PaginationResult, error)
	// 删除历史流程实例
	DeleteHistoricProcessInstance(ctx context.Context, id int64) error
	// 批量删除历史流程实例
	BatchDeleteHistoricProcessInstances(ctx context.Context, processDefinitionKey string, endTimeBefore time.Time) (int64, error)
	// 获取历史变量
	GetHistoricVariables(ctx context.Context, processInstanceID int64) ([]*HistoricVariableInstance, error)
	// 获取流程统计信息
	GetProcessStatistics(ctx context.Context, processDefinitionKey string, startTime, endTime time.Time) (*ProcessStatistics, error)
	// 获取流程趋势分析
	GetProcessTrend(ctx context.Context, processDefinitionKey string, startTime, endTime time.Time, granularity string) ([]*ProcessTrendData, error)
}

// ProcessVariableRepo 流程变量仓储接口
type ProcessVariableRepo interface {
	// 创建流程变量
	Create(ctx context.Context, pv *ent.ProcessVariable) (*ent.ProcessVariable, error)
	// 根据ID获取流程变量
	GetByID(ctx context.Context, id string) (*ent.ProcessVariable, error)
	// 更新流程变量
	Update(ctx context.Context, pv *ent.ProcessVariable) (*ent.ProcessVariable, error)
	// 删除流程变量
	Delete(ctx context.Context, id string) error
	// 根据流程实例ID和变量名获取变量
	GetByProcessInstanceIDAndName(ctx context.Context, processInstanceID, name string) (*ent.ProcessVariable, error)
	// 根据流程实例ID获取所有变量
	ListByProcessInstanceID(ctx context.Context, processInstanceID string) ([]*ent.ProcessVariable, error)
	// 批量设置流程变量
	SetVariables(ctx context.Context, processInstanceID string, variables map[string]interface{}) error
	// 删除流程实例的所有变量
	DeleteByProcessInstanceID(ctx context.Context, processInstanceID string) error
}

// ProcessEventRepo 流程事件仓储接口
type ProcessEventRepo interface {
	// 创建流程事件
	Create(ctx context.Context, pe *ent.ProcessEvent) (*ent.ProcessEvent, error)
	// 根据ID获取流程事件
	GetByID(ctx context.Context, id string) (*ent.ProcessEvent, error)
	// 分页查询流程事件
	List(ctx context.Context, processInstanceID string, opts *QueryOptions) ([]*ent.ProcessEvent, *PaginationResult, error)
	// 根据流程实例ID查询事件
	ListByProcessInstanceID(ctx context.Context, processInstanceID string) ([]*ent.ProcessEvent, error)
	// 根据事件类型查询事件
	ListByEventType(ctx context.Context, eventType string, opts *QueryOptions) ([]*ent.ProcessEvent, *PaginationResult, error)
	// 删除流程事件
	Delete(ctx context.Context, id string) error
	// 删除流程实例的所有事件
	DeleteByProcessInstanceID(ctx context.Context, processInstanceID string) error
}

// CacheRepo 缓存仓储接口
type CacheRepo interface {
	// 设置缓存
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	// 获取缓存
	Get(ctx context.Context, key string) (string, error)
	// 删除缓存
	Delete(ctx context.Context, key string) error
	// 检查缓存是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// 设置缓存过期时间
	Expire(ctx context.Context, key string, expiration time.Duration) error
	// 获取哈希缓存
	HGet(ctx context.Context, key, field string) (string, error)
	// 设置哈希缓存
	HSet(ctx context.Context, key, field string, value interface{}) error
	// 删除哈希缓存字段
	HDel(ctx context.Context, key string, fields ...string) error
	// 获取哈希缓存所有字段
	HGetAll(ctx context.Context, key string) (map[string]string, error)
}

// TransactionRepo 事务仓储接口
type TransactionRepo interface {
	// 执行事务
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	// 开始事务
	Begin(ctx context.Context) (context.Context, error)
	// 提交事务
	Commit(ctx context.Context) error
	// 回滚事务
	Rollback(ctx context.Context) error
}
