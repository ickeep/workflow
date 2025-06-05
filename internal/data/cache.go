// Package data 缓存策略实现
// 提供Redis缓存、本地缓存和分布式缓存的统一接口
package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CacheManager 缓存管理器接口
type CacheManager interface {
	// 基础操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// 批量操作
	MGet(ctx context.Context, keys []string) ([]interface{}, error)
	MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error
	MDelete(ctx context.Context, keys []string) error

	// 高级操作
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 清理操作
	FlushDB(ctx context.Context) error
	FlushAll(ctx context.Context) error

	// 统计信息
	GetStats() CacheStats
}

// CacheStats 缓存统计信息
type CacheStats struct {
	HitCount    int64   `json:"hit_count"`    // 命中次数
	MissCount   int64   `json:"miss_count"`   // 未命中次数
	SetCount    int64   `json:"set_count"`    // 设置次数
	DeleteCount int64   `json:"delete_count"` // 删除次数
	HitRate     float64 `json:"hit_rate"`     // 命中率
	TotalKeys   int64   `json:"total_keys"`   // 总键数量
	MemoryUsage int64   `json:"memory_usage"` // 内存使用量
	Connections int     `json:"connections"`  // 连接数
}

// RedisCacheManager Redis缓存管理器
type RedisCacheManager struct {
	client *redis.Client
	config *CacheConfig
	logger *zap.Logger
	stats  *CacheStats
}

// NewRedisCacheManager 创建Redis缓存管理器
func NewRedisCacheManager(config *CacheConfig, logger *zap.Logger) *RedisCacheManager {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	})

	return &RedisCacheManager{
		client: rdb,
		config: config,
		logger: logger,
		stats:  &CacheStats{},
	}
}

// Get 获取缓存值
func (r *RedisCacheManager) Get(ctx context.Context, key string) (string, error) {
	fullKey := r.buildKey(key)

	start := time.Now()
	val, err := r.client.Get(ctx, fullKey).Result()
	duration := time.Since(start)

	if err == redis.Nil {
		r.stats.MissCount++
		r.logger.Debug("缓存未命中",
			zap.String("key", fullKey),
			zap.Duration("duration", duration),
		)
		return "", fmt.Errorf("缓存键不存在: %s", key)
	}

	if err != nil {
		r.logger.Error("获取缓存失败",
			zap.String("key", fullKey),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return "", err
	}

	r.stats.HitCount++
	r.logger.Debug("缓存命中",
		zap.String("key", fullKey),
		zap.Duration("duration", duration),
		zap.Int("value_length", len(val)),
	)

	return val, nil
}

// Set 设置缓存值
func (r *RedisCacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := r.buildKey(key)

	// 序列化值
	var serializedValue string
	switch v := value.(type) {
	case string:
		serializedValue = v
	case []byte:
		serializedValue = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			r.logger.Error("序列化缓存值失败",
				zap.String("key", fullKey),
				zap.Error(err),
			)
			return err
		}
		serializedValue = string(jsonData)
	}

	// 使用默认过期时间
	if expiration == 0 {
		expiration = r.config.DefaultExpiration
	}

	start := time.Now()
	err := r.client.Set(ctx, fullKey, serializedValue, expiration).Err()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("设置缓存失败",
			zap.String("key", fullKey),
			zap.Duration("duration", duration),
			zap.Duration("expiration", expiration),
			zap.Error(err),
		)
		return err
	}

	r.stats.SetCount++
	r.logger.Debug("缓存设置成功",
		zap.String("key", fullKey),
		zap.Duration("duration", duration),
		zap.Duration("expiration", expiration),
		zap.Int("value_length", len(serializedValue)),
	)

	return nil
}

// Delete 删除缓存
func (r *RedisCacheManager) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)

	start := time.Now()
	result, err := r.client.Del(ctx, fullKey).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("删除缓存失败",
			zap.String("key", fullKey),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	r.stats.DeleteCount++
	r.logger.Debug("缓存删除完成",
		zap.String("key", fullKey),
		zap.Duration("duration", duration),
		zap.Int64("deleted_count", result),
	)

	return nil
}

// Exists 检查缓存是否存在
func (r *RedisCacheManager) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)

	start := time.Now()
	count, err := r.client.Exists(ctx, fullKey).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("检查缓存存在性失败",
			zap.String("key", fullKey),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return false, err
	}

	exists := count > 0
	r.logger.Debug("缓存存在性检查",
		zap.String("key", fullKey),
		zap.Duration("duration", duration),
		zap.Bool("exists", exists),
	)

	return exists, nil
}

// MGet 批量获取缓存
func (r *RedisCacheManager) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	start := time.Now()
	results, err := r.client.MGet(ctx, fullKeys...).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("批量获取缓存失败",
			zap.Strings("keys", fullKeys),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	// 统计命中和未命中
	hitCount := 0
	for _, result := range results {
		if result != nil {
			hitCount++
		}
	}

	r.stats.HitCount += int64(hitCount)
	r.stats.MissCount += int64(len(keys) - hitCount)

	r.logger.Debug("批量获取缓存完成",
		zap.Strings("keys", fullKeys),
		zap.Duration("duration", duration),
		zap.Int("hit_count", hitCount),
		zap.Int("miss_count", len(keys)-hitCount),
	)

	return results, nil
}

// MSet 批量设置缓存
func (r *RedisCacheManager) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()

	// 使用默认过期时间
	if expiration == 0 {
		expiration = r.config.DefaultExpiration
	}

	start := time.Now()

	// 添加批量设置命令
	for key, value := range pairs {
		fullKey := r.buildKey(key)

		// 序列化值
		var serializedValue string
		switch v := value.(type) {
		case string:
			serializedValue = v
		case []byte:
			serializedValue = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				r.logger.Error("序列化批量缓存值失败",
					zap.String("key", fullKey),
					zap.Error(err),
				)
				return err
			}
			serializedValue = string(jsonData)
		}

		pipe.Set(ctx, fullKey, serializedValue, expiration)
	}

	// 执行批量命令
	_, err := pipe.Exec(ctx)
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("批量设置缓存失败",
			zap.Int("pair_count", len(pairs)),
			zap.Duration("duration", duration),
			zap.Duration("expiration", expiration),
			zap.Error(err),
		)
		return err
	}

	r.stats.SetCount += int64(len(pairs))
	r.logger.Debug("批量设置缓存成功",
		zap.Int("pair_count", len(pairs)),
		zap.Duration("duration", duration),
		zap.Duration("expiration", expiration),
	)

	return nil
}

// MDelete 批量删除缓存
func (r *RedisCacheManager) MDelete(ctx context.Context, keys []string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	start := time.Now()
	deletedCount, err := r.client.Del(ctx, fullKeys...).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("批量删除缓存失败",
			zap.Strings("keys", fullKeys),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	r.stats.DeleteCount += deletedCount
	r.logger.Debug("批量删除缓存完成",
		zap.Strings("keys", fullKeys),
		zap.Duration("duration", duration),
		zap.Int64("deleted_count", deletedCount),
	)

	return nil
}

// Increment 增加计数器
func (r *RedisCacheManager) Increment(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := r.buildKey(key)

	start := time.Now()
	result, err := r.client.IncrBy(ctx, fullKey, value).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("增加计数器失败",
			zap.String("key", fullKey),
			zap.Int64("value", value),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return 0, err
	}

	r.logger.Debug("计数器增加成功",
		zap.String("key", fullKey),
		zap.Int64("increment", value),
		zap.Int64("result", result),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// Expire 设置过期时间
func (r *RedisCacheManager) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := r.buildKey(key)

	start := time.Now()
	success, err := r.client.Expire(ctx, fullKey, expiration).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("设置过期时间失败",
			zap.String("key", fullKey),
			zap.Duration("expiration", expiration),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	r.logger.Debug("设置过期时间",
		zap.String("key", fullKey),
		zap.Duration("expiration", expiration),
		zap.Duration("duration", duration),
		zap.Bool("success", success),
	)

	return nil
}

// TTL 获取剩余生存时间
func (r *RedisCacheManager) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.buildKey(key)

	start := time.Now()
	ttl, err := r.client.TTL(ctx, fullKey).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("获取TTL失败",
			zap.String("key", fullKey),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return 0, err
	}

	r.logger.Debug("获取TTL",
		zap.String("key", fullKey),
		zap.Duration("ttl", ttl),
		zap.Duration("duration", duration),
	)

	return ttl, nil
}

// FlushDB 清空当前数据库
func (r *RedisCacheManager) FlushDB(ctx context.Context) error {
	start := time.Now()
	err := r.client.FlushDB(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("清空数据库失败",
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("数据库已清空",
		zap.Duration("duration", duration),
	)

	return nil
}

// FlushAll 清空所有数据库
func (r *RedisCacheManager) FlushAll(ctx context.Context) error {
	start := time.Now()
	err := r.client.FlushAll(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("清空所有数据库失败",
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("所有数据库已清空",
		zap.Duration("duration", duration),
	)

	return nil
}

// GetStats 获取缓存统计信息
func (r *RedisCacheManager) GetStats() CacheStats {
	// 计算命中率
	totalRequests := r.stats.HitCount + r.stats.MissCount
	if totalRequests > 0 {
		r.stats.HitRate = float64(r.stats.HitCount) / float64(totalRequests)
	}

	// 获取Redis信息
	ctx := context.Background()
	info, err := r.client.Info(ctx, "memory").Result()
	if err == nil {
		// 解析内存使用信息 (简化版本)
		r.logger.Debug("Redis内存信息", zap.String("info", info))
	}

	return *r.stats
}

// buildKey 构建完整的缓存键
func (r *RedisCacheManager) buildKey(key string) string {
	return r.config.KeyPrefix + key
}

// Close 关闭Redis连接
func (r *RedisCacheManager) Close() error {
	return r.client.Close()
}

// Ping 检查Redis连接状态
func (r *RedisCacheManager) Ping(ctx context.Context) error {
	start := time.Now()
	pong, err := r.client.Ping(ctx).Result()
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("Redis连接检查失败",
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	r.logger.Debug("Redis连接正常",
		zap.String("response", pong),
		zap.Duration("duration", duration),
	)

	return nil
}
