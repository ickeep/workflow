#!/bin/bash

# 工作流引擎数据备份脚本
# 功能：自动备份PostgreSQL数据库、Redis数据和配置文件
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

# 默认配置
DEFAULT_BACKUP_DIR="/opt/backups/workflow-engine"
DEFAULT_RETENTION_DAYS=30

# 从环境变量获取配置
BACKUP_DIR="${BACKUP_DIR:-$DEFAULT_BACKUP_DIR}"
RETENTION_DAYS="${RETENTION_DAYS:-$DEFAULT_RETENTION_DAYS}"

# 数据库配置
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-workflow_engine}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"

# Redis配置
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

# 时间戳
TIMESTAMP=$(date '+%Y%m%d_%H%M%S')

# =============================================================================
# 辅助函数
# =============================================================================

# 日志函数
log_info() {
  echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 创建备份目录
create_backup_dir() {
  local backup_path="$BACKUP_DIR/$TIMESTAMP"
  log_info "创建备份目录: $backup_path"
  mkdir -p "$backup_path"
  chmod 700 "$backup_path"
  echo "$backup_path"
}

# 备份PostgreSQL数据库
backup_database() {
  local backup_path="$1"
  local db_backup_file="$backup_path/database.sql"

  log_info "开始备份PostgreSQL数据库..."

  export PGPASSWORD="$POSTGRES_PASSWORD"

  if pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" >"$db_backup_file"; then
    log_info "PostgreSQL数据库备份完成: $db_backup_file"
  else
    log_error "PostgreSQL数据库备份失败"
    return 1
  fi

  unset PGPASSWORD
}

# 备份Redis数据
backup_redis() {
  local backup_path="$1"
  local redis_backup_file="$backup_path/redis_dump.rdb"

  log_info "开始备份Redis数据..."

  if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" --rdb "$redis_backup_file"; then
    log_info "Redis数据备份完成: $redis_backup_file"
  else
    log_error "Redis备份失败"
    return 1
  fi
}

# 备份配置文件
backup_configs() {
  local backup_path="$1"
  local config_backup_dir="$backup_path/configs"

  log_info "开始备份配置文件..."

  mkdir -p "$config_backup_dir"

  # 备份主配置目录
  if [[ -d "configs" ]]; then
    cp -r configs/* "$config_backup_dir/"
    log_info "配置文件备份完成: $config_backup_dir"
  fi

  # 备份Docker配置
  for file in docker-compose.yaml Dockerfile; do
    if [[ -f "$file" ]]; then
      cp "$file" "$config_backup_dir/"
    fi
  done
}

# 创建备份归档
create_archive() {
  local backup_path="$1"
  local archive_file="$backup_path.tar.gz"

  log_info "创建备份归档..."

  if tar -czf "$archive_file" -C "$(dirname "$backup_path")" "$(basename "$backup_path")"; then
    log_info "备份归档创建完成: $archive_file"
    rm -rf "$backup_path"
    echo "$archive_file"
  else
    log_error "备份归档创建失败"
    return 1
  fi
}

# 清理旧备份
cleanup_old_backups() {
  log_info "清理 $RETENTION_DAYS 天前的旧备份..."

  if [[ -d "$BACKUP_DIR" ]]; then
    find "$BACKUP_DIR" -name "*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete
    log_info "旧备份清理完成"
  fi
}

# =============================================================================
# 主函数
# =============================================================================

main() {
  log_info "开始工作流引擎数据备份..."

  # 创建备份目录
  backup_path=$(create_backup_dir)

  # 执行备份
  backup_database "$backup_path" || exit 1
  backup_redis "$backup_path" || exit 1
  backup_configs "$backup_path"

  # 创建备份归档
  archive_file=$(create_archive "$backup_path") || exit 1

  # 清理旧备份
  cleanup_old_backups

  log_info "备份完成: $archive_file"
}

# 执行主函数
main "$@"
