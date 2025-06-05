package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad 测试配置加载功能
func TestLoad(t *testing.T) {
	t.Run("加载有效配置文件", func(t *testing.T) {
		// 创建临时配置文件
		configContent := `
server:
  http:
    addr: "0.0.0.0:8000"
    timeout: 1s
  grpc:
    addr: "0.0.0.0:9000"
    timeout: 1s

data:
  database:
    driver: postgres
    source: "postgres://test:test@localhost:5432/test?sslmode=disable"
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 3600s
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    read_timeout: 1s
    write_timeout: 1s

temporal:
  host_port: "localhost:7233"
  namespace: "default"
  task_queue: "workflow-engine-tasks"
  workers:
    max_concurrent_activities: 100
    max_concurrent_workflow_tasks: 100

log:
  level: info
  format: console
  output: stdout

auth:
  secret: "test-secret-key"
  expires: 24h

engine:
  max_concurrent_executions: 1000
  execution_timeout: 30m
  step_timeout: 5m
  retry:
    max_attempts: 3
    initial_interval: 1s
    backoff_coefficient: 2.0
    maximum_interval: 30s
`
		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err, "创建临时配置文件不应该失败")
		defer os.Remove(tmpFile.Name())

		// 写入配置内容
		_, err = tmpFile.WriteString(configContent)
		require.NoError(t, err, "写入配置内容不应该失败")
		tmpFile.Close()

		// 加载配置
		config, err := Load(tmpFile.Name())
		require.NoError(t, err, "加载配置不应该失败")
		require.NotNil(t, config, "配置对象不应该为空")

		// 验证服务器配置
		assert.Equal(t, "0.0.0.0:8000", config.Server.HTTP.Addr, "HTTP 地址应该匹配")
		assert.Equal(t, time.Second, config.Server.HTTP.Timeout, "HTTP 超时时间应该匹配")
		assert.Equal(t, "0.0.0.0:9000", config.Server.GRPC.Addr, "gRPC 地址应该匹配")

		// 验证数据库配置
		assert.Equal(t, "postgres", config.Data.Database.Driver, "数据库驱动应该匹配")
		assert.Contains(t, config.Data.Database.Source, "postgres://", "数据库连接字符串应该包含协议")
		assert.Equal(t, 10, config.Data.Database.MaxIdleConns, "最大空闲连接数应该匹配")

		// 验证 Temporal 配置
		assert.Equal(t, "localhost:7233", config.Temporal.HostPort, "Temporal 地址应该匹配")
		assert.Equal(t, "default", config.Temporal.Namespace, "Temporal 命名空间应该匹配")
		assert.Equal(t, "workflow-engine-tasks", config.Temporal.TaskQueue, "任务队列名称应该匹配")

		// 验证认证配置
		assert.Equal(t, "test-secret-key", config.Auth.Secret, "认证密钥应该匹配")
		assert.Equal(t, 24*time.Hour, config.Auth.Expires, "Token 过期时间应该匹配")

		// 验证引擎配置
		assert.Equal(t, 1000, config.Engine.MaxConcurrentExecutions, "最大并发执行数应该匹配")
		assert.Equal(t, 30*time.Minute, config.Engine.ExecutionTimeout, "执行超时时间应该匹配")
		assert.Equal(t, 3, config.Engine.Retry.MaxAttempts, "最大重试次数应该匹配")
	})

	t.Run("配置文件不存在", func(t *testing.T) {
		// 尝试加载不存在的配置文件
		config, err := Load("nonexistent.yaml")
		assert.Error(t, err, "加载不存在的配置文件应该返回错误")
		assert.Nil(t, config, "配置对象应该为空")
		assert.Contains(t, err.Error(), "读取配置文件失败", "错误信息应该包含读取失败")
	})

	t.Run("无效的YAML格式", func(t *testing.T) {
		// 创建无效的 YAML 文件
		invalidYAML := `
server:
  http:
    addr: "0.0.0.0:8000"
  invalid_yaml: [
`
		tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(invalidYAML)
		require.NoError(t, err)
		tmpFile.Close()

		// 尝试加载无效配置
		config, err := Load(tmpFile.Name())
		assert.Error(t, err, "加载无效 YAML 应该返回错误")
		assert.Nil(t, config, "配置对象应该为空")
		assert.Contains(t, err.Error(), "解析配置文件失败", "错误信息应该包含解析失败")
	})
}

// TestLoadFromEnv 测试从环境变量加载配置
func TestLoadFromEnv(t *testing.T) {
	t.Run("从环境变量覆盖配置", func(t *testing.T) {
		// 设置环境变量
		originalDBSource := os.Getenv("DATABASE_SOURCE")
		originalRedisAddr := os.Getenv("REDIS_ADDR")
		originalAuthSecret := os.Getenv("AUTH_SECRET")

		defer func() {
			// 恢复原始环境变量
			os.Setenv("DATABASE_SOURCE", originalDBSource)
			os.Setenv("REDIS_ADDR", originalRedisAddr)
			os.Setenv("AUTH_SECRET", originalAuthSecret)
		}()

		// 设置测试环境变量
		os.Setenv("DATABASE_SOURCE", "postgres://env:env@localhost:5432/env_db")
		os.Setenv("REDIS_ADDR", "env-redis:6379")
		os.Setenv("AUTH_SECRET", "env-secret")

		// 创建基础配置
		config := &Config{
			Data: DataConfig{
				Database: DatabaseConfig{
					Source: "original-db-source",
				},
				Redis: RedisConfig{
					Addr: "original-redis",
				},
			},
			Auth: AuthConfig{
				Secret: "original-secret",
			},
		}

		// 从环境变量加载
		err := loadFromEnv(config)
		assert.NoError(t, err, "从环境变量加载配置不应该失败")

		// 验证环境变量覆盖了原始配置
		assert.Equal(t, "postgres://env:env@localhost:5432/env_db", config.Data.Database.Source, "数据库连接应该被环境变量覆盖")
		assert.Equal(t, "env-redis:6379", config.Data.Redis.Addr, "Redis 地址应该被环境变量覆盖")
		assert.Equal(t, "env-secret", config.Auth.Secret, "认证密钥应该被环境变量覆盖")
	})
}

// TestValidate 测试配置验证功能
func TestValidate(t *testing.T) {
	t.Run("有效配置验证通过", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				HTTP: HTTPConfig{Addr: "0.0.0.0:8000"},
				GRPC: GRPCConfig{Addr: "0.0.0.0:9000"},
			},
			Data: DataConfig{
				Database: DatabaseConfig{
					Driver: "postgres",
					Source: "postgres://test:test@localhost:5432/test",
				},
			},
			Temporal: TemporalConfig{
				HostPort:  "localhost:7233",
				TaskQueue: "test-queue",
			},
			Auth: AuthConfig{
				Secret: "test-secret",
			},
		}

		err := validate(config)
		assert.NoError(t, err, "有效配置验证应该通过")
	})

	t.Run("缺少HTTP地址", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				HTTP: HTTPConfig{Addr: ""}, // 空地址
				GRPC: GRPCConfig{Addr: "0.0.0.0:9000"},
			},
			Data: DataConfig{
				Database: DatabaseConfig{
					Driver: "postgres",
					Source: "postgres://test:test@localhost:5432/test",
				},
			},
			Temporal: TemporalConfig{
				HostPort:  "localhost:7233",
				TaskQueue: "test-queue",
			},
			Auth: AuthConfig{
				Secret: "test-secret",
			},
		}

		err := validate(config)
		assert.Error(t, err, "缺少 HTTP 地址应该验证失败")
		assert.Contains(t, err.Error(), "HTTP 服务器地址不能为空", "错误信息应该包含 HTTP 地址")
	})

	t.Run("缺少数据库驱动", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				HTTP: HTTPConfig{Addr: "0.0.0.0:8000"},
				GRPC: GRPCConfig{Addr: "0.0.0.0:9000"},
			},
			Data: DataConfig{
				Database: DatabaseConfig{
					Driver: "", // 空驱动
					Source: "postgres://test:test@localhost:5432/test",
				},
			},
			Temporal: TemporalConfig{
				HostPort:  "localhost:7233",
				TaskQueue: "test-queue",
			},
			Auth: AuthConfig{
				Secret: "test-secret",
			},
		}

		err := validate(config)
		assert.Error(t, err, "缺少数据库驱动应该验证失败")
		assert.Contains(t, err.Error(), "数据库驱动不能为空", "错误信息应该包含数据库驱动")
	})

	t.Run("缺少认证密钥", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				HTTP: HTTPConfig{Addr: "0.0.0.0:8000"},
				GRPC: GRPCConfig{Addr: "0.0.0.0:9000"},
			},
			Data: DataConfig{
				Database: DatabaseConfig{
					Driver: "postgres",
					Source: "postgres://test:test@localhost:5432/test",
				},
			},
			Temporal: TemporalConfig{
				HostPort:  "localhost:7233",
				TaskQueue: "test-queue",
			},
			Auth: AuthConfig{
				Secret: "", // 空密钥
			},
		}

		err := validate(config)
		assert.Error(t, err, "缺少认证密钥应该验证失败")
		assert.Contains(t, err.Error(), "认证密钥不能为空", "错误信息应该包含认证密钥")
	})
}

// TestConfigStructure 测试配置结构的完整性
func TestConfigStructure(t *testing.T) {
	t.Run("配置结构字段完整性", func(t *testing.T) {
		config := &Config{}

		// 验证主要配置字段存在
		assert.NotNil(t, &config.Server, "Server 配置字段应该存在")
		assert.NotNil(t, &config.Data, "Data 配置字段应该存在")
		assert.NotNil(t, &config.Temporal, "Temporal 配置字段应该存在")
		assert.NotNil(t, &config.Log, "Log 配置字段应该存在")
		assert.NotNil(t, &config.Auth, "Auth 配置字段应该存在")
		assert.NotNil(t, &config.Engine, "Engine 配置字段应该存在")
	})

	t.Run("默认值设置", func(t *testing.T) {
		config := &Config{
			Engine: EngineConfig{
				Retry: RetryConfig{
					BackoffCoefficient: 2.0,
				},
			},
		}

		// 验证默认值
		assert.Equal(t, 2.0, config.Engine.Retry.BackoffCoefficient, "退避系数默认值应该正确")
	})
}
