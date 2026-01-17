package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
)

// Splitter 分割器 把文档分割成chunk块(因为窗口限制)
var Splitter document.Transformer

func NewSplitter(ctx context.Context) (document.Transformer, error) {
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   1000, // 每个文档块的大小
		OverlapSize: 200,  // 块之间的重叠大小(防止chunk的时候切出歧义导致语义丢失)
		IDGenerator: nil,
	})
	if err != nil {
		return nil, err
	}

	return splitter, nil
}
