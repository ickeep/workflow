// Package repository 提供数据仓储层实现
// 包含缓存相关的数据访问逻辑
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/workflow-engine/workflow-engine/internal/biz"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// cacheRepo 缓存仓储实现
type cacheRepo struct {
	client *redis.Client
	logger *zap.Logger
}

// NewCacheRepo 创建缓存仓储实例
func NewCacheRepo(client *redis.Client, logger *zap.Logger) biz.CacheRepo {
	return &cacheRepo{
		client: client,
		logger: logger,
	}
}

// Set 设置缓存
func (r *cacheRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	r.logger.Debug("设置缓存", zap.String("key", key), zap.Duration("expiration", expiration))

	// 序列化值
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			r.logger.Error("序列化缓存值失败", zap.String("key", key), zap.Error(err))
			return fmt.Errorf("序列化缓存值失败: %w", err)
		}
		data = string(jsonData)
	}

	err := r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		r.logger.Error("设置缓存失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("设置缓存失败: %w", err)
	}

	r.logger.Debug("缓存设置成功", zap.String("key", key))
	return nil
}

// Get 获取缓存
func (r *cacheRepo) Get(ctx context.Context, key string) (string, error) {
	r.logger.Debug("获取缓存", zap.String("key", key))

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Debug("缓存不存在", zap.String("key", key))
			return "", fmt.Errorf("缓存不存在: %s", key)
		}
		r.logger.Error("获取缓存失败", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("获取缓存失败: %w", err)
	}

	r.logger.Debug("缓存获取成功", zap.String("key", key))
	return result, nil
}

// Delete 删除缓存
func (r *cacheRepo) Delete(ctx context.Context, key string) error {
	r.logger.Debug("删除缓存", zap.String("key", key))

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("删除缓存失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	r.logger.Debug("缓存删除成功", zap.String("key", key))
	return nil
}

// Exists 检查缓存是否存在
func (r *cacheRepo) Exists(ctx context.Context, key string) (bool, error) {
	r.logger.Debug("检查缓存是否存在", zap.String("key", key))

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error("检查缓存存在性失败", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("检查缓存存在性失败: %w", err)
	}

	exists := count > 0
	r.logger.Debug("缓存存在性检查完成", zap.String("key", key), zap.Bool("exists", exists))
	return exists, nil
}

// Expire 设置缓存过期时间
func (r *cacheRepo) Expire(ctx context.Context, key string, expiration time.Duration) error {
	r.logger.Debug("设置缓存过期时间", zap.String("key", key), zap.Duration("expiration", expiration))

	ok, err := r.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		r.logger.Error("设置缓存过期时间失败", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("设置缓存过期时间失败: %w", err)
	}

	if !ok {
		r.logger.Warn("设置缓存过期时间失败，缓存可能不存在", zap.String("key", key))
		return fmt.Errorf("设置缓存过期时间失败，缓存可能不存在: %s", key)
	}

	r.logger.Debug("缓存过期时间设置成功", zap.String("key", key))
	return nil
}

// HGet 获取哈希缓存
func (r *cacheRepo) HGet(ctx context.Context, key, field string) (string, error) {
	r.logger.Debug("获取哈希缓存", zap.String("key", key), zap.String("field", field))

	result, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Debug("哈希缓存字段不存在", zap.String("key", key), zap.String("field", field))
			return "", fmt.Errorf("哈希缓存字段不存在: %s.%s", key, field)
		}
		r.logger.Error("获取哈希缓存失败", zap.String("key", key), zap.String("field", field), zap.Error(err))
		return "", fmt.Errorf("获取哈希缓存失败: %w", err)
	}

	r.logger.Debug("哈希缓存获取成功", zap.String("key", key), zap.String("field", field))
	return result, nil
}

// HSet 设置哈希缓存
func (r *cacheRepo) HSet(ctx context.Context, key, field string, value interface{}) error {
	r.logger.Debug("设置哈希缓存", zap.String("key", key), zap.String("field", field))

	// 序列化值
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			r.logger.Error("序列化哈希缓存值失败",
				zap.String("key", key),
				zap.String("field", field),
				zap.Error(err))
			return fmt.Errorf("序列化哈希缓存值失败: %w", err)
		}
		data = string(jsonData)
	}

	err := r.client.HSet(ctx, key, field, data).Err()
	if err != nil {
		r.logger.Error("设置哈希缓存失败",
			zap.String("key", key),
			zap.String("field", field),
			zap.Error(err))
		return fmt.Errorf("设置哈希缓存失败: %w", err)
	}

	r.logger.Debug("哈希缓存设置成功", zap.String("key", key), zap.String("field", field))
	return nil
}

// HDel 删除哈希缓存字段
func (r *cacheRepo) HDel(ctx context.Context, key string, fields ...string) error {
	r.logger.Debug("删除哈希缓存字段", zap.String("key", key), zap.Strings("fields", fields))

	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		r.logger.Error("删除哈希缓存字段失败",
			zap.String("key", key),
			zap.Strings("fields", fields),
			zap.Error(err))
		return fmt.Errorf("删除哈希缓存字段失败: %w", err)
	}

	r.logger.Debug("哈希缓存字段删除成功", zap.String("key", key), zap.Strings("fields", fields))
	return nil
}

// HGetAll 获取哈希缓存所有字段
func (r *cacheRepo) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	r.logger.Debug("获取哈希缓存所有字段", zap.String("key", key))

	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		r.logger.Error("获取哈希缓存所有字段失败", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("获取哈希缓存所有字段失败: %w", err)
	}

	r.logger.Debug("哈希缓存所有字段获取成功",
		zap.String("key", key),
		zap.Int("field_count", len(result)))
	return result, nil
}
