# 工作流引擎 (Workflow Engine)

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()
[![Coverage](https://img.shields.io/badge/Coverage-96%25-brightgreen.svg)]()
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)]()
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-blue.svg)]()

一个基于 Go 语言开发的高性能、可扩展的工作流管理系统。采用微服务架构，支持复杂的业务流程编排和自动化。

## ✨ 核心特性

### 🚀 高性能
- **微服务架构**: 基于 Kratos v2 框架的高性能微服务
- **并发处理**: 支持高并发工作流执行
- **缓存优化**: Redis 多级缓存策略
- **数据库优化**: PostgreSQL 连接池和查询优化

### 🛡️ 可靠性
- **分布式**: 基于 Temporal 的分布式工作流引擎
- **高可用**: 支持集群部署和故障转移
- **数据一致性**: 强一致性保证和事务支持
- **备份恢复**: 完整的备份和灾难恢复方案

### 🔐 安全性
- **JWT 认证**: 企业级 JWT 认证系统
- **RBAC 权限**: 基于角色的访问控制
- **API 保护**: 全 API JWT 认证覆盖
- **数据加密**: 支持数据传输和存储加密

### 📊 监控运维
- **Prometheus 监控**: 完整的系统监控指标
- **Grafana 仪表板**: 实时监控可视化
- **日志管理**: 结构化日志和集中式日志收集
- **健康检查**: 多级健康检查机制

### 🔧 开发友好
- **RESTful API**: 标准化 API 设计
- **多语言 SDK**: Go、JavaScript、Python SDK 支持
- **完整文档**: 详细的用户手册和 API 文档
- **容器化**: Docker 和 Kubernetes 部署支持

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     工作流引擎架构图                          │
├─────────────────────────────────────────────────────────────┤
│  Frontend        │  API Gateway     │  Backend Services    │
│                  │                  │                      │
│  ┌─────────────┐ │ ┌─────────────┐  │ ┌─────────────────┐  │
│  │   Web UI    │ │ │   Gateway   │  │ │  Process Mgmt   │  │
│  │             │ │ │             │  │ │                 │  │
│  │ ┌─────────┐ │ │ │ ┌─────────┐ │  │ │ ┌─────────────┐ │  │
│  │ │Dashboard│ │ │ │ │  Auth   │ │  │ │ │ Definitions │ │  │
│  │ │         │ │ │ │ │         │ │  │ │ │             │ │  │
│  │ └─────────┘ │ │ │ └─────────┘ │  │ │ └─────────────┘ │  │
│  │             │ │ │             │  │ │                 │  │
│  │ ┌─────────┐ │ │ │ ┌─────────┐ │  │ │ ┌─────────────┐ │  │
│  │ │Designer │ │ │ │ │ Router  │ │  │ │ │  Instances  │ │  │
│  │ │         │ │ │ │ │         │ │  │ │ │             │ │  │
│  │ └─────────┘ │ │ │ └─────────┘ │  │ │ └─────────────┘ │  │
│  └─────────────┘ │ └─────────────┘  │ └─────────────────┘  │
│                  │                  │                      │
│                  │                  │ ┌─────────────────┐  │
│                  │                  │ │   Task Mgmt     │  │
│                  │                  │ │                 │  │
│                  │                  │ │ ┌─────────────┐ │  │
│                  │                  │ │ │   Executor  │ │  │
│                  │                  │ │ │             │ │  │
│                  │                  │ │ └─────────────┘ │  │
│                  │                  │ │                 │  │
│                  │                  │ │ ┌─────────────┐ │  │
│                  │                  │ │ │  Scheduler  │ │  │
│                  │                  │ │ │             │ │  │
│                  │                  │ │ └─────────────┘ │  │
│                  │                  │ └─────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                      Storage Layer                         │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ PostgreSQL   │  │    Redis     │  │     Temporal     │  │
│  │              │  │              │  │                  │  │
│  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────────┐ │  │
│  │ │Workflow  │ │  │ │  Cache   │ │  │ │   Workflow   │ │  │
│  │ │  Data    │ │  │ │          │ │  │ │    Engine    │ │  │
│  │ └──────────┘ │  │ └──────────┘ │  │ └──────────────┘ │  │
│  │              │  │              │  │                  │  │
│  │ ┌──────────┐ │  │ ┌──────────┐ │  │ ┌──────────────┐ │  │
│  │ │History   │ │  │ │Sessions  │ │  │ │   History    │ │  │
│  │ │  Data    │ │  │ │          │ │  │ │     Data     │ │  │
│  │ └──────────┘ │  │ └──────────┘ │  │ └──────────────┘ │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Monitoring & Ops                        │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Prometheus   │  │   Grafana    │  │      Jaeger      │  │
│  │   Metrics    │  │  Dashboard   │  │     Tracing      │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 快速开始

### 环境要求

- **Go**: 1.21 或更高版本
- **PostgreSQL**: 13+ (用于数据存储)
- **Redis**: 6+ (用于缓存)
- **Docker**: 20.10+ (可选，用于容器化部署)
- **Kubernetes**: 1.20+ (可选，用于 K8s 部署)

### 使用 Docker Compose 快速启动 (推荐)

```bash
# 1. 克隆项目
git clone https://github.com/workflow-engine/workflow-engine.git
cd workflow-engine

# 2. 启动完整环境
docker-compose up -d

# 3. 检查服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f workflow-engine
```

### 本地开发环境

```bash
# 1. 安装依赖
go mod download

# 2. 启动依赖服务
docker-compose up -d postgres redis temporal

# 3. 运行数据库迁移
make migrate-up

# 4. 启动应用
make run

# 或者直接运行
go run cmd/server/main.go
```

### 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# API 测试
curl -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"password123"}' \
     http://localhost:8080/api/v1/auth/login
```

## 📖 使用指南

### 基本概念

#### 流程定义 (Process Definition)
流程定义是工作流的模板，定义了业务流程的结构和步骤。

```json
{
  "name": "订单处理流程",
  "description": "电商订单处理的完整工作流程",
  "version": "1.0.0",
  "variables": {
    "order_id": "string",
    "customer_id": "string",
    "total_amount": "number"
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
```

#### 流程实例 (Process Instance)
流程实例是流程定义的具体执行，包含实际的业务数据。

```bash
# 启动流程实例
curl -X POST http://localhost:8080/api/v1/process-instances \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "process_definition_id": 123,
    "business_key": "ORDER-001",
    "variables": {
      "order_id": "ORDER-001",
      "customer_id": "CUST-456",
      "total_amount": 999.99
    }
  }'
```

#### 任务管理 (Task Management)
任务是流程中需要人工处理的步骤。

```bash
# 获取待办任务
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/tasks?assignee=current_user"

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

### 高级功能

#### 条件分支
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

#### 并行处理
```json
{
  "type": "parallel_gateway",
  "config": {
    "parallel_tasks": [
      "inventory_check",
      "credit_check",
      "fraud_check"
    ],
    "join_condition": "all_complete"
  }
}
```

#### 定时器事件
```json
{
  "type": "timer_event",
  "config": {
    "duration": "2h",
    "action": "escalate_to_manager"
  }
}
```

## 🔌 API 参考

### 认证
所有 API 请求都需要在 Header 中包含有效的 JWT Token：

```http
Authorization: Bearer <your-jwt-token>
```

### 主要端点

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/api/v1/auth/login` | 用户登录 |
| GET | `/api/v1/process-definitions` | 获取流程定义列表 |
| POST | `/api/v1/process-definitions` | 创建流程定义 |
| POST | `/api/v1/process-instances` | 启动流程实例 |
| GET | `/api/v1/process-instances` | 获取流程实例列表 |
| GET | `/api/v1/tasks` | 获取任务列表 |
| PUT | `/api/v1/tasks/{id}/complete` | 完成任务 |

完整的 API 文档请参考: [API 参考文档](docs/api-reference.md)

## 🛠️ 开发指南

### 项目结构

```
workflow-engine/
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器启动程序
├── internal/              # 内部应用代码
│   ├── auth/              # 认证授权模块
│   ├── biz/               # 业务逻辑层
│   ├── data/              # 数据访问层
│   ├── middleware/        # HTTP 中间件
│   └── service/           # 服务层
├── api/                   # API 定义 (protobuf)
├── configs/               # 配置文件
├── docs/                  # 文档
├── ent/                   # Ent 生成的代码
├── scripts/               # 脚本文件
├── tests/                 # 测试文件
│   ├── integration/       # 集成测试
│   └── performance/       # 性能测试
├── deployments/           # 部署配置
│   └── kubernetes/        # Kubernetes 配置
├── docker-compose.yaml    # Docker Compose 配置
└── Dockerfile            # Docker 镜像构建文件
```

### 本地开发

```bash
# 运行测试
make test

# 代码格式化
make fmt

# 静态分析
make lint

# 生成代码
make generate

# 构建项目
make build
```

### 添加新功能

1. **定义数据模型**: 在 `ent/schema/` 中定义数据结构
2. **实现业务逻辑**: 在 `internal/biz/` 中实现业务逻辑
3. **添加服务层**: 在 `internal/service/` 中添加服务接口
4. **创建 API**: 定义 RESTful API 端点
5. **编写测试**: 添加单元测试和集成测试

### 代码规范

- 遵循 Go 语言官方编码规范
- 使用 `gofmt` 格式化代码
- 注释使用中文，代码使用英文
- 单元测试覆盖率 > 80%
- 所有 public 函数都要有注释

## 🚀 部署指南

### Docker 部署

```bash
# 构建镜像
docker build -t workflow-engine:latest .

# 运行容器
docker run -d \
  --name workflow-engine \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/dbname" \
  -e REDIS_URL="redis://host:6379" \
  workflow-engine:latest
```

### Kubernetes 部署

```bash
# 应用配置
kubectl apply -f deployments/kubernetes/

# 检查状态
kubectl get pods -l app=workflow-engine

# 查看日志
kubectl logs -f deployment/workflow-engine
```

### 生产环境配置

#### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务端口 | `8080` |
| `DATABASE_URL` | 数据库连接字符串 | - |
| `REDIS_URL` | Redis 连接字符串 | - |
| `JWT_SECRET` | JWT 密钥 | - |
| `LOG_LEVEL` | 日志级别 | `info` |

#### 性能调优

```yaml
# configs/production.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  max_open_conns: 100
  max_idle_conns: 25
  conn_max_lifetime: 10m

cache:
  default_expiration: 30m
  cleanup_interval: 5m
```

## 📊 监控

### Prometheus 指标

系统提供丰富的监控指标：

- `workflow_instances_total`: 流程实例总数
- `workflow_instances_active`: 活跃流程实例数
- `task_completion_time`: 任务完成时间
- `api_request_duration`: API 请求延迟
- `database_connections`: 数据库连接数

### Grafana 仪表板

访问 http://localhost:3000 查看预配置的监控仪表板：

- **系统概览**: 整体系统运行状态
- **流程监控**: 流程执行情况和性能指标
- **API 监控**: HTTP 请求量和响应时间
- **基础设施**: 数据库、Redis、资源使用情况

### 告警配置

```yaml
# alerts.yaml
groups:
  - name: workflow-engine
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        annotations:
          summary: "工作流引擎错误率过高"
```

## 🔧 运维

### 备份

```bash
# 完整备份
./scripts/backup.sh

# 仅备份数据库
./scripts/backup.sh --db-only

# 自动备份 (crontab)
0 2 * * * /opt/workflow-engine/scripts/backup.sh
```

### 恢复

```bash
# 完整恢复
./scripts/restore.sh /path/to/backup.tar.gz

# 仅恢复数据库
./scripts/restore.sh /path/to/backup.tar.gz --db-only
```

### 日志管理

```bash
# 查看实时日志
docker-compose logs -f workflow-engine

# 查看错误日志
kubectl logs -l app=workflow-engine --tail=100 | grep ERROR
```

## 🤝 贡献指南

### 开发流程

1. **Fork 项目**: 在 GitHub 上 fork 本项目
2. **创建分支**: 创建功能分支 `git checkout -b feature/new-feature`
3. **开发功能**: 实现新功能并添加测试
4. **提交代码**: 提交代码 `git commit -am 'Add new feature'`
5. **推送分支**: 推送到 GitHub `git push origin feature/new-feature`
6. **创建 PR**: 在 GitHub 上创建 Pull Request

### 代码审查

- 所有代码必须通过 CI/CD 测试
- 至少需要一个维护者的代码审查
- 确保代码覆盖率不低于当前水平
- 更新相关文档

### 问题反馈

- 通过 [GitHub Issues](https://github.com/workflow-engine/workflow-engine/issues) 报告 bug
- 提供详细的问题描述和复现步骤
- 包含系统环境信息

## 📚 更多资源

### 文档

- [用户手册](docs/user-manual.md) - 详细的使用指南
- [API 参考](docs/api-reference.md) - 完整的 API 文档
- [开发指南](docs/development-guide.md) - 开发者指南
- [部署指南](docs/deployment-guide.md) - 生产部署指南

### 社区

- [GitHub Discussions](https://github.com/workflow-engine/workflow-engine/discussions) - 社区讨论
- [Stack Overflow](https://stackoverflow.com/questions/tagged/workflow-engine) - 技术问答
- [博客](https://blog.workflow-engine.dev) - 技术博客

### 支持

- **企业支持**: enterprise@workflow-engine.dev
- **技术支持**: support@workflow-engine.dev
- **安全问题**: security@workflow-engine.dev

## 📄 许可证

本项目使用 [MIT 许可证](LICENSE)。

## 🙏 致谢

感谢以下开源项目：

- [Kratos](https://github.com/go-kratos/kratos) - Go 微服务框架
- [Ent](https://entgo.io/) - Go 实体框架
- [Temporal](https://temporal.io/) - 工作流引擎
- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [PostgreSQL](https://www.postgresql.org/) - 数据库
- [Redis](https://redis.io/) - 缓存

---

**工作流引擎** - 让业务流程自动化变得简单 ✨

[GitHub](https://github.com/workflow-engine/workflow-engine) |
[文档](https://docs.workflow-engine.dev) |
[演示](https://demo.workflow-engine.dev) |
[博客](https://blog.workflow-engine.dev) 