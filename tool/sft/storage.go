package sft

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Manager struct {
	BaseDir string
	mu      sync.Mutex
}

var defaultManager *Manager
var once sync.Once

func GetManager() *Manager {
	once.Do(func() {
		defaultManager = &Manager{BaseDir: "data/sft"}
	})
	return defaultManager
}

func (m *Manager) SaveSample(s *Sample) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 路径示例: data/sft/agent_rag_qa/2026-01-21.jsonl
	dir := filepath.Join(m.BaseDir, fmt.Sprintf("agent_%s", s.AgentID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileName := filepath.Join(dir, "raw_samples.jsonl")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, _ := json.Marshal(s)
	_, err = f.WriteString(string(data) + "\n")
	return err
}
