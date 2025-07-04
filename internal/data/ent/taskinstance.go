// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/workflow-engine/workflow-engine/internal/data/ent/taskinstance"
)

// TaskInstance is the model entity for the TaskInstance schema.
type TaskInstance struct {
	config `json:"-"`
	// ID of the ent.
	// 任务实例ID
	ID int64 `json:"id,omitempty"`
	// 任务名称
	Name string `json:"name,omitempty"`
	// 任务描述
	Description string `json:"description,omitempty"`
	// 任务定义标识
	TaskDefinitionKey string `json:"task_definition_key,omitempty"`
	// 任务分配人
	Assignee string `json:"assignee,omitempty"`
	// 任务拥有者
	Owner string `json:"owner,omitempty"`
	// 委派状态: PENDING, RESOLVED
	Delegation string `json:"delegation,omitempty"`
	// 任务优先级
	Priority int32 `json:"priority,omitempty"`
	// 创建时间
	CreateTime time.Time `json:"create_time,omitempty"`
	// 到期时间
	DueDate *time.Time `json:"due_date,omitempty"`
	// 跟进时间
	FollowUpDate *time.Time `json:"follow_up_date,omitempty"`
	// 表单标识
	FormKey string `json:"form_key,omitempty"`
	// 任务分类
	Category string `json:"category,omitempty"`
	// 父任务ID
	ParentTaskID string `json:"parent_task_id,omitempty"`
	// 执行ID
	ExecutionID string `json:"execution_id,omitempty"`
	// 流程实例ID
	ProcessInstanceID int64 `json:"process_instance_id,omitempty"`
	// 流程定义ID
	ProcessDefinitionID int64 `json:"process_definition_id,omitempty"`
	// 流程定义标识
	ProcessDefinitionKey string `json:"process_definition_key,omitempty"`
	// 案例执行ID
	CaseExecutionID string `json:"case_execution_id,omitempty"`
	// 案例实例ID
	CaseInstanceID string `json:"case_instance_id,omitempty"`
	// 案例定义ID
	CaseDefinitionID string `json:"case_definition_id,omitempty"`
	// 是否挂起
	Suspended bool `json:"suspended,omitempty"`
	// 租户ID
	TenantID string `json:"tenant_id,omitempty"`
	// 创建时间
	CreatedAt time.Time `json:"created_at,omitempty"`
	// 更新时间
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*TaskInstance) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case taskinstance.FieldSuspended:
			values[i] = new(sql.NullBool)
		case taskinstance.FieldID, taskinstance.FieldPriority, taskinstance.FieldProcessInstanceID, taskinstance.FieldProcessDefinitionID:
			values[i] = new(sql.NullInt64)
		case taskinstance.FieldName, taskinstance.FieldDescription, taskinstance.FieldTaskDefinitionKey, taskinstance.FieldAssignee, taskinstance.FieldOwner, taskinstance.FieldDelegation, taskinstance.FieldFormKey, taskinstance.FieldCategory, taskinstance.FieldParentTaskID, taskinstance.FieldExecutionID, taskinstance.FieldProcessDefinitionKey, taskinstance.FieldCaseExecutionID, taskinstance.FieldCaseInstanceID, taskinstance.FieldCaseDefinitionID, taskinstance.FieldTenantID:
			values[i] = new(sql.NullString)
		case taskinstance.FieldCreateTime, taskinstance.FieldDueDate, taskinstance.FieldFollowUpDate, taskinstance.FieldCreatedAt, taskinstance.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the TaskInstance fields.
func (ti *TaskInstance) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case taskinstance.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			ti.ID = int64(value.Int64)
		case taskinstance.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				ti.Name = value.String
			}
		case taskinstance.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				ti.Description = value.String
			}
		case taskinstance.FieldTaskDefinitionKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field task_definition_key", values[i])
			} else if value.Valid {
				ti.TaskDefinitionKey = value.String
			}
		case taskinstance.FieldAssignee:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field assignee", values[i])
			} else if value.Valid {
				ti.Assignee = value.String
			}
		case taskinstance.FieldOwner:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner", values[i])
			} else if value.Valid {
				ti.Owner = value.String
			}
		case taskinstance.FieldDelegation:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field delegation", values[i])
			} else if value.Valid {
				ti.Delegation = value.String
			}
		case taskinstance.FieldPriority:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field priority", values[i])
			} else if value.Valid {
				ti.Priority = int32(value.Int64)
			}
		case taskinstance.FieldCreateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[i])
			} else if value.Valid {
				ti.CreateTime = value.Time
			}
		case taskinstance.FieldDueDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field due_date", values[i])
			} else if value.Valid {
				ti.DueDate = new(time.Time)
				*ti.DueDate = value.Time
			}
		case taskinstance.FieldFollowUpDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field follow_up_date", values[i])
			} else if value.Valid {
				ti.FollowUpDate = new(time.Time)
				*ti.FollowUpDate = value.Time
			}
		case taskinstance.FieldFormKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field form_key", values[i])
			} else if value.Valid {
				ti.FormKey = value.String
			}
		case taskinstance.FieldCategory:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field category", values[i])
			} else if value.Valid {
				ti.Category = value.String
			}
		case taskinstance.FieldParentTaskID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field parent_task_id", values[i])
			} else if value.Valid {
				ti.ParentTaskID = value.String
			}
		case taskinstance.FieldExecutionID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field execution_id", values[i])
			} else if value.Valid {
				ti.ExecutionID = value.String
			}
		case taskinstance.FieldProcessInstanceID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field process_instance_id", values[i])
			} else if value.Valid {
				ti.ProcessInstanceID = value.Int64
			}
		case taskinstance.FieldProcessDefinitionID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field process_definition_id", values[i])
			} else if value.Valid {
				ti.ProcessDefinitionID = value.Int64
			}
		case taskinstance.FieldProcessDefinitionKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field process_definition_key", values[i])
			} else if value.Valid {
				ti.ProcessDefinitionKey = value.String
			}
		case taskinstance.FieldCaseExecutionID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field case_execution_id", values[i])
			} else if value.Valid {
				ti.CaseExecutionID = value.String
			}
		case taskinstance.FieldCaseInstanceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field case_instance_id", values[i])
			} else if value.Valid {
				ti.CaseInstanceID = value.String
			}
		case taskinstance.FieldCaseDefinitionID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field case_definition_id", values[i])
			} else if value.Valid {
				ti.CaseDefinitionID = value.String
			}
		case taskinstance.FieldSuspended:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field suspended", values[i])
			} else if value.Valid {
				ti.Suspended = value.Bool
			}
		case taskinstance.FieldTenantID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field tenant_id", values[i])
			} else if value.Valid {
				ti.TenantID = value.String
			}
		case taskinstance.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ti.CreatedAt = value.Time
			}
		case taskinstance.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ti.UpdatedAt = value.Time
			}
		default:
			ti.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the TaskInstance.
// This includes values selected through modifiers, order, etc.
func (ti *TaskInstance) Value(name string) (ent.Value, error) {
	return ti.selectValues.Get(name)
}

// Update returns a builder for updating this TaskInstance.
// Note that you need to call TaskInstance.Unwrap() before calling this method if this TaskInstance
// was returned from a transaction, and the transaction was committed or rolled back.
func (ti *TaskInstance) Update() *TaskInstanceUpdateOne {
	return NewTaskInstanceClient(ti.config).UpdateOne(ti)
}

// Unwrap unwraps the TaskInstance entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ti *TaskInstance) Unwrap() *TaskInstance {
	_tx, ok := ti.config.driver.(*txDriver)
	if !ok {
		panic("ent: TaskInstance is not a transactional entity")
	}
	ti.config.driver = _tx.drv
	return ti
}

// String implements the fmt.Stringer.
func (ti *TaskInstance) String() string {
	var builder strings.Builder
	builder.WriteString("TaskInstance(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ti.ID))
	builder.WriteString("name=")
	builder.WriteString(ti.Name)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(ti.Description)
	builder.WriteString(", ")
	builder.WriteString("task_definition_key=")
	builder.WriteString(ti.TaskDefinitionKey)
	builder.WriteString(", ")
	builder.WriteString("assignee=")
	builder.WriteString(ti.Assignee)
	builder.WriteString(", ")
	builder.WriteString("owner=")
	builder.WriteString(ti.Owner)
	builder.WriteString(", ")
	builder.WriteString("delegation=")
	builder.WriteString(ti.Delegation)
	builder.WriteString(", ")
	builder.WriteString("priority=")
	builder.WriteString(fmt.Sprintf("%v", ti.Priority))
	builder.WriteString(", ")
	builder.WriteString("create_time=")
	builder.WriteString(ti.CreateTime.Format(time.ANSIC))
	builder.WriteString(", ")
	if v := ti.DueDate; v != nil {
		builder.WriteString("due_date=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	if v := ti.FollowUpDate; v != nil {
		builder.WriteString("follow_up_date=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("form_key=")
	builder.WriteString(ti.FormKey)
	builder.WriteString(", ")
	builder.WriteString("category=")
	builder.WriteString(ti.Category)
	builder.WriteString(", ")
	builder.WriteString("parent_task_id=")
	builder.WriteString(ti.ParentTaskID)
	builder.WriteString(", ")
	builder.WriteString("execution_id=")
	builder.WriteString(ti.ExecutionID)
	builder.WriteString(", ")
	builder.WriteString("process_instance_id=")
	builder.WriteString(fmt.Sprintf("%v", ti.ProcessInstanceID))
	builder.WriteString(", ")
	builder.WriteString("process_definition_id=")
	builder.WriteString(fmt.Sprintf("%v", ti.ProcessDefinitionID))
	builder.WriteString(", ")
	builder.WriteString("process_definition_key=")
	builder.WriteString(ti.ProcessDefinitionKey)
	builder.WriteString(", ")
	builder.WriteString("case_execution_id=")
	builder.WriteString(ti.CaseExecutionID)
	builder.WriteString(", ")
	builder.WriteString("case_instance_id=")
	builder.WriteString(ti.CaseInstanceID)
	builder.WriteString(", ")
	builder.WriteString("case_definition_id=")
	builder.WriteString(ti.CaseDefinitionID)
	builder.WriteString(", ")
	builder.WriteString("suspended=")
	builder.WriteString(fmt.Sprintf("%v", ti.Suspended))
	builder.WriteString(", ")
	builder.WriteString("tenant_id=")
	builder.WriteString(ti.TenantID)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(ti.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ti.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// TaskInstances is a parsable slice of TaskInstance.
type TaskInstances []*TaskInstance
