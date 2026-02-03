package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// MsgToMap 将消息列表转换为 map[string]any，用于 ChatTemplate 的输入
func MsgToMap(ctx context.Context, input []*schema.Message) (map[string]any, error) {
	if len(input) == 0 {
		return map[string]any{}, nil
	}
	// 默认取最后一条消息的内容作为 query
	return map[string]any{
		"query": input[len(input)-1].Content,
	}, nil
}

// MsgsToMsg 取消息列表中的最后一条消息
func MsgsToMsg(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
	if len(input) == 0 {
		return nil, nil
	}
	return input[len(input)-1], nil
}

// MsgToMsgs 将单条消息包装为消息列表
func MsgToMsgs(ctx context.Context, input *schema.Message) ([]*schema.Message, error) {
	if input == nil {
		return nil, nil
	}
	return []*schema.Message{input}, nil
}

// MsgToSQLToolCall 将包含 SQL 的消息包装为符合 mysql_query 工具要求的 Assistant 消息
func MsgToSQLToolCall(ctx context.Context, input *schema.Message) (*schema.Message, error) {
	if input == nil {
		return nil, nil
	}

	// 构造 mysql_query 所需的 JSON 参数: {"sql": "..."}
	params := map[string]string{"sql": input.Content}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sql params: %w", err)
	}

	return schema.AssistantMessage("", []schema.ToolCall{
		{
			ID: "call_sql_exec",
			Function: schema.FunctionCall{
				Name:      "mysql_query",
				Arguments: string(paramsJSON),
			},
		},
	}), nil
}

// StringToMsg 将字符串转换为 User 角色消息
func StringToMsg(ctx context.Context, input string) (*schema.Message, error) {
	return schema.UserMessage(input), nil
}

// MsgsToQuery 从消息列表中提取最后一条消息的内容作为查询字符串
func MsgsToQuery(ctx context.Context, input []*schema.Message) (string, error) {
	if len(input) == 0 {
		return "", nil
	}
	return input[len(input)-1].Content, nil
}
