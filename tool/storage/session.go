package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix = "session"
	sessionTTL    = 24 * time.Hour // 会话24小时过期
)

// SessionContext 会话上下文
type SessionContext struct {
	InterruptID   string `json:"interrupt_id"`
	CheckPointID  string `json:"checkpoint_id"`
	OriginalQuery string `json:"original_query"`
	WaitingRefine bool   `json:"waiting_refine"`
}

// SessionStore Session存储管理器
type SessionStore struct {
	client      *redis.Client
	fallbackMap map[string]*SessionContext // 降级到内存模式
	useFallback bool
}

// NewSessionStore 创建Session存储管理器
func NewSessionStore() *SessionStore {
	client, err := GetRedisClient()
	if err != nil {
		// Redis不可用时使用内存模式
		return &SessionStore{
			fallbackMap: make(map[string]*SessionContext),
			useFallback: true,
		}
	}

	return &SessionStore{
		client:      client,
		fallbackMap: make(map[string]*SessionContext),
		useFallback: false,
	}
}

// SaveSession 保存会话上下文
func (s *SessionStore) SaveSession(ctx context.Context, sessionID string, sc *SessionContext) error {
	// 降级模式
	if s.useFallback {
		s.fallbackMap[sessionID] = sc
		return nil
	}

	// 序列化为JSON
	data, err := json.Marshal(sc)
	if err != nil {
		return fmt.Errorf("failed to marshal session context: %w", err)
	}

	// 保存到Redis
	key := s.makeKey(sessionID)
	err = s.client.Set(ctx, key, data, sessionTTL).Err()
	if err != nil {
		// Redis出错时降级到内存模式
		s.useFallback = true
		return s.SaveSession(ctx, sessionID, sc)
	}

	return nil
}

// GetSession 获取会话上下文
func (s *SessionStore) GetSession(ctx context.Context, sessionID string) (*SessionContext, error) {
	// 降级模式
	if s.useFallback {
		if sc, ok := s.fallbackMap[sessionID]; ok {
			return sc, nil
		}
		return nil, fmt.Errorf("session not found")
	}

	// 从Redis获取
	key := s.makeKey(sessionID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		// Redis出错时降级到内存模式
		s.useFallback = true
		return s.GetSession(ctx, sessionID)
	}

	// 反序列化
	var sc SessionContext
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session context: %w", err)
	}

	return &sc, nil
}

// DeleteSession 删除会话上下文
func (s *SessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	// 降级模式
	if s.useFallback {
		delete(s.fallbackMap, sessionID)
		return nil
	}

	// 从Redis删除
	key := s.makeKey(sessionID)
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		// Redis出错时降级到内存模式
		s.useFallback = true
		return s.DeleteSession(ctx, sessionID)
	}

	return nil
}

// Exists 检查会话是否存在
func (s *SessionStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	// 降级模式
	if s.useFallback {
		_, ok := s.fallbackMap[sessionID]
		return ok, nil
	}

	// 检查Redis
	key := s.makeKey(sessionID)
	n, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		// Redis出错时降级到内存模式
		s.useFallback = true
		return s.Exists(ctx, sessionID)
	}

	return n > 0, nil
}

// UpdateSession 更新会话中的特定字段
func (s *SessionStore) UpdateSession(ctx context.Context, sessionID string, updateFn func(*SessionContext)) error {
	// 获取现有会话
	sc, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// 执行更新函数
	updateFn(sc)

	// 保存更新后的会话
	return s.SaveSession(ctx, sessionID, sc)
}

// makeKey 生成Redis key: session:{sessionID}
func (s *SessionStore) makeKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", sessionPrefix, sessionID)
}

// ListSessions 列出所有会话ID
func (s *SessionStore) ListSessions(ctx context.Context) ([]string, error) {
	// 降级模式
	if s.useFallback {
		ids := make([]string, 0, len(s.fallbackMap))
		for id := range s.fallbackMap {
			ids = append(ids, id)
		}
		return ids, nil
	}

	// Redis模式
	pattern := s.makeKey("*")
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		// Redis出错时降级到内存模式
		s.useFallback = true
		return s.ListSessions(ctx)
	}

	// 提取session ID
	ids := make([]string, 0, len(keys))
	prefix := sessionPrefix + ":"
	for _, key := range keys {
		if len(key) > len(prefix) {
			ids = append(ids, key[len(prefix):])
		}
	}

	return ids, nil
}
