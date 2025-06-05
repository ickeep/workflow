package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProcessDefinition 流程定义表 - 存储流程模板和版本信息
type ProcessDefinition struct {
	ent.Schema
}

// Fields 定义 ProcessDefinition 的字段
func (ProcessDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("流程定义ID"),
		field.String("key").
			NotEmpty().
			Comment("流程唯一标识").
			MaxLen(255),
		field.String("name").
			NotEmpty().
			Comment("流程名称").
			MaxLen(255),
		field.String("category").
			Optional().
			Comment("流程分类").
			MaxLen(100),
		field.Int32("version").
			Default(1).
			Comment("版本号"),
		field.Text("description").
			Optional().
			Comment("流程描述"),
		field.Time("deploy_time").
			Default(time.Now).
			Comment("部署时间"),
		field.Text("resource").
			Optional().
			Comment("流程文件资源"),
		field.JSON("diagram_data", map[string]interface{}{}).
			Optional().
			Comment("流程图数据(JSON)"),
		field.Bool("has_start_form").
			Default(false).
			Comment("是否有启动表单"),
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

// Edges 定义 ProcessDefinition 的边（关系）
func (ProcessDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		// 关系将在所有 schema 定义完成后添加
	}
}

// Indexes 定义 ProcessDefinition 的索引
func (ProcessDefinition) Indexes() []ent.Index {
	return []ent.Index{
		// 流程标识唯一索引
		index.Fields("key", "version").
			Unique(),
		// 租户索引
		index.Fields("tenant_id"),
		// 分类索引
		index.Fields("category"),
		// 部署时间索引
		index.Fields("deploy_time"),
		// 挂起状态索引
		index.Fields("suspended"),
	}
}
