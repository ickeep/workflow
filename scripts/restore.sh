#!/bin/bash

# 工作流引擎数据恢复脚本
# 功能：从备份中恢复PostgreSQL数据库、Redis数据和配置文件
# 作者：工作流引擎开发团队
# 版本：v1.0.0

set -euo pipefail

# =============================================================================
# 配置参数
# =============================================================================

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 数据库配置
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-workflow_engine}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"

# Redis配置
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

# =============================================================================
# 辅助函数
# =============================================================================

# 日志函数
log_info() {
  echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
  echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 显示帮助信息
show_help() {
  cat <<EOF
工作流引擎数据恢复脚本 v1.0.0

用法: $0 <备份文件路径> [选项]

选项:
    -h, --help              显示此帮助信息
    --db-only              仅恢复数据库
    --redis-only           仅恢复Redis数据
    --config-only          仅恢复配置文件
    --force                强制恢复（不询问确认）
    --test                 测试模式（不执行实际恢复）

环境变量:
    POSTGRES_HOST          PostgreSQL主机
    POSTGRES_PORT          PostgreSQL端口
    POSTGRES_DB            数据库名称
    POSTGRES_USER          数据库用户
    POSTGRES_PASSWORD      数据库密码
    REDIS_HOST             Redis主机
    REDIS_PORT             Redis端口

示例:
    # 完整恢复
    $0 /path/to/backup.tar.gz

    # 仅恢复数据库
    $0 /path/to/backup.tar.gz --db-only

    # 强制恢复（不询问确认）
    $0 /path/to/backup.tar.gz --force

EOF
}

# 确认操作
confirm_operation() {
  local message="$1"

  if [[ "$FORCE" == "true" ]]; then
    log_info "强制模式：跳过确认"
    return 0
  fi

  echo -e "${YELLOW}警告: $message${NC}"
  read -p "确定要继续吗？(y/N): " -r
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "操作已取消"
    exit 0
  fi
}

# 检查备份文件
check_backup_file() {
  local backup_file="$1"

  log_info "检查备份文件: $backup_file"

  if [[ ! -f "$backup_file" ]]; then
    log_error "备份文件不存在: $backup_file"
    exit 1
  fi

  # 检查文件格式
  if [[ "$backup_file" == *.tar.gz ]]; then
    if ! tar -tzf "$backup_file" >/dev/null 2>&1; then
      log_error "备份文件格式无效或已损坏"
      exit 1
    fi
  else
    log_error "不支持的备份文件格式，请使用 .tar.gz 格式"
    exit 1
  fi

  log_info "备份文件检查通过"
}

# 解压备份文件
extract_backup() {
  local backup_file="$1"
  local extract_dir="/tmp/workflow-engine-restore-$(date +%s)"

  log_info "解压备份文件到: $extract_dir"

  if [[ "$TEST_MODE" == "false" ]]; then
    mkdir -p "$extract_dir"

    if tar -xzf "$backup_file" -C "$extract_dir"; then
      log_info "备份文件解压完成"

      # 查找解压后的目录
      local backup_dir=$(find "$extract_dir" -maxdepth 1 -type d | grep -v "^$extract_dir$" | head -1)
      if [[ -n "$backup_dir" ]]; then
        echo "$backup_dir"
      else
        log_error "解压后未找到备份目录"
        exit 1
      fi
    else
      log_error "备份文件解压失败"
      exit 1
    fi
  else
    log_info "[TEST] 模拟解压备份文件"
    echo "$extract_dir/test-backup"
  fi
}

# 恢复数据库
restore_database() {
  local backup_dir="$1"
  local db_backup_file="$backup_dir/database.sql"

  log_info "开始恢复PostgreSQL数据库..."

  if [[ ! -f "$db_backup_file" ]]; then
    log_error "数据库备份文件不存在: $db_backup_file"
    return 1
  fi

  confirm_operation "这将覆盖现有的数据库数据！"

  export PGPASSWORD="$POSTGRES_PASSWORD"

  if [[ "$TEST_MODE" == "false" ]]; then
    # 停止相关服务（如果在Docker环境中运行）
    if command -v docker-compose >/dev/null 2>&1; then
      log_info "停止工作流引擎服务..."
      docker-compose stop workflow-engine || true
    fi

    # 创建数据库备份（恢复前的安全措施）
    log_info "创建当前数据库的安全备份..."
    local safety_backup="/tmp/pre-restore-backup-$(date +%s).sql"
    pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" >"$safety_backup" || {
      log_warn "无法创建安全备份，继续恢复..."
    }

    # 恢复数据库
    if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <"$db_backup_file"; then
      log_info "PostgreSQL数据库恢复完成"

      # 清理安全备份
      rm -f "$safety_backup"
    else
      log_error "PostgreSQL数据库恢复失败"
      log_info "安全备份保存在: $safety_backup"
      return 1
    fi
  else
    log_info "[TEST] 模拟PostgreSQL数据库恢复"
  fi

  unset PGPASSWORD
}

# 恢复Redis数据
restore_redis() {
  local backup_dir="$1"
  local redis_backup_file="$backup_dir/redis_dump.rdb"

  log_info "开始恢复Redis数据..."

  if [[ ! -f "$redis_backup_file" ]]; then
    log_error "Redis备份文件不存在: $redis_backup_file"
    return 1
  fi

  confirm_operation "这将覆盖现有的Redis数据！"

  if [[ "$TEST_MODE" == "false" ]]; then
    # 停止Redis服务
    log_info "停止Redis服务..."
    if command -v systemctl >/dev/null 2>&1; then
      systemctl stop redis-server || true
    elif command -v service >/dev/null 2>&1; then
      service redis-server stop || true
    elif command -v docker-compose >/dev/null 2>&1; then
      docker-compose stop redis || true
    fi

    # 获取Redis数据目录
    local redis_data_dir
    if command -v redis-cli >/dev/null 2>&1; then
      redis_data_dir=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" CONFIG GET dir 2>/dev/null | tail -1 || echo "/var/lib/redis")
    else
      redis_data_dir="/var/lib/redis"
    fi

    # 备份当前Redis数据
    if [[ -f "$redis_data_dir/dump.rdb" ]]; then
      cp "$redis_data_dir/dump.rdb" "$redis_data_dir/dump.rdb.backup.$(date +%s)"
    fi

    # 恢复Redis数据
    if cp "$redis_backup_file" "$redis_data_dir/dump.rdb"; then
      log_info "Redis数据文件恢复完成"

      # 启动Redis服务
      log_info "启动Redis服务..."
      if command -v systemctl >/dev/null 2>&1; then
        systemctl start redis-server
      elif command -v service >/dev/null 2>&1; then
        service redis-server start
      elif command -v docker-compose >/dev/null 2>&1; then
        docker-compose start redis
      fi

      log_info "Redis数据恢复完成"
    else
      log_error "Redis数据恢复失败"
      return 1
    fi
  else
    log_info "[TEST] 模拟Redis数据恢复"
  fi
}

# 恢复配置文件
restore_configs() {
  local backup_dir="$1"
  local config_backup_dir="$backup_dir/configs"

  log_info "开始恢复配置文件..."

  if [[ ! -d "$config_backup_dir" ]]; then
    log_error "配置备份目录不存在: $config_backup_dir"
    return 1
  fi

  confirm_operation "这将覆盖现有的配置文件！"

  if [[ "$TEST_MODE" == "false" ]]; then
    # 备份当前配置
    if [[ -d "configs" ]]; then
      mv "configs" "configs.backup.$(date +%s)"
    fi

    # 恢复配置文件
    if cp -r "$config_backup_dir" "configs"; then
      log_info "配置文件恢复完成"
    else
      log_error "配置文件恢复失败"
      return 1
    fi

    # 恢复Docker配置文件
    for file in docker-compose.yaml Dockerfile; do
      local backup_file="$config_backup_dir/$file"
      if [[ -f "$backup_file" ]]; then
        cp "$backup_file" "./"
        log_info "恢复文件: $file"
      fi
    done
  else
    log_info "[TEST] 模拟配置文件恢复"
  fi
}

# 验证恢复结果
verify_restore() {
  log_info "验证恢复结果..."

  if [[ "$TEST_MODE" == "false" ]]; then
    # 检查数据库连接
    if [[ "$RESTORE_TYPE" == "all" || "$RESTORE_TYPE" == "db" ]]; then
      export PGPASSWORD="$POSTGRES_PASSWORD"
      if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1;" >/dev/null 2>&1; then
        log_info "数据库连接验证通过"
      else
        log_error "数据库连接验证失败"
      fi
      unset PGPASSWORD
    fi

    # 检查Redis连接
    if [[ "$RESTORE_TYPE" == "all" || "$RESTORE_TYPE" == "redis" ]]; then
      if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping >/dev/null 2>&1; then
        log_info "Redis连接验证通过"
      else
        log_error "Redis连接验证失败"
      fi
    fi

    # 检查配置文件
    if [[ "$RESTORE_TYPE" == "all" || "$RESTORE_TYPE" == "config" ]]; then
      if [[ -f "configs/config.yaml" ]]; then
        log_info "配置文件验证通过"
      else
        log_warn "主配置文件不存在"
      fi
    fi
  else
    log_info "[TEST] 模拟验证恢复结果"
  fi

  log_info "恢复验证完成"
}

# 清理临时文件
cleanup() {
  if [[ -n "${TEMP_DIR:-}" && -d "$TEMP_DIR" ]]; then
    log_info "清理临时文件: $TEMP_DIR"
    rm -rf "$TEMP_DIR"
  fi
}

# =============================================================================
# 主函数
# =============================================================================

main() {
  local backup_file="$1"

  log_info "开始工作流引擎数据恢复..."
  log_info "备份文件: $backup_file"
  log_info "恢复类型: $RESTORE_TYPE"

  # 设置清理函数
  trap cleanup EXIT

  # 检查备份文件
  check_backup_file "$backup_file"

  # 解压备份文件
  TEMP_DIR=$(extract_backup "$backup_file")

  # 执行恢复
  if [[ "$RESTORE_TYPE" == "all" || "$RESTORE_TYPE" == "db" ]]; then
    restore_database "$TEMP_DIR" || {
      log_error "数据库恢复失败"
      exit 1
    }
  fi

  if [[ "$RESTORE_TYPE" == "all" || "$RESTORE_TYPE" == "redis" ]]; then
    restore_redis "$TEMP_DIR" || {
      log_error "Redis恢复失败"
      exit 1
    }
  fi

  if [[ "$RESTORE_TYPE" == "all" || "$RESTORE_TYPE" == "config" ]]; then
    restore_configs "$TEMP_DIR" || {
      log_error "配置文件恢复失败"
      exit 1
    }
  fi

  # 验证恢复结果
  verify_restore

  log_info "数据恢复完成！"

  # 提醒用户重启服务
  if [[ "$TEST_MODE" == "false" ]]; then
    log_info "请重启工作流引擎服务以应用恢复的数据："
    log_info "  Docker Compose: docker-compose restart"
    log_info "  Kubernetes: kubectl rollout restart deployment/workflow-engine"
  fi
}

# =============================================================================
# 命令行参数解析
# =============================================================================

# 检查参数
if [[ $# -eq 0 ]]; then
  log_error "请指定备份文件路径"
  show_help
  exit 1
fi

# 默认值
RESTORE_TYPE="all"
FORCE="false"
TEST_MODE="false"

# 第一个参数是备份文件
BACKUP_FILE="$1"
shift

# 解析其他参数
while [[ $# -gt 0 ]]; do
  case $1 in
  -h | --help)
    show_help
    exit 0
    ;;
  --db-only)
    RESTORE_TYPE="db"
    shift
    ;;
  --redis-only)
    RESTORE_TYPE="redis"
    shift
    ;;
  --config-only)
    RESTORE_TYPE="config"
    shift
    ;;
  --force)
    FORCE="true"
    shift
    ;;
  --test)
    TEST_MODE="true"
    shift
    ;;
  *)
    log_error "未知参数: $1"
    show_help
    exit 1
    ;;
  esac
done

# 执行主函数
main "$BACKUP_FILE"
