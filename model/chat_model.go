package model

import (
	"context"
	"fmt"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

type ChatModelFactory func(ctx context.Context) (model.BaseChatModel, error)

var CM model.BaseChatModel
var modelRegistry = map[string]ChatModelFactory{
	"ark": func(ctx context.Context) (model.BaseChatModel, error) {
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: config.Cfg.ArkConf.ArkKey,
			Model:  config.Cfg.ArkConf.ArkChatModel,
		})
	},
	"openai": func(ctx context.Context) (model.BaseChatModel, error) {
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey: config.Cfg.OpenAIConf.OpenAIKey,
			Model:  config.Cfg.OpenAIConf.OpenAIChatModel,
		})
	},
}

// NewChatModel 根据配置创建 ChatModel 实例
func NewChatModel(ctx context.Context) (model.BaseChatModel, error) {
	create, ok := modelRegistry[config.Cfg.ChatModelType]
	if !ok {
		return nil, fmt.Errorf("不支持的 ChatModel 类型: %s", config.Cfg.ChatModelType)
	}

	return create(ctx)
}
