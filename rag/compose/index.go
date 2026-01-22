package compose

import (
	"context"
	"go-agent/rag/tools/indexer"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildIndexingGraph 创建检索图
func BuildIndexingGraph(ctx context.Context) (compose.Runnable[[]*schema.Document, []string], error) {
	const (
		MilvusIndexer = "MilvusIndexer"
		ESIndexer     = "ESIndexer"
	)

	// 创建图
	g := compose.NewGraph[[]*schema.Document, []string]()

	milvus, err := indexer.GetIndexer(ctx, "milvus")
	if err != nil {
		return nil, err
	}
	es, err := indexer.GetIndexer(ctx, "es")
	if err != nil {
		return nil, err
	}

	// 添加节点
	_ = g.AddIndexerNode(MilvusIndexer, milvus)
	_ = g.AddIndexerNode(ESIndexer, es)

	// 添加边
	_ = g.AddEdge(compose.START, MilvusIndexer)
	_ = g.AddEdge(compose.START, ESIndexer)
	_ = g.AddEdge(MilvusIndexer, compose.END)
	_ = g.AddEdge(ESIndexer, compose.END)

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
