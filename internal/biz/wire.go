// Package biz 提供业务逻辑层功能
// Wire 依赖注入配置
//go:build wireinject
// +build wireinject

package biz

import (
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProviderSet 业务逻辑层的依赖注入提供器集合
var ProviderSet = wire.NewSet(
	NewProcessDefinitionUseCase,
	NewProcessInstanceUseCase,
	NewTaskInstanceUseCase,
	NewEventMessageUseCase,
	NewHistoricDataUseCase,
)

// NewBizContainer 创建业务逻辑容器
func NewBizContainer(
	processDefRepo ProcessDefinitionRepo,
	processInstanceRepo ProcessInstanceRepo,
	taskInstanceRepo TaskInstanceRepo,
	variableRepo ProcessVariableRepo,
	eventRepo ProcessEventRepo,
	historicRepo HistoricProcessInstanceRepo,
	cache CacheRepo,
	logger *zap.Logger,
) *BizContainer {
	return &BizContainer{
		ProcessDefinition: NewProcessDefinitionUseCase(processDefRepo, cache, logger),
		ProcessInstance:   NewProcessInstanceUseCase(processInstanceRepo, processDefRepo, variableRepo, cache, logger),
		TaskInstance:      NewTaskInstanceUseCase(taskInstanceRepo, processInstanceRepo, variableRepo, cache, logger),
		EventMessage:      NewEventMessageUseCase(eventRepo, cache, logger),
		HistoricData:      NewHistoricDataUseCase(historicRepo, cache, logger),
	}
}

// BizContainer 业务逻辑容器
type BizContainer struct {
	ProcessDefinition *ProcessDefinitionUseCase
	ProcessInstance   *ProcessInstanceUseCase
	TaskInstance      *TaskInstanceUseCase
	EventMessage      *EventMessageUseCase
	HistoricData      *HistoricDataUseCase
}
