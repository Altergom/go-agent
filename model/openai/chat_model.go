package openai

import (
	"context"
	"go-agent/config"
	"go-agent/model"

	"github.com/cloudwego/eino-ext/components/model/openai"
	model2 "github.com/cloudwego/eino/components/model"
)

func InitChatModel() {
	model.RegisterChatModel("openai", func(ctx context.Context) (model2.BaseChatModel, error) {
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey: config.Cfg.OpenAIConf.OpenAIKey,
			Model:  config.Cfg.OpenAIConf.OpenAIChatModel,
		})
	})
}
