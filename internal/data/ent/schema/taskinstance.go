package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TaskInstance 任务实例表 - 存储用户任务和服务任务信息
type TaskInstance struct {
	ent.Schema
}

// Fields 定义 TaskInstance 的字段
func (TaskInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("任务实例ID"),
		field.String("name").
			Optional().
			Comment("任务名称").
			MaxLen(255),
		field.Text("description").
			Optional().
			Comment("任务描述"),
		field.String("task_definition_key").
			NotEmpty().
			Comment("任务定义标识").
			MaxLen(255),
		field.String("assignee").
			Optional().
			Comment("任务分配人").
			MaxLen(255),
		field.String("owner").
			Optional().
			Comment("任务拥有者").
			MaxLen(255),
		field.String("delegation").
			Optional().
			Comment("委派状态: PENDING, RESOLVED").
			MaxLen(50),
		field.Int32("priority").
			Default(50).
			Comment("任务优先级"),
		field.Time("create_time").
			Default(time.Now).
			Comment("创建时间"),
		field.Time("due_date").
			Optional().
			Nillable().
			Comment("到期时间"),
		field.Time("follow_up_date").
			Optional().
			Nillable().
			Comment("跟进时间"),
		field.String("form_key").
			Optional().
			Comment("表单标识").
			MaxLen(255),
		field.String("category").
			Optional().
			Comment("任务分类").
			MaxLen(255),
		field.String("parent_task_id").
			Optional().
			Comment("父任务ID").
			MaxLen(255),
		field.String("execution_id").
			Optional().
			Comment("执行ID").
			MaxLen(255),
		field.Int64("process_instance_id").
			Comment("流程实例ID"),
		field.Int64("process_definition_id").
			Comment("流程定义ID"),
		field.String("process_definition_key").
			NotEmpty().
			Comment("流程定义标识").
			MaxLen(255),
		field.String("case_execution_id").
			Optional().
			Comment("案例执行ID").
			MaxLen(255),
		field.String("case_instance_id").
			Optional().
			Comment("案例实例ID").
			MaxLen(255),
		field.String("case_definition_id").
			Optional().
			Comment("案例定义ID").
			MaxLen(255),
		field.Bool("suspended").
			Default(false).
			Comment("是否挂起"),
		field.String("tenant_id").
			Default("default").
			Comment("租户ID").
			MaxLen(100),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

// Edges 定义 TaskInstance 的边（关系）
func (TaskInstance) Edges() []ent.Edge {
	return []ent.Edge{
		// 关系将在所有 schema 定义完成后添加
	}
}

// Indexes 定义 TaskInstance 的索引
func (TaskInstance) Indexes() []ent.Index {
	return []ent.Index{
		// 任务分配人索引
		index.Fields("assignee"),
		// 任务拥有者索引
		index.Fields("owner"),
		// 流程实例索引
		index.Fields("process_instance_id"),
		// 流程定义索引
		index.Fields("process_definition_id"),
		index.Fields("process_definition_key"),
		// 任务定义索引
		index.Fields("task_definition_key"),
		// 创建时间索引
		index.Fields("create_time"),
		// 到期时间索引
		index.Fields("due_date"),
		// 优先级索引
		index.Fields("priority"),
		// 挂起状态索引
		index.Fields("suspended"),
		// 租户索引
		index.Fields("tenant_id"),
		// 委派状态索引
		index.Fields("delegation"),
		// 父任务索引
		index.Fields("parent_task_id"),
		// 执行ID索引
		index.Fields("execution_id"),
		// 分类索引
		index.Fields("category"),
		// 复合索引：分配人+流程实例
		index.Fields("assignee", "process_instance_id"),
		// 复合索引：租户+分配人
		index.Fields("tenant_id", "assignee"),
	}
}
