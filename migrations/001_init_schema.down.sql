-- 回滚初始数据库模式
-- 按创建顺序的逆序删除

-- 删除视图
DROP VIEW IF EXISTS v_process_statistics;
DROP VIEW IF EXISTS v_active_tasks;

-- 删除触发器
DROP TRIGGER IF EXISTS update_process_variables_update_time ON process_variables;
DROP TRIGGER IF EXISTS update_process_definitions_updated_at ON process_definitions;

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除分区 (如果存在)
DROP TABLE IF EXISTS historic_process_instances_y2024m01 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m02 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m03 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m04 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m05 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m06 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m07 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m08 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m09 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m10 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m11 CASCADE;
DROP TABLE IF EXISTS historic_process_instances_y2024m12 CASCADE;

DROP TABLE IF EXISTS historic_task_instances_y2024m01 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m02 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m03 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m04 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m05 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m06 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m07 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m08 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m09 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m10 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m11 CASCADE;
DROP TABLE IF EXISTS historic_task_instances_y2024m12 CASCADE;

DROP TABLE IF EXISTS event_logs_y2024m01 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m02 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m03 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m04 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m05 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m06 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m07 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m08 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m09 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m10 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m11 CASCADE;
DROP TABLE IF EXISTS event_logs_y2024m12 CASCADE;

-- 删除表 (按依赖关系的逆序)
DROP TABLE IF EXISTS jobs CASCADE;
DROP TABLE IF EXISTS deployment_resources CASCADE;
DROP TABLE IF EXISTS deployments CASCADE;
DROP TABLE IF EXISTS event_logs CASCADE;
DROP TABLE IF EXISTS historic_task_instances CASCADE;
DROP TABLE IF EXISTS historic_process_instances CASCADE;
DROP TABLE IF EXISTS process_variables CASCADE;
DROP TABLE IF EXISTS task_instances CASCADE;
DROP TABLE IF EXISTS process_instances CASCADE;
DROP TABLE IF EXISTS process_definitions CASCADE;

-- 删除枚举类型
DROP TYPE IF EXISTS event_type;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS process_instance_status;
DROP TYPE IF EXISTS process_definition_status;

-- 删除扩展 (谨慎删除，因为可能被其他应用使用)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp"; 