-- 创建初始数据库模式
-- 工作流引擎核心表结构

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建枚举类型
CREATE TYPE process_definition_status AS ENUM ('draft', 'active', 'suspended', 'deprecated');
CREATE TYPE process_instance_status AS ENUM ('running', 'suspended', 'completed', 'terminated', 'failed');
CREATE TYPE task_status AS ENUM ('created', 'assigned', 'completed', 'canceled', 'failed');
CREATE TYPE event_type AS ENUM ('start', 'end', 'user_task', 'service_task', 'gateway', 'timer', 'message');

-- 1. 流程定义表
CREATE TABLE process_definitions (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    category VARCHAR(100),
    resource_name VARCHAR(255),
    resource_content TEXT,
    deployment_id VARCHAR(100),
    status process_definition_status NOT NULL DEFAULT 'draft',
    tenant_id VARCHAR(100) DEFAULT 'default',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    
    -- 索引
    CONSTRAINT uk_process_definition_key_version UNIQUE (key, version, tenant_id)
);

-- 2. 流程实例表
CREATE TABLE process_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    process_definition_id BIGINT NOT NULL REFERENCES process_definitions(id),
    process_definition_key VARCHAR(255) NOT NULL,
    process_definition_version INTEGER NOT NULL,
    business_key VARCHAR(255),
    status process_instance_status NOT NULL DEFAULT 'running',
    start_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_in_millis BIGINT,
    start_user_id VARCHAR(100),
    super_process_instance_id UUID,
    tenant_id VARCHAR(100) DEFAULT 'default',
    delete_reason VARCHAR(4000),
    
    -- 索引
    INDEX idx_process_instance_definition_id (process_definition_id),
    INDEX idx_process_instance_business_key (business_key),
    INDEX idx_process_instance_status (status),
    INDEX idx_process_instance_start_time (start_time),
    INDEX idx_process_instance_tenant (tenant_id)
);

-- 3. 任务实例表
CREATE TABLE task_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    task_definition_key VARCHAR(255),
    process_instance_id UUID REFERENCES process_instances(id),
    process_definition_id BIGINT REFERENCES process_definitions(id),
    assignee VARCHAR(100),
    candidate_users TEXT, -- JSON 数组
    candidate_groups TEXT, -- JSON 数组
    status task_status NOT NULL DEFAULT 'created',
    priority INTEGER DEFAULT 50,
    due_date TIMESTAMP WITH TIME ZONE,
    follow_up_date TIMESTAMP WITH TIME ZONE,
    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    claim_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_in_millis BIGINT,
    form_key VARCHAR(255),
    tenant_id VARCHAR(100) DEFAULT 'default',
    parent_task_id UUID,
    
    -- 索引
    INDEX idx_task_assignee (assignee),
    INDEX idx_task_process_instance (process_instance_id),
    INDEX idx_task_status (status),
    INDEX idx_task_create_time (create_time),
    INDEX idx_task_tenant (tenant_id)
);

-- 4. 流程变量表
CREATE TABLE process_variables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    process_instance_id UUID REFERENCES process_instances(id),
    task_id UUID REFERENCES task_instances(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- string, integer, boolean, json, etc.
    value_text TEXT,
    value_double DOUBLE PRECISION,
    value_long BIGINT,
    value_bytes BYTEA,
    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 索引
    INDEX idx_variable_process_instance (process_instance_id),
    INDEX idx_variable_task (task_id),
    INDEX idx_variable_name (name),
    CONSTRAINT uk_process_variable UNIQUE (process_instance_id, task_id, name)
);

-- 5. 历史流程实例表
CREATE TABLE historic_process_instances (
    id UUID PRIMARY KEY,
    process_definition_id BIGINT NOT NULL,
    process_definition_key VARCHAR(255) NOT NULL,
    process_definition_version INTEGER NOT NULL,
    process_definition_name VARCHAR(255),
    business_key VARCHAR(255),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_in_millis BIGINT,
    start_user_id VARCHAR(100),
    start_activity_id VARCHAR(255),
    end_activity_id VARCHAR(255),
    super_process_instance_id UUID,
    tenant_id VARCHAR(100) DEFAULT 'default',
    delete_reason VARCHAR(4000),
    
    -- 索引
    INDEX idx_hist_proc_inst_definition_id (process_definition_id),
    INDEX idx_hist_proc_inst_business_key (business_key),
    INDEX idx_hist_proc_inst_start_time (start_time),
    INDEX idx_hist_proc_inst_end_time (end_time),
    INDEX idx_hist_proc_inst_tenant (tenant_id)
);

-- 6. 历史任务实例表
CREATE TABLE historic_task_instances (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    task_definition_key VARCHAR(255),
    process_instance_id UUID,
    process_definition_id BIGINT,
    assignee VARCHAR(100),
    owner VARCHAR(100),
    priority INTEGER DEFAULT 50,
    due_date TIMESTAMP WITH TIME ZONE,
    create_time TIMESTAMP WITH TIME ZONE NOT NULL,
    claim_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_in_millis BIGINT,
    delete_reason VARCHAR(4000),
    form_key VARCHAR(255),
    tenant_id VARCHAR(100) DEFAULT 'default',
    parent_task_id UUID,
    
    -- 索引
    INDEX idx_hist_task_assignee (assignee),
    INDEX idx_hist_task_process_instance (process_instance_id),
    INDEX idx_hist_task_create_time (create_time),
    INDEX idx_hist_task_end_time (end_time),
    INDEX idx_hist_task_tenant (tenant_id)
);

-- 7. 事件日志表
CREATE TABLE event_logs (
    id BIGSERIAL PRIMARY KEY,
    process_instance_id UUID,
    task_id UUID,
    event_type event_type NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    event_data JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id VARCHAR(100),
    tenant_id VARCHAR(100) DEFAULT 'default',
    
    -- 索引
    INDEX idx_event_log_process_instance (process_instance_id),
    INDEX idx_event_log_task (task_id),
    INDEX idx_event_log_type (event_type),
    INDEX idx_event_log_timestamp (timestamp),
    INDEX idx_event_log_tenant (tenant_id)
);

-- 8. 部署表
CREATE TABLE deployments (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(255),
    deploy_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    source VARCHAR(255),
    tenant_id VARCHAR(100) DEFAULT 'default',
    
    -- 索引
    INDEX idx_deployment_tenant (tenant_id),
    INDEX idx_deployment_deploy_time (deploy_time)
);

-- 9. 资源表
CREATE TABLE deployment_resources (
    id BIGSERIAL PRIMARY KEY,
    deployment_id VARCHAR(100) NOT NULL REFERENCES deployments(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100),
    content BYTEA,
    generated BOOLEAN DEFAULT FALSE,
    
    -- 索引
    INDEX idx_resource_deployment (deployment_id),
    CONSTRAINT uk_resource_deployment_name UNIQUE (deployment_id, name)
);

-- 10. 作业表 (用于异步任务)
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(100) NOT NULL,
    process_instance_id UUID,
    task_id UUID,
    job_handler_type VARCHAR(255),
    job_handler_configuration TEXT,
    exclusive BOOLEAN DEFAULT TRUE,
    execution_id VARCHAR(100),
    retries INTEGER DEFAULT 3,
    exception_stack_trace TEXT,
    exception_message VARCHAR(4000),
    due_date TIMESTAMP WITH TIME ZONE,
    create_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    tenant_id VARCHAR(100) DEFAULT 'default',
    
    -- 索引
    INDEX idx_job_process_instance (process_instance_id),
    INDEX idx_job_due_date (due_date),
    INDEX idx_job_tenant (tenant_id),
    INDEX idx_job_retries (retries)
);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为 process_definitions 表创建更新时间触发器
CREATE TRIGGER update_process_definitions_updated_at
    BEFORE UPDATE ON process_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为 process_variables 表创建更新时间触发器
CREATE TRIGGER update_process_variables_update_time
    BEFORE UPDATE ON process_variables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建分区表 (用于历史数据)
-- 按月分区历史流程实例
SELECT create_range_partitions('historic_process_instances', 'start_time', 
    '2024-01-01'::date, interval '1 month', 12);

-- 按月分区历史任务实例
SELECT create_range_partitions('historic_task_instances', 'create_time', 
    '2024-01-01'::date, interval '1 month', 12);

-- 按月分区事件日志
SELECT create_range_partitions('event_logs', 'timestamp', 
    '2024-01-01'::date, interval '1 month', 12);

-- 插入初始数据
INSERT INTO process_definitions (key, name, description, category, status, created_by) VALUES
('sample-process', '示例流程', '这是一个示例流程定义', '示例', 'active', 'system'),
('approval-process', '审批流程', '通用审批流程模板', '审批', 'active', 'system'),
('notification-process', '通知流程', '消息通知流程模板', '通知', 'active', 'system');

-- 创建视图
CREATE VIEW v_active_tasks AS
SELECT 
    t.id,
    t.name,
    t.description,
    t.assignee,
    t.create_time,
    t.due_date,
    t.priority,
    pi.business_key,
    pd.name as process_name,
    pd.category
FROM task_instances t
JOIN process_instances pi ON t.process_instance_id = pi.id
JOIN process_definitions pd ON t.process_definition_id = pd.id
WHERE t.status IN ('created', 'assigned');

CREATE VIEW v_process_statistics AS
SELECT 
    pd.key,
    pd.name,
    pd.category,
    COUNT(pi.id) as total_instances,
    COUNT(CASE WHEN pi.status = 'running' THEN 1 END) as running_instances,
    COUNT(CASE WHEN pi.status = 'completed' THEN 1 END) as completed_instances,
    COUNT(CASE WHEN pi.status = 'failed' THEN 1 END) as failed_instances,
    AVG(CASE WHEN pi.duration_in_millis IS NOT NULL THEN pi.duration_in_millis END) as avg_duration_millis
FROM process_definitions pd
LEFT JOIN process_instances pi ON pd.id = pi.process_definition_id
GROUP BY pd.id, pd.key, pd.name, pd.category;

-- 创建性能优化索引
CREATE INDEX CONCURRENTLY idx_process_instances_compound 
ON process_instances (process_definition_key, status, start_time);

CREATE INDEX CONCURRENTLY idx_task_instances_compound 
ON task_instances (assignee, status, create_time);

CREATE INDEX CONCURRENTLY idx_process_variables_compound 
ON process_variables (process_instance_id, name, type);

-- 创建全文搜索索引
CREATE INDEX CONCURRENTLY idx_process_definitions_search 
ON process_definitions USING gin(to_tsvector('simple', name || ' ' || coalesce(description, '')));

CREATE INDEX CONCURRENTLY idx_task_instances_search 
ON task_instances USING gin(to_tsvector('simple', name || ' ' || coalesce(description, '')));

COMMENT ON TABLE process_definitions IS '流程定义表，存储工作流模板';
COMMENT ON TABLE process_instances IS '流程实例表，存储运行中的工作流实例';
COMMENT ON TABLE task_instances IS '任务实例表，存储用户任务和服务任务';
COMMENT ON TABLE process_variables IS '流程变量表，存储流程执行过程中的变量';
COMMENT ON TABLE historic_process_instances IS '历史流程实例表，存储已完成的流程实例';
COMMENT ON TABLE historic_task_instances IS '历史任务实例表，存储已完成的任务实例';
COMMENT ON TABLE event_logs IS '事件日志表，记录流程执行过程中的所有事件';
COMMENT ON TABLE deployments IS '部署表，管理流程定义的部署版本';
COMMENT ON TABLE deployment_resources IS '部署资源表，存储部署的资源文件';
COMMENT ON TABLE jobs IS '作业表，管理异步任务和定时任务'; 