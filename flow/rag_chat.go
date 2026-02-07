package flow

import (
	"context"
	"fmt"
	"go-agent/model/chat_model"
	"go-agent/rag/rag_tools"
	"go-agent/tool/memory"
	"sync"

	"go-agent/rag/rag_flow"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const RewritePrompt = `参考以下背景摘要和最近对话，将用户最后一次提问重写为一个独立的、适合向量检索的搜索语句。
背景摘要: %s
最近对话: %s
用户提问: %s
重写后的搜索语句（直接输出语句）: `

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

var (
	cachedRAGChatFlow  compose.Runnable[RAGChatInput, *schema.Message]
	ragChatFlowOnce    sync.Once
	ragChatFlowInitErr error
)

// InitRAGChatFlow 在应用启动时编译并缓存RAG对话图
func InitRAGChatFlow(ctx context.Context, store memory.Store, taskModel model.BaseChatModel) error {
	ragChatFlowOnce.Do(func() {
		cachedRAGChatFlow, ragChatFlowInitErr = buildRAGChatFlow(ctx, store, taskModel)
	})
	return ragChatFlowInitErr
}

func GetRAGChatFlow() (compose.Runnable[RAGChatInput, *schema.Message], error) {
	if cachedRAGChatFlow == nil {
		return nil, fmt.Errorf("RAGChatFlow 未初始化，请先调用 InitRAGChatFlow")
	}
	return cachedRAGChatFlow, nil
}

func buildRAGChatFlow(ctx context.Context, store memory.Store, taskModel model.BaseChatModel) (compose.Runnable[RAGChatInput, *schema.Message], error) {
	const (
		PreProcess = "preProcess"
		Rewrite    = "rewrite"
		Retrieve   = "retrieve"
		Chat       = "chat"
	)

	cm, _ := chat_model.GetChatModel(ctx, "rewrite")
	sm := &memory.Summarizer{Model: taskModel, MaxHistoryLen: 3}

	retrieverSubGraph, err := rag_flow.BuildRetrieverGraph(ctx)
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

			newQuery, err := rag_tools.Rewrite(ctx, state.Session.Summary, RewritePrompt, state.Session.History, in.Query, cm)
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

	// string -> []*schema.Message 转换节点（Rewrite 输出 string，Retrieve 子图入参是 []*schema.Message）
	_ = g.AddLambdaNode("QueryToMsgs", compose.InvokableLambda(func(ctx context.Context, query string) ([]*schema.Message, error) {
		return []*schema.Message{schema.UserMessage(query)}, nil
	}))

	_ = g.AddGraphNode(Retrieve, retrieverSubGraph)

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
	chat, err := chat_model.GetChatModel(ctx, "ark")
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(Chat, chat, compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *GraphState) ([]*schema.Message, error) {
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
	_ = g.AddEdge(Rewrite, "QueryToMsgs")
	_ = g.AddEdge("QueryToMsgs", Retrieve)
	_ = g.AddEdge(Retrieve, "ConstructMessages")
	_ = g.AddEdge("ConstructMessages", Chat)
	_ = g.AddEdge(Chat, compose.END)

	return g.Compile(ctx, compose.WithGraphName("RAGGraphOptimized"))
}
