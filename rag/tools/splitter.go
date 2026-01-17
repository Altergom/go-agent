package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
)

var Splitter document.Transformer

func NewSplitter(ctx context.Context) (document.Transformer, error) {
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   10,
		OverlapSize: 100,
		IDGenerator: nil,
	})
	if err != nil {
		return nil, err
	}

	return splitter, nil
}
