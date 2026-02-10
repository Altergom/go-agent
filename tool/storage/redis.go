package storage

import (
	"context"
	"fmt"
	"go-agent/config"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
	redisErr    error
)

// InitRedis 初始化Redis连接
func InitRedis(ctx context.Context) error {
	redisOnce.Do(func() {
		if config.Cfg == nil {
			redisErr = fmt.Errorf("config not initialized")
			return
		}

		db, _ := strconv.Atoi(config.Cfg.RedisConf.DB)
		redisClient = redis.NewClient(&redis.Options{
			Addr:         config.Cfg.RedisConf.Addr,
			Password:     config.Cfg.RedisConf.Password,
			DB:           db,
			PoolSize:     10,
			MinIdleConns: 2,
			MaxRetries:   3,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		})

		// 测试连接
		if err := redisClient.Ping(ctx).Err(); err != nil {
			redisErr = fmt.Errorf("failed to connect to Redis: %w", err)
			return
		}

		log.Println("Redis 连接已建立")
	})

	return redisErr
}

// GetRedisClient 获取Redis客户端
func GetRedisClient() (*redis.Client, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized, call InitRedis first")
	}
	return redisClient, nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

// Set 设置键值对，可选过期时间
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键对应的值
func Get(ctx context.Context, key string) (string, error) {
	client, err := GetRedisClient()
	if err != nil {
		return "", err
	}
	return client.Get(ctx, key).Result()
}

// Delete 删除键
func Delete(ctx context.Context, keys ...string) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	client, err := GetRedisClient()
	if err != nil {
		return false, err
	}
	n, err := client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置键的过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.Expire(ctx, key, expiration).Err()
}

// HSet 设置哈希表字段的值
func HSet(ctx context.Context, key string, field string, value interface{}) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希表字段的值
func HGet(ctx context.Context, key string, field string) (string, error) {
	client, err := GetRedisClient()
	if err != nil {
		return "", err
	}
	return client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希表所有字段和值
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	client, err := GetRedisClient()
	if err != nil {
		return nil, err
	}
	return client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希表字段
func HDel(ctx context.Context, key string, fields ...string) error {
	client, err := GetRedisClient()
	if err != nil {
		return err
	}
	return client.HDel(ctx, key, fields...).Err()
}
