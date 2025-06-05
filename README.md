# Workflow Engine

基于 Go 语言开发的工作流引擎，支持通过 JSON 配置定义和执行复杂的业务工作流。

## 技术栈

- **Go**: 1.21+
- **框架**: [Kratos v2](https://github.com/go-kratos/kratos) - Go 微服务框架
- **ORM**: [Ent](https://entgo.io/zh/docs/getting-started/) - Go 实体框架
- **工作流引擎**: [Temporal](https://docs.temporal.io/) - 分布式工作流编排
- **数据库**: PostgreSQL / MySQL
- **配置格式**: JSON / YAML

## 核心功能

- ✅ 工作流模板管理（创建、查询、编辑）
- ✅ 工作流执行引擎
- ✅ 实时状态监控
- ✅ 工作流控制操作（启动、暂停、恢复、终止）
- ✅ RESTful API 接口
- ✅ Web 管理界面

## 快速开始

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 13+

### 本地开发

```bash
# 克隆项目
git clone <repository-url>
cd workflow-engine

# 启动依赖服务
docker-compose up -d

# 安装依赖
go mod tidy

# 生成代码
make generate

# 运行服务
make run
```

## 项目结构

```
workflow-engine/
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器启动程序
├── internal/              # 内部应用代码
│   ├── biz/              # 业务逻辑层
│   ├── data/             # 数据访问层
│   ├── service/          # 服务层 (gRPC/HTTP)
│   └── server/           # 服务器配置
├── api/                   # API 定义 (protobuf)
│   ├── workflow/         # 工作流相关 API
│   └── template/         # 模板相关 API
├── configs/              # 配置文件
├── pkg/                  # 可重用的包
├── ent/                  # Ent 生成的代码
├── workflows/            # Temporal 工作流定义
├── migrations/           # 数据库迁移文件
├── docs/                 # 文档
└── scripts/              # 脚本文件
```

## API 文档

启动服务后，可通过以下地址访问：

- Swagger UI: http://localhost:8000/swagger/
- API 文档: http://localhost:8000/docs/

## 工作流配置示例

```json
{
  "name": "用户注册流程",
  "description": "处理用户注册的完整流程",
  "version": "1.0.0",
  "timeout": "300s",
  "steps": [
    {
      "id": "validate_user",
      "name": "验证用户信息",
      "type": "activity",
      "config": {
        "activity_name": "ValidateUserActivity",
        "timeout": "30s",
        "retry_policy": {
          "max_attempts": 3,
          "initial_interval": "1s"
        }
      },
      "next": ["send_email"]
    },
    {
      "id": "send_email",
      "name": "发送欢迎邮件",
      "type": "activity",
      "config": {
        "activity_name": "SendEmailActivity",
        "timeout": "10s"
      },
      "next": []
    }
  ],
  "variables": {
    "user_id": "string",
    "email": "string",
    "username": "string"
  }
}
```

## 参考文档

### 官方文档
- **Kratos 框架**: [https://github.com/go-kratos/kratos](https://github.com/go-kratos/kratos) - Go 微服务框架官方仓库
- **Ent ORM**: [https://entgo.io/zh/docs/getting-started/](https://entgo.io/zh/docs/getting-started/) - Ent 入门指南（中文版）
- **Temporal 工作流**: [https://docs.temporal.io/](https://docs.temporal.io/) - Temporal 官方文档
- **工作流引擎手册**: [https://workflow-engine-book.shuwoom.com/](https://workflow-engine-book.shuwoom.com/) - 工作流引擎实现参考

### 技术资源
- [Kratos 官方教程](https://go-kratos.dev/docs/)
- [Ent Schema 定义指南](https://entgo.io/zh/docs/schema-def/)
- [Temporal Go SDK](https://docs.temporal.io/docs/go/introduction)
- [Protocol Buffers 语法指南](https://developers.google.com/protocol-buffers/docs/proto3)

## 开发指南

请查看 [.cursorrules](.cursorrules) 文件了解详细的开发规范和指导。

## 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。 