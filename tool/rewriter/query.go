package rewriter

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type QueryRewriter struct {
	Model model.BaseChatModel
}

const RewritePrompt = `参考以下背景摘要和最近对话，将用户最后一次提问重写为一个独立的、适合向量检索的搜索语句。
背景摘要: %s
最近对话: %s
用户提问: %s
重写后的搜索语句（直接输出语句）: `

func (qr *QueryRewriter) Rephrase(ctx context.Context, summary string, history []*schema.Message, query string) (string, error) {
	historyText := ""
	for _, m := range history {
		historyText += fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}

	finalPrompt := fmt.Sprintf(RewritePrompt, summary, historyText, query)

	resp, err := qr.Model.Generate(ctx, []*schema.Message{
		schema.UserMessage(finalPrompt),
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
