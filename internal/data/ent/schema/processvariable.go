package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProcessVariable 流程变量表 - 存储流程实例和任务的变量数据
type ProcessVariable struct {
	ent.Schema
}

// Fields 定义 ProcessVariable 的字段
func (ProcessVariable) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("变量ID"),
		field.String("name").
			NotEmpty().
			Comment("变量名称").
			MaxLen(255),
		field.String("type").
			NotEmpty().
			Comment("变量类型: string, integer, boolean, date, json, bytes").
			MaxLen(50),
		field.Text("text_value").
			Optional().
			Comment("文本值"),
		field.Text("text_value2").
			Optional().
			Comment("文本值2(用于长文本)"),
		field.Int64("long_value").
			Optional().
			Comment("长整型值"),
		field.Float("double_value").
			Optional().
			Comment("浮点型值"),
		field.Bytes("byte_array_value").
			Optional().
			Comment("字节数组值"),
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
		field.String("case_execution_id").
			Optional().
			Comment("案例执行ID").
			MaxLen(255),
		field.String("case_instance_id").
			Optional().
			Comment("案例实例ID").
			MaxLen(255),
		field.Int64("task_id").
			Optional().
			Comment("任务ID"),
		field.String("activity_instance_id").
			Optional().
			Comment("活动实例ID").
			MaxLen(255),
		field.String("tenant_id").
			Default("default").
			Comment("租户ID").
			MaxLen(100),
		field.Int32("sequence_counter").
			Default(1).
			Comment("序列计数器"),
		field.Bool("concurrent_local").
			Default(false).
			Comment("是否并发本地变量"),
		field.String("scope_id").
			Optional().
			Comment("作用域ID").
			MaxLen(255),
		field.String("scope_type").
			Optional().
			Comment("作用域类型").
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

// Edges 定义 ProcessVariable 的边（关系）
func (ProcessVariable) Edges() []ent.Edge {
	return []ent.Edge{
		// 关系将在所有 schema 定义完成后添加
	}
}

// Indexes 定义 ProcessVariable 的索引
func (ProcessVariable) Indexes() []ent.Index {
	return []ent.Index{
		// 变量名索引
		index.Fields("name"),
		// 变量类型索引
		index.Fields("type"),
		// 流程实例索引
		index.Fields("process_instance_id"),
		// 流程定义索引
		index.Fields("process_definition_id"),
		// 任务索引
		index.Fields("task_id"),
		// 执行ID索引
		index.Fields("execution_id"),
		// 租户索引
		index.Fields("tenant_id"),
		// 作用域索引
		index.Fields("scope_id", "scope_type"),
		// 活动实例索引
		index.Fields("activity_instance_id"),
		// 复合索引：流程实例+变量名
		index.Fields("process_instance_id", "name"),
		// 复合索引：任务+变量名
		index.Fields("task_id", "name"),
		// 复合索引：租户+流程实例
		index.Fields("tenant_id", "process_instance_id"),
		// 复合索引：执行ID+变量名
		index.Fields("execution_id", "name"),
	}
}
