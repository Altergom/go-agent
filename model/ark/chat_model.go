package ark

import (
	"context"
	"go-agent/config"
	"go-agent/model"

	"github.com/cloudwego/eino-ext/components/model/ark"
	model2 "github.com/cloudwego/eino/components/model"
)

func InitChatModel() {
	model.RegisterChatModel("ark", func(ctx context.Context) (model2.BaseChatModel, error) {
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: config.Cfg.ArkConf.ArkKey,
			Model:  config.Cfg.ArkConf.ArkChatModel,
		})
	})
}
