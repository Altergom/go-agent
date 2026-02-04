package embedding_model

import (
	"context"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/embedding/gemini"
	"github.com/cloudwego/eino/components/embedding"
	"google.golang.org/genai"
)

func initGemini() {
	registerEmbeddingModel("gemini", func(ctx context.Context) (embedding.Embedder, error) {
		cli, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: config.Cfg.GeminiConf.GeminiKey,
		})
		if err != nil {
			return nil, err
		}

		return gemini.NewEmbedder(ctx, &gemini.EmbeddingConfig{
			Client: cli,
			Model:  config.Cfg.GeminiConf.GeminiEmbedding,
		})
	})
}
