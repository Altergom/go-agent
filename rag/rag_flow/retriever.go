package rag_flow

import (
	"context"
	"fmt"
	"go-agent/rag/rag_tools/retriever"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	MilvusRetriever = "MilvusRetriever"
	ESRetriever     = "ESRetriever"
	Reranker        = "Reranker"
)

// BuildRetrieverGraph 仅负责检索，输入 query，输出文档列表
func BuildRetrieverGraph(ctx context.Context) (*compose.Graph[[]*schema.Message, []*schema.Document], error) {
	g := compose.NewGraph[[]*schema.Message, []*schema.Document]()

	// 构建召回节点
	milvus, err := retriever.GetRetriever(ctx, "milvus")
	if err != nil {
		return nil, err
	}
	es, err := retriever.GetRetriever(ctx, "es")
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode("QueryExtract", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (string, error) {
		if len(input) == 0 {
			return "", fmt.Errorf("empty input message")
		}
		return input[len(input)-1].Content, nil
	}))

	_ = g.AddRetrieverNode(MilvusRetriever, milvus, compose.WithOutputKey("milvus_retriever"))
	_ = g.AddRetrieverNode(ESRetriever, es, compose.WithOutputKey("es_retriever"))

	// ... (rest of the nodes) ...

	// 构建节点指向
	_ = g.AddEdge(compose.START, "QueryExtract")
	_ = g.AddEdge("QueryExtract", MilvusRetriever)
	_ = g.AddEdge("QueryExtract", ESRetriever)
	_ = g.AddEdge(MilvusRetriever, Reranker)
	_ = g.AddEdge(ESRetriever, Reranker)
	_ = g.AddEdge(Reranker, compose.END)

	return g, nil
}
