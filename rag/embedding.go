package rag

import (
	"context"
	"go-agent/config"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/embedding"
)

var Embedding embedding.Embedder

func NewEmbedding(ctx context.Context) (embedding.Embedder, error) {
	emb, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: config.Cfg.ArkConf.ArkKey,
		Model:  config.Cfg.ArkConf.ArkEmbeddingModel,
	})
	if err != nil {
		return nil, err
	}
	return emb, nil
}
