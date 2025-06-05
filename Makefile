.PHONY: init generate build run test clean docker-up docker-down migrate

# 变量定义
APP_NAME := workflow-engine
DOCKER_COMPOSE := docker-compose
GO_FILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./ent/*")

# 初始化项目
init:
	@echo "初始化项目..."
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install entgo.io/ent/cmd/ent@latest

# 生成代码
generate:
	@echo "生成代码..."
	@if [ -d "api" ]; then \
		find api -name "*.proto" -exec protoc --proto_path=. \
			--proto_path=./third_party \
			--go_out=paths=source_relative:. \
			--go-http_out=paths=source_relative:. \
			--go-grpc_out=paths=source_relative:. \
			--go-errors_out=paths=source_relative:. {} \; ; \
	fi
	@if [ -f "ent/generate.go" ]; then \
		go generate ./ent; \
	fi

# 构建应用
build:
	@echo "构建应用..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) cmd/server/main.go

# 运行应用
run:
	@echo "运行应用..."
	go run cmd/server/main.go -conf configs/config.yaml

# 运行测试
test:
	@echo "运行测试..."
	go test -v -race -cover ./...

# 运行基准测试
bench:
	@echo "运行基准测试..."
	go test -bench=. -benchmem ./...

# 代码格式化
fmt:
	@echo "格式化代码..."
	gofmt -s -w $(GO_FILES)
	go mod tidy

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run

# 清理
clean:
	@echo "清理..."
	rm -rf bin/
	go clean -cache

# 启动依赖服务
docker-up:
	@echo "启动依赖服务..."
	$(DOCKER_COMPOSE) up -d

# 停止依赖服务
docker-down:
	@echo "停止依赖服务..."
	$(DOCKER_COMPOSE) down

# 数据库迁移
migrate-up:
	@echo "执行数据库迁移..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "回滚数据库迁移..."
	go run cmd/migrate/main.go down

# 创建新的迁移文件
migrate-create:
	@read -p "输入迁移文件名: " name; \
	go run cmd/migrate/main.go create $$name

# 开发环境一键启动
dev: docker-up generate
	@echo "开发环境启动完成"

# 生产构建
prod-build:
	@echo "生产环境构建..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/$(APP_NAME) cmd/server/main.go

# Docker 构建
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t $(APP_NAME):latest .

# 帮助
help:
	@echo "可用的命令:"
	@echo "  init          - 初始化项目和安装工具"
	@echo "  generate      - 生成 protobuf 和 ent 代码"
	@echo "  build         - 构建应用程序"
	@echo "  run           - 运行应用程序"
	@echo "  test          - 运行测试"
	@echo "  bench         - 运行基准测试"
	@echo "  fmt           - 格式化代码"
	@echo "  lint          - 代码检查"
	@echo "  clean         - 清理构建文件"
	@echo "  docker-up     - 启动依赖服务"
	@echo "  docker-down   - 停止依赖服务"
	@echo "  migrate-up    - 执行数据库迁移"
	@echo "  migrate-down  - 回滚数据库迁移"
	@echo "  migrate-create- 创建新的迁移文件"
	@echo "  dev           - 开发环境一键启动"
	@echo "  prod-build    - 生产环境构建"
	@echo "  docker-build  - 构建 Docker 镜像"
	@echo "  help          - 显示帮助信息" 