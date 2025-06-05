// Package service 事件消息服务实现
// 处理事件和消息相关的API请求并调用业务逻辑
package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/workflow-engine/workflow-engine/internal/biz"
)

// EventMessageService 事件消息服务
// 负责处理事件和消息相关的API请求，连接HTTP/gRPC层与业务逻辑层
type EventMessageService struct {
	uc     *biz.EventMessageUseCase
	logger *zap.Logger
}

// NewEventMessageService 创建事件消息服务
func NewEventMessageService(
	uc *biz.EventMessageUseCase,
	logger *zap.Logger,
) *EventMessageService {
	return &EventMessageService{
		uc:     uc,
		logger: logger,
	}
}

// PublishEvent 发布事件
// 发布流程事件到事件总线
func (s *EventMessageService) PublishEvent(ctx context.Context, event *biz.ProcessEvent) error {
	s.logger.Info("服务层: 发布事件",
		zap.String("event_type", event.EventType),
		zap.String("process_instance_id", event.ProcessInstanceID))

	// 参数验证
	if err := s.validateProcessEvent(event); err != nil {
		s.logger.Error("发布事件参数验证失败", zap.Error(err))
		return err
	}

	err := s.uc.PublishEvent(ctx, event)
	if err != nil {
		s.logger.Error("发布事件失败",
			zap.String("event_type", event.EventType),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "发布事件失败")
	}

	s.logger.Info("服务层: 发布事件成功",
		zap.String("event_type", event.EventType),
		zap.String("process_instance_id", event.ProcessInstanceID))
	return nil
}

// SendSignal 发送信号
// 向流程实例发送信号事件
func (s *EventMessageService) SendSignal(ctx context.Context, signal *biz.SignalEvent) error {
	s.logger.Info("服务层: 发送信号",
		zap.String("signal_name", signal.SignalName),
		zap.String("process_instance_id", signal.ProcessInstanceID))

	// 参数验证
	if err := s.validateSignalEvent(signal); err != nil {
		s.logger.Error("发送信号参数验证失败", zap.Error(err))
		return err
	}

	err := s.uc.SendSignal(ctx, signal)
	if err != nil {
		s.logger.Error("发送信号失败",
			zap.String("signal_name", signal.SignalName),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "发送信号失败")
	}

	s.logger.Info("服务层: 发送信号成功",
		zap.String("signal_name", signal.SignalName),
		zap.String("process_instance_id", signal.ProcessInstanceID))
	return nil
}

// SendMessage 发送消息
// 向流程实例发送消息事件
func (s *EventMessageService) SendMessage(ctx context.Context, message *biz.MessageEvent) error {
	s.logger.Info("服务层: 发送消息",
		zap.String("message_name", message.MessageName),
		zap.String("process_instance_id", message.ProcessInstanceID))

	// 参数验证
	if err := s.validateMessageEvent(message); err != nil {
		s.logger.Error("发送消息参数验证失败", zap.Error(err))
		return err
	}

	err := s.uc.SendMessage(ctx, message)
	if err != nil {
		s.logger.Error("发送消息失败",
			zap.String("message_name", message.MessageName),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "发送消息失败")
	}

	s.logger.Info("服务层: 发送消息成功",
		zap.String("message_name", message.MessageName),
		zap.String("process_instance_id", message.ProcessInstanceID))
	return nil
}

// ScheduleTimer 调度定时器
// 创建定时器事件
func (s *EventMessageService) ScheduleTimer(ctx context.Context, timer *biz.TimerEvent) error {
	s.logger.Info("服务层: 调度定时器",
		zap.String("timer_id", timer.TimerID),
		zap.String("process_instance_id", timer.ProcessInstanceID),
		zap.Time("due_date", timer.DueDate))

	// 参数验证
	if err := s.validateTimerEvent(timer); err != nil {
		s.logger.Error("调度定时器参数验证失败", zap.Error(err))
		return err
	}

	err := s.uc.ScheduleTimer(ctx, timer)
	if err != nil {
		s.logger.Error("调度定时器失败",
			zap.String("timer_id", timer.TimerID),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "调度定时器失败")
	}

	s.logger.Info("服务层: 调度定时器成功",
		zap.String("timer_id", timer.TimerID),
		zap.String("process_instance_id", timer.ProcessInstanceID))
	return nil
}

// CancelTimer 取消定时器
// 取消已调度的定时器事件
func (s *EventMessageService) CancelTimer(ctx context.Context, timerID string) error {
	s.logger.Info("服务层: 取消定时器", zap.String("timer_id", timerID))

	if timerID == "" {
		s.logger.Error("定时器ID不能为空")
		return NewServiceError(ErrCodeBadRequest, "定时器ID不能为空")
	}

	err := s.uc.CancelTimer(ctx, timerID)
	if err != nil {
		s.logger.Error("取消定时器失败",
			zap.String("timer_id", timerID),
			zap.Error(err))
		return WrapError(err, ErrCodeInternalError, "取消定时器失败")
	}

	s.logger.Info("服务层: 取消定时器成功", zap.String("timer_id", timerID))
	return nil
}

// GetProcessEvents 获取流程事件列表
// 获取指定流程实例的所有事件
func (s *EventMessageService) GetProcessEvents(ctx context.Context, processInstanceID string) ([]*biz.ProcessEvent, error) {
	s.logger.Debug("服务层: 获取流程事件列表",
		zap.String("process_instance_id", processInstanceID))

	if processInstanceID == "" {
		s.logger.Error("流程实例ID不能为空")
		return nil, NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}

	result, err := s.uc.GetProcessEvents(ctx, processInstanceID)
	if err != nil {
		s.logger.Error("获取流程事件列表失败",
			zap.String("process_instance_id", processInstanceID),
			zap.Error(err))
		return nil, WrapError(err, ErrCodeInternalError, "获取流程事件列表失败")
	}

	s.logger.Info("服务层: 获取流程事件列表成功",
		zap.String("process_instance_id", processInstanceID),
		zap.Int("event_count", len(result)))
	return result, nil
}

// validateProcessEvent 验证流程事件参数
func (s *EventMessageService) validateProcessEvent(event *biz.ProcessEvent) error {
	if event.EventType == "" {
		return NewServiceError(ErrCodeBadRequest, "事件类型不能为空")
	}
	if event.ProcessInstanceID == "" {
		return NewServiceError(ErrCodeBadRequest, "流程实例ID不能为空")
	}
	return nil
}

// validateSignalEvent 验证信号事件参数
func (s *EventMessageService) validateSignalEvent(signal *biz.SignalEvent) error {
	if signal.SignalName == "" {
		return NewServiceError(ErrCodeBadRequest, "信号名称不能为空")
	}
	return nil
}

// validateMessageEvent 验证消息事件参数
func (s *EventMessageService) validateMessageEvent(message *biz.MessageEvent) error {
	if message.MessageName == "" {
		return NewServiceError(ErrCodeBadRequest, "消息名称不能为空")
	}
	return nil
}

// validateTimerEvent 验证定时器事件参数
func (s *EventMessageService) validateTimerEvent(timer *biz.TimerEvent) error {
	if timer.TimerID == "" {
		return NewServiceError(ErrCodeBadRequest, "定时器ID不能为空")
	}
	if timer.DueDate.IsZero() {
		return NewServiceError(ErrCodeBadRequest, "定时器到期时间不能为空")
	}
	return nil
}
