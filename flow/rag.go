package flow

import (
	"context"
	"fmt"
	"go-agent/model/chat_model"
	"go-agent/tool/memory"
	"go-agent/tool/rewriter"

	compose2 "go-agent/rag/compose"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type RAGChatInput struct {
	SessionID string
	Query     string
}

// internalState 用于在图节点间传递复杂的中间状态
type internalState struct {
	Input   RAGChatInput
	Session *memory.Session
	Query   string
	Docs    []*schema.Document
}

func BuildRAGChatFlow(ctx context.Context, store memory.Store, taskModel model.BaseChatModel) (compose.Runnable[RAGChatInput, *schema.Message], error) {
	qr := &rewriter.QueryRewriter{Model: taskModel}
	sm := &memory.Summarizer{Model: taskModel, MaxHistoryLen: 3}

	retrieverSubGraph, err := compose2.BuildRetrieverGraph(ctx)
	if err != nil {
		return nil, err
	}

	g := compose.NewGraph[RAGChatInput, *schema.Message]()

	// 节点 1: 预处理与加载记忆
	_ = g.AddLambdaNode("PreProcess", compose.InvokableLambda(func(ctx context.Context, in RAGChatInput) (*internalState, error) {
		sess, _ := store.Get(ctx, in.SessionID)
		return &internalState{Input: in, Session: sess}, nil
	}))

	// 节点 2: 查询重写
	_ = g.AddLambdaNode("Rewrite", compose.InvokableLambda(func(ctx context.Context, state *internalState) (*internalState, error) {
		newQuery, _ := qr.Rephrase(ctx, state.Session.Summary, state.Session.History, state.Input.Query)
		state.Query = newQuery
		return state, nil
	}))

	// 节点 3: 嵌套检索子图
	_ = g.AddLambdaNode("Retrieve", compose.InvokableLambda(func(ctx context.Context, state *internalState) (*internalState, error) {
		docs, err := retrieverSubGraph.Invoke(ctx, state.Query)
		if err != nil {
			return nil, err
		}
		state.Docs = docs
		return state, nil
	}))

	// 节点 4: 核心对话生成
	_ = g.AddLambdaNode("Chat", compose.InvokableLambda(func(ctx context.Context, state *internalState) (*internalState, error) {
		messages := make([]*schema.Message, 0)
		// 注入长期记忆
		if state.Session.Summary != "" {
			messages = append(messages, schema.SystemMessage("背景摘要: "+state.Session.Summary))
		}
		// 注入短期记忆
		messages = append(messages, state.Session.History...)

		// 注入检索知识
		knowledge := "参考知识:\n"
		for _, doc := range state.Docs {
			knowledge += doc.Content + "\n"
		}
		messages = append(messages, schema.UserMessage(knowledge+state.Input.Query))

		resp, err := chat_model.CM.Generate(ctx, messages)
		if err != nil {
			return nil, err
		}

		// 更新历史状态
		state.Session.History = append(state.Session.History, schema.UserMessage(state.Input.Query))
		state.Session.History = append(state.Session.History, resp)

		// 自动触发压缩与持久化
		go func(s *memory.Session) {
			bgCtx := context.Background()
			_ = sm.Compress(bgCtx, s)
			_ = store.Save(bgCtx, s.ID, s)
		}(state.Session)

		return state, nil
	}))

	// 节点 5: 类型转换节点
	_ = g.AddLambdaNode("FormatOutput", compose.InvokableLambda(func(ctx context.Context, state *internalState) (*schema.Message, error) {
		if len(state.Session.History) == 0 {
			return nil, fmt.Errorf("empty history")
		}
		return state.Session.History[len(state.Session.History)-1], nil
	}))

	// 连边
	_ = g.AddEdge(compose.START, "PreProcess")
	_ = g.AddEdge("PreProcess", "Rewrite")
	_ = g.AddEdge("Rewrite", "Retrieve")
	_ = g.AddEdge("Retrieve", "Chat")
	_ = g.AddEdge("Chat", compose.END)

	// 编译并提取最终结果
	return g.Compile(ctx, compose.WithGraphName("RAGGraph"))

}
