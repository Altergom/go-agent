package flow

import (
	"context"
	"fmt"
	"strings"

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

func BuildSQLReact(ctx context.Context, cm model.BaseChatModel, mcpToolsNode *compose.ToolsNode) (compose.Runnable[string, string], error) {
	g := compose.NewGraph[string, string](
		compose.WithGenLocalState(func(ctx context.Context) *SQLFlowState {
			return &SQLFlowState{Step: Intent}
		}),
	)

	// 意图识别
	_ = g.AddLambdaNode(Intent, compose.InvokableLambda(func(ctx context.Context, input string) (output []*schema.Message, err error) {
		var prompt string
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			if state.Step == "VALIDATE" {
				prompt = fmt.Sprintf("用户问题: %s\n生成的SQL: %s\n请判断该SQL是否能准确回答问题。只需回答 YES 或 NO。", state.Query, state.CurrentSQL)
			} else {
				state.Query = input
				prompt = fmt.Sprintf("识别以下用户的意图: %s", input)
			}
			return nil
		})

		res, err := cm.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
		if err != nil {
			return nil, err
		}

		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			if state.Step == "VALIDATE" {
				state.IsValid = strings.Contains(strings.ToUpper(res.Content), "YES")
			}
			return nil
		})

		return []*schema.Message{schema.UserMessage(res.Content)}, nil
	}))

	// 类型转换节点
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

	// ReAct 智能体节点
	_ = g.AddChatModelNode(ReAct, cm,
		compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *SQLFlowState) ([]*schema.Message, error) {
			state.Step = ReAct
			return append(state.History, schema.UserMessage(state.Query)), nil
		}),
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, state *SQLFlowState) (*schema.Message, error) {
			if len(out.ToolCalls) > 0 {
				state.CurrentSQL = out.ToolCalls[0].Function.Arguments
				state.LastMessage = out
			}
			return out, nil
		}),
	)

	_ = g.AddLambdaNode(PreExecution, compose.InvokableLambda(func(ctx context.Context, _ []*schema.Message) (*schema.Message, error) {
		var lastMsg *schema.Message
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			lastMsg = state.LastMessage
			return nil
		})
		return lastMsg, nil
	}))

	// 工具执行节点
	_ = g.AddToolsNode(MCPExecute, mcpToolsNode)

	_ = g.AddEdge(compose.START, Intent)
	_ = g.AddEdge(Intent, ReAct)
	_ = g.AddEdge(MCPExecute, ReAct) // 关键：工具执行结果流回模型

	_ = g.AddBranch(ReAct, compose.NewGraphBranch(func(ctx context.Context, out *schema.Message) (string, error) {
		if len(out.ToolCalls) > 0 {
			return Trans, nil
		}

		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			if state.CurrentSQL != "" {
				state.Step = "VALIDATE"
			}
			return nil
		})

		return Trans, nil // 统一走 Trans 转换类型
	}, map[string]bool{Trans: true}))

	_ = g.AddBranch(Trans, compose.NewGraphBranch(func(ctx context.Context, content string) (string, error) {
		var isValidate bool
		_ = compose.ProcessState[*SQLFlowState](ctx, func(ctx context.Context, state *SQLFlowState) error {
			isValidate = state.Step == "VALIDATE"
			return nil
		})

		if isValidate {
			return Intent, nil
		}
		return compose.END, nil
	}, map[string]bool{Intent: true, compose.END: true}))

	_ = g.AddBranch(Intent, compose.NewGraphBranch(func(ctx context.Context, res []*schema.Message) (string, error) {
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
			return PreExecution, nil // 校验通过，去执行
		}

		return ReAct, nil // 校验不通过，回 ReAct 修正
	}, map[string]bool{ReAct: true, PreExecution: true}))

	_ = g.AddEdge(PreExecution, MCPExecute)

	return g.Compile(ctx)
}
