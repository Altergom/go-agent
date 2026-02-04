package embedding_model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
)

type EmbeddingModelFactory func(ctx context.Context) (embedding.Embedder, error)

var embeddingModelRegistry = make(map[string]EmbeddingModelFactory)

func init() {
	initArk()
	initOpenAI()
	initQwen()
	initGemini()
}

// registerEmbeddingModel 注册嵌入模型进入工厂
func registerEmbeddingModel(name string, factory EmbeddingModelFactory) {
	embeddingModelRegistry[name] = factory
}

func GetEmbeddingModel(ctx context.Context, name string) (embedding.Embedder, error) {
	create, ok := embeddingModelRegistry[name]
	if !ok {
		return nil, fmt.Errorf("不支持的嵌入模型类型: %s", name)
	}

	return create(ctx)
}
