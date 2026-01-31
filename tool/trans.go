package tool

import (
	"github.com/cloudwego/eino/schema"
)

// MsgToMap 将消息列表转换为 map[string]any，用于 ChatTemplate 的输入
func MsgToMap(input []*schema.Message) (map[string]any, error) {
	if len(input) == 0 {
		return map[string]any{}, nil
	}
	// 默认取最后一条消息的内容作为 query
	return map[string]any{
		"query": input[len(input)-1].Content,
	}, nil
}

// MsgsToMsg 取消息列表中的最后一条消息
func MsgsToMsg(input []*schema.Message) (*schema.Message, error) {
	if len(input) == 0 {
		return nil, nil
	}
	return input[len(input)-1], nil
}

// MsgToMsgs 将单条消息包装为消息列表
func MsgToMsgs(input *schema.Message) ([]*schema.Message, error) {
	if input == nil {
		return nil, nil
	}
	return []*schema.Message{input}, nil
}

// StringToMsg 将字符串转换为 User 角色消息
func StringToMsg(input string) (*schema.Message, error) {
	return schema.UserMessage(input), nil
}
