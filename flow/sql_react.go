package flow

import (
	"context"
	"fmt"
	"go-agent/model/chat_model"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type SQLFlowState struct {
	Query      string
	History    []*schema.Message
	CurrentSQL string
	IsValid    bool
	Step       string
}

const (
	Intent     = "Intent"
	ReAct      = "ReAct"
	MCPExecute = "MCPExecute"
)

func BuildSQLReact(ctx context.Context, cm model.BaseChatModel, mcpToolsNode *compose.ToolsNode) (compose.Runnable[string, *schema.Message], error) {
	g := compose.NewGraph[string, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *SQLFlowState {
			return &SQLFlowState{Step: Intent}
		}),
	)

	_ = g.AddLambdaNode(Intent, compose.InvokableLambda(func(ctx context.Context, input string) (output []*schema.Message, err error) {
		var prompt string
		err = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			if state.Step == "VALIDATE" {
				prompt = fmt.Sprintf("用户问题: %s\n生成的SQL: %s\n请判断该SQL是否能准确回答问题。只需回答 YES 或 NO。", state.Query, state.CurrentSQL)
			} else {
				state.Query = input
				prompt = fmt.Sprintf("识别以下用户的意图: %s", input)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		res, err := chat_model.CM.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
		if err != nil {
			return nil, err
		}

		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			if state.Step == "VALIDATE" {
				state.IsValid = res.Content == "YES"
			}
			return nil
		})

		return []*schema.Message{schema.UserMessage(res.Content)}, nil
	}))

	_ = g.AddChatModelNode(ReAct, chat_model.CM, compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *SQLFlowState) ([]*schema.Message, error) {
		state.Step = ReAct
		return append(state.History, schema.UserMessage(state.Query)), nil
	}),
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *SQLFlowState) (*schema.Message, error) {
			if len(out.ToolCalls) > 0 {
				state.CurrentSQL = out.ToolCalls[0].Function.Arguments
			}
			return out, nil
		}),
	)

	_ = g.AddToolsNode(MCPExecute, mcpToolsNode)

	_ = g.AddEdge(compose.START, Intent)

	// 意图识别后进入 ReAct
	_ = g.AddEdge(Intent, ReAct)

	// ReAct 后的分支路由
	_ = g.AddBranch(ReAct, compose.NewGraphBranch(func(ctx context.Context, out *schema.Message) (string, error) {
		var hasSQL bool
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			if state.CurrentSQL != "" {
				hasSQL = true
				state.Step = "VALIDATE"
			}
			return nil
		})

		if hasSQL {
			return Intent, nil
		}
		return compose.END, nil
	}, map[string]bool{Intent: true, compose.END: true}))

	// 意图识别后的分支路由
	_ = g.AddBranch(Intent, compose.NewGraphBranch(func(ctx context.Context, res string) (string, error) {
		var isValid bool
		var isValidationPhase bool
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			isValid = state.IsValid
			isValidationPhase = state.Step == "VALIDATE"
			return nil
		})

		if !isValidationPhase {
			return ReAct, nil
		}

		if isValid {
			return MCPExecute, nil
		}
		return ReAct, nil
	}, map[string]bool{ReAct: true, MCPExecute: true}))

	_ = g.AddEdge(MCPExecute, compose.END)

	return g.Compile(ctx)
}
