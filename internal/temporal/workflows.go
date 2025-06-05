package temporal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ProcessWorkflowInput 流程工作流输入参数
type ProcessWorkflowInput struct {
	ProcessDefinitionID int64                  `json:"process_definition_id"`
	BusinessKey         string                 `json:"business_key"`
	Variables           map[string]interface{} `json:"variables"`
	Initiator           string                 `json:"initiator"`
}

// ProcessWorkflowResult 流程工作流结果
type ProcessWorkflowResult struct {
	ProcessInstanceID int64                  `json:"process_instance_id"`
	Status            string                 `json:"status"`
	Result            map[string]interface{} `json:"result"`
	EndTime           time.Time              `json:"end_time"`
}

// TaskWorkflowInput 任务工作流输入参数
type TaskWorkflowInput struct {
	TaskID    int64                  `json:"task_id"`
	Assignee  string                 `json:"assignee"`
	Variables map[string]interface{} `json:"variables"`
}

// TaskWorkflowResult 任务工作流结果
type TaskWorkflowResult struct {
	TaskID    int64                  `json:"task_id"`
	Status    string                 `json:"status"`
	Result    map[string]interface{} `json:"result"`
	Completed bool                   `json:"completed"`
}

// ApprovalWorkflowInput 审批工作流输入参数
type ApprovalWorkflowInput struct {
	RequestID  string                 `json:"request_id"`
	Requestor  string                 `json:"requestor"`
	Approvers  []string               `json:"approvers"`
	Content    map[string]interface{} `json:"content"`
	Deadline   time.Duration          `json:"deadline"`
	RequireAll bool                   `json:"require_all"`
}

// ApprovalWorkflowResult 审批工作流结果
type ApprovalWorkflowResult struct {
	RequestID string    `json:"request_id"`
	Status    string    `json:"status"`
	Approved  bool      `json:"approved"`
	Comments  string    `json:"comments"`
	EndTime   time.Time `json:"end_time"`
}

// ProcessWorkflow 流程工作流
func ProcessWorkflow(ctx workflow.Context, input ProcessWorkflowInput) (*ProcessWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("开始执行流程工作流", "process_definition_id", input.ProcessDefinitionID, "business_key", input.BusinessKey)

	// 设置工作流选项
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// 1. 验证数据
	var validateResult ValidateDataResult
	err := workflow.ExecuteActivity(ctx, ValidateDataActivity, ValidateDataInput{
		ProcessDefinitionID: input.ProcessDefinitionID,
		Variables:           input.Variables,
	}).Get(ctx, &validateResult)
	if err != nil {
		logger.Error("数据验证失败", "error", err)
		return &ProcessWorkflowResult{
			Status:  "failed",
			Result:  map[string]interface{}{"error": err.Error()},
			EndTime: workflow.Now(ctx),
		}, nil
	}

	if !validateResult.Valid {
		logger.Warn("数据验证未通过", "errors", validateResult.Errors)
		return &ProcessWorkflowResult{
			Status:  "failed",
			Result:  map[string]interface{}{"validation_errors": validateResult.Errors},
			EndTime: workflow.Now(ctx),
		}, nil
	}

	// 2. 更新状态为运行中
	err = workflow.ExecuteActivity(ctx, UpdateStatusActivity, UpdateStatusInput{
		ProcessInstanceID: validateResult.ProcessInstanceID,
		Status:            "running",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("更新状态失败", "error", err)
		return &ProcessWorkflowResult{
			Status:  "failed",
			Result:  map[string]interface{}{"error": err.Error()},
			EndTime: workflow.Now(ctx),
		}, nil
	}

	// 3. 发送开始通知
	err = workflow.ExecuteActivity(ctx, SendNotificationActivity, SendNotificationInput{
		Type:      "process_started",
		Recipient: input.Initiator,
		Data: map[string]interface{}{
			"process_instance_id": validateResult.ProcessInstanceID,
			"business_key":        input.BusinessKey,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("发送通知失败", "error", err)
		// 通知失败不影响流程继续
	}

	// 4. 等待完成信号或超时
	selector := workflow.NewSelector(ctx)
	var result ProcessWorkflowResult

	// 设置超时
	timerFuture := workflow.NewTimer(ctx, time.Hour*24) // 24小时超时
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		result = ProcessWorkflowResult{
			ProcessInstanceID: validateResult.ProcessInstanceID,
			Status:            "timeout",
			Result:            map[string]interface{}{"message": "工作流执行超时"},
			EndTime:           workflow.Now(ctx),
		}
	})

	// 等待完成信号
	completeSignal := workflow.GetSignalChannel(ctx, "complete")
	selector.AddReceive(completeSignal, func(c workflow.ReceiveChannel, more bool) {
		var signalData map[string]interface{}
		c.Receive(ctx, &signalData)

		result = ProcessWorkflowResult{
			ProcessInstanceID: validateResult.ProcessInstanceID,
			Status:            "completed",
			Result:            signalData,
			EndTime:           workflow.Now(ctx),
		}
	})

	// 等待其中一个条件满足
	selector.Select(ctx)

	// 5. 更新最终状态
	err = workflow.ExecuteActivity(ctx, UpdateStatusActivity, UpdateStatusInput{
		ProcessInstanceID: result.ProcessInstanceID,
		Status:            result.Status,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("更新最终状态失败", "error", err)
	}

	// 6. 发送完成通知
	err = workflow.ExecuteActivity(ctx, SendNotificationActivity, SendNotificationInput{
		Type:      "process_completed",
		Recipient: input.Initiator,
		Data: map[string]interface{}{
			"process_instance_id": result.ProcessInstanceID,
			"status":              result.Status,
			"result":              result.Result,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("发送完成通知失败", "error", err)
	}

	logger.Info("流程工作流执行完成", "status", result.Status)
	return &result, nil
}

// TaskWorkflow 任务工作流
func TaskWorkflow(ctx workflow.Context, input TaskWorkflowInput) (*TaskWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("开始执行任务工作流", "task_id", input.TaskID, "assignee", input.Assignee)

	// 设置活动选项
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// 等待任务完成信号
	completeSignal := workflow.GetSignalChannel(ctx, "task_complete")
	selector := workflow.NewSelector(ctx)

	var result TaskWorkflowResult

	// 设置超时
	timerFuture := workflow.NewTimer(ctx, time.Hour*72) // 72小时超时
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		result = TaskWorkflowResult{
			TaskID:    input.TaskID,
			Status:    "timeout",
			Result:    map[string]interface{}{"message": "任务执行超时"},
			Completed: false,
		}
	})

	// 等待完成信号
	selector.AddReceive(completeSignal, func(c workflow.ReceiveChannel, more bool) {
		var signalData map[string]interface{}
		c.Receive(ctx, &signalData)

		result = TaskWorkflowResult{
			TaskID:    input.TaskID,
			Status:    "completed",
			Result:    signalData,
			Completed: true,
		}
	})

	// 等待其中一个条件满足
	selector.Select(ctx)

	logger.Info("任务工作流执行完成", "task_id", input.TaskID, "status", result.Status)
	return &result, nil
}

// ApprovalWorkflow 审批工作流
func ApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) (*ApprovalWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("开始执行审批工作流", "request_id", input.RequestID, "approvers", input.Approvers)

	// 设置活动选项
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// 发送审批请求
	for _, approver := range input.Approvers {
		err := workflow.ExecuteActivity(ctx, SendNotificationActivity, SendNotificationInput{
			Type:      "approval_request",
			Recipient: approver,
			Data: map[string]interface{}{
				"request_id": input.RequestID,
				"requestor":  input.Requestor,
				"content":    input.Content,
				"deadline":   workflow.Now(ctx).Add(input.Deadline),
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("发送审批通知失败", "approver", approver, "error", err)
		}
	}

	// 收集审批结果
	approvalSignal := workflow.GetSignalChannel(ctx, "approval")
	selector := workflow.NewSelector(ctx)

	approvals := make(map[string]bool)
	var comments []string

	// 设置截止时间
	deadline := workflow.NewTimer(ctx, input.Deadline)
	selector.AddFuture(deadline, func(f workflow.Future) {
		// 超时处理
	})

	// 接收审批信号
	selector.AddReceive(approvalSignal, func(c workflow.ReceiveChannel, more bool) {
		var approval ApprovalData
		c.Receive(ctx, &approval)

		approvals[approval.Approver] = approval.Approved
		if approval.Comments != "" {
			comments = append(comments, approval.Comments)
		}
	})

	// 等待所有审批或超时
	for {
		selector.Select(ctx)

		// 检查是否所有人都审批了
		if len(approvals) >= len(input.Approvers) {
			break
		}

		// 检查是否超时
		if deadline.IsReady() {
			break
		}

		// 如果不需要所有人同意，有一个人拒绝就可以结束
		if !input.RequireAll {
			for _, approved := range approvals {
				if !approved {
					break
				}
			}
		}
	}

	// 计算最终结果
	var finalApproved bool
	if input.RequireAll {
		// 需要所有人同意
		finalApproved = len(approvals) == len(input.Approvers)
		for _, approved := range approvals {
			if !approved {
				finalApproved = false
				break
			}
		}
	} else {
		// 只需要一个人同意
		for _, approved := range approvals {
			if approved {
				finalApproved = true
				break
			}
		}
	}

	var status string
	if len(approvals) < len(input.Approvers) && deadline.IsReady() {
		status = "timeout"
	} else if finalApproved {
		status = "approved"
	} else {
		status = "rejected"
	}

	result := &ApprovalWorkflowResult{
		RequestID: input.RequestID,
		Status:    status,
		Approved:  finalApproved,
		Comments:  joinComments(comments),
		EndTime:   workflow.Now(ctx),
	}

	logger.Info("审批工作流执行完成", "request_id", input.RequestID, "status", status, "approved", finalApproved)
	return result, nil
}

// ApprovalData 审批信号数据
type ApprovalData struct {
	Approver string `json:"approver"`
	Approved bool   `json:"approved"`
	Comments string `json:"comments"`
}

// joinComments 合并评论
func joinComments(comments []string) string {
	if len(comments) == 0 {
		return ""
	}

	result := comments[0]
	for i := 1; i < len(comments); i++ {
		result += "; " + comments[i]
	}
	return result
}
