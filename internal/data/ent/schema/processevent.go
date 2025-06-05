package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProcessEvent 流程事件表 - 存储流程执行过程中的事件信息
type ProcessEvent struct {
	ent.Schema
}

// Fields 定义 ProcessEvent 的字段
func (ProcessEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("事件ID"),
		field.String("event_type").
			NotEmpty().
			Comment("事件类型: PROCESS_STARTED, PROCESS_COMPLETED, TASK_CREATED, TASK_COMPLETED, etc.").
			MaxLen(100),
		field.String("event_name").
			Optional().
			Comment("事件名称").
			MaxLen(255),
		field.String("execution_id").
			Optional().
			Comment("执行ID").
			MaxLen(255),
		field.Int64("process_instance_id").
			Optional().
			Comment("流程实例ID"),
		field.Int64("process_definition_id").
			Optional().
			Comment("流程定义ID"),
		field.String("process_definition_key").
			Optional().
			Comment("流程定义标识").
			MaxLen(255),
		field.Int64("task_id").
			Optional().
			Comment("任务ID"),
		field.String("activity_id").
			Optional().
			Comment("活动ID").
			MaxLen(255),
		field.String("activity_name").
			Optional().
			Comment("活动名称").
			MaxLen(255),
		field.String("activity_type").
			Optional().
			Comment("活动类型").
			MaxLen(100),
		field.String("user_id").
			Optional().
			Comment("用户ID").
			MaxLen(255),
		field.Time("timestamp").
			Default(time.Now).
			Comment("事件时间戳"),
		field.JSON("event_data", map[string]interface{}{}).
			Optional().
			Comment("事件数据(JSON)"),
		field.String("correlation_id").
			Optional().
			Comment("关联ID").
			MaxLen(255),
		field.String("message_name").
			Optional().
			Comment("消息名称").
			MaxLen(255),
		field.String("signal_name").
			Optional().
			Comment("信号名称").
			MaxLen(255),
		field.String("job_id").
			Optional().
			Comment("作业ID").
			MaxLen(255),
		field.String("job_type").
			Optional().
			Comment("作业类型").
			MaxLen(100),
		field.String("job_handler_type").
			Optional().
			Comment("作业处理器类型").
			MaxLen(100),
		field.String("tenant_id").
			Default("default").
			Comment("租户ID").
			MaxLen(100),
		field.String("deployment_id").
			Optional().
			Comment("部署ID").
			MaxLen(255),
		field.String("sequence_counter").
			Optional().
			Comment("序列计数器").
			MaxLen(100),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
	}
}

// Edges 定义 ProcessEvent 的边（关系）
func (ProcessEvent) Edges() []ent.Edge {
	return []ent.Edge{
		// 关系将在所有 schema 定义完成后添加
	}
}

// Indexes 定义 ProcessEvent 的索引
func (ProcessEvent) Indexes() []ent.Index {
	return []ent.Index{
		// 事件类型索引
		index.Fields("event_type"),
		// 时间戳索引
		index.Fields("timestamp"),
		// 流程实例索引
		index.Fields("process_instance_id"),
		// 流程定义索引
		index.Fields("process_definition_id"),
		index.Fields("process_definition_key"),
		// 任务索引
		index.Fields("task_id"),
		// 活动索引
		index.Fields("activity_id"),
		// 用户索引
		index.Fields("user_id"),
		// 执行ID索引
		index.Fields("execution_id"),
		// 租户索引
		index.Fields("tenant_id"),
		// 关联ID索引
		index.Fields("correlation_id"),
		// 消息名称索引
		index.Fields("message_name"),
		// 信号名称索引
		index.Fields("signal_name"),
		// 作业索引
		index.Fields("job_id"),
		index.Fields("job_type"),
		// 复合索引：流程实例+事件类型
		index.Fields("process_instance_id", "event_type"),
		// 复合索引：租户+事件类型
		index.Fields("tenant_id", "event_type"),
		// 复合索引：时间戳范围查询
		index.Fields("timestamp", "event_type"),
		// 复合索引：用户+时间戳
		index.Fields("user_id", "timestamp"),
	}
}
