package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProcessInstance 流程实例表 - 存储运行中的流程实例信息
type ProcessInstance struct {
	ent.Schema
}

// Fields 定义 ProcessInstance 的字段
func (ProcessInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("流程实例ID"),
		field.String("business_key").
			Optional().
			Comment("业务标识").
			MaxLen(255),
		field.Int64("process_definition_id").
			Comment("流程定义ID"),
		field.String("process_definition_key").
			NotEmpty().
			Comment("流程定义标识").
			MaxLen(255),
		field.String("process_definition_name").
			Optional().
			Comment("流程定义名称").
			MaxLen(255),
		field.Int32("process_definition_version").
			Comment("流程定义版本"),
		field.String("deployment_id").
			Optional().
			Comment("部署ID").
			MaxLen(255),
		field.String("start_user_id").
			Optional().
			Comment("启动用户ID").
			MaxLen(255),
		field.Time("start_time").
			Default(time.Now).
			Comment("启动时间"),
		field.Time("end_time").
			Optional().
			Nillable().
			Comment("结束时间"),
		field.Int64("duration").
			Optional().
			Comment("持续时间(毫秒)"),
		field.String("delete_reason").
			Optional().
			Comment("删除原因").
			MaxLen(500),
		field.String("super_process_instance_id").
			Optional().
			Comment("父流程实例ID").
			MaxLen(255),
		field.String("root_process_instance_id").
			Optional().
			Comment("根流程实例ID").
			MaxLen(255),
		field.Bool("suspended").
			Default(false).
			Comment("是否挂起"),
		field.String("tenant_id").
			Default("default").
			Comment("租户ID").
			MaxLen(100),
		field.String("name").
			Optional().
			Comment("流程实例名称").
			MaxLen(255),
		field.Text("description").
			Optional().
			Comment("流程实例描述"),
		field.String("callback_id").
			Optional().
			Comment("回调ID").
			MaxLen(255),
		field.String("callback_type").
			Optional().
			Comment("回调类型").
			MaxLen(100),
		field.String("reference_id").
			Optional().
			Comment("引用ID").
			MaxLen(255),
		field.String("reference_type").
			Optional().
			Comment("引用类型").
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

// Edges 定义 ProcessInstance 的边（关系）
func (ProcessInstance) Edges() []ent.Edge {
	return []ent.Edge{
		// 关系将在所有 schema 定义完成后添加
	}
}

// Indexes 定义 ProcessInstance 的索引
func (ProcessInstance) Indexes() []ent.Index {
	return []ent.Index{
		// 业务标识索引
		index.Fields("business_key"),
		// 流程定义索引
		index.Fields("process_definition_id"),
		index.Fields("process_definition_key"),
		// 启动用户索引
		index.Fields("start_user_id"),
		// 启动时间索引
		index.Fields("start_time"),
		// 结束时间索引
		index.Fields("end_time"),
		// 挂起状态索引
		index.Fields("suspended"),
		// 租户索引
		index.Fields("tenant_id"),
		// 父流程实例索引
		index.Fields("super_process_instance_id"),
		// 根流程实例索引
		index.Fields("root_process_instance_id"),
		// 回调索引
		index.Fields("callback_id"),
		// 引用索引
		index.Fields("reference_id", "reference_type"),
	}
}
