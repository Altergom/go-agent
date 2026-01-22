package memory

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type Summarizer struct {
	Model         model.BaseChatModel
	MaxHistoryLen int
}

const SummaryPrompt = `你是一个记忆管理助手。
请将 <previous_summary> 和 <older_messages> 合并为一个新的、精炼的摘要。
要求：
1. 保留核心事实、用户的偏好、以及尚未解决的问题。
2. 丢弃寒暄和冗余的中间过程。
3. 保持摘要的连贯性。

<previous_summary>: %s
<older_messages>: %s
新的摘要：`

func (s *Summarizer) Compress(ctx context.Context, sess *Session) error {
	// 只有当历史消息超过一定长度才触发
	if len(sess.History) <= s.MaxHistoryLen {
		return nil
	}

	// 划分：前N轮需要压缩，后M轮保留为短期记忆
	// 比如保留最近3轮
	toCompress := sess.History[:len(sess.History)-3]
	sess.History = sess.History[len(sess.History)-3:]

	// 格式化待压缩的消息
	historyText := ""
	for _, m := range toCompress {
		historyText += fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}

	// 调用LLM生成新摘要
	prompt := fmt.Sprintf(SummaryPrompt, sess.Summary, historyText)
	resp, err := s.Model.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return err
	}

	sess.Summary = resp.Content
	return nil
}
