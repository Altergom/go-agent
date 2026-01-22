package sft

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cloudwego/eino/schema"
)

type Manager struct {
	BaseDir string
	mu      sync.Mutex
}

type ExportOptions struct {
	MinScore    float64 // 最低评分过滤
	OnlyLabeled bool    // 是否只导出已标注样本
	Format      string  // 导出格式
}

var defaultManager *Manager
var once sync.Once

func GetManager() *Manager {
	once.Do(func() {
		// 基础目录
		defaultManager = &Manager{BaseDir: "data/sft"}
	})
	return defaultManager
}

func (m *Manager) SaveSample(s *Sample) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Join(m.BaseDir, fmt.Sprintf("agent_%s", s.AgentID), "samples")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 使用 ID 作为文件名，实现覆盖更新
	fileName := filepath.Join(dir, fmt.Sprintf("%s.json", s.ID))

	// os.O_TRUNC 会在文件存在时清空它，实现更新
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ") // 独立文件建议格式化，方便查看
	return encoder.Encode(s)
}

// ExportToJSONL 将散落的独立 JSON 文件聚合成一个训练用的 JSONL 文件
func (m *Manager) ExportToJSONL(agentID string, outputPath string, opts ExportOptions) (int, error) {
	samplesDir := filepath.Join(m.BaseDir, fmt.Sprintf("agent_%s", agentID), "samples")

	// 读取所有样本文件
	files, err := os.ReadDir(samplesDir)
	if err != nil {
		return 0, err
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer outFile.Close()

	count := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// 加载样本
		data, _ := os.ReadFile(filepath.Join(samplesDir, file.Name()))
		var s Sample
		json.Unmarshal(data, &s)

		// 3. 过滤逻辑
		if opts.OnlyLabeled && !s.IsAnnotated {
			continue
		}
		if s.TeacherScore < opts.MinScore {
			continue
		}

		// 格式转换
		// 训练的目标是：输入上下文，输出教师的 Correction
		trainingItem := map[string]interface{}{
			"messages": append(s.Messages[:len(s.Messages)-1], &schema.Message{
				Role:    schema.Assistant,
				Content: s.Correction, // 使用教师修正后的答案作为训练 Label
			}),
		}

		// 写入 JSONL
		line, _ := json.Marshal(trainingItem)
		outFile.WriteString(string(line) + "\n")
		count++
	}

	return count, nil
}
