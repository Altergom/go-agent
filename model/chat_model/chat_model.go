package chat_model

import (
	"context"
	"fmt"
	"go-agent/config"

	"github.com/cloudwego/eino/components/model"
)

type ChatModelFactory func(ctx context.Context) (model.BaseChatModel, error)

var chatModelRegistry = make(map[string]ChatModelFactory)
var CM model.BaseChatModel

// NewChatModel 根据配置创建 ChatModel 实例
func NewChatModel(ctx context.Context) (model.BaseChatModel, error) {
	initArk()
	initOpenAI()
	initQwen()
	create, ok := chatModelRegistry[config.Cfg.ChatModelType]
	if !ok {
		return nil, fmt.Errorf("不支持的 ChatModel 类型: %s", config.Cfg.ChatModelType)
	}

	return create(ctx)
}

// registerChatModel 注册聊天模型进入工厂
func registerChatModel(name string, factory ChatModelFactory) {
	chatModelRegistry[name] = factory
}
