package compose

import (
	"context"
	"go-agent/rag/tools"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

// BuildIndexingGraph 创建检索图
func BuildIndexingGraph(ctx context.Context) (compose.Runnable[document.Source, []string], error) {
	const (
		FileLoader    = "FileLoader"
		TextSplitter  = "TextSplitter"
		MilvusIndexer = "MilvusIndexer"
	)

	// 初始化组件
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		return nil, err
	}

	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   0,
		OverlapSize: 0,
		IDGenerator: nil,
	})
	if err != nil {
		return nil, err
	}

	indexer, err := tools.NewIndexer(ctx)
	if err != nil {
		return nil, err
	}

	// 创建图
	g := compose.NewGraph[document.Source, []string]()

	// 添加节点
	_ = g.AddLoaderNode(FileLoader, loader)
	_ = g.AddDocumentTransformerNode(TextSplitter, splitter)
	_ = g.AddIndexerNode(MilvusIndexer, indexer)

	// 添加边
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, TextSplitter)
	_ = g.AddEdge(TextSplitter, MilvusIndexer)
	_ = g.AddEdge(MilvusIndexer, compose.END)

	// 编译图
	r, err := g.Compile(
		ctx,
		compose.WithGraphName("RAGIndexing"),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}
