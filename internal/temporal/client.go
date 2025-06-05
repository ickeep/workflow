package temporal

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/workflow-engine/workflow-engine/pkg/config"
)

// Client Temporal客户端封装
type Client struct {
	client.Client
	config config.TemporalConfig
}

// NewClient 创建新的Temporal客户端
func NewClient(cfg config.TemporalConfig) (*Client, error) {
	// 创建客户端选项
	options := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	}

	// 创建Temporal客户端
	c, err := client.Dial(options)
	if err != nil {
		return nil, fmt.Errorf("创建Temporal客户端失败: %w", err)
	}

	return &Client{
		Client: c,
		config: cfg,
	}, nil
}

// StartWorker 启动Worker
func (c *Client) StartWorker(ctx context.Context) error {
	// 创建Worker
	w := worker.New(c.Client, c.config.TaskQueue, worker.Options{})

	// 注册工作流
	w.RegisterWorkflow(ProcessWorkflow)
	w.RegisterWorkflow(TaskWorkflow)
	w.RegisterWorkflow(ApprovalWorkflow)

	// 注册活动
	w.RegisterActivity(ValidateDataActivity)
	w.RegisterActivity(SendNotificationActivity)
	w.RegisterActivity(UpdateStatusActivity)
	w.RegisterActivity(ApprovalActivity)

	log.Printf("启动Temporal Worker, 任务队列: %s", c.config.TaskQueue)

	// 启动Worker
	return w.Run(worker.InterruptCh())
}

// ExecuteWorkflow 执行工作流
func (c *Client) ExecuteWorkflow(ctx context.Context, workflowID string, workflowType string, input interface{}) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: c.config.TaskQueue,
	}

	// 根据工作流类型选择对应的工作流函数
	var workflowFunc interface{}
	switch workflowType {
	case "process":
		workflowFunc = ProcessWorkflow
	case "task":
		workflowFunc = TaskWorkflow
	case "approval":
		workflowFunc = ApprovalWorkflow
	default:
		return nil, fmt.Errorf("不支持的工作流类型: %s", workflowType)
	}

	return c.Client.ExecuteWorkflow(ctx, options, workflowFunc, input)
}

// GetWorkflowResult 获取工作流执行结果
func (c *Client) GetWorkflowResult(ctx context.Context, workflowID string, runID string) (interface{}, error) {
	workflowRun := c.Client.GetWorkflow(ctx, workflowID, runID)

	var result interface{}
	err := workflowRun.Get(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("获取工作流结果失败: %w", err)
	}

	return result, nil
}

// TerminateWorkflow 终止工作流
func (c *Client) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string) error {
	return c.Client.TerminateWorkflow(ctx, workflowID, runID, reason)
}

// CancelWorkflow 取消工作流
func (c *Client) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	return c.Client.CancelWorkflow(ctx, workflowID, runID)
}

// SignalWorkflow 发送信号给工作流
func (c *Client) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	return c.Client.SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// QueryWorkflow 查询工作流状态
func (c *Client) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (interface{}, error) {
	response, err := c.Client.QueryWorkflow(ctx, workflowID, runID, queryType, args...)
	if err != nil {
		return nil, fmt.Errorf("查询工作流失败: %w", err)
	}

	var result interface{}
	err = response.Get(&result)
	if err != nil {
		return nil, fmt.Errorf("获取查询结果失败: %w", err)
	}

	return result, nil
}

// Close 关闭客户端
func (c *Client) Close() {
	c.Client.Close()
}
