package sft

import (
	"context"
	"encoding/json"
	"fmt"
	"go-agent/model/chat_model"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type teacherResponse struct {
	Score     float64 `json:"score"`     // 评分
	Reasoning string  `json:"reasoning"` // 理由
	Corrected string  `json:"corrected"` // 修正后的回答
}

func Annotate(ctx context.Context, s *Sample) error {
	teacherModel, _ := chat_model.GetChatModel(context.Background(), "deepseek")

	// 构建标注提示词
	prompt := fmt.Sprintf(`你是一个专业的数据标注导师。请根据提供的对话内容，对 AI 的回答进行评估和修正。
---
对话历史：
%v
AI 的回答：
%s
---
请严格按照以下 JSON 格式输出，不要包含任何其他文字：
{
  "score": 0.0到1.0之间的评分,
  "reasoning": "你的简要评价和改进建议",
  "corrected": "你修正后的更优回答"
}`, s.Messages[:len(s.Messages)-1], s.Messages[len(s.Messages)-1].Content)

	msgs := []*schema.Message{
		schema.SystemMessage("你只输出 JSON 格式的内容。"),
		schema.UserMessage(prompt),
	}

	resp, err := teacherModel.Generate(ctx, msgs)
	if err != nil {
		return err
	}

	// 解析模型输出（处理可能的 Markdown 代码块包裹）
	content := resp.Content
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var tr teacherResponse
	if err := json.Unmarshal([]byte(content), &tr); err != nil {
		return fmt.Errorf("解析教师模型返回失败: %v, content: %s", err, content)
	}

	// 更新 Sample 内容
	s.TeacherScore = tr.Score
	s.TeacherReasoning = tr.Reasoning
	s.Correction = tr.Corrected
	s.IsAnnotated = true

	return nil
}
