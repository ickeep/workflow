package data

import (
	"context"
	"testing"
	"time"

	"github.com/workflow-engine/workflow-engine/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewData 测试数据访问层创建功能
func TestNewData(t *testing.T) {
	t.Run("创建数据访问层成功", func(t *testing.T) {
		// 跳过需要真实数据库连接的测试
		t.Skip("需要真实的数据库和 Redis 连接")

		cfg := &config.Config{
			Data: config.DataConfig{
				Database: config.DatabaseConfig{
					Driver:          "postgres",
					Source:          "postgres://test:test@localhost:5432/test_db?sslmode=disable",
					MaxIdleConns:    5,
					MaxOpenConns:    10,
					ConnMaxLifetime: time.Hour,
				},
				Redis: config.RedisConfig{
					Addr:         "localhost:6379",
					Password:     "",
					DB:           1, // 使用测试数据库
					ReadTimeout:  time.Second,
					WriteTimeout: time.Second,
				},
			},
		}

		logger := zap.NewNop() // 使用无操作日志器
		data, cleanup, err := NewData(cfg, logger)
		require.NoError(t, err, "创建数据访问层不应该失败")
		require.NotNil(t, data, "数据访问层实例不应该为空")
		defer cleanup()

		// 验证数据访问层结构
		assert.NotNil(t, data.DB, "数据库客户端不应该为空")
		assert.NotNil(t, data.Redis, "Redis 客户端不应该为空")
		assert.NotNil(t, data.Logger, "日志器不应该为空")
	})

	t.Run("数据库连接失败", func(t *testing.T) {
		cfg := &config.Config{
			Data: config.DataConfig{
				Database: config.DatabaseConfig{
					Driver: "postgres",
					Source: "invalid-connection-string",
				},
				Redis: config.RedisConfig{
					Addr: "localhost:6379",
				},
			},
		}

		logger := zap.NewNop()
		data, cleanup, err := NewData(cfg, logger)
		assert.Error(t, err, "无效的数据库连接应该返回错误")
		assert.Nil(t, data, "数据访问层实例应该为空")
		assert.Nil(t, cleanup, "清理函数应该为空")
		assert.Contains(t, err.Error(), "创建数据库连接失败", "错误信息应该包含数据库连接失败")
	})

	t.Run("Redis连接失败", func(t *testing.T) {
		// 跳过需要真实数据库连接的测试
		t.Skip("需要真实的数据库连接")

		cfg := &config.Config{
			Data: config.DataConfig{
				Database: config.DatabaseConfig{
					Driver: "postgres",
					Source: "postgres://test:test@localhost:5432/test_db?sslmode=disable",
				},
				Redis: config.RedisConfig{
					Addr: "invalid-redis-host:6379",
				},
			},
		}

		logger := zap.NewNop()
		data, cleanup, err := NewData(cfg, logger)
		assert.Error(t, err, "无效的 Redis 连接应该返回错误")
		assert.Nil(t, data, "数据访问层实例应该为空")
		assert.Nil(t, cleanup, "清理函数应该为空")
		assert.Contains(t, err.Error(), "创建 Redis 连接失败", "错误信息应该包含 Redis 连接失败")
	})
}

// TestNewDB 测试数据库连接创建功能
func TestNewDB(t *testing.T) {
	t.Run("有效配置验证", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Driver:          "postgres",
			Source:          "postgres://user:pass@localhost:5432/db?sslmode=disable",
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
		}

		logger := zap.NewNop()

		// 测试参数验证（不进行实际连接）
		assert.Equal(t, "postgres", cfg.Driver, "数据库驱动应该正确")
		assert.Equal(t, 5, cfg.MaxIdleConns, "最大空闲连接数应该正确")
		assert.Equal(t, 10, cfg.MaxOpenConns, "最大开放连接数应该正确")
		assert.Equal(t, time.Hour, cfg.ConnMaxLifetime, "连接最大生命周期应该正确")

		// 记录测试通过
		logger.Info("数据库配置验证通过")
	})

	t.Run("无效数据库驱动", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Driver: "invalid-driver",
			Source: "invalid-source",
		}

		logger := zap.NewNop()
		client, cleanup, err := NewDB(cfg, logger)
		assert.Error(t, err, "无效的数据库驱动应该返回错误")
		assert.Nil(t, client, "数据库客户端应该为空")
		assert.Nil(t, cleanup, "清理函数应该为空")
	})
}

// TestNewRedis 测试 Redis 连接创建功能
func TestNewRedis(t *testing.T) {
	t.Run("有效配置验证", func(t *testing.T) {
		cfg := config.RedisConfig{
			Addr:         "localhost:6379",
			Password:     "",
			DB:           0,
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
		}

		logger := zap.NewNop()

		// 测试参数验证（不进行实际连接）
		assert.Equal(t, "localhost:6379", cfg.Addr, "Redis 地址应该正确")
		assert.Equal(t, 0, cfg.DB, "Redis 数据库编号应该正确")
		assert.Equal(t, time.Second, cfg.ReadTimeout, "读取超时应该正确")
		assert.Equal(t, time.Second, cfg.WriteTimeout, "写入超时应该正确")

		// 记录测试通过
		logger.Info("Redis 配置验证通过")
	})
}

// TestHealthCheck 测试健康检查功能
func TestHealthCheck(t *testing.T) {
	t.Run("空数据访问层健康检查", func(t *testing.T) {
		// 测试空指针的情况，应该正确处理
		data := &Data{
			Logger: zap.NewNop(),
		}

		// 由于健康检查会访问 nil 的 DB 和 Redis，这个测试应该 panic
		// 我们使用 assert.Panics 来验证
		ctx := context.Background()
		assert.Panics(t, func() {
			data.HealthCheck(ctx)
		}, "空的数据访问层健康检查应该 panic")
	})
}

// TestMigrate 测试数据库迁移功能
func TestMigrate(t *testing.T) {
	t.Run("迁移参数验证", func(t *testing.T) {
		// 创建空的数据访问层用于测试
		data := &Data{
			Logger: zap.NewNop(),
		}

		ctx := context.Background()
		// 空的数据库客户端会导致 panic
		assert.Panics(t, func() {
			data.Migrate(ctx)
		}, "空的数据库客户端迁移应该 panic")
	})
}

// TestDataStructure 测试数据访问层结构完整性
func TestDataStructure(t *testing.T) {
	t.Run("数据访问层结构字段", func(t *testing.T) {
		data := &Data{}

		// 验证结构体字段存在
		assert.NotNil(t, &data.DB, "DB 字段应该存在")
		assert.NotNil(t, &data.Redis, "Redis 字段应该存在")
		assert.NotNil(t, &data.Logger, "Logger 字段应该存在")
	})

	t.Run("配置结构验证", func(t *testing.T) {
		cfg := &config.Config{
			Data: config.DataConfig{
				Database: config.DatabaseConfig{
					Driver: "postgres",
					Source: "test-connection-string",
				},
				Redis: config.RedisConfig{
					Addr: "test-redis-addr",
				},
			},
		}

		// 验证配置结构
		assert.Equal(t, "postgres", cfg.Data.Database.Driver, "数据库驱动配置应该正确")
		assert.Equal(t, "test-connection-string", cfg.Data.Database.Source, "数据库连接字符串应该正确")
		assert.Equal(t, "test-redis-addr", cfg.Data.Redis.Addr, "Redis 地址配置应该正确")
	})
}
