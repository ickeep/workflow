-- Workflow Engine 数据库初始化脚本
-- 创建基础数据库和用户权限配置

-- 确保数据库使用 UTF8 编码
SET client_encoding = 'UTF8';

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- 创建时间戳函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建索引管理注释
COMMENT ON DATABASE workflow_engine IS 'Workflow Engine 流程引擎数据库';

-- 设置时区
SET timezone = 'UTC';

-- 预创建一些基础表结构会在 Ent 迁移时自动创建
-- 这里只做一些基础配置 