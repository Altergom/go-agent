package chat_model

import (
	"context"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/components/model"
)

func initDeepSeek() {
	registerChatModel("deepseek", func(ctx context.Context) (model.BaseChatModel, error) {
		return deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			APIKey:  config.Cfg.DeepSeekConf.DeepSeekKey,
			Model:   config.Cfg.DeepSeekConf.DeepSeekChatModel,
			BaseURL: config.Cfg.DeepSeekConf.BaseUrl,
		})
	})
}
