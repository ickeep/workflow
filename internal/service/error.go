// Package service 服务层错误定义
// 定义服务层的错误类型和错误码常量
package service

import (
	"fmt"
)

// ServiceError 服务层错误类型
// 统一的服务层错误结构，包含错误码和错误信息
type ServiceError struct {
	Code    int    `json:"code"`    // 错误码
	Message string `json:"message"` // 错误信息
	Details string `json:"details"` // 错误详情
}

// Error 实现 error 接口
func (e *ServiceError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewServiceError 创建服务层错误
func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
	}
}

// NewServiceErrorWithDetails 创建带详情的服务层错误
func NewServiceErrorWithDetails(code int, message, details string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// 错误码常量定义
const (
	// 成功响应
	ErrCodeSuccess = 200

	// 客户端错误 (4xx)
	ErrCodeBadRequest       = 400 // 请求参数错误
	ErrCodeUnauthorized     = 401 // 未授权
	ErrCodeForbidden        = 403 // 禁止访问
	ErrCodeNotFound         = 404 // 资源未找到
	ErrCodeMethodNotAllowed = 405 // 方法不允许
	ErrCodeConflict         = 409 // 资源冲突
	ErrCodeValidationError  = 422 // 参数验证错误
	ErrCodeTooManyRequests  = 429 // 请求过于频繁

	// 服务器错误 (5xx)
	ErrCodeInternalError      = 500 // 内部服务器错误
	ErrCodeNotImplemented     = 501 // 功能未实现
	ErrCodeBadGateway         = 502 // 网关错误
	ErrCodeServiceUnavailable = 503 // 服务不可用
	ErrCodeGatewayTimeout     = 504 // 网关超时

	// 业务错误 (6xx)
	ErrCodeBusinessError        = 600 // 通用业务错误
	ErrCodeWorkflowNotFound     = 601 // 工作流未找到
	ErrCodeWorkflowSuspended    = 602 // 工作流已挂起
	ErrCodeTaskNotFound         = 603 // 任务未找到
	ErrCodeTaskAlreadyClaimed   = 604 // 任务已被认领
	ErrCodeTaskNotAssigned      = 605 // 任务未分配
	ErrCodeProcessNotStarted    = 606 // 流程未启动
	ErrCodeProcessAlreadyEnded  = 607 // 流程已结束
	ErrCodeInvalidVariable      = 608 // 无效的变量
	ErrCodeInvalidConfiguration = 609 // 无效的配置
)

// 错误信息映射
var ErrorMessages = map[int]string{
	ErrCodeSuccess:              "成功",
	ErrCodeBadRequest:           "请求参数错误",
	ErrCodeUnauthorized:         "未授权访问",
	ErrCodeForbidden:            "禁止访问",
	ErrCodeNotFound:             "资源未找到",
	ErrCodeMethodNotAllowed:     "方法不允许",
	ErrCodeConflict:             "资源冲突",
	ErrCodeValidationError:      "参数验证错误",
	ErrCodeTooManyRequests:      "请求过于频繁",
	ErrCodeInternalError:        "内部服务器错误",
	ErrCodeNotImplemented:       "功能未实现",
	ErrCodeBadGateway:           "网关错误",
	ErrCodeServiceUnavailable:   "服务不可用",
	ErrCodeGatewayTimeout:       "网关超时",
	ErrCodeBusinessError:        "业务处理错误",
	ErrCodeWorkflowNotFound:     "工作流未找到",
	ErrCodeWorkflowSuspended:    "工作流已挂起",
	ErrCodeTaskNotFound:         "任务未找到",
	ErrCodeTaskAlreadyClaimed:   "任务已被认领",
	ErrCodeTaskNotAssigned:      "任务未分配",
	ErrCodeProcessNotStarted:    "流程未启动",
	ErrCodeProcessAlreadyEnded:  "流程已结束",
	ErrCodeInvalidVariable:      "无效的变量",
	ErrCodeInvalidConfiguration: "无效的配置",
}

// GetErrorMessage 根据错误码获取错误信息
func GetErrorMessage(code int) string {
	if message, exists := ErrorMessages[code]; exists {
		return message
	}
	return "未知错误"
}

// IsClientError 判断是否为客户端错误
func IsClientError(code int) bool {
	return code >= 400 && code < 500
}

// IsServerError 判断是否为服务器错误
func IsServerError(code int) bool {
	return code >= 500 && code < 600
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(code int) bool {
	return code >= 600 && code < 700
}

// WrapError 包装错误为服务层错误
func WrapError(err error, code int, message string) *ServiceError {
	if serviceErr, ok := err.(*ServiceError); ok {
		return serviceErr
	}
	return &ServiceError{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}

// 预定义常用错误
var (
	ErrBadRequest           = NewServiceError(ErrCodeBadRequest, "请求参数错误")
	ErrUnauthorized         = NewServiceError(ErrCodeUnauthorized, "未授权访问")
	ErrForbidden            = NewServiceError(ErrCodeForbidden, "禁止访问")
	ErrNotFound             = NewServiceError(ErrCodeNotFound, "资源未找到")
	ErrMethodNotAllowed     = NewServiceError(ErrCodeMethodNotAllowed, "方法不允许")
	ErrConflict             = NewServiceError(ErrCodeConflict, "资源冲突")
	ErrValidationError      = NewServiceError(ErrCodeValidationError, "参数验证错误")
	ErrTooManyRequests      = NewServiceError(ErrCodeTooManyRequests, "请求过于频繁")
	ErrInternalError        = NewServiceError(ErrCodeInternalError, "内部服务器错误")
	ErrNotImplemented       = NewServiceError(ErrCodeNotImplemented, "功能未实现")
	ErrBadGateway           = NewServiceError(ErrCodeBadGateway, "网关错误")
	ErrServiceUnavailable   = NewServiceError(ErrCodeServiceUnavailable, "服务不可用")
	ErrGatewayTimeout       = NewServiceError(ErrCodeGatewayTimeout, "网关超时")
	ErrBusinessError        = NewServiceError(ErrCodeBusinessError, "业务处理错误")
	ErrWorkflowNotFound     = NewServiceError(ErrCodeWorkflowNotFound, "工作流未找到")
	ErrWorkflowSuspended    = NewServiceError(ErrCodeWorkflowSuspended, "工作流已挂起")
	ErrTaskNotFound         = NewServiceError(ErrCodeTaskNotFound, "任务未找到")
	ErrTaskAlreadyClaimed   = NewServiceError(ErrCodeTaskAlreadyClaimed, "任务已被认领")
	ErrTaskNotAssigned      = NewServiceError(ErrCodeTaskNotAssigned, "任务未分配")
	ErrProcessNotStarted    = NewServiceError(ErrCodeProcessNotStarted, "流程未启动")
	ErrProcessAlreadyEnded  = NewServiceError(ErrCodeProcessAlreadyEnded, "流程已结束")
	ErrInvalidVariable      = NewServiceError(ErrCodeInvalidVariable, "无效的变量")
	ErrInvalidConfiguration = NewServiceError(ErrCodeInvalidConfiguration, "无效的配置")
)
