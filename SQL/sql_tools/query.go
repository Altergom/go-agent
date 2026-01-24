package sql_tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func Rewrite(ctx context.Context, prompt string, history []*schema.Document, query string, cm model.BaseChatModel) (string, error) {
	historyText := ""
	for _, m := range history {
		historyText += fmt.Sprintf("[%s]: %s\n", m.ID, m.Content)
	}

	finalPrompt := fmt.Sprintf(prompt, historyText, query)

	resp, err := cm.Generate(ctx, []*schema.Message{
		schema.UserMessage(finalPrompt),
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
