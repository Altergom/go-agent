package chat_model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
)

type ChatModelFactory func(ctx context.Context) (model.BaseChatModel, error)

var chatModelRegistry = make(map[string]ChatModelFactory)

func init() {
	initArk()
	initOpenAI()
	initQwen()
	initDeepSeek()
	initGemini()
}

// registerChatModel 注册聊天模型进入工厂
func registerChatModel(name string, factory ChatModelFactory) {
	chatModelRegistry[name] = factory
}

func GetChatModel(ctx context.Context, name string) (model.BaseChatModel, error) {
	create, ok := chatModelRegistry[name]
	if !ok {
		return nil, fmt.Errorf("不支持的模型类型: %s", name)
	}

	return create(ctx)
}
