package qwen

import (
	"context"
	"go-agent/config"
	"go-agent/model"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	model2 "github.com/cloudwego/eino/components/model"
)

func InitChatModel() {
	model.RegisterChatModel("qwen", func(ctx context.Context) (model2.BaseChatModel, error) {
		return qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			APIKey:  config.Cfg.QwenConf.QwenKey,
			Model:   config.Cfg.QwenConf.QwenChatModel,
		})
	})
}
