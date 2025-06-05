# 工作流引擎 RESTful API 文档

## 概述

工作流引擎提供完整的RESTful API，支持流程定义管理、流程实例运行时控制、任务管理、历史数据查询等功能。

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **内容类型**: `application/json`
- **字符编码**: `UTF-8`

## 统一响应格式

所有API响应都使用统一的JSON格式：

```json
{
  "code": 200,
  "message": "成功",
  "data": {...},
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**响应字段说明**:
- `code`: 响应码 (200=成功, 400=客户端错误, 500=服务器错误)
- `message`: 响应消息 (中文描述)
- `data`: 响应数据 (成功时包含具体数据)
- `error`: 错误信息 (失败时包含错误详情)
- `timestamp`: 响应时间戳

## 1. 流程定义管理

### 1.1 查询流程定义列表

**请求**:
```http
GET /api/v1/process-definitions
```

**查询参数**:
- `page` (可选): 页码，默认1
- `page_size` (可选): 每页大小，默认20，最大100
- `key` (可选): 流程定义Key
- `name` (可选): 流程定义名称
- `category` (可选): 流程分类

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "items": [
      {
        "id": "1",
        "key": "sample-process",
        "name": "示例流程",
        "version": 1,
        "category": "业务流程",
        "suspended": false,
        "deploy_time": "2024-01-01T12:00:00Z",
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 1.2 创建流程定义

**请求**:
```http
POST /api/v1/process-definitions
Content-Type: application/json

{
  "key": "new-process",
  "name": "新流程",
  "description": "这是一个新的流程定义",
  "resource": "{\"version\": \"1.0\", \"steps\": [...]}",
  "category": "业务流程"
}
```

**请求参数**:
- `key` (必须): 流程定义Key，全局唯一
- `name` (必须): 流程定义名称
- `description` (可选): 流程描述
- `resource` (必须): 流程定义资源 (JSON格式)
- `category` (可选): 流程分类

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "2",
    "key": "new-process",
    "name": "新流程",
    "version": 1,
    "status": "created"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 1.3 获取流程定义详情

**请求**:
```http
GET /api/v1/process-definitions/{id}
```

**路径参数**:
- `id`: 流程定义ID

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "1",
    "key": "sample-process",
    "name": "示例流程",
    "version": 1,
    "description": "这是一个示例流程",
    "resource": "{\"version\": \"1.0\", \"steps\": [...]}",
    "category": "业务流程",
    "suspended": false,
    "deploy_time": "2024-01-01T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 1.4 更新流程定义

**请求**:
```http
PUT /api/v1/process-definitions/{id}
Content-Type: application/json

{
  "name": "更新后的流程名称",
  "description": "更新后的描述",
  "resource": "{\"version\": \"1.1\", \"steps\": [...]}",
  "category": "更新后的分类"
}
```

### 1.5 删除流程定义

**请求**:
```http
DELETE /api/v1/process-definitions/{id}
```

### 1.6 部署流程定义

**请求**:
```http
POST /api/v1/process-definitions/{id}/deploy
```

## 2. 流程实例管理

### 2.1 启动流程实例

**请求**:
```http
POST /api/v1/process-instances
Content-Type: application/json

{
  "process_definition_id": "1",
  "business_key": "business-001",
  "variables": {
    "applicant": "张三",
    "amount": 10000
  },
  "start_user_id": "user-1"
}
```

**请求参数**:
- `process_definition_id` (可选): 流程定义ID
- `process_definition_key` (可选): 流程定义Key (与ID二选一)
- `business_key` (可选): 业务键
- `variables` (可选): 初始流程变量
- `start_user_id` (可选): 启动用户ID

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "inst-1",
    "process_definition_id": "1",
    "process_definition_key": "sample-process",
    "business_key": "business-001",
    "status": "running",
    "start_time": "2024-01-01T12:00:00Z",
    "start_user_id": "user-1"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 2.2 查询流程实例列表

**请求**:
```http
GET /api/v1/process-instances?page=1&page_size=20&status=running
```

**查询参数**:
- `page` (可选): 页码
- `page_size` (可选): 每页大小
- `process_definition_key` (可选): 流程定义Key
- `business_key` (可选): 业务键
- `status` (可选): 实例状态 (running|suspended|completed|terminated)
- `start_user_id` (可选): 启动用户ID

### 2.3 获取流程实例详情

**请求**:
```http
GET /api/v1/process-instances/{id}
```

### 2.4 挂起流程实例

**请求**:
```http
POST /api/v1/process-instances/{id}/suspend
```

### 2.5 激活流程实例

**请求**:
```http
POST /api/v1/process-instances/{id}/activate
```

### 2.6 终止流程实例

**请求**:
```http
POST /api/v1/process-instances/{id}/terminate
Content-Type: application/json

{
  "reason": "业务需要终止"
}
```

## 3. 任务管理

### 3.1 查询任务列表

**请求**:
```http
GET /api/v1/tasks?assignee=user-1&status=pending
```

**查询参数**:
- `page` (可选): 页码
- `page_size` (可选): 每页大小
- `assignee` (可选): 任务处理人
- `candidate_user` (可选): 候选用户
- `candidate_group` (可选): 候选组
- `process_instance_id` (可选): 流程实例ID
- `status` (可选): 任务状态

### 3.2 获取任务详情

**请求**:
```http
GET /api/v1/tasks/{id}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "task-1",
    "name": "审批任务",
    "description": "请审批此申请",
    "assignee": "user-1",
    "status": "pending",
    "process_instance_id": "inst-1",
    "activity_id": "UserTask_1",
    "create_time": "2024-01-01T12:00:00Z",
    "due_date": "2024-01-02T12:00:00Z",
    "priority": 50
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 3.3 认领任务

**请求**:
```http
POST /api/v1/tasks/{id}/claim
Content-Type: application/json

{
  "user_id": "user-1"
}
```

### 3.4 完成任务

**请求**:
```http
POST /api/v1/tasks/{id}/complete
Content-Type: application/json

{
  "variables": {
    "approved": true,
    "comment": "同意申请"
  },
  "comment": "审批完成"
}
```

### 3.5 委派任务

**请求**:
```http
POST /api/v1/tasks/{id}/delegate
Content-Type: application/json

{
  "delegate_id": "user-2",
  "comment": "委派给其他人处理"
}
```

## 4. 历史数据查询

### 4.1 查询历史流程实例列表

**请求**:
```http
GET /api/v1/history/process-instances?process_definition_key=sample-process
```

**查询参数**:
- `page` (可选): 页码
- `page_size` (可选): 每页大小
- `process_definition_key` (可选): 流程定义Key
- `business_key` (可选): 业务键
- `start_time_after` (可选): 开始时间之后 (ISO 8601格式)
- `start_time_before` (可选): 开始时间之前
- `end_time_after` (可选): 结束时间之后
- `end_time_before` (可选): 结束时间之前

### 4.2 获取历史流程实例详情

**请求**:
```http
GET /api/v1/history/process-instances/{id}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": "hist-1",
    "process_definition_id": "1",
    "process_definition_key": "sample-process",
    "business_key": "business-001",
    "start_time": "2024-01-01T12:00:00Z",
    "end_time": "2024-01-01T14:30:00Z",
    "duration": "2h30m",
    "status": "completed",
    "start_user_id": "user-1",
    "delete_reason": null
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 5. 健康检查

### 5.1 健康检查

**请求**:
```http
GET /health
```

**响应示例**:
```json
{
  "code": 200,
  "message": "成功",
  "data": {
    "status": "healthy",
    "service": "workflow-engine",
    "version": "1.0.0"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 5.2 就绪检查

**请求**:
```http
GET /ready
```

## 错误码说明

| 错误码 | 说明 | 示例 |
|--------|------|------|
| 200 | 成功 | 操作成功完成 |
| 400 | 请求参数错误 | 缺少必需参数或参数格式错误 |
| 401 | 未授权 | 需要认证或认证失败 |
| 403 | 禁止访问 | 权限不足 |
| 404 | 资源不存在 | 流程定义、实例或任务不存在 |
| 500 | 内部服务器错误 | 服务器处理异常 |
| 503 | 服务不可用 | 服务暂时不可用 |

## 使用示例

### 完整的流程执行示例

1. **创建流程定义**:
```bash
curl -X POST http://localhost:8080/api/v1/process-definitions \
  -H "Content-Type: application/json" \
  -d '{
    "key": "approval-process",
    "name": "审批流程",
    "resource": "{\"steps\": [...]}"
  }'
```

2. **启动流程实例**:
```bash
curl -X POST http://localhost:8080/api/v1/process-instances \
  -H "Content-Type: application/json" \
  -d '{
    "process_definition_key": "approval-process",
    "business_key": "req-001",
    "variables": {"amount": 5000}
  }'
```

3. **查询待处理任务**:
```bash
curl "http://localhost:8080/api/v1/tasks?assignee=user-1&status=pending"
```

4. **完成任务**:
```bash
curl -X POST http://localhost:8080/api/v1/tasks/task-1/complete \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {"approved": true},
    "comment": "审批通过"
  }'
```

5. **查询历史记录**:
```bash
curl "http://localhost:8080/api/v1/history/process-instances?business_key=req-001"
``` 