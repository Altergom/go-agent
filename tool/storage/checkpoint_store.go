package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/redis/go-redis/v9"
)

const (
	checkpointPrefix = "checkpoint"
	checkpointTTL    = 24 * time.Hour // 24小时过期
)

// RedisCheckPointStore 实现 compose.CheckPointStore 接口
type RedisCheckPointStore struct {
	client      *redis.Client
	fallbackMap map[string][]byte // 降级到内存模式
	useFallback bool
}

// NewRedisCheckPointStore 创建Redis CheckPoint存储
func NewRedisCheckPointStore() compose.CheckPointStore {
	client, err := GetRedisClient()
	if err != nil {
		// Redis不可用时使用内存模式
		return &RedisCheckPointStore{
			fallbackMap: make(map[string][]byte),
			useFallback: true,
		}
	}

	return &RedisCheckPointStore{
		client:      client,
		fallbackMap: make(map[string][]byte),
		useFallback: false,
	}
}

// Get 获取checkpoint
func (r *RedisCheckPointStore) Get(ctx context.Context, checkpointID string) ([]byte, bool, error) {
	// 降级模式
	if r.useFallback {
		if data, ok := r.fallbackMap[checkpointID]; ok {
			return data, true, nil
		}
		return nil, false, fmt.Errorf("no fallback checkpoint found for %s", checkpointID)
	}

	// Redis模式
	key := r.makeKey(checkpointID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, fmt.Errorf("no checkpoint found for %s", checkpointID)
		}
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.Get(ctx, checkpointID)
	}

	return data, true, nil
}

// Set 保存checkpoint
func (r *RedisCheckPointStore) Set(ctx context.Context, checkpointID string, checkpoint []byte) error {
	// 降级模式
	if r.useFallback {
		if _, ok := r.fallbackMap[checkpointID]; !ok {
			r.fallbackMap[checkpointID] = []byte{}
		}
		r.fallbackMap[checkpointID] = checkpoint
		return nil
	}

	// Redis模式
	key := r.makeKey(checkpointID)
	err := r.client.Set(ctx, key, checkpoint, checkpointTTL).Err()
	if err != nil {
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.Set(ctx, checkpointID, checkpoint)
	}

	return nil
}

// Delete 删除checkpoint
func (r *RedisCheckPointStore) Delete(ctx context.Context, checkpointID string) error {
	// 降级模式
	if r.useFallback {
		if _, ok := r.fallbackMap[checkpointID]; ok {
			delete(r.fallbackMap, checkpointID)
		}
		return nil
	}

	// Redis模式
	key := r.makeKey(checkpointID)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.Delete(ctx, checkpointID)
	}

	return nil
}

// makeKey 生成Redis key
func (r *RedisCheckPointStore) makeKey(checkpointID string) string {
	return fmt.Sprintf("%s:%s", checkpointPrefix, checkpointID)
}

// GetMetadata 获取checkpoint元数据
func (r *RedisCheckPointStore) GetMetadata(ctx context.Context, checkpointID string) (map[string]interface{}, bool, error) {
	data, exist, err := r.Get(ctx, checkpointID)
	if err != nil {
		return nil, exist, err
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, exist, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, exist, nil
}
