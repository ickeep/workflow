// Package biz 提供业务逻辑层功能
// 定义数据传输对象 (DTO) 和请求响应结构
package biz

import (
	"time"
)

// 流程定义相关的请求响应结构

// CreateProcessDefinitionRequest 创建流程定义请求
type CreateProcessDefinitionRequest struct {
	Key         string `json:"key" validate:"required"`      // 流程唯一标识
	Name        string `json:"name" validate:"required"`     // 流程名称
	Description string `json:"description"`                  // 流程描述
	Category    string `json:"category"`                     // 流程分类
	Resource    string `json:"resource" validate:"required"` // 流程资源(JSON格式)
	TenantID    string `json:"tenant_id"`                    // 租户ID
	Version     int32  `json:"-"`                            // 版本号(内部使用)
}

// UpdateProcessDefinitionRequest 更新流程定义请求
type UpdateProcessDefinitionRequest struct {
	Name        string `json:"name"`        // 流程名称
	Description string `json:"description"` // 流程描述
	Category    string `json:"category"`    // 流程分类
	Resource    string `json:"resource"`    // 流程资源(JSON格式)
}

// ProcessDefinitionResponse 流程定义响应
type ProcessDefinitionResponse struct {
	ID          string    `json:"id"`          // 流程定义ID
	Key         string    `json:"key"`         // 流程唯一标识
	Name        string    `json:"name"`        // 流程名称
	Description string    `json:"description"` // 流程描述
	Category    string    `json:"category"`    // 流程分类
	Version     int32     `json:"version"`     // 版本号
	Resource    string    `json:"resource"`    // 流程资源
	Suspended   bool      `json:"suspended"`   // 是否挂起
	TenantID    string    `json:"tenant_id"`   // 租户ID
	DeployTime  time.Time `json:"deploy_time"` // 部署时间
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// ListProcessDefinitionsRequest 查询流程定义列表请求
type ListProcessDefinitionsRequest struct {
	// 分页参数
	Page     int `json:"page" validate:"min=1"`      // 页码，从1开始
	PageSize int `json:"page_size" validate:"min=1"` // 每页大小

	// 排序参数
	OrderBy string `json:"order_by"` // 排序字段：name, created_at, version
	Order   string `json:"order"`    // 排序方向：asc, desc

	// 搜索参数
	Search string `json:"search"` // 搜索关键词

	// 过滤参数
	Name        string     `json:"name"`         // 按名称过滤
	Category    string     `json:"category"`     // 按分类过滤
	Status      string     `json:"status"`       // 按状态过滤：active, suspended
	CreatedFrom *time.Time `json:"created_from"` // 创建时间起始
	CreatedTo   *time.Time `json:"created_to"`   // 创建时间结束
}

// ListProcessDefinitionsResponse 查询流程定义列表响应
type ListProcessDefinitionsResponse struct {
	Items      []*ProcessDefinitionResponse `json:"items"`      // 流程定义列表
	Pagination *PaginationResult            `json:"pagination"` // 分页信息
}

// 流程实例相关的请求响应结构

// StartProcessInstanceRequest 启动流程实例请求
type StartProcessInstanceRequest struct {
	ProcessDefinitionID  string                 `json:"process_definition_id"`  // 流程定义ID
	ProcessDefinitionKey string                 `json:"process_definition_key"` // 流程定义键
	BusinessKey          string                 `json:"business_key"`           // 业务键
	Variables            map[string]interface{} `json:"variables"`              // 流程变量
	Name                 string                 `json:"name"`                   // 实例名称
	Description          string                 `json:"description"`            // 实例描述
	TenantID             string                 `json:"tenant_id"`              // 租户ID
}

// ProcessInstanceResponse 流程实例响应
type ProcessInstanceResponse struct {
	ID                  string                 `json:"id"`                    // 流程实例ID
	ProcessDefinitionID string                 `json:"process_definition_id"` // 流程定义ID
	BusinessKey         string                 `json:"business_key"`          // 业务键
	StartUserID         string                 `json:"start_user_id"`         // 启动用户ID
	StartTime           time.Time              `json:"start_time"`            // 启动时间
	EndTime             *time.Time             `json:"end_time"`              // 结束时间
	Duration            *int64                 `json:"duration"`              // 持续时间(毫秒)
	DeleteReason        string                 `json:"delete_reason"`         // 删除原因
	ActivityID          string                 `json:"activity_id"`           // 当前活动节点ID
	Name                string                 `json:"name"`                  // 实例名称
	Description         string                 `json:"description"`           // 实例描述
	IsActive            bool                   `json:"is_active"`             // 是否激活
	IsEnded             bool                   `json:"is_ended"`              // 是否结束
	IsSuspended         bool                   `json:"is_suspended"`          // 是否挂起
	TenantID            string                 `json:"tenant_id"`             // 租户ID
	Variables           map[string]interface{} `json:"variables"`             // 流程变量
	CreatedAt           time.Time              `json:"created_at"`            // 创建时间
	UpdatedAt           time.Time              `json:"updated_at"`            // 更新时间
}

// ListProcessInstancesRequest 查询流程实例列表请求
type ListProcessInstancesRequest struct {
	// 分页参数
	Page     int `json:"page" validate:"min=1"`      // 页码
	PageSize int `json:"page_size" validate:"min=1"` // 每页大小

	// 排序参数
	OrderBy string `json:"order_by"` // 排序字段
	Order   string `json:"order"`    // 排序方向

	// 搜索参数
	Search string `json:"search"` // 搜索关键词

	// 过滤参数
	ProcessDefinitionID string     `json:"process_definition_id"` // 按流程定义ID过滤
	BusinessKey         string     `json:"business_key"`          // 按业务键过滤
	StartUserID         string     `json:"start_user_id"`         // 按启动用户过滤
	IsActive            *bool      `json:"is_active"`             // 按激活状态过滤
	IsEnded             *bool      `json:"is_ended"`              // 按结束状态过滤
	IsSuspended         *bool      `json:"is_suspended"`          // 按挂起状态过滤
	StartedFrom         *time.Time `json:"started_from"`          // 开始时间起始
	StartedTo           *time.Time `json:"started_to"`            // 开始时间结束
}

// ListProcessInstancesResponse 查询流程实例列表响应
type ListProcessInstancesResponse struct {
	Items      []*ProcessInstanceResponse `json:"items"`      // 流程实例列表
	Pagination *PaginationResult          `json:"pagination"` // 分页信息
}

// 任务实例相关的请求响应结构

// TaskInstanceResponse 任务实例响应
type TaskInstanceResponse struct {
	ID                  string                 `json:"id"`                    // 任务ID
	ProcessInstanceID   string                 `json:"process_instance_id"`   // 流程实例ID
	ProcessDefinitionID string                 `json:"process_definition_id"` // 流程定义ID
	Name                string                 `json:"name"`                  // 任务名称
	Description         string                 `json:"description"`           // 任务描述
	TaskDefinitionKey   string                 `json:"task_definition_key"`   // 任务定义键
	Priority            int32                  `json:"priority"`              // 优先级
	CreateTime          time.Time              `json:"create_time"`           // 创建时间
	ClaimTime           *time.Time             `json:"claim_time"`            // 认领时间
	DueDate             *time.Time             `json:"due_date"`              // 到期时间
	Category            string                 `json:"category"`              // 任务分类
	Owner               string                 `json:"owner"`                 // 拥有者
	Assignee            string                 `json:"assignee"`              // 委派人
	Delegation          string                 `json:"delegation"`            // 委派状态
	FormKey             string                 `json:"form_key"`              // 表单键
	IsSuspended         bool                   `json:"is_suspended"`          // 是否挂起
	TenantID            string                 `json:"tenant_id"`             // 租户ID
	Variables           map[string]interface{} `json:"variables"`             // 任务变量
	CreatedAt           time.Time              `json:"created_at"`            // 创建时间
	UpdatedAt           time.Time              `json:"updated_at"`            // 更新时间
}

// CompleteTaskRequest 完成任务请求
type CompleteTaskRequest struct {
	Variables map[string]interface{} `json:"variables"` // 任务变量
	Comment   string                 `json:"comment"`   // 完成备注
}

// ClaimTaskRequest 认领任务请求
type ClaimTaskRequest struct {
	AssigneeID string `json:"assignee_id" validate:"required"` // 认领人ID
}

// DelegateTaskRequest 委派任务请求
type DelegateTaskRequest struct {
	DelegateID string `json:"delegate_id" validate:"required"` // 委派人ID
	Comment    string `json:"comment"`                         // 委派备注
}

// ListTaskInstancesRequest 查询任务实例列表请求
type ListTaskInstancesRequest struct {
	// 分页参数
	Page     int `json:"page" validate:"min=1"`      // 页码
	PageSize int `json:"page_size" validate:"min=1"` // 每页大小

	// 排序参数
	OrderBy string `json:"order_by"` // 排序字段
	Order   string `json:"order"`    // 排序方向

	// 搜索参数
	Search string `json:"search"` // 搜索关键词

	// 过滤参数
	ProcessInstanceID string     `json:"process_instance_id"` // 按流程实例ID过滤
	AssigneeID        string     `json:"assignee_id"`         // 按执行人过滤
	Owner             string     `json:"owner"`               // 按拥有者过滤
	Category          string     `json:"category"`            // 按分类过滤
	IsSuspended       *bool      `json:"is_suspended"`        // 按挂起状态过滤
	CreatedFrom       *time.Time `json:"created_from"`        // 创建时间起始
	CreatedTo         *time.Time `json:"created_to"`          // 创建时间结束
}

// ListTaskInstancesResponse 查询任务实例列表响应
type ListTaskInstancesResponse struct {
	Items      []*TaskInstanceResponse `json:"items"`      // 任务实例列表
	Pagination *PaginationResult       `json:"pagination"` // 分页信息
}

// 历史数据相关的请求响应结构

// HistoricProcessInstanceResponse 历史流程实例响应
type HistoricProcessInstanceResponse struct {
	ID                       string                 `json:"id"`                         // 历史ID
	ProcessInstanceID        string                 `json:"process_instance_id"`        // 流程实例ID
	ProcessDefinitionID      string                 `json:"process_definition_id"`      // 流程定义ID
	ProcessDefinitionKey     string                 `json:"process_definition_key"`     // 流程定义键
	ProcessDefinitionName    string                 `json:"process_definition_name"`    // 流程定义名称
	ProcessDefinitionVersion int32                  `json:"process_definition_version"` // 流程定义版本
	BusinessKey              string                 `json:"business_key"`               // 业务键
	StartTime                time.Time              `json:"start_time"`                 // 开始时间
	EndTime                  *time.Time             `json:"end_time"`                   // 结束时间
	Duration                 *int64                 `json:"duration"`                   // 持续时间
	StartUserID              string                 `json:"start_user_id"`              // 启动用户
	DeleteReason             string                 `json:"delete_reason"`              // 删除原因
	TenantID                 string                 `json:"tenant_id"`                  // 租户ID
	Name                     string                 `json:"name"`                       // 名称
	Description              string                 `json:"description"`                // 描述
	Variables                map[string]interface{} `json:"variables"`                  // 流程变量
	CreatedAt                time.Time              `json:"created_at"`                 // 创建时间
	UpdatedAt                time.Time              `json:"updated_at"`                 // 更新时间
}

// ListHistoricProcessInstancesRequest 查询历史流程实例列表请求
type ListHistoricProcessInstancesRequest struct {
	// 分页参数
	Page     int `json:"page" validate:"min=1"`      // 页码
	PageSize int `json:"page_size" validate:"min=1"` // 每页大小

	// 排序参数
	OrderBy        string `json:"order_by"`        // 排序字段
	OrderDirection string `json:"order_direction"` // 排序方向

	// 过滤参数
	ProcessDefinitionID  string     `json:"process_definition_id"`  // 按流程定义ID过滤
	ProcessDefinitionKey string     `json:"process_definition_key"` // 按流程定义键过滤
	BusinessKey          string     `json:"business_key"`           // 按业务键过滤
	StartUserID          string     `json:"start_user_id"`          // 按启动用户过滤
	State                string     `json:"state"`                  // 按状态过滤
	StartTimeAfter       *time.Time `json:"start_time_after"`       // 开始时间之后
	StartTimeBefore      *time.Time `json:"start_time_before"`      // 开始时间之前
	EndTimeAfter         *time.Time `json:"end_time_after"`         // 结束时间之后
	EndTimeBefore        *time.Time `json:"end_time_before"`        // 结束时间之前
	TenantID             string     `json:"tenant_id"`              // 租户ID
}

// ListHistoricProcessInstancesResponse 查询历史流程实例列表响应
type ListHistoricProcessInstancesResponse struct {
	Items    []*HistoricProcessInstanceResponse `json:"items"`     // 历史流程实例列表
	Total    int                                `json:"total"`     // 总记录数
	Page     int                                `json:"page"`      // 当前页码
	PageSize int                                `json:"page_size"` // 每页大小
}

// HistoricProcessInstanceFilter 历史流程实例过滤条件
type HistoricProcessInstanceFilter struct {
	ProcessDefinitionID  string     `json:"process_definition_id,omitempty"`  // 按流程定义ID过滤
	ProcessDefinitionKey string     `json:"process_definition_key,omitempty"` // 按流程定义键过滤
	BusinessKey          string     `json:"business_key,omitempty"`           // 按业务键过滤
	StartUserID          string     `json:"start_user_id,omitempty"`          // 按启动用户过滤
	State                string     `json:"state,omitempty"`                  // 按状态过滤
	StartTimeAfter       *time.Time `json:"start_time_after,omitempty"`       // 开始时间之后
	StartTimeBefore      *time.Time `json:"start_time_before,omitempty"`      // 开始时间之前
	EndTimeAfter         *time.Time `json:"end_time_after,omitempty"`         // 结束时间之后
	EndTimeBefore        *time.Time `json:"end_time_before,omitempty"`        // 结束时间之前
	TenantID             string     `json:"tenant_id,omitempty"`              // 租户ID
	Page                 int        `json:"page"`                             // 页码
	PageSize             int        `json:"page_size"`                        // 每页大小
	OrderBy              string     `json:"order_by"`                         // 排序字段
	OrderDirection       string     `json:"order_direction"`                  // 排序方向
}

// ProcessStatisticsRequest 流程统计请求
type ProcessStatisticsRequest struct {
	ProcessDefinitionKey string    `json:"process_definition_key"` // 流程定义键
	StartTime            time.Time `json:"start_time"`             // 统计开始时间
	EndTime              time.Time `json:"end_time"`               // 统计结束时间
}

// ProcessStatisticsResponse 流程统计响应
type ProcessStatisticsResponse struct {
	ProcessDefinitionKey string        `json:"process_definition_key"` // 流程定义键
	StartTime            time.Time     `json:"start_time"`             // 统计开始时间
	EndTime              time.Time     `json:"end_time"`               // 统计结束时间
	TotalInstances       int64         `json:"total_instances"`        // 总实例数
	CompletedInstances   int64         `json:"completed_instances"`    // 已完成实例数
	ActiveInstances      int64         `json:"active_instances"`       // 活跃实例数
	SuspendedInstances   int64         `json:"suspended_instances"`    // 挂起实例数
	TerminatedInstances  int64         `json:"terminated_instances"`   // 终止实例数
	CompletionRate       float64       `json:"completion_rate"`        // 完成率
	AverageDuration      time.Duration `json:"average_duration"`       // 平均持续时间
	MinDuration          time.Duration `json:"min_duration"`           // 最短持续时间
	MaxDuration          time.Duration `json:"max_duration"`           // 最长持续时间
}

// ProcessTrendRequest 流程趋势请求
type ProcessTrendRequest struct {
	ProcessDefinitionKey string    `json:"process_definition_key"` // 流程定义键
	StartTime            time.Time `json:"start_time"`             // 开始时间
	EndTime              time.Time `json:"end_time"`               // 结束时间
	Granularity          string    `json:"granularity"`            // 时间粒度: hour, day, week, month
}

// ProcessTrendResponse 流程趋势响应
type ProcessTrendResponse struct {
	ProcessDefinitionKey string              `json:"process_definition_key"` // 流程定义键
	StartTime            time.Time           `json:"start_time"`             // 开始时间
	EndTime              time.Time           `json:"end_time"`               // 结束时间
	Granularity          string              `json:"granularity"`            // 时间粒度
	TrendData            []*ProcessTrendData `json:"trend_data"`             // 趋势数据
}

// ProcessTrendData 流程趋势数据
type ProcessTrendData struct {
	Time               time.Time `json:"time"`                // 时间点
	StartedInstances   int64     `json:"started_instances"`   // 启动实例数
	CompletedInstances int64     `json:"completed_instances"` // 完成实例数
	ActiveInstances    int64     `json:"active_instances"`    // 活跃实例数
}

// BatchDeleteHistoricProcessInstancesRequest 批量删除历史流程实例请求
type BatchDeleteHistoricProcessInstancesRequest struct {
	ProcessDefinitionKey string    `json:"process_definition_key"` // 流程定义键
	EndTimeBefore        time.Time `json:"end_time_before"`        // 结束时间之前
}

// BatchDeleteHistoricProcessInstancesResponse 批量删除历史流程实例响应
type BatchDeleteHistoricProcessInstancesResponse struct {
	DeletedCount int64 `json:"deleted_count"` // 删除数量
}

// HistoricVariableInstance 历史变量实例
type HistoricVariableInstance struct {
	ID          string   `json:"id"`           // 变量ID
	Name        string   `json:"name"`         // 变量名
	Type        string   `json:"type"`         // 变量类型
	TextValue   *string  `json:"text_value"`   // 文本值
	LongValue   *int64   `json:"long_value"`   // 长整型值
	DoubleValue *float64 `json:"double_value"` // 浮点型值
}

// ProcessStatistics 流程统计数据
type ProcessStatistics struct {
	TotalInstances      int64         `json:"total_instances"`      // 总实例数
	CompletedInstances  int64         `json:"completed_instances"`  // 已完成实例数
	ActiveInstances     int64         `json:"active_instances"`     // 活跃实例数
	SuspendedInstances  int64         `json:"suspended_instances"`  // 挂起实例数
	TerminatedInstances int64         `json:"terminated_instances"` // 终止实例数
	AverageDuration     time.Duration `json:"average_duration"`     // 平均持续时间
	MinDuration         time.Duration `json:"min_duration"`         // 最短持续时间
	MaxDuration         time.Duration `json:"max_duration"`         // 最长持续时间
}

// 通用响应结构

// BaseResponse 基础响应结构
type BaseResponse struct {
	Code    int    `json:"code"`    // 响应码
	Message string `json:"message"` // 响应消息
	Success bool   `json:"success"` // 是否成功
}

// DataResponse 数据响应结构
type DataResponse struct {
	BaseResponse
	Data interface{} `json:"data"` // 响应数据
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	BaseResponse
	Error   string      `json:"error"`             // 错误信息
	Details interface{} `json:"details,omitempty"` // 错误详情
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	BaseResponse
	Data interface{} `json:"data,omitempty"` // 响应数据
}

// 业务常量定义

// 流程实例状态
const (
	ProcessInstanceStatusActive    = "active"    // 激活
	ProcessInstanceStatusSuspended = "suspended" // 挂起
	ProcessInstanceStatusCompleted = "completed" // 完成
	ProcessInstanceStatusDeleted   = "deleted"   // 删除
)

// 任务状态
const (
	TaskStatusCreated   = "created"   // 已创建
	TaskStatusClaimed   = "claimed"   // 已认领
	TaskStatusCompleted = "completed" // 已完成
	TaskStatusCancelled = "cancelled" // 已取消
	TaskStatusSuspended = "suspended" // 挂起
)

// 任务委派状态
const (
	TaskDelegationPending   = "pending"   // 待处理
	TaskDelegationResolved  = "resolved"  // 已解决
	TaskDelegationCancelled = "cancelled" // 已取消
)

// 响应码定义
const (
	ResponseCodeSuccess            = 200 // 成功
	ResponseCodeBadRequest         = 400 // 请求错误
	ResponseCodeUnauthorized       = 401 // 未授权
	ResponseCodeForbidden          = 403 // 禁止访问
	ResponseCodeNotFound           = 404 // 未找到
	ResponseCodeConflict           = 409 // 冲突
	ResponseCodeValidationError    = 422 // 验证错误
	ResponseCodeInternalError      = 500 // 内部错误
	ResponseCodeServiceUnavailable = 503 // 服务不可用
)
