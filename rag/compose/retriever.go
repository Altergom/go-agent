package compose

import (
	"context"
	"fmt"
	"go-agent/model"
	"go-agent/rag/tools"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildRAGGraph 构建 RAG 检索和生成流程
func BuildRAGGraph(ctx context.Context) (compose.Runnable[map[string]any, *schema.Message], error) {
	const (
		QueryExtractor  = "QueryExtractor"
		MilvusRetriever = "MilvusRetriever"
		ChatTemplate    = "ChatTemplate"
		ChatModel       = "ChatModel"
	)

	// 初始化组件
	retriever, err := tools.NewRetriever(ctx)
	if err != nil {
		return nil, err
	}

	// 创建 ChatTemplate，包含检索到的文档上下文
	chatTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`你是一个有用的助手。请基于以下检索到的文档内容回答用户的问题。
如果文档中没有相关信息，请说明你不知道。

检索到的文档：
{documents}`),
		schema.UserMessage("{query}"),
	)

	// 创建图
	g := compose.NewGraph[map[string]any, *schema.Message]()

	//  提取用户查询
	_ = g.AddLambdaNode(QueryExtractor,
		compose.InvokableLambda(func(ctx context.Context, input map[string]any) (string, error) {
			query, ok := input["query"].(string)
			if !ok {
				return "", fmt.Errorf("query not found in input")
			}
			return query, nil
		}),
		compose.WithNodeName("ExtractQuery"),
		compose.WithOutputKey("query"),
	)

	// 检索相关文档
	_ = g.AddRetrieverNode(MilvusRetriever, retriever,
		compose.WithOutputKey("documents"),
		compose.WithNodeName("MilvusRetriever"),
	)

	// 格式化检索到的文档为字符串
	_ = g.AddLambdaNode("FormatDocuments",
		compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) (string, error) {
			var result string
			for i, doc := range docs {
				result += fmt.Sprintf("文档 %d (相似度: %.2f):\n%s\n\n",
					i+1, doc.Score, doc.Content)
			}
			return result, nil
		}),
		compose.WithOutputKey("documents"),
	)

	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplate)

	_ = g.AddChatModelNode(ChatModel, model.Md.ChatModel,
		compose.WithNodeName("ChatModel"),
	)

	_ = g.AddEdge(compose.START, QueryExtractor)
	_ = g.AddEdge(QueryExtractor, MilvusRetriever)

	_ = g.AddEdge(MilvusRetriever, "FormatDocuments")
	_ = g.AddEdge("FormatDocuments", ChatTemplate)

	_ = g.AddEdge(QueryExtractor, ChatTemplate)

	_ = g.AddEdge(ChatTemplate, ChatModel)
	_ = g.AddEdge(ChatModel, compose.END)

	r, err := g.Compile(
		ctx,
		compose.WithGraphName("RAGRetrieval"),
		compose.WithNodeTriggerMode(compose.AllPredecessor),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}
