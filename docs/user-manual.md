# 工作流引擎用户手册

## 概述

工作流引擎是一个基于 Go 语言开发的高性能、可扩展的工作流管理系统。本系统采用微服务架构，支持复杂的业务流程编排和自动化。

### 核心特性

- **可视化流程设计**: 支持通过 JSON 配置定义复杂的工作流程
- **高可用性**: 基于 Temporal 的分布式工作流引擎
- **强一致性**: 保证流程执行的原子性和一致性
- **可扩展性**: 支持水平扩展和高并发处理
- **监控告警**: 完整的监控体系和实时告警
- **权限控制**: 基于 JWT 的细粒度权限管理
- **API 优先**: RESTful API 设计，便于集成

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- PostgreSQL 13+ 
- Redis 6+
- Docker & Docker Compose (可选)

### 安装部署

#### 1. 使用 Docker Compose (推荐)

```bash
# 克隆项目
git clone https://github.com/yourorg/workflow-engine.git
cd workflow-engine

# 启动完整环境
docker-compose up -d

# 检查服务状态
docker-compose ps
```

#### 2. 本地开发环境

```bash
# 安装依赖
go mod download

# 启动数据库
docker-compose up -d postgres redis

# 运行数据库迁移
make migrate-up

# 启动服务
make run
```

### 验证安装

访问以下地址验证服务是否正常运行：

- **健康检查**: http://localhost:8080/health
- **API 文档**: http://localhost:8080/swagger/index.html
- **监控面板**: http://localhost:3000 (Grafana, admin/admin)

## 系统架构

### 技术栈

- **Web 框架**: Kratos v2
- **数据库**: PostgreSQL (主数据库) + Redis (缓存)
- **工作流引擎**: Temporal
- **认证授权**: JWT
- **监控**: Prometheus + Grafana
- **容器化**: Docker & Kubernetes

### 核心模块

1. **流程定义模块**: 管理工作流模板和版本
2. **流程运行时模块**: 控制流程实例的生命周期  
3. **任务管理模块**: 处理人工任务和系统任务
4. **历史数据模块**: 存储和查询历史执行记录
5. **用户权限模块**: 管理用户、角色和权限

## 流程定义

### JSON 配置格式

工作流通过 JSON 格式定义，基本结构如下：

```json
{
  "name": "订单处理流程",
  "description": "电商订单从创建到完成的完整流程",
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
        "timeout": "30s",
        "retry_policy": {
          "max_attempts": 3,
          "backoff": "exponential"
        }
      },
      "next": ["payment_check"]
    },
    {
      "id": "payment_check", 
      "name": "支付检查",
      "type": "user_task",
      "config": {
        "assignee": "finance_team",
        "form": {
          "payment_verified": "boolean",
          "notes": "string"
        }
      },
      "next": ["fulfill_order", "cancel_order"]
    }
  ]
}
```

### 步骤类型

#### 1. 服务任务 (Service Task)
自动执行的后台任务：

```json
{
  "type": "service_task",
  "config": {
    "service_name": "EmailService",
    "method": "sendWelcomeEmail",
    "input": {
      "email": "${user.email}",
      "name": "${user.name}"
    },
    "timeout": "10s"
  }
}
```

#### 2. 用户任务 (User Task)
需要人工处理的任务：

```json
{
  "type": "user_task", 
  "config": {
    "assignee": "manager",
    "candidate_groups": ["approval_team"],
    "form": {
      "approved": "boolean",
      "comments": "text"
    },
    "due_date": "2h"
  }
}
```

#### 3. 网关任务 (Gateway)
条件分支控制：

```json
{
  "type": "exclusive_gateway",
  "config": {
    "conditions": [
      {
        "expression": "${amount > 1000}",
        "next": "manager_approval"
      },
      {
        "expression": "${amount <= 1000}",
        "next": "auto_approve"
      }
    ]
  }
}
```

## API 使用指南

### 认证

所有 API 请求都需要在 Header 中包含有效的 JWT Token：

```http
Authorization: Bearer <your-jwt-token>
```

### 获取访问令牌

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password123"
  }'
```

响应示例：
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

### 流程定义管理

#### 创建流程定义

```bash
curl -X POST http://localhost:8080/api/v1/process-definitions \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试流程",
    "description": "这是一个测试流程",
    "version": "1.0.0",
    "config": { ... }
  }'
```

#### 查询流程定义

```bash
# 获取所有流程定义
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/process-definitions?page=1&size=10"

# 获取特定流程定义
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/process-definitions/123"
```

### 流程实例管理

#### 启动流程实例

```bash
curl -X POST http://localhost:8080/api/v1/process-instances \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "process_definition_id": 123,
    "business_key": "ORDER-001",
    "variables": {
      "order_id": "ORDER-001",
      "amount": 999.99
    }
  }'
```

#### 查询流程实例

```bash
# 获取所有流程实例
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/process-instances?status=running"

# 获取特定流程实例详情
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/process-instances/456"
```

#### 控制流程实例

```bash
# 暂停流程实例
curl -X PUT http://localhost:8080/api/v1/process-instances/456/suspend \
  -H "Authorization: Bearer <token>"

# 恢复流程实例  
curl -X PUT http://localhost:8080/api/v1/process-instances/456/activate \
  -H "Authorization: Bearer <token>"

# 终止流程实例
curl -X DELETE http://localhost:8080/api/v1/process-instances/456 \
  -H "Authorization: Bearer <token>"
```

### 任务管理

#### 查询待办任务

```bash
# 获取当前用户的待办任务
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/tasks?assignee=current_user"

# 获取候选任务
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/tasks?candidate_user=current_user"
```

#### 处理任务

```bash
# 认领任务
curl -X PUT http://localhost:8080/api/v1/tasks/789/claim \
  -H "Authorization: Bearer <token>"

# 完成任务
curl -X PUT http://localhost:8080/api/v1/tasks/789/complete \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "approved": true,
      "comments": "审批通过"
    }
  }'
```

## 监控与运维

### 系统监控

#### Prometheus 指标

系统提供丰富的 Prometheus 指标：

```bash
# 查看所有指标
curl http://localhost:8080/metrics

# 主要指标类型：
# - workflow_instances_total: 流程实例总数
# - workflow_instances_active: 活跃流程实例数  
# - task_completion_time: 任务完成时间
# - api_request_duration: API 请求延迟
# - database_connections: 数据库连接数
```

#### Grafana 仪表板

访问 http://localhost:3000 查看预配置的监控仪表板：

- **系统概览**: 整体系统运行状态
- **流程监控**: 流程执行情况和性能指标
- **API 监控**: HTTP 请求量和响应时间
- **基础设施**: 数据库、Redis、资源使用情况

### 日志管理

#### 日志级别

系统支持以下日志级别：
- `DEBUG`: 详细调试信息
- `INFO`: 一般信息记录  
- `WARN`: 警告信息
- `ERROR`: 错误信息

#### 配置日志级别

```yaml
# configs/config.yaml
log:
  level: info
  format: json
  output: stdout
```

#### 查看日志

```bash
# Docker 环境查看日志
docker-compose logs -f workflow-engine

# Kubernetes 环境查看日志  
kubectl logs -f deployment/workflow-engine
```

### 备份与恢复

#### 数据库备份

```bash
# 备份 PostgreSQL 数据库
docker exec postgres-container pg_dump -U postgres workflow_engine > backup.sql

# 恢复数据库
docker exec -i postgres-container psql -U postgres workflow_engine < backup.sql
```

#### Redis 备份

```bash
# 备份 Redis 数据
docker exec redis-container redis-cli BGSAVE

# 复制备份文件
docker cp redis-container:/data/dump.rdb ./redis-backup.rdb
```

## 故障排查

### 常见问题

#### 1. 服务启动失败

**症状**: 服务无法启动，日志显示连接错误

**排查步骤**:
```bash
# 检查依赖服务状态
docker-compose ps

# 检查网络连接
telnet localhost 5432  # PostgreSQL
telnet localhost 6379  # Redis

# 检查配置文件
cat configs/config.yaml
```

#### 2. 流程执行卡住

**症状**: 流程实例状态长时间未更新

**排查步骤**:
```bash
# 查看流程实例详情
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/process-instances/456"

# 检查 Temporal 工作流状态
curl http://localhost:7233/namespaces/default/workflows/workflow-id

# 查看相关任务
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/tasks?process_instance_id=456"
```

#### 3. API 响应慢

**症状**: API 请求响应时间过长

**排查步骤**:
```bash
# 检查系统资源使用
docker stats

# 查看数据库连接
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/health/database"

# 分析慢查询日志
docker logs postgres-container | grep "slow query"
```

### 性能优化

#### 数据库优化

1. **索引优化**:
   ```sql
   -- 为常用查询字段添加索引
   CREATE INDEX idx_process_instances_status ON process_instances(status);
   CREATE INDEX idx_tasks_assignee ON tasks(assignee);
   ```

2. **连接池调优**:
   ```yaml
   database:
     max_open_conns: 100
     max_idle_conns: 25
     conn_max_lifetime: 10m
   ```

#### 缓存优化

1. **Redis 配置**:
   ```yaml
   cache:
     default_expiration: 30m
     cleanup_interval: 5m
     enable_compression: true
   ```

2. **缓存策略**:
   - 流程定义: 长期缓存
   - 用户信息: 中期缓存
   - 实时数据: 短期缓存

## 安全指南

### 认证与授权

#### JWT 配置

```yaml
jwt:
  secret_key: "your-super-secret-key"
  expiration_time: 24h
  refresh_time: 7d
  enable_blacklist: true
```

#### 权限系统

系统采用基于角色的访问控制 (RBAC)：

- **角色**: admin, manager, user
- **权限**: process:read, process:write, task:assign, etc.

#### API 安全

1. **HTTPS**: 生产环境必须启用 HTTPS
2. **速率限制**: 防止 API 滥用
3. **输入验证**: 严格验证所有输入参数
4. **SQL 注入防护**: 使用参数化查询

### 网络安全

#### 防火墙配置

```bash
# 仅开放必要端口
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS  
ufw allow 22/tcp    # SSH (限制来源IP)
```

#### 容器安全

```dockerfile
# 使用非 root 用户运行
USER 1001:1001

# 只读文件系统
RUN mount -o remount,ro /
```

## 扩展开发

### 自定义任务类型

实现自定义任务类型的步骤：

1. **定义任务接口**:
   ```go
   type CustomTask interface {
       Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
   }
   ```

2. **注册任务类型**:
   ```go
   taskRegistry.Register("custom_email", &EmailTask{})
   ```

3. **配置任务**:
   ```json
   {
     "type": "custom_email",
     "config": {
       "template": "welcome",
       "recipients": ["user@example.com"]
     }
   }
   ```

### 集成第三方服务

#### Webhook 集成

```go
type WebhookTask struct {
    URL     string
    Method  string
    Headers map[string]string
}

func (w *WebhookTask) Execute(ctx context.Context, input map[string]interface{}) error {
    // 发送 HTTP 请求到第三方服务
    return sendHTTPRequest(w.URL, w.Method, input)
}
```

#### 消息队列集成

```go
type MessageQueueTask struct {
    QueueName string
    Message   interface{}
}

func (m *MessageQueueTask) Execute(ctx context.Context, input map[string]interface{}) error {
    // 发送消息到队列
    return messageQueue.Send(m.QueueName, input)
}
```

## 最佳实践

### 流程设计原则

1. **单一职责**: 每个步骤只负责一个明确的业务功能
2. **幂等性**: 步骤可以安全地重复执行
3. **超时设置**: 为所有步骤设置合理的超时时间
4. **错误处理**: 明确定义错误处理和重试策略

### 性能考虑

1. **批量操作**: 使用批量 API 减少网络开销
2. **异步处理**: 长时间运行的任务使用异步模式
3. **缓存策略**: 合理使用缓存减少数据库压力
4. **分页查询**: 大数据量查询必须使用分页

### 运维建议

1. **监控告警**: 设置关键指标的监控告警
2. **日志分析**: 定期分析日志发现潜在问题
3. **容量规划**: 根据业务增长合理规划资源
4. **灾备方案**: 制定完整的灾难恢复计划

## 社区支持

### 获取帮助

- **GitHub Issues**: https://github.com/yourorg/workflow-engine/issues
- **文档站点**: https://workflow-engine-docs.example.com
- **邮件列表**: workflow-engine@example.com

### 贡献代码

1. Fork 项目仓库
2. 创建功能分支
3. 提交 Pull Request
4. 通过代码审查

### 许可证

本项目使用 MIT 许可证，详情请查看 [LICENSE](../LICENSE) 文件。

---

**版本**: v1.0.0  
**更新时间**: 2024年12月  
**文档维护**: 工作流引擎开发团队 