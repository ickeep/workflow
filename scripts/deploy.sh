#!/bin/bash

# 工作流引擎部署脚本
# 支持 Docker Compose 和 Kubernetes 两种部署方式

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
PROJECT_NAME="workflow-engine"
DOCKER_IMAGE="$PROJECT_NAME:latest"
NAMESPACE="workflow-engine"
DEPLOY_TYPE=${1:-"docker-compose"} # docker-compose 或 kubernetes

# 工具函数
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# 检查必要工具
check_prerequisites() {
  log_info "检查部署前置条件..."

  # 检查 Docker
  if ! command -v docker &>/dev/null; then
    log_error "Docker 未安装，请先安装 Docker"
    exit 1
  fi

  # 检查 Docker Compose
  if [[ "$DEPLOY_TYPE" == "docker-compose" ]]; then
    if ! command -v docker-compose &>/dev/null; then
      log_error "Docker Compose 未安装，请先安装 Docker Compose"
      exit 1
    fi
  fi

  # 检查 kubectl (如果是 Kubernetes 部署)
  if [[ "$DEPLOY_TYPE" == "kubernetes" ]]; then
    if ! command -v kubectl &>/dev/null; then
      log_error "kubectl 未安装，请先安装 kubectl"
      exit 1
    fi

    # 检查 Kubernetes 连接
    if ! kubectl cluster-info &>/dev/null; then
      log_error "无法连接到 Kubernetes 集群"
      exit 1
    fi
  fi

  log_success "前置条件检查通过"
}

# 构建 Docker 镜像
build_image() {
  log_info "构建 Docker 镜像..."

  # 确保我们在项目根目录
  cd "$(dirname "$0")/.."

  # 构建镜像
  docker build -t $DOCKER_IMAGE .

  if [ $? -eq 0 ]; then
    log_success "Docker 镜像构建成功: $DOCKER_IMAGE"
  else
    log_error "Docker 镜像构建失败"
    exit 1
  fi
}

# 运行测试
run_tests() {
  log_info "运行集成测试..."

  # 运行单元测试
  go test -v ./internal/... -short

  if [ $? -eq 0 ]; then
    log_success "单元测试通过"
  else
    log_warning "单元测试失败，继续部署..."
  fi

  # 运行集成测试
  go test -v ./tests/integration/... -run TestAPITestSuite/TestHealthCheck

  if [ $? -eq 0 ]; then
    log_success "集成测试通过"
  else
    log_warning "集成测试失败，继续部署..."
  fi
}

# Docker Compose 部署
deploy_docker_compose() {
  log_info "使用 Docker Compose 部署..."

  # 停止现有服务
  docker-compose down --remove-orphans

  # 启动核心服务
  log_info "启动核心服务 (postgres, redis, temporal)..."
  docker-compose up -d postgres redis temporal temporal-web

  # 等待服务启动
  log_info "等待依赖服务启动..."
  sleep 30

  # 检查服务健康状态
  check_docker_services() {
    local services=("postgres" "redis" "temporal")
    for service in "${services[@]}"; do
      if docker-compose ps $service | grep -q "healthy\|Up"; then
        log_success "$service 服务运行正常"
      else
        log_error "$service 服务启动失败"
        docker-compose logs $service
        return 1
      fi
    done
  }

  check_docker_services

  # 启动应用服务
  log_info "启动工作流引擎应用..."
  docker-compose up -d workflow-engine

  # 等待应用启动
  sleep 20

  # 健康检查
  local max_attempts=30
  local attempt=1

  while [ $attempt -le $max_attempts ]; do
    if curl -f http://localhost:8080/health &>/dev/null; then
      log_success "工作流引擎应用启动成功"
      break
    else
      log_info "等待应用启动... ($attempt/$max_attempts)"
      sleep 5
      ((attempt++))
    fi
  done

  if [ $attempt -gt $max_attempts ]; then
    log_error "应用启动超时"
    docker-compose logs workflow-engine
    exit 1
  fi

  # 显示服务状态
  log_info "部署完成，服务状态："
  docker-compose ps

  log_success "Docker Compose 部署成功！"
  log_info "访问地址："
  log_info "  - 工作流引擎 API: http://localhost:8080"
  log_info "  - Temporal Web UI: http://localhost:8088"
  log_info "  - PostgreSQL: localhost:5432"
  log_info "  - Redis: localhost:6379"
}

# Kubernetes 部署
deploy_kubernetes() {
  log_info "使用 Kubernetes 部署..."

  # 创建命名空间
  kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

  # 创建 PostgreSQL 初始化脚本 ConfigMap
  kubectl create configmap postgres-init-scripts \
    --from-file=migrations/ \
    --namespace=$NAMESPACE \
    --dry-run=client -o yaml | kubectl apply -f -

  # 应用 Kubernetes 配置
  kubectl apply -f deployments/kubernetes/workflow-engine-deployment.yaml

  # 等待部署完成
  log_info "等待部署完成..."

  # 等待 PostgreSQL 就绪
  kubectl wait --for=condition=available deployment/postgres --timeout=300s -n $NAMESPACE
  log_success "PostgreSQL 部署完成"

  # 等待 Redis 就绪
  kubectl wait --for=condition=available deployment/redis --timeout=300s -n $NAMESPACE
  log_success "Redis 部署完成"

  # 等待 Temporal 就绪
  kubectl wait --for=condition=available deployment/temporal --timeout=300s -n $NAMESPACE
  log_success "Temporal 部署完成"

  # 等待工作流引擎就绪
  kubectl wait --for=condition=available deployment/workflow-engine --timeout=300s -n $NAMESPACE
  log_success "工作流引擎部署完成"

  # 显示部署状态
  log_info "部署状态："
  kubectl get pods -n $NAMESPACE
  kubectl get services -n $NAMESPACE

  # 获取服务访问地址
  local app_port=$(kubectl get service workflow-engine-service -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "80")

  log_success "Kubernetes 部署成功！"
  log_info "访问地址："
  log_info "  - 工作流引擎 API: kubectl port-forward -n $NAMESPACE service/workflow-engine-service 8080:80"
  log_info "  - Temporal Web UI: kubectl port-forward -n $NAMESPACE service/temporal-web-service 8088:8088"
}

# 部署监控 (仅限 Docker Compose)
deploy_monitoring() {
  if [[ "$DEPLOY_TYPE" == "docker-compose" ]]; then
    log_info "部署监控组件..."
    docker-compose --profile monitoring up -d prometheus grafana jaeger

    log_info "监控服务访问地址："
    log_info "  - Prometheus: http://localhost:9090"
    log_info "  - Grafana: http://localhost:3000 (admin/admin)"
    log_info "  - Jaeger: http://localhost:16686"
  fi
}

# 部署后验证
post_deploy_verification() {
  log_info "执行部署后验证..."

  local health_url
  if [[ "$DEPLOY_TYPE" == "docker-compose" ]]; then
    health_url="http://localhost:8080/health"
  else
    # Kubernetes 需要端口转发
    kubectl port-forward -n $NAMESPACE service/workflow-engine-service 8080:80 &
    local port_forward_pid=$!
    sleep 5
    health_url="http://localhost:8080/health"
  fi

  # 健康检查
  if curl -f $health_url &>/dev/null; then
    log_success "健康检查通过"
  else
    log_error "健康检查失败"
    return 1
  fi

  # API 测试
  local api_url="${health_url%/health}/api/v1/process-definitions"
  if curl -f $api_url &>/dev/null; then
    log_success "API 接口测试通过"
  else
    log_error "API 接口测试失败"
    return 1
  fi

  # 清理端口转发 (仅限 Kubernetes)
  if [[ "$DEPLOY_TYPE" == "kubernetes" && -n "$port_forward_pid" ]]; then
    kill $port_forward_pid 2>/dev/null || true
  fi

  log_success "部署后验证完成"
}

# 清理函数
cleanup() {
  log_info "执行清理操作..."

  if [[ "$DEPLOY_TYPE" == "docker-compose" ]]; then
    docker-compose down --remove-orphans
    log_success "Docker Compose 服务已停止"
  else
    kubectl delete namespace $NAMESPACE --ignore-not-found=true
    log_success "Kubernetes 资源已清理"
  fi
}

# 显示帮助信息
show_help() {
  echo "工作流引擎部署脚本"
  echo ""
  echo "用法: $0 [COMMAND] [OPTIONS]"
  echo ""
  echo "命令:"
  echo "  docker-compose    使用 Docker Compose 部署 (默认)"
  echo "  kubernetes        使用 Kubernetes 部署"
  echo "  monitoring        部署监控组件 (仅限 Docker Compose)"
  echo "  cleanup           清理部署"
  echo "  test              运行测试"
  echo "  help              显示帮助信息"
  echo ""
  echo "环境变量:"
  echo "  SKIP_TESTS=true   跳过测试步骤"
  echo "  SKIP_BUILD=true   跳过镜像构建步骤"
  echo ""
  echo "示例:"
  echo "  $0 docker-compose    # Docker Compose 部署"
  echo "  $0 kubernetes        # Kubernetes 部署"
  echo "  $0 monitoring        # 部署监控"
  echo "  $0 cleanup           # 清理环境"
}

# 主函数
main() {
  case "$1" in
  "docker-compose" | "")
    DEPLOY_TYPE="docker-compose"
    check_prerequisites
    [[ "$SKIP_BUILD" != "true" ]] && build_image
    [[ "$SKIP_TESTS" != "true" ]] && run_tests
    deploy_docker_compose
    post_deploy_verification
    ;;
  "kubernetes")
    DEPLOY_TYPE="kubernetes"
    check_prerequisites
    [[ "$SKIP_BUILD" != "true" ]] && build_image
    [[ "$SKIP_TESTS" != "true" ]] && run_tests
    deploy_kubernetes
    post_deploy_verification
    ;;
  "monitoring")
    deploy_monitoring
    ;;
  "cleanup")
    cleanup
    ;;
  "test")
    run_tests
    ;;
  "help" | "-h" | "--help")
    show_help
    ;;
  *)
    log_error "未知命令: $1"
    show_help
    exit 1
    ;;
  esac
}

# 执行主函数
main "$@"
