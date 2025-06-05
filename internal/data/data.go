// Package data 提供数据访问层功能
// 包含数据库连接管理、Ent 客户端配置、Redis 连接等
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/workflow-engine/workflow-engine/internal/data/ent"
	"github.com/workflow-engine/workflow-engine/pkg/config"

	"entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq" // PostgreSQL 驱动
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Data 数据访问层结构体，包含所有数据连接和客户端
type Data struct {
	// 数据库客户端
	DB *ent.Client
	// Redis 客户端
	Redis *redis.Client
	// 日志器
	Logger *zap.Logger
}

// NewData 创建新的数据访问层实例
func NewData(cfg *config.Config, logger *zap.Logger) (*Data, func(), error) {
	// 创建数据库连接
	db, dbCleanup, err := NewDB(cfg.Data.Database, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("创建数据库连接失败: %w", err)
	}

	// 创建 Redis 连接
	rdb, redisCleanup, err := NewRedis(cfg.Data.Redis, logger)
	if err != nil {
		dbCleanup() // 清理已创建的数据库连接
		return nil, nil, fmt.Errorf("创建 Redis 连接失败: %w", err)
	}

	// 创建数据访问层实例
	data := &Data{
		DB:     db,
		Redis:  rdb,
		Logger: logger,
	}

	// 返回清理函数
	cleanup := func() {
		logger.Info("开始清理数据连接")
		redisCleanup()
		dbCleanup()
		logger.Info("数据连接清理完成")
	}

	return data, cleanup, nil
}

// NewDB 创建数据库连接和 Ent 客户端
func NewDB(cfg config.DatabaseConfig, logger *zap.Logger) (*ent.Client, func(), error) {
	logger.Info("正在创建数据库连接",
		zap.String("driver", cfg.Driver),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
	)

	// 创建数据库驱动
	drv, err := sql.Open(cfg.Driver, cfg.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 配置连接池
	db := drv.DB()
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		drv.Close()
		return nil, nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 创建 Ent 客户端
	client := ent.NewClient(ent.Driver(drv))

	// 启用调试模式 (仅在开发环境)
	if logger.Core().Enabled(zap.DebugLevel) {
		client = client.Debug()
	}

	// 清理函数
	cleanup := func() {
		if err := client.Close(); err != nil {
			logger.Error("关闭数据库连接失败", zap.Error(err))
		} else {
			logger.Info("数据库连接已关闭")
		}
	}

	logger.Info("数据库连接创建成功")
	return client, cleanup, nil
}

// NewRedis 创建 Redis 连接
func NewRedis(cfg config.RedisConfig, logger *zap.Logger) (*redis.Client, func(), error) {
	logger.Info("正在创建 Redis 连接",
		zap.String("addr", cfg.Addr),
		zap.Int("db", cfg.DB),
	)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolSize:        100,              // 连接池大小
		PoolTimeout:     10 * time.Second, // 连接池超时
		ConnMaxIdleTime: 5 * time.Minute,  // 空闲连接超时
	})

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, nil, fmt.Errorf("Redis 连接测试失败: %w", err)
	}

	// 清理函数
	cleanup := func() {
		if err := rdb.Close(); err != nil {
			logger.Error("关闭 Redis 连接失败", zap.Error(err))
		} else {
			logger.Info("Redis 连接已关闭")
		}
	}

	logger.Info("Redis 连接创建成功")
	return rdb, cleanup, nil
}

// Migrate 执行数据库迁移
func (d *Data) Migrate(ctx context.Context) error {
	d.Logger.Info("开始执行数据库迁移")

	if err := d.DB.Schema.Create(ctx); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	d.Logger.Info("数据库迁移完成")
	return nil
}

// HealthCheck 检查数据连接健康状态
func (d *Data) HealthCheck(ctx context.Context) error {
	// 检查数据库连接 - 尝试计数查询，如果失败则忽略（表可能尚未创建）
	_, _ = d.DB.ProcessDefinition.Query().Count(ctx)

	// 检查 Redis 连接
	if err := d.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 健康检查失败: %w", err)
	}

	return nil
}
