package sft

import "github.com/cloudwego/eino/schema"

type Sample struct {
	ID               string            `json:"id"`
	SessionID        string            `json:"session_id"` // 关联一次完整的对话
	AgentID          string            `json:"agent_id"`
	NodeName         string            `json:"node_name"`
	Component        string            `json:"component"`
	ModelType        string            `json:"model_type"`
	Messages         []*schema.Message `json:"messages"`         // 转换后的对话历史
	Context          []string          `json:"context"`          // 检索到的原始片段
	Label            int               `json:"label"`            // 0: 未标注, 1: 优秀, -1: 差评
	Correction       string            `json:"correction"`       // 教师模型修正后的回答内容
	TeacherScore     float64           `json:"teacher_score"`    // 教师打分
	TeacherReasoning string            `json:"teacher_response"` // 教师打分理由
	IsAnnotated      bool              `json:"is_annotated"`     // 是否已标注
	IsSpeculative    bool              `json:"is_speculative"`   // 是否启用了投机采样
	SpecHitRate      float64           `json:"spec_hit_rate"`    // 投机命中率
	InferenceMs      int64             `json:"inference_ms"`     // 总推理耗时
	Timestamp        int64             `json:"timestamp"`
}
