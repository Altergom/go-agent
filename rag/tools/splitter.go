package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
)

// Splitter 分割器 把文档分割成chunk块(因为窗口限制)
var Splitter document.Transformer

func NewSplitter(ctx context.Context) (document.Transformer, error) {
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   1000, // 每个文档块的大小
		OverlapSize: 200,  // 块之间的重叠大小(防止chunk的时候切出歧义导致语义丢失)
		IDGenerator: func(ctx context.Context, originalID string, splitIndex int) string {
			// 如果原始ID为空，使用默认前缀
			if originalID == "" {
				originalID = "doc"
			}
			// 为每个chunk生成唯一ID: 原始ID_索引
			return fmt.Sprintf("%s_chunk_%d", originalID, splitIndex)
		},
	})
	if err != nil {
		return nil, err
	}

	return splitter, nil
}
