package flow

import (
	"context"
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

// GraphState 存储图运行时的状态
type GraphState struct {
	Input   RAGChatInput
	Session *memory.Session
	Query   string
	Docs    []*schema.Document
}

func BuildRAGChatFlow(ctx context.Context, store memory.Store, taskModel model.BaseChatModel) (compose.Runnable[RAGChatInput, *schema.Message], error) {
	const (
		PreProcess = "preProcess"
		Rewrite    = "rewrite"
		Retrieve   = "retrieve"
		Chat       = "chat"
	)

	qr := &rewriter.QueryRewriter{Model: taskModel}
	sm := &memory.Summarizer{Model: taskModel, MaxHistoryLen: 3}

	retrieverSubGraph, err := compose2.BuildRetrieverGraph(ctx)
	if err != nil {
		return nil, err
	}

	g := compose.NewGraph[RAGChatInput, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *GraphState {
			return &GraphState{}
		}),
	)

	_ = g.AddLambdaNode(PreProcess, compose.InvokableLambda(func(ctx context.Context, in RAGChatInput) (RAGChatInput, error) {
		_ = compose.ProcessState[*GraphState](ctx, func(ctx context.Context, state *GraphState) error {
			state.Input = in
			sess, _ := store.Get(ctx, in.SessionID)
			state.Session = sess
			return nil
		})
		return in, nil
	}))

	_ = g.AddLambdaNode(Rewrite, compose.InvokableLambda(func(ctx context.Context, in RAGChatInput) (string, error) {
		var query string
		_ = compose.ProcessState[*GraphState](ctx, func(ctx context.Context, state *GraphState) error {
			if len(state.Session.History) == 0 && state.Session.Summary == "" {
				state.Query = in.Query
				query = in.Query
				return nil
			}

			newQuery, err := qr.Rephrase(ctx, state.Session.Summary, state.Session.History, in.Query)
			if err != nil {
				state.Query = in.Query
				query = in.Query
				return nil
			}
			state.Query = newQuery
			query = newQuery
			return nil
		})
		return query, nil
	}))

	_ = g.AddLambdaNode(Retrieve, compose.InvokableLambda(func(ctx context.Context, query string) ([]*schema.Document, error) {
		docs, err := retrieverSubGraph.Invoke(ctx, query)
		if err != nil {
			return nil, err
		}
		_ = compose.ProcessState[*GraphState](ctx, func(ctx context.Context, state *GraphState) error {
			state.Docs = docs
			return nil
		})
		return docs, nil
	}))

	// []*schema.Document转为[]*schema.Message
	_ = g.AddLambdaNode("ConstructMessages", compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) ([]*schema.Message, error) {
		var messages []*schema.Message

		// 获取状态中的历史和原始输入
		err := compose.ProcessState[*GraphState](ctx, func(ctx context.Context, state *GraphState) error {
			if state.Session.Summary != "" {
				messages = append(messages, schema.SystemMessage("背景摘要: "+state.Session.Summary))
			}
			messages = append(messages, state.Session.History...)

			knowledge := "参考知识:\n"
			for _, doc := range docs {
				knowledge += doc.Content + "\n"
			}
			messages = append(messages, schema.UserMessage(knowledge+state.Input.Query))
			return nil
		})

		return messages, err
	}))

	// 对话生成
	_ = g.AddChatModelNode(Chat, chat_model.CM, compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *GraphState) ([]*schema.Message, error) {
		return in, nil
	}),
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *GraphState) (*schema.Message, error) {
			state.Session.History = append(state.Session.History, schema.UserMessage(state.Input.Query))
			state.Session.History = append(state.Session.History, out)

			go func(s *memory.Session) {
				bgCtx := context.Background()
				_ = sm.Compress(bgCtx, s)
				_ = store.Save(bgCtx, s.ID, s)
			}(state.Session)

			return out, nil
		}),
	)

	_ = g.AddEdge(compose.START, PreProcess)
	_ = g.AddEdge(PreProcess, Rewrite)
	_ = g.AddEdge(Rewrite, Retrieve)
	_ = g.AddEdge(Retrieve, "ConstructMessages")
	_ = g.AddEdge("ConstructMessages", Chat)
	_ = g.AddEdge(Chat, compose.END)

	return g.Compile(ctx, compose.WithGraphName("RAGGraphOptimized"))
}
