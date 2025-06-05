// Package service 服务层依赖注入配置
// 使用Wire进行依赖注入，连接业务逻辑层和服务层
//go:build wireinject
// +build wireinject

package service

import (
	"github.com/google/wire"
	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/biz"
)

// ServiceSet 服务层依赖注入集合
// 包含所有服务层的构造函数
var ServiceSet = wire.NewSet(
	NewProcessDefinitionService,
	NewProcessInstanceService,
	NewTaskInstanceService,
	NewHistoricDataService,
	NewEventMessageService,
)

// ProcessDefinitionService 流程定义服务依赖注入
func NewProcessDefinitionService(
	uc *biz.ProcessDefinitionUseCase,
	logger *zap.Logger,
) *ProcessDefinitionService {
	return &ProcessDefinitionService{
		uc:     uc,
		logger: logger,
	}
}

// ProcessInstanceService 流程实例服务依赖注入
func NewProcessInstanceService(
	uc *biz.ProcessInstanceUseCase,
	logger *zap.Logger,
) *ProcessInstanceService {
	return &ProcessInstanceService{
		uc:     uc,
		logger: logger,
	}
}

// TaskInstanceService 任务实例服务依赖注入
func NewTaskInstanceService(
	uc *biz.TaskInstanceUseCase,
	logger *zap.Logger,
) *TaskInstanceService {
	return &TaskInstanceService{
		uc:     uc,
		logger: logger,
	}
}

// HistoricDataService 历史数据服务依赖注入
func NewHistoricDataService(
	uc *biz.HistoricDataUseCase,
	logger *zap.Logger,
) *HistoricDataService {
	return &HistoricDataService{
		uc:     uc,
		logger: logger,
	}
}

// EventMessageService 事件消息服务依赖注入
func NewEventMessageService(
	uc *biz.EventMessageUseCase,
	logger *zap.Logger,
) *EventMessageService {
	return &EventMessageService{
		uc:     uc,
		logger: logger,
	}
}

// ServiceContainer 服务容器
// 包含所有服务层实例
type ServiceContainer struct {
	ProcessDefinitionService *ProcessDefinitionService
	ProcessInstanceService   *ProcessInstanceService
	TaskInstanceService      *TaskInstanceService
	HistoricDataService      *HistoricDataService
	EventMessageService      *EventMessageService
	Logger                   *zap.Logger
}

// NewServiceContainer 创建服务容器
// 通过Wire自动注入所有依赖
func NewServiceContainer(
	processDefService *ProcessDefinitionService,
	processInstService *ProcessInstanceService,
	taskService *TaskInstanceService,
	historicService *HistoricDataService,
	eventService *EventMessageService,
	logger *zap.Logger,
) *ServiceContainer {
	return &ServiceContainer{
		ProcessDefinitionService: processDefService,
		ProcessInstanceService:   processInstService,
		TaskInstanceService:      taskService,
		HistoricDataService:      historicService,
		EventMessageService:      eventService,
		Logger:                   logger,
	}
}

// InitializeServiceContainer 初始化服务容器
// 使用Wire生成的代码进行依赖注入
func InitializeServiceContainer(
	processDefUC *biz.ProcessDefinitionUseCase,
	processInstUC *biz.ProcessInstanceUseCase,
	taskUC *biz.TaskInstanceUseCase,
	historicUC *biz.HistoricDataUseCase,
	eventUC *biz.EventMessageUseCase,
	logger *zap.Logger,
) *ServiceContainer {
	wire.Build(
		ServiceSet,
		NewServiceContainer,
	)
	return &ServiceContainer{}
}
