package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// HistoricProcessInstance 历史流程实例表 - 存储已完成的流程实例信息
type HistoricProcessInstance struct {
	ent.Schema
}

// Fields 定义 HistoricProcessInstance 的字段
func (HistoricProcessInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("历史流程实例ID"),
		field.String("process_instance_id").
			NotEmpty().
			Comment("流程实例ID").
			MaxLen(255),
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
			Comment("启动时间"),
		field.Time("end_time").
			Optional().
			Nillable().
			Comment("结束时间"),
		field.Int64("duration").
			Optional().
			Comment("持续时间(毫秒)"),
		field.String("start_activity_id").
			Optional().
			Comment("启动活动ID").
			MaxLen(255),
		field.String("end_activity_id").
			Optional().
			Comment("结束活动ID").
			MaxLen(255),
		field.String("super_process_instance_id").
			Optional().
			Comment("父流程实例ID").
			MaxLen(255),
		field.String("root_process_instance_id").
			Optional().
			Comment("根流程实例ID").
			MaxLen(255),
		field.String("super_case_instance_id").
			Optional().
			Comment("父案例实例ID").
			MaxLen(255),
		field.String("case_instance_id").
			Optional().
			Comment("案例实例ID").
			MaxLen(255),
		field.String("delete_reason").
			Optional().
			Comment("删除原因").
			MaxLen(500),
		field.String("tenant_id").
			Default("default").
			Comment("租户ID").
			MaxLen(100),
		field.String("state").
			Optional().
			Comment("流程状态: ACTIVE, SUSPENDED, COMPLETED, EXTERNALLY_TERMINATED, INTERNALLY_TERMINATED").
			MaxLen(50),
		field.String("removal_time").
			Optional().
			Comment("移除时间").
			MaxLen(255),
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

// Edges 定义 HistoricProcessInstance 的边（关系）
func (HistoricProcessInstance) Edges() []ent.Edge {
	return []ent.Edge{
		// 关系将在所有 schema 定义完成后添加
	}
}

// Indexes 定义 HistoricProcessInstance 的索引
func (HistoricProcessInstance) Indexes() []ent.Index {
	return []ent.Index{
		// 流程实例ID唯一索引
		index.Fields("process_instance_id").
			Unique(),
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
		// 持续时间索引
		index.Fields("duration"),
		// 租户索引
		index.Fields("tenant_id"),
		// 状态索引
		index.Fields("state"),
		// 父流程实例索引
		index.Fields("super_process_instance_id"),
		// 根流程实例索引
		index.Fields("root_process_instance_id"),
		// 复合索引：租户+流程定义
		index.Fields("tenant_id", "process_definition_key"),
		// 复合索引：启动时间范围查询
		index.Fields("start_time", "end_time"),
	}
}
