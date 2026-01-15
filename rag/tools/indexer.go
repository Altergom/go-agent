package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
)

var Indexer *milvus.Indexer

func NewIndexer(ctx context.Context) (*milvus.Indexer, error) {
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:    Milvus,
		Embedding: Embedding,
	})
	if err != nil {
		return nil, err
	}
	return indexer, nil
}
