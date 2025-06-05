# 工作流引擎 API 参考文档

## 概述

工作流引擎提供完整的 RESTful API，支持流程定义管理、流程实例控制、任务处理等核心功能。所有 API 都遵循统一的响应格式和错误处理机制。

### API 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: Bearer Token (JWT)
- **API 版本**: v1.0.0

### 统一响应格式

#### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体数据内容
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

#### 分页响应

```json
{
  "code": 0,
  "message": "success", 
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "size": 10,
      "total": 100,
      "pages": 10
    }
  },
  "timestamp": "2024-12-19T10:30:00Z"
}
```

#### 错误响应

```json
{
  "code": 40001,
  "message": "参数验证失败",
  "error": "详细错误信息",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

### 错误码说明

| 错误码 | 说明 | HTTP状态码 |
|--------|------|------------|
| 0 | 成功 | 200 |
| 10001 | 参数验证失败 | 400 |
| 10002 | 请求格式错误 | 400 |
| 20001 | 资源不存在 | 404 |
| 20002 | 资源已存在 | 409 |
| 30001 | 业务逻辑错误 | 422 |
| 40001 | 认证失败 | 401 |
| 40301 | 权限不足 | 403 |
| 50001 | 内部服务器错误 | 500 |
| 50002 | 数据库错误 | 500 |

## 认证授权

### 登录认证

**POST** `/auth/login`

用户登录获取访问令牌。

#### 请求参数

```json
{
  "username": "admin",
  "password": "password123"
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "expires_at": "2024-12-20T10:30:00Z"
  }
}
```

### 刷新令牌

**POST** `/auth/refresh`

使用刷新令牌获取新的访问令牌。

#### 请求参数

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 退出登录

**POST** `/auth/logout`

吊销访问令牌。

#### 请求头

```http
Authorization: Bearer <access_token>
```

## 流程定义管理

### 创建流程定义

**POST** `/process-definitions`

创建新的流程定义。

#### 权限要求

- `process:write`

#### 请求参数

```json
{
  "name": "订单处理流程",
  "description": "电商订单从创建到完成的完整流程",
  "version": "1.0.0",
  "category": "业务流程",
  "config": {
    "name": "订单处理流程",
    "description": "电商订单处理的完整工作流程",
    "version": "1.0.0",
    "variables": {
      "order_id": "string",
      "customer_id": "string",
      "total_amount": "number",
      "status": "string"
    },
    "steps": [
      {
        "id": "validate_order",
        "name": "验证订单",
        "type": "service_task",
        "config": {
          "service_name": "OrderValidationService",
          "method": "validate",
          "timeout": "30s"
        },
        "next": ["payment_check"]
      }
    ]
  }
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "流程定义创建成功",
  "data": {
    "id": 123,
    "name": "订单处理流程",
    "description": "电商订单从创建到完成的完整流程",
    "version": "1.0.0",
    "category": "业务流程",
    "status": "active",
    "config": { ... },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### 获取流程定义列表

**GET** `/process-definitions`

获取流程定义列表，支持分页和过滤。

#### 权限要求

- `process:read`

#### 查询参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页大小，默认10 |
| name | string | 否 | 流程名称模糊查询 |
| category | string | 否 | 流程分类 |
| status | string | 否 | 流程状态：active, inactive |
| sort | string | 否 | 排序字段：name, created_at, updated_at |
| order | string | 否 | 排序方向：asc, desc |

#### 请求示例

```bash
GET /api/v1/process-definitions?page=1&size=10&name=订单&status=active
```

#### 响应示例

```json
{
  "code": 0,
  "message": "获取流程定义列表成功",
  "data": {
    "items": [
      {
        "id": 123,
        "name": "订单处理流程",
        "description": "电商订单处理流程",
        "version": "1.0.0",
        "category": "业务流程",
        "status": "active",
        "created_at": "2024-12-19T10:30:00Z",
        "updated_at": "2024-12-19T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "size": 10,
      "total": 1,
      "pages": 1
    }
  }
}
```

### 获取流程定义详情

**GET** `/process-definitions/{id}`

获取指定流程定义的详细信息。

#### 权限要求

- `process:read`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 流程定义ID |

#### 响应示例

```json
{
  "code": 0,
  "message": "获取流程定义详情成功",
  "data": {
    "id": 123,
    "name": "订单处理流程",
    "description": "电商订单处理流程",
    "version": "1.0.0",
    "category": "业务流程", 
    "status": "active",
    "config": {
      "name": "订单处理流程",
      "variables": { ... },
      "steps": [ ... ]
    },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### 更新流程定义

**PUT** `/process-definitions/{id}`

更新指定的流程定义。

#### 权限要求

- `process:write`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 流程定义ID |

#### 请求参数

```json
{
  "name": "订单处理流程 v2",
  "description": "更新后的订单处理流程",
  "version": "2.0.0",
  "category": "业务流程",
  "config": { ... }
}
```

### 删除流程定义

**DELETE** `/process-definitions/{id}`

删除指定的流程定义。

#### 权限要求

- `process:delete`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 流程定义ID |

### 部署流程定义

**POST** `/process-definitions/{id}/deploy`

部署流程定义，使其可以被实例化。

#### 权限要求

- `process:deploy`

#### 路径参数

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 流程定义ID |

## 流程实例管理

### 启动流程实例

**POST** `/process-instances`

启动新的流程实例。

#### 权限要求

- `instance:create`

#### 请求参数

```json
{
  "process_definition_id": 123,
  "business_key": "ORDER-001",
  "variables": {
    "order_id": "ORDER-001",
    "customer_id": "CUST-456",
    "total_amount": 999.99,
    "priority": "high"
  }
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "流程实例启动成功",
  "data": {
    "id": 456,
    "process_definition_id": 123,
    "business_key": "ORDER-001",
    "status": "running",
    "current_activity": "validate_order",
    "variables": { ... },
    "started_at": "2024-12-19T10:30:00Z",
    "started_by": "user123"
  }
}
```

### 获取流程实例列表

**GET** `/process-instances`

获取流程实例列表。

#### 权限要求

- `instance:read`

#### 查询参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页大小，默认10 |
| process_definition_id | int | 否 | 流程定义ID |
| business_key | string | 否 | 业务键 |
| status | string | 否 | 实例状态：running, completed, suspended, terminated |
| started_by | string | 否 | 启动人 |
| started_after | string | 否 | 启动时间范围（开始） |
| started_before | string | 否 | 启动时间范围（结束） |

#### 请求示例

```bash
GET /api/v1/process-instances?status=running&process_definition_id=123
```

### 获取流程实例详情

**GET** `/process-instances/{id}`

获取指定流程实例的详细信息。

#### 权限要求

- `instance:read`

#### 响应示例

```json
{
  "code": 0,
  "message": "获取流程实例详情成功",
  "data": {
    "id": 456,
    "process_definition_id": 123,
    "process_definition_name": "订单处理流程",
    "business_key": "ORDER-001",
    "status": "running",
    "current_activity": "payment_check",
    "variables": {
      "order_id": "ORDER-001",
      "customer_id": "CUST-456",
      "total_amount": 999.99,
      "validation_result": "passed"
    },
    "started_at": "2024-12-19T10:30:00Z",
    "started_by": "user123",
    "ended_at": null,
    "duration": 1800
  }
}
```

### 暂停流程实例

**PUT** `/process-instances/{id}/suspend`

暂停指定的流程实例。

#### 权限要求

- `instance:control`

### 恢复流程实例

**PUT** `/process-instances/{id}/activate`

恢复暂停的流程实例。

#### 权限要求

- `instance:control`

### 终止流程实例

**DELETE** `/process-instances/{id}`

终止指定的流程实例。

#### 权限要求

- `instance:control`

### 获取流程实例变量

**GET** `/process-instances/{id}/variables`

获取流程实例的所有变量。

#### 权限要求

- `instance:read`

#### 响应示例

```json
{
  "code": 0,
  "message": "获取流程变量成功",
  "data": {
    "variables": {
      "order_id": "ORDER-001",
      "customer_id": "CUST-456",
      "total_amount": 999.99,
      "status": "processing"
    }
  }
}
```

### 设置流程实例变量

**PUT** `/process-instances/{id}/variables`

设置流程实例的变量。

#### 权限要求

- `instance:write`

#### 请求参数

```json
{
  "variables": {
    "status": "approved",
    "approval_time": "2024-12-19T11:00:00Z"
  }
}
```

## 任务管理

### 获取任务列表

**GET** `/tasks`

获取任务列表，支持多种过滤条件。

#### 权限要求

- `task:read`

#### 查询参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页大小，默认10 |
| assignee | string | 否 | 任务分配人，"current_user"表示当前用户 |
| candidate_user | string | 否 | 候选人 |
| candidate_group | string | 否 | 候选组 |
| process_instance_id | int | 否 | 流程实例ID |
| status | string | 否 | 任务状态：created, assigned, completed |
| created_after | string | 否 | 创建时间范围（开始） |
| created_before | string | 否 | 创建时间范围（结束） |

#### 请求示例

```bash
GET /api/v1/tasks?assignee=current_user&status=created
```

#### 响应示例

```json
{
  "code": 0,
  "message": "获取任务列表成功",
  "data": {
    "items": [
      {
        "id": 789,
        "name": "支付检查",
        "description": "检查订单支付状态",
        "process_instance_id": 456,
        "process_definition_name": "订单处理流程",
        "assignee": "user123",
        "candidate_users": [],
        "candidate_groups": ["finance_team"],
        "status": "created",
        "priority": "normal",
        "due_date": "2024-12-19T14:30:00Z",
        "form_data": {
          "payment_verified": "boolean",
          "notes": "string"
        },
        "created_at": "2024-12-19T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "size": 10,
      "total": 1,
      "pages": 1
    }
  }
}
```

### 获取任务详情

**GET** `/tasks/{id}`

获取指定任务的详细信息。

#### 权限要求

- `task:read`

#### 响应示例

```json
{
  "code": 0,
  "message": "获取任务详情成功",
  "data": {
    "id": 789,
    "name": "支付检查",
    "description": "检查订单支付状态并确认",
    "process_instance_id": 456,
    "process_definition_id": 123,
    "process_definition_name": "订单处理流程",
    "business_key": "ORDER-001",
    "assignee": "user123",
    "candidate_users": [],
    "candidate_groups": ["finance_team"],
    "status": "created",
    "priority": "normal",
    "due_date": "2024-12-19T14:30:00Z",
    "form_data": {
      "payment_verified": "boolean",
      "amount": "number",
      "notes": "string"
    },
    "variables": {
      "order_id": "ORDER-001",
      "total_amount": 999.99
    },
    "created_at": "2024-12-19T10:30:00Z",
    "updated_at": "2024-12-19T10:30:00Z"
  }
}
```

### 认领任务

**PUT** `/tasks/{id}/claim`

认领指定的任务。

#### 权限要求

- `task:claim`

#### 响应示例

```json
{
  "code": 0,
  "message": "任务认领成功",
  "data": {
    "id": 789,
    "assignee": "user123",
    "status": "assigned",
    "claimed_at": "2024-12-19T10:45:00Z"
  }
}
```

### 释放任务

**PUT** `/tasks/{id}/unclaim`

释放已认领的任务。

#### 权限要求

- `task:unclaim`

### 完成任务

**PUT** `/tasks/{id}/complete`

完成指定的任务。

#### 权限要求

- `task:complete`

#### 请求参数

```json
{
  "variables": {
    "payment_verified": true,
    "amount": 999.99,
    "notes": "支付验证通过，金额无误"
  }
}
```

#### 响应示例

```json
{
  "code": 0,
  "message": "任务完成成功",
  "data": {
    "id": 789,
    "status": "completed",
    "completed_at": "2024-12-19T11:00:00Z",
    "completed_by": "user123"
  }
}
```

### 委派任务

**PUT** `/tasks/{id}/delegate`

将任务委派给其他用户。

#### 权限要求

- `task:delegate`

#### 请求参数

```json
{
  "assignee": "user456",
  "reason": "专业领域由专人处理"
}
```

### 设置任务变量

**PUT** `/tasks/{id}/variables`

设置任务的局部变量。

#### 权限要求

- `task:write`

#### 请求参数

```json
{
  "variables": {
    "notes": "已联系客户确认",
    "contact_time": "2024-12-19T10:45:00Z"
  }
}
```

## 历史数据查询

### 获取历史流程实例

**GET** `/history/process-instances`

查询历史流程实例数据。

#### 权限要求

- `history:read`

#### 查询参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页大小，默认10 |
| process_definition_id | int | 否 | 流程定义ID |
| business_key | string | 否 | 业务键 |
| status | string | 否 | 最终状态 |
| started_by | string | 否 | 启动人 |
| started_after | string | 否 | 启动时间范围（开始） |
| started_before | string | 否 | 启动时间范围（结束） |
| finished_after | string | 否 | 结束时间范围（开始） |
| finished_before | string | 否 | 结束时间范围（结束） |

### 获取历史任务

**GET** `/history/tasks`

查询历史任务数据。

#### 权限要求

- `history:read`

#### 查询参数

类似于当前任务查询，但包含已完成的任务。

### 获取历史变量

**GET** `/history/variables`

查询历史变量变更记录。

#### 权限要求

- `history:read`

## 系统管理

### 健康检查

**GET** `/health`

检查系统健康状态。

#### 响应示例

```json
{
  "code": 0,
  "message": "系统运行正常",
  "data": {
    "status": "healthy",
    "timestamp": "2024-12-19T10:30:00Z",
    "services": {
      "database": "healthy",
      "redis": "healthy",
      "temporal": "healthy"
    },
    "version": "1.0.0",
    "uptime": 86400
  }
}
```

### 系统指标

**GET** `/metrics`

获取 Prometheus 格式的系统指标。

### 系统信息

**GET** `/info`

获取系统基本信息。

#### 权限要求

- `system:read`

#### 响应示例

```json
{
  "code": 0,
  "message": "获取系统信息成功",
  "data": {
    "name": "workflow-engine",
    "version": "1.0.0",
    "build_time": "2024-12-19T08:00:00Z",
    "git_commit": "abc123def456",
    "go_version": "go1.21.0",
    "environment": "production",
    "features": {
      "auth": true,
      "cache": true,
      "monitoring": true
    }
  }
}
```

## 错误处理

### 常见错误示例

#### 参数验证错误

```json
{
  "code": 10001,
  "message": "参数验证失败",
  "error": "name字段是必需的",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

#### 资源不存在

```json
{
  "code": 20001,
  "message": "资源不存在",
  "error": "流程定义ID 999 不存在",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

#### 权限不足

```json
{
  "code": 40301,
  "message": "权限不足",
  "error": "缺少 process:write 权限",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

#### 业务逻辑错误

```json
{
  "code": 30001,
  "message": "业务逻辑错误", 
  "error": "流程实例已完成，无法暂停",
  "timestamp": "2024-12-19T10:30:00Z"
}
```

## SDK 示例

### Go SDK

```go
package main

import (
    "context"
    "fmt"
    "github.com/yourorg/workflow-engine-sdk-go"
)

func main() {
    // 创建客户端
    client := workflowsdk.NewClient("http://localhost:8080", "your-jwt-token")
    
    // 启动流程实例
    instance, err := client.ProcessInstances.Start(context.Background(), &workflowsdk.StartProcessInstanceRequest{
        ProcessDefinitionID: 123,
        BusinessKey:        "ORDER-001",
        Variables: map[string]interface{}{
            "order_id": "ORDER-001",
            "amount":   999.99,
        },
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("流程实例ID: %d\n", instance.ID)
}
```

### JavaScript SDK

```javascript
import { WorkflowClient } from '@yourorg/workflow-engine-sdk-js';

const client = new WorkflowClient({
  baseURL: 'http://localhost:8080',
  token: 'your-jwt-token'
});

// 启动流程实例
const instance = await client.processInstances.start({
  processDefinitionId: 123,
  businessKey: 'ORDER-001',
  variables: {
    order_id: 'ORDER-001',
    amount: 999.99
  }
});

console.log('流程实例ID:', instance.id);
```

### Python SDK

```python
from workflow_engine_sdk import WorkflowClient

# 创建客户端
client = WorkflowClient(
    base_url='http://localhost:8080',
    token='your-jwt-token'
)

# 启动流程实例
instance = client.process_instances.start(
    process_definition_id=123,
    business_key='ORDER-001',
    variables={
        'order_id': 'ORDER-001',
        'amount': 999.99
    }
)

print(f'流程实例ID: {instance.id}')
```

## 版本更新日志

### v1.0.0 (2024-12-19)

#### 新增功能
- 完整的 RESTful API
- JWT 认证授权
- 流程定义管理
- 流程实例控制
- 任务处理
- 历史数据查询
- 系统监控

#### API 变更
- 初始版本发布

---

**文档版本**: v1.0.0  
**最后更新**: 2024年12月19日  
**API 版本**: v1.0.0 