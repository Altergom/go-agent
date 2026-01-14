package rag

import (
	"context"
	"go-agent/config"
	"strconv"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
)

var Retriever *milvus.Retriever

func NewRetriever(ctx context.Context) (*milvus.Retriever, error) {
	topK, _ := strconv.Atoi(config.Cfg.MilvusConf.TopK)
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
