// Package data 数据库性能优化配置
// 提供数据库连接池、查询优化和缓存策略的配置管理
package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// DatabaseConfig 数据库优化配置
type DatabaseConfig struct {
	// 连接池配置
	MaxOpenConns    int           `yaml:"max_open_conns"`     // 最大打开连接数
	MaxIdleConns    int           `yaml:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`  // 连接最大生存时间
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"` // 连接最大空闲时间

	// 查询优化配置
	QueryTimeout       time.Duration `yaml:"query_timeout"`        // 查询超时时间
	SlowQueryTime      time.Duration `yaml:"slow_query_time"`      // 慢查询阈值
	EnablePreparedStmt bool          `yaml:"enable_prepared_stmt"` // 启用预编译语句

	// 事务配置
	TxTimeout        time.Duration      `yaml:"tx_timeout"`         // 事务超时时间
	TxIsolationLevel sql.IsolationLevel `yaml:"tx_isolation_level"` // 事务隔离级别
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// Redis配置
	Addr         string        `yaml:"addr"`           // Redis地址
	Password     string        `yaml:"password"`       // Redis密码
	DB           int           `yaml:"db"`             // Redis数据库编号
	PoolSize     int           `yaml:"pool_size"`      // 连接池大小
	MinIdleConns int           `yaml:"min_idle_conns"` // 最小空闲连接数
	DialTimeout  time.Duration `yaml:"dial_timeout"`   // 连接超时时间
	ReadTimeout  time.Duration `yaml:"read_timeout"`   // 读取超时时间
	WriteTimeout time.Duration `yaml:"write_timeout"`  // 写入超时时间
	IdleTimeout  time.Duration `yaml:"idle_timeout"`   // 空闲超时时间

	// 缓存策略配置
	DefaultExpiration time.Duration `yaml:"default_expiration"` // 默认过期时间
	CleanupInterval   time.Duration `yaml:"cleanup_interval"`   // 清理间隔
	EnableCompression bool          `yaml:"enable_compression"` // 启用压缩
	KeyPrefix         string        `yaml:"key_prefix"`         // 键前缀
}

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
	db     *sql.DB
	cache  *redis.Client
	logger *zap.Logger
	config *DatabaseConfig
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer(db *sql.DB, cache *redis.Client, logger *zap.Logger, config *DatabaseConfig) *QueryOptimizer {
	return &QueryOptimizer{
		db:     db,
		cache:  cache,
		logger: logger,
		config: config,
	}
}

// OptimizeDatabase 优化数据库配置
func (o *QueryOptimizer) OptimizeDatabase() error {
	o.logger.Info("开始优化数据库配置")

	// 设置连接池参数
	o.db.SetMaxOpenConns(o.config.MaxOpenConns)
	o.db.SetMaxIdleConns(o.config.MaxIdleConns)
	o.db.SetConnMaxLifetime(o.config.ConnMaxLifetime)
	o.db.SetConnMaxIdleTime(o.config.ConnMaxIdleTime)

	o.logger.Info("数据库连接池配置已优化",
		zap.Int("max_open_conns", o.config.MaxOpenConns),
		zap.Int("max_idle_conns", o.config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", o.config.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", o.config.ConnMaxIdleTime),
	)

	return nil
}

// ExecuteOptimizedQuery 执行优化查询
func (o *QueryOptimizer) ExecuteOptimizedQuery(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// 添加查询超时
	queryCtx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout)
	defer cancel()

	start := time.Now()
	rows, err := o.db.QueryContext(queryCtx, query, args...)
	duration := time.Since(start)

	// 记录慢查询
	if duration > o.config.SlowQueryTime {
		o.logger.Warn("检测到慢查询",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Any("args", args),
		)
	}

	// 记录查询指标
	o.logger.Debug("查询执行完成",
		zap.String("query", query),
		zap.Duration("duration", duration),
		zap.Error(err),
	)

	return rows, err
}

// ExecuteOptimizedQueryRow 执行优化单行查询
func (o *QueryOptimizer) ExecuteOptimizedQueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	// 添加查询超时
	queryCtx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout)
	defer cancel()

	start := time.Now()
	row := o.db.QueryRowContext(queryCtx, query, args...)
	duration := time.Since(start)

	// 记录慢查询
	if duration > o.config.SlowQueryTime {
		o.logger.Warn("检测到慢查询",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Any("args", args),
		)
	}

	// 记录查询指标
	o.logger.Debug("单行查询执行完成",
		zap.String("query", query),
		zap.Duration("duration", duration),
	)

	return row
}

// ExecuteOptimizedExec 执行优化更新操作
func (o *QueryOptimizer) ExecuteOptimizedExec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// 添加查询超时
	queryCtx, cancel := context.WithTimeout(ctx, o.config.QueryTimeout)
	defer cancel()

	start := time.Now()
	result, err := o.db.ExecContext(queryCtx, query, args...)
	duration := time.Since(start)

	// 记录慢查询
	if duration > o.config.SlowQueryTime {
		o.logger.Warn("检测到慢更新操作",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Any("args", args),
		)
	}

	// 记录更新指标
	if result != nil {
		rowsAffected, _ := result.RowsAffected()
		o.logger.Debug("更新操作执行完成",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Int64("rows_affected", rowsAffected),
			zap.Error(err),
		)
	}

	return result, err
}

// BeginOptimizedTx 开始优化事务
func (o *QueryOptimizer) BeginOptimizedTx(ctx context.Context) (*sql.Tx, error) {
	// 添加事务超时
	txCtx, cancel := context.WithTimeout(ctx, o.config.TxTimeout)
	defer cancel()

	opts := &sql.TxOptions{
		Isolation: o.config.TxIsolationLevel,
		ReadOnly:  false,
	}

	start := time.Now()
	tx, err := o.db.BeginTx(txCtx, opts)
	duration := time.Since(start)

	o.logger.Debug("事务开始",
		zap.Duration("duration", duration),
		zap.String("isolation_level", opts.Isolation.String()),
		zap.Error(err),
	)

	return tx, err
}

// GetConnectionStats 获取连接池统计信息
func (o *QueryOptimizer) GetConnectionStats() sql.DBStats {
	stats := o.db.Stats()

	o.logger.Info("数据库连接池统计",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	return stats
}

// 预定义的优化配置
var (
	// DefaultDatabaseConfig 默认数据库配置
	DefaultDatabaseConfig = &DatabaseConfig{
		MaxOpenConns:       50,
		MaxIdleConns:       10,
		ConnMaxLifetime:    5 * time.Minute,
		ConnMaxIdleTime:    1 * time.Minute,
		QueryTimeout:       30 * time.Second,
		SlowQueryTime:      1 * time.Second,
		EnablePreparedStmt: true,
		TxTimeout:          30 * time.Second,
		TxIsolationLevel:   sql.LevelReadCommitted,
	}

	// ProductionDatabaseConfig 生产环境数据库配置
	ProductionDatabaseConfig = &DatabaseConfig{
		MaxOpenConns:       100,
		MaxIdleConns:       25,
		ConnMaxLifetime:    10 * time.Minute,
		ConnMaxIdleTime:    2 * time.Minute,
		QueryTimeout:       15 * time.Second,
		SlowQueryTime:      500 * time.Millisecond,
		EnablePreparedStmt: true,
		TxTimeout:          30 * time.Second,
		TxIsolationLevel:   sql.LevelReadCommitted,
	}

	// DefaultCacheConfig 默认缓存配置
	DefaultCacheConfig = &CacheConfig{
		Addr:              "localhost:6379",
		Password:          "",
		DB:                0,
		PoolSize:          10,
		MinIdleConns:      3,
		DialTimeout:       5 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       5 * time.Minute,
		DefaultExpiration: 1 * time.Hour,
		CleanupInterval:   10 * time.Minute,
		EnableCompression: false,
		KeyPrefix:         "workflow:",
	}

	// ProductionCacheConfig 生产环境缓存配置
	ProductionCacheConfig = &CacheConfig{
		Addr:              "redis-service:6379",
		Password:          "",
		DB:                0,
		PoolSize:          20,
		MinIdleConns:      5,
		DialTimeout:       5 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       10 * time.Minute,
		DefaultExpiration: 30 * time.Minute,
		CleanupInterval:   5 * time.Minute,
		EnableCompression: true,
		KeyPrefix:         "workflow:",
	}
)
