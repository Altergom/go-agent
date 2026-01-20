package chat_model

import (
	"context"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	model2 "github.com/cloudwego/eino/components/model"
)

func initQwen() {
	registerChatModel("qwen", func(ctx context.Context) (model2.BaseChatModel, error) {
		return qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			APIKey:  config.Cfg.QwenConf.QwenKey,
			Model:   config.Cfg.QwenConf.QwenChatModel,
		})
	})
}
