package model

import (
	"context"
	"fmt"
	"go-agent/config"

	"github.com/cloudwego/eino/components/embedding"
)

type EmbeddingModelFactory func(ctx context.Context) (embedding.Embedder, error)

var embeddingModelRegistry = make(map[string]EmbeddingModelFactory)
var Embedding embedding.Embedder

func NewEmbeddingModel(ctx context.Context) (embedding.Embedder, error) {
	create, ok := embeddingModelRegistry[config.Cfg.EmbeddingModelType]
	if !ok {
		return nil, fmt.Errorf("不支持的 EmbeddingModel 类型: %s", config.Cfg.EmbeddingModelType)
	}

	return create(ctx)
}

// RegisterEmbeddingModel 注册嵌入模型进入工厂
func RegisterEmbeddingModel(name string, factory EmbeddingModelFactory) {
	embeddingModelRegistry[name] = factory
}
