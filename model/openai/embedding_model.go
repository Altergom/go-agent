package openai

import (
	"context"
	"go-agent/config"
	"go-agent/model"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
)

func InitEmbeddingModel() {
	model.RegisterEmbeddingModel("openai", func(ctx context.Context) (embedding.Embedder, error) {
		emb, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
			APIKey: config.Cfg.OpenAIConf.OpenAIKey,
			Model:  config.Cfg.OpenAIConf.OpenAIEmbedding,
		})
		if err != nil {
			return nil, err
		}

		return emb, nil
	})
}
