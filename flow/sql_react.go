package flow

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type SQLFlowState struct {
	Query       string
	History     []*schema.Message
	CurrentSQL  string
	IsValid     bool
	Step        string
	LastMessage *schema.Message
}

const (
	Intent       = "Intent"
	ReAct        = "ReAct"
	MCPExecute   = "MCPExecute"
	Trans        = "Trans"
	PreExecution = "PreExecution"
)

func BuildSQLReact(ctx context.Context, cm model.BaseChatModel, mcpToolsNode *compose.ToolsNode, store compose.CheckPointStore) (compose.Runnable[string, string], error) {
	g := compose.NewGraph[string, string](
		compose.WithGenLocalState(func(ctx context.Context) *SQLFlowState {
			return &SQLFlowState{Step: "START"}
		}),
	)

	_ = g.AddLambdaNode(Intent, compose.InvokableLambda(func(ctx context.Context, input string) (output []*schema.Message, err error) {
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			state.Query = input
			return nil
		})
		return []*schema.Message{schema.UserMessage(input)}, nil
	}))

	_ = g.AddLambdaNode(Trans, compose.InvokableLambda(func(ctx context.Context, input any) (string, error) {
		switch v := input.(type) {
		case *schema.Message:
			return v.Content, nil
		case []*schema.Message:
			if len(v) > 0 {
				return v[len(v)-1].Content, nil
			}
		case string:
			return v, nil
		}
		return "", nil
	}))

	_ = g.AddChatModelNode(ReAct, cm,
		compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *SQLFlowState) ([]*schema.Message, error) {
			state.Step = ReAct
			if len(state.History) == 0 {
				state.History = append(state.History, schema.UserMessage(state.Query))
			}
			return state.History, nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *SQLFlowState) (*schema.Message, error) {
			state.History = append(state.History, out)
			if len(out.ToolCalls) > 0 {
				state.CurrentSQL = out.ToolCalls[0].Function.Arguments
				state.LastMessage = out
			}
			return out, nil
		}),
	)

	_ = g.AddLambdaNode(PreExecution, compose.InvokableLambda(func(ctx context.Context, _ any) (*schema.Message, error) {
		var lastMsg *schema.Message
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			lastMsg = state.LastMessage
			return nil
		})
		return lastMsg, nil
	}))

	_ = g.AddToolsNode(MCPExecute, mcpToolsNode, compose.WithStatePostHandler(func(ctx context.Context, out []*schema.Message, state *SQLFlowState) ([]*schema.Message, error) {
		state.History = append(state.History, out...)
		return out, nil
	}))

	_ = g.AddEdge(compose.START, Intent)
	_ = g.AddEdge(Intent, ReAct)
	_ = g.AddEdge(MCPExecute, ReAct)

	_ = g.AddBranch(ReAct, compose.NewGraphBranch(func(ctx context.Context, out *schema.Message) (string, error) {
		if len(out.ToolCalls) > 0 {
			return PreExecution, nil
		}
		return Trans, nil
	}, map[string]bool{PreExecution: true, Trans: true}))

	_ = g.AddEdge(PreExecution, MCPExecute)
	_ = g.AddEdge(Trans, compose.END)

	return g.Compile(ctx,
		compose.WithCheckPointStore(store),
		compose.WithInterruptBeforeNodes([]string{MCPExecute}),
	)
}
