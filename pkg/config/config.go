// Package config 提供配置管理功能
// 支持从 YAML 文件和环境变量加载配置
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用程序主配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`   // 服务器配置
	Data     DataConfig     `yaml:"data"`     // 数据存储配置
	Temporal TemporalConfig `yaml:"temporal"` // Temporal 配置
	Log      LogConfig      `yaml:"log"`      // 日志配置
	Auth     AuthConfig     `yaml:"auth"`     // 认证配置
	Engine   EngineConfig   `yaml:"engine"`   // 工作流引擎配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http"` // HTTP 服务器配置
	GRPC GRPCConfig `yaml:"grpc"` // gRPC 服务器配置
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Addr    string        `yaml:"addr"`    // 监听地址
	Timeout time.Duration `yaml:"timeout"` // 请求超时时间
}

// GRPCConfig gRPC 服务器配置
type GRPCConfig struct {
	Addr    string        `yaml:"addr"`    // 监听地址
	Timeout time.Duration `yaml:"timeout"` // 请求超时时间
}

// DataConfig 数据存储配置
type DataConfig struct {
	Database DatabaseConfig `yaml:"database"` // 数据库配置
	Redis    RedisConfig    `yaml:"redis"`    // Redis 配置
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`            // 数据库驱动
	Source          string        `yaml:"source"`            // 连接字符串
	MaxIdleConns    int           `yaml:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int           `yaml:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"` // 连接最大生命周期
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr         string        `yaml:"addr"`          // Redis 地址
	Password     string        `yaml:"password"`      // 密码
	DB           int           `yaml:"db"`            // 数据库编号
	ReadTimeout  time.Duration `yaml:"read_timeout"`  // 读取超时
	WriteTimeout time.Duration `yaml:"write_timeout"` // 写入超时
}

// TemporalConfig Temporal 配置
type TemporalConfig struct {
	HostPort  string        `yaml:"host_port"`  // Temporal 服务地址
	Namespace string        `yaml:"namespace"`  // 命名空间
	TaskQueue string        `yaml:"task_queue"` // 任务队列名称
	Workers   WorkersConfig `yaml:"workers"`    // Worker 配置
}

// WorkersConfig Worker 配置
type WorkersConfig struct {
	MaxConcurrentActivities    int `yaml:"max_concurrent_activities"`     // 最大并发活动数
	MaxConcurrentWorkflowTasks int `yaml:"max_concurrent_workflow_tasks"` // 最大并发工作流任务数
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`  // 日志级别
	Format string `yaml:"format"` // 日志格式
	Output string `yaml:"output"` // 输出目标
}

// AuthConfig 认证配置
type AuthConfig struct {
	Secret  string        `yaml:"secret"`  // JWT 密钥
	Expires time.Duration `yaml:"expires"` // Token 过期时间
}

// EngineConfig 工作流引擎配置
type EngineConfig struct {
	MaxConcurrentExecutions int           `yaml:"max_concurrent_executions"` // 最大并发执行数
	ExecutionTimeout        time.Duration `yaml:"execution_timeout"`         // 执行超时时间
	StepTimeout             time.Duration `yaml:"step_timeout"`              // 步骤超时时间
	Retry                   RetryConfig   `yaml:"retry"`                     // 重试配置
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts        int           `yaml:"max_attempts"`        // 最大重试次数
	InitialInterval    time.Duration `yaml:"initial_interval"`    // 初始重试间隔
	BackoffCoefficient float64       `yaml:"backoff_coefficient"` // 退避系数
	MaximumInterval    time.Duration `yaml:"maximum_interval"`    // 最大重试间隔
}

// Load 从指定文件加载配置
func Load(configFile string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML 配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 从环境变量覆盖配置
	if err := loadFromEnv(&config); err != nil {
		return nil, fmt.Errorf("从环境变量加载配置失败: %w", err)
	}

	// 验证配置
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) error {
	// 数据库配置
	if dbSource := os.Getenv("DATABASE_SOURCE"); dbSource != "" {
		config.Data.Database.Source = dbSource
	}

	// Redis 配置
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		config.Data.Redis.Addr = redisAddr
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Data.Redis.Password = redisPassword
	}

	// Temporal 配置
	if temporalHost := os.Getenv("TEMPORAL_HOST_PORT"); temporalHost != "" {
		config.Temporal.HostPort = temporalHost
	}
	if temporalNamespace := os.Getenv("TEMPORAL_NAMESPACE"); temporalNamespace != "" {
		config.Temporal.Namespace = temporalNamespace
	}

	// 认证配置
	if authSecret := os.Getenv("AUTH_SECRET"); authSecret != "" {
		config.Auth.Secret = authSecret
	}

	return nil
}

// validate 验证配置的有效性
func validate(config *Config) error {
	// 验证服务器配置
	if config.Server.HTTP.Addr == "" {
		return fmt.Errorf("HTTP 服务器地址不能为空")
	}
	if config.Server.GRPC.Addr == "" {
		return fmt.Errorf("gRPC 服务器地址不能为空")
	}

	// 验证数据库配置
	if config.Data.Database.Driver == "" {
		return fmt.Errorf("数据库驱动不能为空")
	}
	if config.Data.Database.Source == "" {
		return fmt.Errorf("数据库连接字符串不能为空")
	}

	// 验证 Temporal 配置
	if config.Temporal.HostPort == "" {
		return fmt.Errorf("Temporal 服务地址不能为空")
	}
	if config.Temporal.TaskQueue == "" {
		return fmt.Errorf("Temporal 任务队列名称不能为空")
	}

	// 验证认证配置
	if config.Auth.Secret == "" {
		return fmt.Errorf("认证密钥不能为空")
	}

	return nil
}
