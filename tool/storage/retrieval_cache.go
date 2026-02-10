package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

const (
	embeddingCachePrefix = "embedding"
	retrievalCachePrefix = "retrieval"
	embeddingCacheTTL    = 1 * time.Hour
	retrievalCacheTTL    = 1 * time.Hour
)

// RetrievalCache 检索缓存管理器
type RetrievalCache struct {
	client      *redis.Client
	fallbackEmb map[string][]float64          // 降级模式：Embedding缓存
	fallbackDoc map[string][]*schema.Document // 降级模式：召回结果缓存
	useFallback bool
}

// NewRetrievalCache 创建检索缓存管理器
func NewRetrievalCache() *RetrievalCache {
	client, err := GetRedisClient()
	if err != nil {
		// Redis不可用时使用内存模式
		return &RetrievalCache{
			fallbackEmb: make(map[string][]float64),
			fallbackDoc: make(map[string][]*schema.Document),
			useFallback: true,
		}
	}

	return &RetrievalCache{
		client:      client,
		fallbackEmb: make(map[string][]float64),
		fallbackDoc: make(map[string][]*schema.Document),
		useFallback: false,
	}
}

// hashQuery 对查询字符串计算哈希值
func (r *RetrievalCache) hashQuery(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:])
}

// GetEmbedding 获取缓存的Embedding向量
func (r *RetrievalCache) GetEmbedding(ctx context.Context, query string) ([]float64, bool) {
	hash := r.hashQuery(query)

	// 降级模式
	if r.useFallback {
		if vec, ok := r.fallbackEmb[hash]; ok {
			return vec, true
		}
		return nil, false
	}

	// Redis模式
	key := fmt.Sprintf("%s:%s", embeddingCachePrefix, hash)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.GetEmbedding(ctx, query)
	}

	// 反序列化
	var vec []float64
	if err := json.Unmarshal(data, &vec); err != nil {
		return nil, false
	}

	return vec, true
}

// SetEmbedding 缓存Embedding向量
func (r *RetrievalCache) SetEmbedding(ctx context.Context, query string, embedding []float64) error {
	hash := r.hashQuery(query)

	// 降级模式
	if r.useFallback {
		r.fallbackEmb[hash] = embedding
		return nil
	}

	// 序列化
	data, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	// 保存到Redis
	key := fmt.Sprintf("%s:%s", embeddingCachePrefix, hash)
	err = r.client.Set(ctx, key, data, embeddingCacheTTL).Err()
	if err != nil {
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.SetEmbedding(ctx, query, embedding)
	}

	return nil
}

// GetRetrieval 获取缓存的召回结果
func (r *RetrievalCache) GetRetrieval(ctx context.Context, query string) ([]*schema.Document, bool) {
	hash := r.hashQuery(query)

	// 降级模式
	if r.useFallback {
		if docs, ok := r.fallbackDoc[hash]; ok {
			return docs, true
		}
		return nil, false
	}

	// Redis模式
	key := fmt.Sprintf("%s:%s", retrievalCachePrefix, hash)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.GetRetrieval(ctx, query)
	}

	// 反序列化
	var docs []*schema.Document
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, false
	}

	return docs, true
}

// SetRetrieval 缓存召回结果
func (r *RetrievalCache) SetRetrieval(ctx context.Context, query string, docs []*schema.Document) error {
	hash := r.hashQuery(query)

	// 降级模式
	if r.useFallback {
		r.fallbackDoc[hash] = docs
		return nil
	}

	// 序列化
	data, err := json.Marshal(docs)
	if err != nil {
		return fmt.Errorf("failed to marshal documents: %w", err)
	}

	// 保存到Redis
	key := fmt.Sprintf("%s:%s", retrievalCachePrefix, hash)
	err = r.client.Set(ctx, key, data, retrievalCacheTTL).Err()
	if err != nil {
		// Redis出错时降级到内存模式
		r.useFallback = true
		return r.SetRetrieval(ctx, query, docs)
	}

	return nil
}

// InvalidateRetrieval 使指定查询的召回缓存失效
func (r *RetrievalCache) InvalidateRetrieval(ctx context.Context, query string) error {
	hash := r.hashQuery(query)

	// 降级模式
	if r.useFallback {
		delete(r.fallbackDoc, hash)
		return nil
	}

	// Redis模式
	key := fmt.Sprintf("%s:%s", retrievalCachePrefix, hash)
	return r.client.Del(ctx, key).Err()
}

// InvalidateEmbedding 使指定查询的Embedding缓存失效
func (r *RetrievalCache) InvalidateEmbedding(ctx context.Context, query string) error {
	hash := r.hashQuery(query)

	// 降级模式
	if r.useFallback {
		delete(r.fallbackEmb, hash)
		return nil
	}

	// Redis模式
	key := fmt.Sprintf("%s:%s", embeddingCachePrefix, hash)
	return r.client.Del(ctx, key).Err()
}

// ClearAllCache 清空所有缓存
func (r *RetrievalCache) ClearAllCache(ctx context.Context) error {
	// 降级模式
	if r.useFallback {
		r.fallbackEmb = make(map[string][]float64)
		r.fallbackDoc = make(map[string][]*schema.Document)
		return nil
	}

	// Redis模式：清除所有embedding和retrieval key
	patterns := []string{
		fmt.Sprintf("%s:*", embeddingCachePrefix),
		fmt.Sprintf("%s:*", retrievalCachePrefix),
	}

	for _, pattern := range patterns {
		keys, err := r.client.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}
		if len(keys) > 0 {
			r.client.Del(ctx, keys...)
		}
	}

	return nil
}

// GetStats 获取缓存统计信息
func (r *RetrievalCache) GetStats(ctx context.Context) map[string]interface{} {
	stats := map[string]interface{}{
		"mode": "memory",
	}

	if r.useFallback {
		stats["embedding_count"] = len(r.fallbackEmb)
		stats["retrieval_count"] = len(r.fallbackDoc)
	} else {
		stats["mode"] = "redis"
		// 可以添加Redis统计信息
	}

	return stats
}
