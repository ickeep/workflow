# Workflow Engine 开发规划文档

## 1. 项目概述

### 1.1 项目简介
基于 Go 语言开发的分布式工作流引擎，支持通过 JSON 配置定义复杂的业务工作流，提供完整的工作流生命周期管理。

### 1.2 技术栈
- **后端框架**: Kratos v2 (https://github.com/go-kratos/kratos)
- **ORM框架**: Ent (https://entgo.io/zh/docs/getting-started/)
- **工作流引擎**: Temporal (https://docs.temporal.io/)
- **数据库**: PostgreSQL
- **配置格式**: JSON/YAML
- **API协议**: HTTP/gRPC (Protocol Buffers)
- **参考实现**: https://workflow-engine-book.shuwoom.com/

### 1.3 核心功能
- ✅ 工作流模板管理（CRUD）
- ✅ 工作流执行引擎
- ✅ 实时状态监控
- ✅ 工作流控制操作
- ✅ 执行历史记录
- ✅ RESTful API 接口

## 2. 架构设计

### 2.1 整体架构设计
基于[流程引擎原理与实践](https://workflow-engine-book.shuwoom.com/)的设计理念，采用分层架构和事件驱动模式：

```
┌─────────────────────────────────────────────────────────────────┐
│                         应用层 (Application Layer)              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Web Portal │  │  Mobile App │  │ External API│             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                         接口层 (Interface Layer)                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ REST API    │  │  GraphQL    │  │    gRPC     │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                         服务层 (Service Layer)                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ 流程管理服务 │  │ 流程执行服务 │  │ 流程监控服务 │             │
│  │(Process Mgmt)│  │(Execution)  │  │(Monitoring) │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                      流程引擎核心 (Workflow Engine Core)          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ 流程解析器   │  │ 执行引擎     │  │ 事件总线     │             │
│  │(Parser)     │  │(Executor)   │  │(Event Bus)  │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ 任务调度器   │  │ 状态管理器   │  │ 规则引擎     │             │
│  │(Scheduler)  │  │(State Mgr)  │  │(Rule Engine)│             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                        数据层 (Data Layer)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ 流程定义表   │  │ 流程实例表   │  │ 任务实例表   │             │
│  │(Process Def)│  │(Process Ins)│  │(Task Ins)   │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ 执行历史表   │  │ 变量存储表   │  │ 事件日志表   │             │
│  │(History)    │  │(Variables)  │  │(Event Log)  │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                                  │
┌─────────────────────────────────▼─────────────────────────────────┐
│                      基础设施层 (Infrastructure)                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │ PostgreSQL  │  │   Redis     │  │  Temporal   │             │
│  │(主数据库)    │  │  (缓存)      │  │(工作流引擎)  │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 模块设计
基于[流程引擎原理与实践](https://workflow-engine-book.shuwoom.com/)第4章核心组件设计：

```
workflow-engine/
├── cmd/server/                           # 应用启动入口
├── internal/
│   ├── biz/                             # 业务逻辑层
│   │   ├── definition/                  # 流程定义业务逻辑
│   │   ├── instance/                    # 流程实例业务逻辑
│   │   ├── task/                        # 任务业务逻辑
│   │   ├── history/                     # 历史数据业务逻辑
│   │   └── variable/                    # 变量管理业务逻辑
│   ├── data/                            # 数据访问层
│   │   ├── ent/                         # Ent 数据模型
│   │   │   ├── schema/                  # 数据模型定义
│   │   │   └── migrate/                 # 数据库迁移
│   │   └── repo/                        # 仓储实现
│   │       ├── definition/              # 流程定义仓储
│   │       ├── instance/                # 流程实例仓储
│   │       ├── task/                    # 任务仓储
│   │       └── history/                 # 历史仓储
│   ├── service/                         # 服务层
│   │   ├── definition/                  # 流程定义服务
│   │   ├── runtime/                     # 运行时服务
│   │   ├── task/                        # 任务服务
│   │   ├── history/                     # 历史服务
│   │   └── management/                  # 管理服务
│   ├── server/                          # 服务器配置
│   │   ├── http/                        # HTTP 服务器
│   │   └── grpc/                        # gRPC 服务器
│   └── engine/                          # 流程引擎核心
│       ├── parser/                      # 流程解析器
│       │   ├── bpmn/                    # BPMN 解析器
│       │   └── json/                    # JSON 解析器
│       ├── executor/                    # 执行引擎
│       │   ├── activity/                # 活动执行器
│       │   ├── gateway/                 # 网关执行器
│       │   └── event/                   # 事件执行器
│       ├── scheduler/                   # 任务调度器
│       │   ├── timer/                   # 定时器
│       │   └── job/                     # 作业管理
│       ├── state/                       # 状态管理器
│       │   ├── machine/                 # 状态机
│       │   └── transition/              # 状态转换
│       ├── rule/                        # 规则引擎
│       │   ├── expression/              # 表达式引擎
│       │   └── condition/               # 条件判断
│       └── event/                       # 事件总线
│           ├── publisher/               # 事件发布
│           ├── subscriber/              # 事件订阅
│           └── handler/                 # 事件处理
├── api/                                 # API 定义
│   ├── definition/                      # 流程定义 API
│   │   └── v1/                         # 版本1
│   ├── runtime/                         # 运行时 API
│   │   └── v1/                         # 版本1
│   ├── task/                           # 任务 API
│   │   └── v1/                         # 版本1
│   ├── history/                        # 历史 API
│   │   └── v1/                         # 版本1
│   └── management/                     # 管理 API
│       └── v1/                         # 版本1
├── workflows/                          # Temporal 工作流定义
│   ├── activities/                     # 活动定义
│   ├── workflows/                      # 工作流定义
│   └── workers/                        # Worker 配置
├── configs/                            # 配置文件
│   ├── config.yaml                     # 主配置文件
│   ├── database.yaml                   # 数据库配置
│   ├── temporal.yaml                   # Temporal 配置
│   └── logging.yaml                    # 日志配置
├── pkg/                                # 公共包
│   ├── temporal/                       # Temporal 客户端
│   │   ├── client/                     # 客户端封装
│   │   └── worker/                     # Worker 管理
│   ├── config/                         # 配置解析
│   ├── cache/                          # 缓存管理
│   ├── logger/                         # 日志管理
│   ├── errors/                         # 错误定义
│   ├── validator/                      # 数据验证
│   └── utils/                          # 工具函数
├── deployments/                        # 部署配置
│   ├── docker/                         # Docker 配置
│   ├── k8s/                           # Kubernetes 配置
│   └── helm/                          # Helm Charts
├── scripts/                           # 脚本文件
│   ├── build/                         # 构建脚本
│   ├── deploy/                        # 部署脚本
│   └── migration/                     # 数据库迁移脚本
├── docs/                              # 文档
│   ├── api/                           # API 文档
│   ├── design/                        # 设计文档
│   └── user/                          # 用户手册
└── tests/                             # 测试代码
    ├── unit/                          # 单元测试
    ├── integration/                   # 集成测试
    └── e2e/                           # 端到端测试
```

## 3. 核心数据模型设计

基于[流程引擎原理与实践](https://workflow-engine-book.shuwoom.com/)第6章核心表结构设计，定义以下核心数据模型：

### 3.1 流程定义 (ProcessDefinition)
```go
// ProcessDefinition 流程定义表 - 存储流程模板和版本信息
type ProcessDefinition struct {
    ID          int64     `json:"id"`            // 流程定义ID
    Key         string    `json:"key"`           // 流程唯一标识
    Name        string    `json:"name"`          // 流程名称
    Category    string    `json:"category"`      // 流程分类
    Version     int32     `json:"version"`       // 版本号
    Description string    `json:"description"`   // 流程描述
    DeployTime  time.Time `json:"deploy_time"`   // 部署时间
    Resource    string    `json:"resource"`      // 流程文件资源
    DiagramData string    `json:"diagram_data"`  // 流程图数据(JSON)
    HasStartForm bool     `json:"has_start_form"`// 是否有启动表单
    Suspended   bool      `json:"suspended"`     // 是否挂起
    TenantID    string    `json:"tenant_id"`     // 租户ID
    CreatedAt   time.Time `json:"created_at"`    // 创建时间
    UpdatedAt   time.Time `json:"updated_at"`    // 更新时间
}
```

### 3.2 流程实例 (ProcessInstance)
```go
// ProcessInstance 流程实例表 - 存储流程执行实例
type ProcessInstance struct {
    ID                 int64      `json:"id"`                    // 流程实例ID
    ProcessDefinitionID int64     `json:"process_definition_id"` // 流程定义ID
    BusinessKey        string     `json:"business_key"`          // 业务主键
    StartUserID        string     `json:"start_user_id"`         // 启动用户ID
    StartTime          time.Time  `json:"start_time"`            // 启动时间
    EndTime            *time.Time `json:"end_time"`              // 结束时间
    Duration           *int64     `json:"duration"`              // 持续时间(毫秒)
    DeleteReason       string     `json:"delete_reason"`         // 删除原因
    ActivityID         string     `json:"activity_id"`           // 当前活动节点ID
    Name               string     `json:"name"`                  // 实例名称
    Description        string     `json:"description"`           // 实例描述
    LocalizedName      string     `json:"localized_name"`        // 本地化名称
    LocalizedDesc      string     `json:"localized_desc"`        // 本地化描述
    LockTime           *time.Time `json:"lock_time"`             // 锁定时间
    IsActive           bool       `json:"is_active"`             // 是否激活
    IsEnded            bool       `json:"is_ended"`              // 是否结束
    IsSuspended        bool       `json:"is_suspended"`          // 是否挂起
    TenantID           string     `json:"tenant_id"`             // 租户ID
    CallbackID         string     `json:"callback_id"`           // 回调ID
    CallbackType       string     `json:"callback_type"`         // 回调类型
    CreatedAt          time.Time  `json:"created_at"`            // 创建时间
    UpdatedAt          time.Time  `json:"updated_at"`            // 更新时间
}
```

### 3.3 任务实例 (TaskInstance)
```go
// TaskInstance 任务实例表 - 存储用户任务和活动任务
type TaskInstance struct {
    ID                 int64      `json:"id"`                    // 任务ID
    ProcessInstanceID  int64      `json:"process_instance_id"`   // 流程实例ID
    ProcessDefinitionID int64     `json:"process_definition_id"` // 流程定义ID
    ExecutionID        string     `json:"execution_id"`          // 执行ID
    Name               string     `json:"name"`                  // 任务名称
    Description        string     `json:"description"`           // 任务描述
    TaskDefinitionKey  string     `json:"task_definition_key"`   // 任务定义键
    Priority           int32      `json:"priority"`              // 优先级
    CreateTime         time.Time  `json:"create_time"`           // 创建时间
    ClaimTime          *time.Time `json:"claim_time"`            // 认领时间
    DueDate            *time.Time `json:"due_date"`              // 到期时间
    Category           string     `json:"category"`              // 任务分类
    ParentTaskID       *int64     `json:"parent_task_id"`        // 父任务ID
    Owner              string     `json:"owner"`                 // 拥有者
    Assignee           string     `json:"assignee"`              // 委派人
    Delegation         string     `json:"delegation"`            // 委派状态
    SuspensionState    int32      `json:"suspension_state"`      // 挂起状态
    TaskDefinitionID   string     `json:"task_definition_id"`    // 任务定义ID
    FormKey            string     `json:"form_key"`              // 表单键
    IsSuspended        bool       `json:"is_suspended"`          // 是否挂起
    TenantID           string     `json:"tenant_id"`             // 租户ID
    CreatedAt          time.Time  `json:"created_at"`            // 创建时间
    UpdatedAt          time.Time  `json:"updated_at"`            // 更新时间
}
```

### 3.4 执行历史 (HistoricProcessInstance)
```go
// HistoricProcessInstance 历史流程实例表 - 存储已完成的流程实例
type HistoricProcessInstance struct {
    ID                  int64      `json:"id"`                     // 历史ID
    ProcessInstanceID   int64      `json:"process_instance_id"`    // 流程实例ID
    ProcessDefinitionID int64      `json:"process_definition_id"`  // 流程定义ID
    ProcessDefinitionKey string    `json:"process_definition_key"` // 流程定义键
    ProcessDefinitionName string   `json:"process_definition_name"`// 流程定义名称
    ProcessDefinitionVersion int32 `json:"process_definition_version"` // 流程定义版本
    DeploymentID        string     `json:"deployment_id"`          // 部署ID
    BusinessKey         string     `json:"business_key"`           // 业务键
    StartTime           time.Time  `json:"start_time"`             // 开始时间
    EndTime             *time.Time `json:"end_time"`               // 结束时间
    Duration            *int64     `json:"duration"`               // 持续时间
    StartUserID         string     `json:"start_user_id"`          // 启动用户
    StartActivityID     string     `json:"start_activity_id"`      // 启动活动ID
    EndActivityID       string     `json:"end_activity_id"`        // 结束活动ID
    SuperProcessInstanceID *int64  `json:"super_process_instance_id"` // 父流程实例ID
    DeleteReason        string     `json:"delete_reason"`          // 删除原因
    TenantID            string     `json:"tenant_id"`              // 租户ID
    Name                string     `json:"name"`                   // 名称
    LocalizedName       string     `json:"localized_name"`         // 本地化名称
    Description         string     `json:"description"`            // 描述
    LocalizedDesc       string     `json:"localized_desc"`         // 本地化描述
    CallbackID          string     `json:"callback_id"`            // 回调ID
    CallbackType        string     `json:"callback_type"`          // 回调类型
    CreatedAt           time.Time  `json:"created_at"`             // 创建时间
    UpdatedAt           time.Time  `json:"updated_at"`             // 更新时间
}
```

### 3.5 流程变量 (ProcessVariable)
```go
// ProcessVariable 流程变量表 - 存储流程执行过程中的变量
type ProcessVariable struct {
    ID                int64     `json:"id"`                   // 变量ID
    ProcessInstanceID int64     `json:"process_instance_id"`  // 流程实例ID
    TaskID            *int64    `json:"task_id"`              // 任务ID(可选)
    Name              string    `json:"name"`                 // 变量名
    Type              string    `json:"type"`                 // 变量类型
    Value             string    `json:"value"`                // 变量值(JSON)
    Scope             string    `json:"scope"`                // 作用域(global/local)
    SerializerName    string    `json:"serializer_name"`      // 序列化器名称
    IsActive          bool      `json:"is_active"`            // 是否激活
    CreatedAt         time.Time `json:"created_at"`           // 创建时间
    UpdatedAt         time.Time `json:"updated_at"`           // 更新时间
}
```

### 3.6 事件日志 (ProcessEvent)
```go
// ProcessEvent 流程事件表 - 记录流程执行过程中的事件
type ProcessEvent struct {
    ID                int64     `json:"id"`                   // 事件ID
    ProcessInstanceID int64     `json:"process_instance_id"`  // 流程实例ID
    TaskID            *int64    `json:"task_id"`              // 任务ID(可选)
    ActivityID        string    `json:"activity_id"`          // 活动ID
    EventType         string    `json:"event_type"`           // 事件类型
    EventName         string    `json:"event_name"`           // 事件名称
    EventData         string    `json:"event_data"`           // 事件数据(JSON)
    UserID            string    `json:"user_id"`              // 用户ID
    Timestamp         time.Time `json:"timestamp"`            // 事件时间戳
    Source            string    `json:"source"`               // 事件源
    CorrelationID     string    `json:"correlation_id"`       // 关联ID
    CreatedAt         time.Time `json:"created_at"`           // 创建时间
}
```

## 4. API 设计

### 4.1 模板管理 API
```protobuf
service TemplateService {
    // 创建工作流模板
    rpc CreateTemplate(CreateTemplateRequest) returns (CreateTemplateResponse);
    // 获取工作流模板
    rpc GetTemplate(GetTemplateRequest) returns (GetTemplateResponse);
    // 更新工作流模板
    rpc UpdateTemplate(UpdateTemplateRequest) returns (UpdateTemplateResponse);
    // 删除工作流模板
    rpc DeleteTemplate(DeleteTemplateRequest) returns (DeleteTemplateResponse);
    // 列出工作流模板
    rpc ListTemplates(ListTemplatesRequest) returns (ListTemplatesResponse);
}
```

### 4.2 工作流执行 API
```protobuf
service ExecutionService {
    // 启动工作流
    rpc StartExecution(StartExecutionRequest) returns (StartExecutionResponse);
    // 获取执行详情
    rpc GetExecution(GetExecutionRequest) returns (GetExecutionResponse);
    // 列出执行记录
    rpc ListExecutions(ListExecutionsRequest) returns (ListExecutionsResponse);
    // 终止工作流
    rpc TerminateExecution(TerminateExecutionRequest) returns (TerminateExecutionResponse);
    // 暂停工作流
    rpc PauseExecution(PauseExecutionRequest) returns (PauseExecutionResponse);
    // 恢复工作流
    rpc ResumeExecution(ResumeExecutionRequest) returns (ResumeExecutionResponse);
}
```

### 4.3 历史记录 API
```protobuf
service HistoryService {
    // 获取执行历史
    rpc GetExecutionHistory(GetExecutionHistoryRequest) returns (GetExecutionHistoryResponse);
    // 获取步骤详情
    rpc GetStepDetail(GetStepDetailRequest) returns (GetStepDetailResponse);
}
```

## 5. 流程定义配置格式

基于[流程引擎原理与实践](https://workflow-engine-book.shuwoom.com/)第3章流程建模和第5章事件驱动机制，定义JSON流程配置格式：

### 5.1 流程定义配置
```json
{
    "process": {
        "id": "user-registration-process",
        "key": "userRegistration",
        "name": "用户注册流程",
        "description": "处理用户注册的完整流程",
        "version": "1.0.0",
        "category": "user-management",
        "tenant_id": "default",
        "start_form_key": "user-registration-form",
        "executable": true,
        "timeout": "PT5M",
        "variables": {
            "user_id": {
                "type": "string",
                "scope": "global",
                "required": true
            },
            "email": {
                "type": "string", 
                "scope": "global",
                "required": true
            },
            "username": {
                "type": "string",
                "scope": "global", 
                "required": true
            },
            "verification_code": {
                "type": "string",
                "scope": "local",
                "required": false
            }
        },
        "listeners": [
            {
                "event": "start",
                "type": "execution",
                "class": "com.example.ProcessStartListener"
            },
            {
                "event": "end",
                "type": "execution", 
                "class": "com.example.ProcessEndListener"
            }
        ]
    },
    "elements": [
        {
            "id": "start_event",
            "name": "开始事件",
            "type": "startEvent",
            "config": {
                "form_key": "start-form",
                "initiator": "starter"
            },
            "outgoing": ["validate_user_task"]
        },
        {
            "id": "validate_user_task",
            "name": "验证用户信息",
            "type": "userTask",
            "config": {
                "assignee": "${initiator}",
                "form_key": "validate-user-form",
                "priority": "50",
                "due_date": "PT1H",
                "category": "validation"
            },
            "listeners": [
                {
                    "event": "create",
                    "type": "task",
                    "class": "com.example.TaskCreateListener"
                },
                {
                    "event": "complete", 
                    "type": "task",
                    "class": "com.example.TaskCompleteListener"
                }
            ],
            "incoming": ["start_event"],
            "outgoing": ["email_gateway"]
        },
        {
            "id": "email_gateway",
            "name": "邮箱验证网关",
            "type": "exclusiveGateway",
            "config": {
                "default_flow": "reject_path"
            },
            "incoming": ["validate_user_task"],
            "outgoing": ["email_valid_path", "reject_path"]
        },
        {
            "id": "email_valid_path",
            "name": "邮箱有效路径",
            "type": "sequenceFlow",
            "config": {
                "condition": "${userValid == true}"
            },
            "source": "email_gateway",
            "target": "send_email_task"
        },
        {
            "id": "send_email_task",
            "name": "发送验证邮件",
            "type": "serviceTask",
            "config": {
                "implementation": "temporal_activity",
                "activity_name": "SendEmailActivity",
                "timeout": "PT30S",
                "retry_policy": {
                    "max_attempts": 3,
                    "initial_interval": "PT1S",
                    "backoff_coefficient": 2.0,
                    "maximum_interval": "PT30S"
                },
                "input_mapping": {
                    "email": "${email}",
                    "template": "verification_email",
                    "user_id": "${user_id}"
                },
                "output_mapping": {
                    "verification_code": "${result.code}",
                    "email_sent": "${result.success}"
                }
            },
            "listeners": [
                {
                    "event": "start",
                    "type": "execution",
                    "class": "com.example.EmailStartListener"
                }
            ],
            "incoming": ["email_valid_path"],
            "outgoing": ["wait_verification_event"]
        },
        {
            "id": "wait_verification_event",
            "name": "等待邮箱验证",
            "type": "intermediateCatchEvent",
            "config": {
                "signal_ref": "email_verified",
                "timeout": "PT2H",
                "timeout_action": "cancel"
            },
            "incoming": ["send_email_task"],
            "outgoing": ["complete_registration_task"]
        },
        {
            "id": "complete_registration_task",
            "name": "完成注册",
            "type": "serviceTask",
            "config": {
                "implementation": "temporal_activity",
                "activity_name": "CompleteRegistrationActivity",
                "timeout": "PT10S",
                "input_mapping": {
                    "user_id": "${user_id}",
                    "email": "${email}",
                    "username": "${username}"
                },
                "output_mapping": {
                    "registration_id": "${result.id}",
                    "success": "${result.success}"
                }
            },
            "incoming": ["wait_verification_event"],
            "outgoing": ["end_event"]
        },
        {
            "id": "reject_path",
            "name": "拒绝路径",
            "type": "sequenceFlow",
            "config": {
                "condition": "${userValid == false}"
            },
            "source": "email_gateway",
            "target": "reject_registration_task"
        },
        {
            "id": "reject_registration_task", 
            "name": "拒绝注册",
            "type": "serviceTask",
            "config": {
                "implementation": "temporal_activity",
                "activity_name": "RejectRegistrationActivity",
                "timeout": "PT5S",
                "input_mapping": {
                    "user_id": "${user_id}",
                    "reason": "用户信息验证失败"
                }
            },
            "incoming": ["reject_path"],
            "outgoing": ["end_event"]
        },
        {
            "id": "end_event",
            "name": "结束事件",
            "type": "endEvent",
            "config": {},
            "incoming": ["complete_registration_task", "reject_registration_task"]
        }
    ],
    "message_definitions": [
        {
            "id": "email_verified_message",
            "name": "邮箱验证消息",
            "structure": {
                "user_id": "string",
                "verification_code": "string",
                "verified": "boolean"
            }
        }
    ],
    "signal_definitions": [
        {
            "id": "email_verified",
            "name": "邮箱已验证信号",
            "scope": "global"
        }
    ],
    "error_definitions": [
        {
            "id": "validation_error",
            "name": "验证错误",
            "error_code": "VALIDATION_FAILED"
        },
        {
            "id": "email_send_error",
            "name": "邮件发送错误",
            "error_code": "EMAIL_SEND_FAILED"
        }
    ]
}
```

### 5.2 支持的节点类型
基于BPMN 2.0标准和该书的设计理念：

#### 事件类型 (Events)
- **startEvent**: 流程开始事件
- **endEvent**: 流程结束事件  
- **intermediateCatchEvent**: 中间捕获事件
- **intermediateThrowEvent**: 中间抛出事件
- **boundaryEvent**: 边界事件

#### 任务类型 (Tasks)
- **userTask**: 用户任务
- **serviceTask**: 服务任务
- **scriptTask**: 脚本任务
- **businessRuleTask**: 业务规则任务
- **receiveTask**: 接收任务
- **manualTask**: 手工任务

#### 网关类型 (Gateways)
- **exclusiveGateway**: 排他网关
- **parallelGateway**: 并行网关
- **inclusiveGateway**: 包容网关
- **eventBasedGateway**: 事件网关

#### 连接类型 (Connections)
- **sequenceFlow**: 顺序流
- **messageFlow**: 消息流
- **association**: 关联

## 6. 开发阶段规划

### 阶段 1: 项目基础设施 (第1-2周)
- [ ] 初始化 Kratos 项目结构
- [ ] 配置开发环境 (Docker Compose)
- [ ] 设置数据库连接和 Ent Schema
- [ ] 配置 Temporal 连接
- [ ] 实现基础中间件 (日志、错误处理、认证)

### 阶段 2: 模板管理模块 (第3-4周)
- [ ] 定义模板相关的 Protocol Buffers
- [ ] 实现模板 Ent Schema
- [ ] 开发模板 CRUD 业务逻辑
- [ ] 实现模板服务层
- [ ] 编写模板管理 API
- [ ] 添加模板配置验证逻辑
- [ ] 编写单元测试

### 阶段 3: 工作流执行引擎 (第5-7周)
- [ ] 定义 Temporal 工作流和活动
- [ ] 实现工作流配置解析器
- [ ] 开发工作流执行逻辑
- [ ] 实现工作流状态管理
- [ ] 集成 Temporal 客户端
- [ ] 实现工作流控制操作 (启动、暂停、终止)
- [ ] 编写集成测试

### 阶段 4: 执行历史和监控 (第8-9周)
- [ ] 实现执行历史记录
- [ ] 开发状态查询 API
- [ ] 实现实时状态推送
- [ ] 添加执行统计功能
- [ ] 实现日志聚合
- [ ] 编写性能测试

### 阶段 5: 优化和部署 (第10-11周)
- [ ] 性能优化和缓存策略
- [ ] 安全性增强 (JWT 认证、权限控制)
- [ ] 编写部署文档
- [ ] 配置 CI/CD 流水线
- [ ] 生产环境部署
- [ ] 压力测试和调优

### 阶段 6: 文档和维护 (第12周)
- [ ] 完善 API 文档
- [ ] 编写用户使用手册
- [ ] 代码重构和优化
- [ ] 监控告警配置
- [ ] 备份恢复方案

## 7. 开发任务清单

### 7.1 基础设施任务
```bash
# 任务1: 初始化项目
kratos new workflow-engine
cd workflow-engine

# 任务2: 配置依赖
go mod tidy
go get entgo.io/ent/cmd/ent
go get go.temporal.io/sdk

# 任务3: 项目结构调整
mkdir -p internal/{biz,data,service}/{template,execution,workflow}
mkdir -p api/{template,execution,workflow}
mkdir -p workflows/
mkdir -p pkg/{temporal,config,utils}

# 任务4: Docker Compose 配置
touch docker-compose.yaml
```

### 7.2 代码生成任务
```bash
# 任务1: 生成 Ent Schema
ent init Template Execution ExecutionHistory

# 任务2: 生成 Protocol Buffers
protoc --go_out=. --go-grpc_out=. api/**/*.proto

# 任务3: 生成 Kratos HTTP
protoc --go-http_out=. api/**/*.proto
```

### 7.3 核心开发任务

#### Template 模块
- [ ] `internal/data/ent/schema/template.go` - Ent Schema 定义
- [ ] `internal/data/repo/template.go` - 模板数据仓储
- [ ] `internal/biz/template/template.go` - 模板业务逻辑
- [ ] `internal/service/template/template.go` - 模板服务实现
- [ ] `api/template/v1/template.proto` - 模板 API 定义

#### Execution 模块
- [ ] `internal/data/ent/schema/execution.go` - 执行 Schema 定义
- [ ] `internal/data/repo/execution.go` - 执行数据仓储
- [ ] `internal/biz/execution/execution.go` - 执行业务逻辑
- [ ] `internal/service/execution/execution.go` - 执行服务实现
- [ ] `api/execution/v1/execution.proto` - 执行 API 定义

#### Workflow 模块
- [ ] `workflows/registration.go` - 注册工作流定义
- [ ] `workflows/activities.go` - 活动定义
- [ ] `pkg/temporal/client.go` - Temporal 客户端
- [ ] `pkg/config/parser.go` - 配置解析器

### 7.4 测试任务
- [ ] 模板管理单元测试
- [ ] 工作流执行集成测试
- [ ] API 接口测试
- [ ] 性能压力测试
- [ ] 端到端测试

## 8. 流程引擎核心组件实现

基于[流程引擎原理与实践](https://workflow-engine-book.shuwoom.com/)第4章核心组件设计：

### 8.1 流程解析器 (Process Parser)
```go
// 流程解析器接口 - 解析BPMN/JSON格式的流程定义
type ProcessParser interface {
    // 解析流程定义
    Parse(data []byte, format string) (*ProcessModel, error)
    // 验证流程定义
    Validate(model *ProcessModel) error
    // 生成流程图
    GenerateDiagram(model *ProcessModel) ([]byte, error)
}

// JSON格式解析器实现
type JsonParser struct {
    validator *ProcessValidator
    logger    log.Logger
}

// 流程模型
type ProcessModel struct {
    Process   ProcessInfo     `json:"process"`
    Elements  []ProcessElement `json:"elements"`
    Messages  []MessageDef    `json:"message_definitions"`
    Signals   []SignalDef     `json:"signal_definitions"`
    Errors    []ErrorDef      `json:"error_definitions"`
}
```

### 8.2 执行引擎 (Execution Engine)
```go
// 执行引擎接口 - 负责流程实例的执行
type ExecutionEngine interface {
    // 启动流程实例
    StartProcess(ctx context.Context, req *StartProcessRequest) (*ProcessInstance, error)
    // 继续执行流程
    ContinueExecution(ctx context.Context, instanceID int64, token string) error
    // 处理任务完成
    CompleteTask(ctx context.Context, taskID int64, variables map[string]interface{}) error
    // 处理信号事件
    SignalEvent(ctx context.Context, signal string, data map[string]interface{}) error
}

// 执行引擎实现
type DefaultExecutionEngine struct {
    processRepo     ProcessDefinitionRepo
    instanceRepo    ProcessInstanceRepo
    taskRepo        TaskInstanceRepo
    stateManager    StateManager
    eventBus        EventBus
    scheduler       TaskScheduler
    logger          log.Logger
}
```

### 8.3 状态管理器 (State Manager)  
```go
// 状态管理器接口 - 管理流程实例和任务的状态
type StateManager interface {
    // 创建执行令牌
    CreateToken(ctx context.Context, instanceID int64, elementID string) (*ExecutionToken, error)
    // 移动令牌
    MoveToken(ctx context.Context, token *ExecutionToken, targetElement string) error
    // 处理网关
    ProcessGateway(ctx context.Context, token *ExecutionToken, gateway *GatewayElement) error
    // 处理并行分支
    ProcessParallelSplit(ctx context.Context, token *ExecutionToken, branches []string) error
    // 处理并行合并
    ProcessParallelJoin(ctx context.Context, tokens []*ExecutionToken) (*ExecutionToken, error)
}

// 执行令牌 - 表示流程执行的当前位置
type ExecutionToken struct {
    ID            string    `json:"id"`
    InstanceID    int64     `json:"instance_id"`
    ElementID     string    `json:"element_id"`
    ParentTokenID *string   `json:"parent_token_id"`
    IsActive      bool      `json:"is_active"`
    Variables     map[string]interface{} `json:"variables"`
    CreatedAt     time.Time `json:"created_at"`
}
```

### 8.4 事件总线 (Event Bus)
```go
// 事件总线接口 - 基于第5章事件驱动机制设计
type EventBus interface {
    // 发布事件
    Publish(ctx context.Context, event *ProcessEvent) error
    // 订阅事件
    Subscribe(eventType string, handler EventHandler) error
    // 取消订阅
    Unsubscribe(eventType string, handler EventHandler) error
}

// 事件处理器接口
type EventHandler interface {
    Handle(ctx context.Context, event *ProcessEvent) error
    GetEventTypes() []string
}

// 流程事件
type ProcessEvent struct {
    ID            string                 `json:"id"`
    Type          string                 `json:"type"`
    Source        string                 `json:"source"`
    InstanceID    int64                  `json:"instance_id"`
    ElementID     string                 `json:"element_id"`
    Data          map[string]interface{} `json:"data"`
    Timestamp     time.Time              `json:"timestamp"`
    CorrelationID string                 `json:"correlation_id"`
}

// 事件类型常量
const (
    EventProcessStarted    = "process.started"
    EventProcessCompleted  = "process.completed"
    EventTaskCreated      = "task.created"
    EventTaskCompleted    = "task.completed"
    EventActivityStarted  = "activity.started"
    EventActivityCompleted = "activity.completed"
    EventSignalReceived   = "signal.received"
    EventTimerFired       = "timer.fired"
)
```

### 8.5 任务调度器 (Task Scheduler)
```go
// 任务调度器接口 - 基于第8章分布式任务调度设计
type TaskScheduler interface {
    // 调度定时任务
    ScheduleTimer(ctx context.Context, timer *TimerDefinition) error
    // 取消定时任务
    CancelTimer(ctx context.Context, timerID string) error
    // 调度异步任务
    ScheduleAsyncTask(ctx context.Context, task *AsyncTask) error
    // 处理任务超时
    HandleTimeout(ctx context.Context, taskID string) error
}

// 定时器定义
type TimerDefinition struct {
    ID         string    `json:"id"`
    Type       string    `json:"type"`         // date, duration, cycle
    Expression string    `json:"expression"`   // ISO 8601格式
    InstanceID int64     `json:"instance_id"`
    ElementID  string    `json:"element_id"`
    CreatedAt  time.Time `json:"created_at"`
}

// 异步任务
type AsyncTask struct {
    ID         string                 `json:"id"`
    Type       string                 `json:"type"`
    InstanceID int64                  `json:"instance_id"`
    ElementID  string                 `json:"element_id"`
    Input      map[string]interface{} `json:"input"`
    Timeout    time.Duration          `json:"timeout"`
    RetryCount int                    `json:"retry_count"`
    CreatedAt  time.Time              `json:"created_at"`
}
```

### 8.6 规则引擎 (Rule Engine)
```go
// 规则引擎接口 - 处理网关条件和业务规则
type RuleEngine interface {
    // 评估条件表达式
    EvaluateCondition(ctx context.Context, expression string, variables map[string]interface{}) (bool, error)
    // 评估业务规则
    EvaluateBusinessRule(ctx context.Context, ruleKey string, input map[string]interface{}) (map[string]interface{}, error)
    // 注册自定义函数
    RegisterFunction(name string, fn interface{}) error
}

// 表达式引擎实现
type ExpressionEngine struct {
    functions map[string]interface{}
    logger    log.Logger
}

// 支持的表达式语法示例
// ${variable_name}                    - 变量引用
// ${user.age > 18}                   - 比较表达式  
// ${status == 'approved'}            - 字符串比较
// ${amount >= 1000 && category == 'vip'} - 逻辑表达式
// ${fn:customFunction(param1, param2)} - 自定义函数调用
```

### 8.7 Kratos 框架集成
- 使用 Kratos Wire 进行依赖注入
- 配置 HTTP/gRPC 服务器
- 实现中间件链 (认证、日志、限流、监控)
- 配置服务发现和负载均衡
- 集成 OpenTelemetry 链路追踪

### 8.8 Ent ORM 集成
- Schema 定义和关联关系
- 数据库迁移管理  
- 查询优化和分页
- 事务处理和并发控制
- 软删除和审计日志

### 8.9 Temporal 工作流引擎集成  
- Worker 配置和注册
- 工作流和活动定义
- 错误处理和重试策略
- 信号和查询处理
- 版本管理和热更新

### 8.10 缓存和性能优化
- Redis 缓存流程定义
- 流程实例状态缓存
- 数据库连接池优化
- 异步事件处理
- 批量操作优化

## 9. 质量保证

### 9.1 代码质量
- 代码覆盖率 > 80%
- 使用 golangci-lint 进行静态检查
- 代码审查流程
- 文档注释完善

### 9.2 性能要求
- API 响应时间 < 200ms
- 支持并发 1000+ 工作流
- 数据库查询优化
- 内存使用监控

### 9.3 安全要求
- JWT 认证和授权
- 输入验证和过滤
- SQL 注入防护
- 敏感数据加密

## 10. 监控和运维

### 10.1 监控指标
- 服务健康检查
- 工作流执行统计
- 系统资源监控
- 错误率和延迟监控

### 10.2 日志管理
- 结构化日志输出
- 日志级别配置
- 日志聚合和分析
- 错误追踪

### 10.3 备份策略
- 数据库定期备份
- 配置文件版本控制
- 灾难恢复计划
- 数据一致性检查

---

## 开发进度跟踪

### 当前状态
- **当前阶段**: 阶段 2 已完成 ✅
- **完成度**: 33.3% (2/6 阶段)
- **下一步**: 阶段 3 - 业务逻辑层开发

### 已完成阶段

#### ✅ 阶段 1: 项目基础设施开发 (第1-2周)
**完成时间**: 2024年12月
**完成内容**:
- ✅ 初始化 Kratos 项目结构
- ✅ 设置 Docker Compose 开发环境 (PostgreSQL, Redis, Temporal)
- ✅ 设置 Ent Schema 基础 (6个核心数据模型)
- ✅ 创建配置管理 (支持 YAML 和环境变量)
- ✅ 创建基础单元测试 (配置管理测试覆盖率 100%)
- ✅ 创建 Dockerfile 和部署配置

**验证结果**:
- 所有单元测试通过 ✅
- Docker Compose 配置验证通过 ✅
- Ent 代码生成成功 ✅
- 项目结构符合规范 ✅

**生成的核心文件**:
- `go.mod` - Go 模块定义
- `configs/config.yaml` - 主配置文件
- `docker-compose.yaml` - 开发环境配置
- `internal/data/ent/schema/` - 6个数据模型 Schema
- `pkg/config/` - 配置管理包
- `deployments/docker/Dockerfile` - 容器化配置
- `.dockerignore` - Docker 忽略文件

#### ✅ 阶段 2: 数据访问层开发 (第3-4周)
**完成时间**: 2024年12月
**完成内容**:
- ✅ 数据库连接管理 (`internal/data/data.go`)
- ✅ Repository 接口定义 (`internal/biz/repository.go`)
- ✅ 流程定义 Repository 实现 (`internal/data/repository/process_definition.go`)
- ✅ 缓存 Repository 实现 (`internal/data/repository/cache.go`)
- ✅ 数据库健康检查和迁移功能
- ✅ Redis 缓存集成 (支持基本缓存和哈希缓存)
- ✅ 完整的单元测试覆盖

**验证结果**:
- 数据访问层测试通过 ✅
- Repository 层测试通过 ✅
- 数据库连接和缓存功能验证通过 ✅
- 代码符合项目规范 (中文注释、错误处理、日志记录) ✅

**生成的核心文件**:
- `internal/data/data.go` - 数据访问层核心
- `internal/biz/repository.go` - Repository 接口定义
- `internal/data/repository/process_definition.go` - 流程定义仓储实现
- `internal/data/repository/cache.go` - 缓存仓储实现
- `internal/data/data_test.go` - 数据访问层测试
- `internal/data/repository/process_definition_test.go` - Repository 测试

---

## 注意事项

1. **开发优先级**: 严格按照阶段顺序进行，确保每个阶段完成后再进入下一阶段
2. **代码规范**: 严格遵循 `.cursorrules` 中定义的编码规范
3. **测试驱动**: 每个模块开发完成后立即编写对应的测试用例
4. **文档同步**: 代码变更时同步更新相关文档
5. **性能考虑**: 在设计阶段就要考虑性能优化和扩展性
6. **安全第一**: 所有外部接口都要进行安全验证和防护 