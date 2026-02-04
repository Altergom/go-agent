package chat_model

import (
	"context"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"google.golang.org/genai"
)

func initGemini() {
	registerChatModel("gemini", func(ctx context.Context) (model.BaseChatModel, error) {
		cli, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: config.Cfg.GeminiConf.GeminiKey,
		})
		if err != nil {
			return nil, err
		}

		return gemini.NewChatModel(ctx, &gemini.Config{
			Client:      cli,
			Model:       config.Cfg.GeminiConf.GeminiChatModel,
			MaxTokens:   nil,
			Temperature: nil,
		})
	})
}
