package chat_model

import (
	"context"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

func initArk() {
	registerChatModel("ark", func(ctx context.Context) (model.BaseChatModel, error) {
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: config.Cfg.ArkConf.ArkKey,
			Model:  config.Cfg.ArkConf.ArkChatModel,
		})
	})
}
