package sft

import "github.com/cloudwego/eino/schema"

type Sample struct {
	ID         string            `json:"id"`
	SessionID  string            `json:"session_id"` // 关联一次完整的对话
	AgentID    string            `json:"agent_id"`
	NodeName   string            `json:"node_name"`
	Component  string            `json:"component"`
	ModelType  string            `json:"model_type"`
	Messages   []*schema.Message `json:"messages"`   // 转换后的对话历史
	Context    []string          `json:"context"`    // 检索到的原始片段
	Label      int               `json:"label"`      // 0: 未标注, 1: 优秀, -1: 差评
	Correction string            `json:"correction"` // 人工修正后的回答内容
	Timestamp  int64             `json:"timestamp"`
}
