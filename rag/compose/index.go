package compose

import (
	"context"
	"go-agent/rag/tools"
	"log"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildIndexingGraph 创建检索图
func BuildIndexingGraph(ctx context.Context) (compose.Runnable[document.Source, []string], error) {
	const (
		FileLoader     = "FileLoader"
		TextSplitter   = "TextSplitter"
		MilvusIndexer  = "MilvusIndexer"
		DocumentParser = "DocumentParser"
		DebugChunks    = "DebugChunks"
	)

	// 创建图
	g := compose.NewGraph[document.Source, []string]()

	// 添加节点
	_ = g.AddLoaderNode(FileLoader, tools.Loader)
	_ = g.AddDocumentTransformerNode(TextSplitter, tools.Splitter)
	_ = g.AddIndexerNode(MilvusIndexer, tools.Indexer)
	_ = g.AddLambdaNode(DocumentParser, compose.InvokableLambda(BuildParseNode))
	_ = g.AddLambdaNode(DebugChunks, compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error) {
		for i, doc := range docs {
			contentPreview := doc.Content
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			log.Printf("待嵌入chunk[%d] ID=%s content=%q metadata=%v", i, doc.ID, contentPreview, doc.MetaData)
		}
		return docs, nil
	}))

	// 添加边
	_ = g.AddEdge(compose.START, FileLoader)
	_ = g.AddEdge(FileLoader, DocumentParser)
	_ = g.AddEdge(DocumentParser, TextSplitter)
	_ = g.AddEdge(TextSplitter, DebugChunks)
	_ = g.AddEdge(DebugChunks, MilvusIndexer)
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
