# 多阶段构建 Dockerfile
# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
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

# 第二阶段：运行阶段
FROM alpine:3.18

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/workflow-engine .

# 复制配置文件
COPY --from=builder /app/configs/ ./configs/

# 复制数据库迁移文件
COPY --from=builder /app/migrations/ ./migrations/

# 复制工作流定义文件
COPY --from=builder /app/workflows/ ./workflows/

# 设置权限
RUN chown -R appuser:appgroup /app && \
    chmod +x ./workflow-engine

# 创建数据目录
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app/data

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 设置环境变量
ENV APP_ENV=production
ENV LOG_LEVEL=info
ENV HTTP_PORT=8080

# 启动命令
CMD ["./workflow-engine"] 