package memory

import (
	"context"
	"sync"
)

type Store interface {
	Get(ctx context.Context, sessionID string) (*Session, error)
	Save(ctx context.Context, sessionID string, session *Session) error
}

type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]*Session
}

func NewMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]*Session)}
}

func (s *InMemoryStore) Get(ctx context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sess, ok := s.data[id]; ok {
		return sess, nil
	}

	return &Session{ID: id}, nil
}

func (s *InMemoryStore) Save(ctx context.Context, id string, sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = sess

	return nil
}
