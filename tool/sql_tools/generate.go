package sql_tools

import (
	"context"
	"go-agent/model/chat_model"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type SQLState struct {
	Query        string
	GeneratedSQL string
	Result       string
}

func SQLGenerate(ctx context.Context, prompt string) (string, error) {
	_ = compose.ProcessState[*SQLState](ctx, func(ctx context.Context, state *SQLState) error {
		state.Query = prompt
		return nil
	})

	systemMsg := &schema.Message{
		Role: schema.System,
		Content: `你是一个专业的 MySQL 专家。
你的任务是根据用户提供的【数据处理需求描述】生成一条准确的 SQL 语句。
要求：
1. 必须使用标准 MySQL 语法。
2. 仅输出 SQL 语句，严禁包含任何 Markdown 格式（如 ` + "```" + `）、解释说明或多余字符。
3. 确保 SQL 逻辑严密，处理好时间范围和数值单位。
4. 如果输入的需求描述中包含表名或字段名的显式指定，请严格遵守。`,
	}

	// 将意图识别Agent传来的prompt作为User Message
	userMsg := &schema.Message{
		Role:    schema.User,
		Content: prompt,
	}

	cm, err := chat_model.GetChatModel(ctx, "SQL_Expert")
	if err != nil {
		return "", err
	}

	resp, err := cm.Generate(ctx, []*schema.Message{systemMsg, userMsg})
	if err != nil {
		return "", err
	}

	_ = compose.ProcessState[*SQLState](ctx, func(ctx context.Context, state *SQLState) error {
		state.GeneratedSQL = resp.Content
		return nil
	})

	return resp.Content, nil
}
