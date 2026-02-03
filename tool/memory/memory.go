package memory

import (
	"context"
	"sync"

	"github.com/cloudwego/eino/compose"
)

type inMemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func (s *inMemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.data[key]
	return d, ok, nil
}

func (s *inMemoryStore) Set(ctx context.Context, key string, val []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
	return nil
}

func NewInMemoryStore() compose.CheckPointStore {
	return &inMemoryStore{data: make(map[string][]byte)}
}
