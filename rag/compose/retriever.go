package compose

import (
	"context"
	"go-agent/rag/tools"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildRetrieverGraph 仅负责检索，输入 query，输出文档列表
func BuildRetrieverGraph(ctx context.Context) (compose.Runnable[string, []*schema.Document], error) {
	const (
		MilvusRetriever = "MilvusRetriever"
	)

	g := compose.NewGraph[string, []*schema.Document]()

	// 直接复用全局初始化的 Retriever
	_ = g.AddRetrieverNode(MilvusRetriever, tools.Retriever)

	_ = g.AddEdge(compose.START, MilvusRetriever)
	_ = g.AddEdge(MilvusRetriever, compose.END)

	r, err := g.Compile(
		ctx,
		compose.WithGraphName("RAGRetriever"),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}
