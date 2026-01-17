package compose

import (
	"context"
	"go-agent/rag/tools"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

// BuildIndexingGraph 创建检索图
func BuildIndexingGraph(ctx context.Context) (compose.Runnable[document.Source, []string], error) {
	const (
		FileLoader     = "FileLoader"
		TextSplitter   = "TextSplitter"
		MilvusIndexer  = "MilvusIndexer"
		DocumentParser = "DocumentParser"
	)

	// 创建图
	g := compose.NewGraph[document.Source, []string]()

	// 添加节点
	_ = g.AddLoaderNode(FileLoader, tools.Loader)
	_ = g.AddDocumentTransformerNode(TextSplitter, tools.Splitter)
	_ = g.AddIndexerNode(MilvusIndexer, tools.Indexer)
	_ = g.AddLambdaNode(DocumentParser, compose.InvokableLambda(BuildParseNode))

	// 添加边
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, DocumentParser)
	_ = g.AddEdge(DocumentParser, TextSplitter)
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
