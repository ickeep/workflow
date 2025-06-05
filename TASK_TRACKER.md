# Workflow Engine 开发任务追踪表

## 任务执行状态

- ✅ 已完成
- 🚧 进行中  
- ⏸️ 暂停
- ❌ 失败
- ⏳ 待开始

## 阶段 1: 项目基础设施 (第1-2周)

| 任务ID | 任务描述 | 优先级 | 状态 | 负责模块 | 预计时间 | 完成时间 |
|--------|----------|--------|------|----------|----------|----------|
| T1.1 | 初始化 Kratos 项目结构 | P0 | ⏳ | 基础设施 | 4h | - |
| T1.2 | 配置开发环境 (Docker Compose) | P0 | ⏳ | 基础设施 | 6h | - |
| T1.3 | 设置数据库连接和 Ent Schema | P0 | ⏳ | 数据层 | 8h | - |
| T1.4 | 配置 Temporal 连接 | P0 | ⏳ | 工作流 | 6h | - |
| T1.5 | 实现基础中间件 | P1 | ⏳ | 服务层 | 12h | - |

### 任务详细说明

#### T1.1 初始化 Kratos 项目结构
```bash
# 执行命令
kratos new workflow-engine
cd workflow-engine

# 调整项目结构
mkdir -p internal/{biz,data,service}/{template,execution,workflow}
mkdir -p api/{template,execution,workflow}
mkdir -p workflows/
mkdir -p pkg/{temporal,config,utils}

# 验证条件
- [ ] 项目结构符合规划
- [ ] go.mod 文件正确
- [ ] 基础配置文件存在
```

#### T1.2 配置开发环境
```yaml
# docker-compose.yaml 配置内容
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: workflow_engine
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  temporal:
    image: temporalio/auto-setup:1.20
    environment:
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=postgres
      - POSTGRES_PWD=postgres
      - POSTGRES_SEEDS=postgres
    ports:
      - "7233:7233"
    depends_on:
      - postgres

  temporal-web:
    image: temporalio/web:1.20
    environment:
      - TEMPORAL_GRPC_ENDPOINT=temporal:7233
    ports:
      - "8088:8088"
    depends_on:
      - temporal

volumes:
  postgres_data:
```

## 阶段 2: 模板管理模块 (第3-4周)

| 任务ID | 任务描述 | 优先级 | 状态 | 负责模块 | 预计时间 | 完成时间 |
|--------|----------|--------|------|----------|----------|----------|
| T2.1 | 定义模板相关的 Protocol Buffers | P0 | ⏳ | API层 | 6h | - |
| T2.2 | 实现模板 Ent Schema | P0 | ⏳ | 数据层 | 8h | - |
| T2.3 | 开发模板 CRUD 业务逻辑 | P0 | ⏳ | 业务层 | 12h | - |
| T2.4 | 实现模板服务层 | P0 | ⏳ | 服务层 | 10h | - |
| T2.5 | 编写模板管理 API | P0 | ⏳ | 服务层 | 8h | - |
| T2.6 | 添加模板配置验证逻辑 | P1 | ⏳ | 业务层 | 6h | - |
| T2.7 | 编写单元测试 | P1 | ⏳ | 测试 | 12h | - |

### 任务详细说明

#### T2.1 定义模板相关的 Protocol Buffers
```protobuf
// api/template/v1/template.proto
syntax = "proto3";

package api.template.v1;

option go_package = "github.com/workflow-engine/api/template/v1;v1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

// 创建模板请求
message CreateTemplateRequest {
    string name = 1;        // 模板名称
    string description = 2; // 模板描述
    string version = 3;     // 版本号
    string config = 4;      // JSON配置
}

// 模板信息
message Template {
    int64 id = 1;
    string name = 2;
    string description = 3;
    string version = 4;
    string config = 5;
    string status = 6;
    google.protobuf.Timestamp created_at = 7;
    google.protobuf.Timestamp updated_at = 8;
}

// 模板服务
service TemplateService {
    rpc CreateTemplate(CreateTemplateRequest) returns (CreateTemplateResponse) {
        option (google.api.http) = {
            post: "/api/v1/templates"
            body: "*"
        };
    }
    
    rpc GetTemplate(GetTemplateRequest) returns (GetTemplateResponse) {
        option (google.api.http) = {
            get: "/api/v1/templates/{id}"
        };
    }
    
    rpc UpdateTemplate(UpdateTemplateRequest) returns (UpdateTemplateResponse) {
        option (google.api.http) = {
            put: "/api/v1/templates/{id}"
            body: "*"
        };
    }
    
    rpc DeleteTemplate(DeleteTemplateRequest) returns (DeleteTemplateResponse) {
        option (google.api.http) = {
            delete: "/api/v1/templates/{id}"
        };
    }
    
    rpc ListTemplates(ListTemplatesRequest) returns (ListTemplatesResponse) {
        option (google.api.http) = {
            get: "/api/v1/templates"
        };
    }
}
```

#### T2.2 实现模板 Ent Schema
```go
// internal/data/ent/schema/template.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "time"
)

// Template 工作流模板
type Template struct {
    ent.Schema
}

// Fields 字段定义
func (Template) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("id").
            Positive().
            Comment("模板ID"),
        field.String("name").
            NotEmpty().
            MaxLen(255).
            Comment("模板名称"),
        field.String("description").
            Optional().
            MaxLen(1000).
            Comment("模板描述"),
        field.String("version").
            NotEmpty().
            MaxLen(50).
            Comment("版本号"),
        field.Text("config").
            NotEmpty().
            Comment("JSON配置"),
        field.Enum("status").
            Values("active", "inactive").
            Default("active").
            Comment("状态"),
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

// Indexes 索引定义
func (Template) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("name", "version").Unique(),
        index.Fields("status"),
        index.Fields("created_at"),
    }
}
```

## 阶段 3: 工作流执行引擎 (第5-7周)

| 任务ID | 任务描述 | 优先级 | 状态 | 负责模块 | 预计时间 | 完成时间 |
|--------|----------|--------|------|----------|----------|----------|
| T3.1 | 定义 Temporal 工作流和活动 | P0 | ⏳ | 工作流 | 16h | - |
| T3.2 | 实现工作流配置解析器 | P0 | ⏳ | 公共包 | 12h | - |
| T3.3 | 开发工作流执行逻辑 | P0 | ⏳ | 业务层 | 20h | - |
| T3.4 | 实现工作流状态管理 | P0 | ⏳ | 业务层 | 16h | - |
| T3.5 | 集成 Temporal 客户端 | P0 | ⏳ | 公共包 | 12h | - |
| T3.6 | 实现工作流控制操作 | P0 | ⏳ | 服务层 | 14h | - |
| T3.7 | 编写集成测试 | P1 | ⏳ | 测试 | 16h | - |

## 阶段 4: 执行历史和监控 (第8-9周)

| 任务ID | 任务描述 | 优先级 | 状态 | 负责模块 | 预计时间 | 完成时间 |
|--------|----------|--------|------|----------|----------|----------|
| T4.1 | 实现执行历史记录 | P0 | ⏳ | 数据层 | 10h | - |
| T4.2 | 开发状态查询 API | P0 | ⏳ | 服务层 | 8h | - |
| T4.3 | 实现实时状态推送 | P1 | ⏳ | 服务层 | 12h | - |
| T4.4 | 添加执行统计功能 | P1 | ⏳ | 业务层 | 10h | - |
| T4.5 | 实现日志聚合 | P1 | ⏳ | 公共包 | 8h | - |
| T4.6 | 编写性能测试 | P1 | ⏳ | 测试 | 12h | - |

## 阶段 5: 优化和部署 (第10-11周)

| 任务ID | 任务描述 | 优先级 | 状态 | 负责模块 | 预计时间 | 完成时间 |
|--------|----------|--------|------|----------|----------|----------|
| T5.1 | 性能优化和缓存策略 | P1 | ⏳ | 优化 | 16h | - |
| T5.2 | 安全性增强 | P0 | ⏳ | 安全 | 12h | - |
| T5.3 | 编写部署文档 | P1 | ⏳ | 文档 | 8h | - |
| T5.4 | 配置 CI/CD 流水线 | P1 | ⏳ | 部署 | 10h | - |
| T5.5 | 生产环境部署 | P0 | ⏳ | 部署 | 12h | - |
| T5.6 | 压力测试和调优 | P1 | ⏳ | 测试 | 16h | - |

## 阶段 6: 文档和维护 (第12周)

| 任务ID | 任务描述 | 优先级 | 状态 | 负责模块 | 预计时间 | 完成时间 |
|--------|----------|--------|------|----------|----------|----------|
| T6.1 | 完善 API 文档 | P1 | ⏳ | 文档 | 12h | - |
| T6.2 | 编写用户使用手册 | P1 | ⏳ | 文档 | 16h | - |
| T6.3 | 代码重构和优化 | P2 | ⏳ | 优化 | 20h | - |
| T6.4 | 监控告警配置 | P1 | ⏳ | 运维 | 8h | - |
| T6.5 | 备份恢复方案 | P1 | ⏳ | 运维 | 10h | - |

## 关键里程碑

| 里程碑 | 描述 | 预计完成时间 | 验收标准 |
|--------|------|--------------|----------|
| M1 | 基础设施完成 | 第2周末 | 项目能够启动，数据库连通，Temporal 连接正常 |
| M2 | 模板管理完成 | 第4周末 | 模板 CRUD 功能完整，API 可用，测试通过 |
| M3 | 工作流引擎完成 | 第7周末 | 能够执行简单工作流，状态管理正常 |
| M4 | 监控系统完成 | 第9周末 | 历史记录功能完整，监控指标可用 |
| M5 | 生产就绪 | 第11周末 | 性能达标，安全验证通过，部署文档完整 |
| M6 | 项目交付 | 第12周末 | 文档完善，代码质量达标，监控告警配置完成 |

## 风险控制

| 风险项 | 风险等级 | 影响 | 缓解措施 |
|--------|----------|------|----------|
| Temporal 集成复杂度高 | 高 | 延期 2-3 周 | 提前调研，准备备选方案 |
| 性能不达标 | 中 | 延期 1-2 周 | 提前进行性能测试，优化关键路径 |
| 数据库设计变更 | 中 | 延期 1 周 | 仔细设计 Schema，使用迁移脚本 |
| 团队成员不熟悉技术栈 | 低 | 延期 1 周 | 提供培训文档，代码评审 |

## Background Agent 执行指南

### 自动化任务优先级
1. **P0 任务**: 必须完成，阻塞后续开发
2. **P1 任务**: 重要功能，影响用户体验
3. **P2 任务**: 优化改进，可以后续迭代

### 执行规则
1. 严格按照阶段顺序执行任务
2. 每个任务完成后更新状态表格
3. 遇到阻塞问题时及时记录和上报
4. 所有代码必须通过测试才能标记为完成
5. 遵循 `.cursorrules` 中定义的代码规范

### 质量检查点
- 代码提交前进行 lint 检查
- 单元测试覆盖率 > 80%
- 集成测试通过
- 性能指标达标
- 安全漏洞扫描通过

---

**更新日期**: {{ date }}  
**版本**: v1.0  
**维护者**: Workflow Engine Team 