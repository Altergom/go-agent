package rag_flow

import (
	"context"
	"fmt"
	"go-agent/rag/rag_tools/indexer"
	document2 "go-agent/tool/document"
	"sync"

	"github.com/cloudwego/eino/components/document"

	"github.com/cloudwego/eino/compose"
)

const (
	Milvus   = "Milvus"
	ES       = "ES"
	Splitter = "Splitter"
	Parser   = "Parser"
	Loader   = "Loader"
)

var (
	cachedIndexingGraph  compose.Runnable[document.Source, []string]
	indexingGraphOnce    sync.Once
	indexingGraphInitErr error
)

// InitIndexingGraph 在应用启动时编译并缓存索引图
func InitIndexingGraph(ctx context.Context) error {
	indexingGraphOnce.Do(func() {
		cachedIndexingGraph, indexingGraphInitErr = buildIndexingGraph(ctx)
	})
	return indexingGraphInitErr
}

func GetIndexingGraph() (compose.Runnable[document.Source, []string], error) {
	if cachedIndexingGraph == nil {
		return nil, fmt.Errorf("IndexingGraph 未初始化，请先调用 InitIndexingGraph")
	}
	return cachedIndexingGraph, nil
}

// buildIndexingGraph 创建索引图
func buildIndexingGraph(ctx context.Context) (compose.Runnable[document.Source, []string], error) {
	// 创建图
	g := compose.NewGraph[document.Source, []string]()

	milvus, err := indexer.GetIndexer(ctx, "milvus")
	if err != nil {
		return nil, err
	}
	es, err := indexer.GetIndexer(ctx, "es")
	if err != nil {
		return nil, err
	}

	// 添加节点
	_ = g.AddLoaderNode(Loader, document2.Loader)
	_ = g.AddLambdaNode(Parser, compose.InvokableLambda(BuildParseNode))
	_ = g.AddDocumentTransformerNode(Splitter, document2.Splitter)
	_ = g.AddIndexerNode(Milvus, milvus, compose.WithOutputKey("milvus_res"))
	_ = g.AddIndexerNode(ES, es, compose.WithOutputKey("es_res"))
	_ = g.AddLambdaNode("Merge", compose.InvokableLambda(func(ctx context.Context, input map[string]any) ([]string, error) {
		var allIDs []string
		for _, ids := range input {
			i, _ := ids.([]string)
			allIDs = append(allIDs, i...)
		}
		return allIDs, nil
	}))

	// 设置边
	_ = g.AddEdge(compose.START, Loader)
	_ = g.AddEdge(Loader, Parser)
	_ = g.AddEdge(Parser, Splitter)
	_ = g.AddEdge(Splitter, Milvus)
	_ = g.AddEdge(Splitter, ES)
	_ = g.AddEdge(Milvus, "Merge")
	_ = g.AddEdge(ES, "Merge")
	_ = g.AddEdge("Merge", compose.END)

	// 编译图
	r, err := g.Compile(
		ctx,
		compose.WithGraphName("RAGIndexing"),
		compose.WithNodeTriggerMode(compose.AllPredecessor),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}
