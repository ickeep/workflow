# Workflow Engine Dockerfile
# 使用多阶段构建优化镜像大小

# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o workflow-engine \
    ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S workflow && \
    adduser -u 1001 -S workflow -G workflow

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/workflow-engine .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p /app/logs && \
    chown -R workflow:workflow /app

# 切换到非 root 用户
USER workflow

# 暴露端口
EXPOSE 8000 9000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# 设置环境变量
ENV CONFIG_FILE=/app/configs/config.yaml

# 启动应用程序
CMD ["./workflow-engine", "-config", "/app/configs/config.yaml"] 