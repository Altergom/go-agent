package tools

import (
	"context"
	"go-agent/config"
	"strconv"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
)

// Retriever 召回器 召回miLvus存储结果(就是MySQL的select)
var Retriever *milvus.Retriever

func NewRetriever(ctx context.Context) (*milvus.Retriever, error) {
	topK, err := strconv.Atoi(config.Cfg.MilvusConf.TopK)
	if err != nil || topK <= 0 {
		topK = 10
	}
	ret, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:    Milvus,
		Embedding: Embedding,
		TopK:      topK,
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}
