package flow

import (
	"context"
	"fmt"
	"go-agent/SQL/sql_tools"
	"go-agent/model/chat_model"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	Intention_Recognition = "Intention_Recognition"
	Chat                  = "Chat"
	SQL                   = "SQL"
)

func BuildFinalGraph(store compose.CheckPointStore) (compose.Runnable[string, string], error) {
	ctx := context.Background()

	// 获取 MCP 工具并初始化 ToolsNode
	mcpTools, err := sql_tools.GetMCPTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("get mcp tools fail: %w", err)
	}

	// 构造 ToolsNode
	mcpToolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: mcpTools,
	})
	if err != nil {
		return nil, fmt.Errorf("new tools node fail: %w", err)
	}

	// 将工具绑定到聊天模型
	if binder, ok := chat_model.CM.(interface {
		BindTools([]*schema.ToolInfo) error
	}); ok {
		toolInfos := make([]*schema.ToolInfo, 0, len(mcpTools))
		for _, t := range mcpTools {
			info, err := t.Info(ctx)
			if err != nil {
				return nil, fmt.Errorf("get tool info fail: %w", err)
			}
			toolInfos = append(toolInfos, info)
		}
		_ = binder.BindTools(toolInfos)
	}

	sqlAgent, err := BuildSQLReact(ctx, chat_model.CM, mcpToolsNode, store)
	if err != nil {
		return nil, fmt.Errorf("build sql agent fail: %w", err)
	}

	g := compose.NewGraph[string, string](
		compose.WithGenLocalState(func(ctx context.Context) *string {
			return new(string)
		}),
	)

	_ = g.AddLambdaNode(Intention_Recognition, compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		// 优先检查是否是恢复流程 (无论是自身中断还是子节点中断)
		if isResume, _, _ := compose.GetResumeContext[any](ctx); isResume {
			fmt.Println(">>> Intention_Recognition: Resume flow detected, routing to SQL")
			return SQL, nil
		}

		// 如果是确认指令，直接跳转到 SQL 节点，且不覆盖之前的 Query 状态
		if input == "YES" || input == "执行" || input == "批准执行" {
			fmt.Printf(">>> Intention_Recognition: Confirmation '%s' detected, routing to SQL\n", input)
			return SQL, nil
		}

		// 只有新请求才更新状态
		_ = compose.ProcessState[*string](ctx, func(ctx context.Context, state *string) error {
			*state = input
			return nil
		})

		prompt := fmt.Sprintf("请分析用户输入的意图。如果是关于数据库查询、数据统计、报表需求，回答 'SQL'；否则回答 'Chat'。\n用户输入: %s\n意图:", input)
		res, err := chat_model.CM.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
		if err != nil {
			return Chat, nil
		}

		if strings.Contains(strings.ToUpper(res.Content), "SQL") {
			return SQL, nil
		}
		return Chat, nil
	}))

	_ = g.AddLambdaNode(Chat, compose.InvokableLambda(func(ctx context.Context, _ string) (string, error) {
		var query string
		_ = compose.ProcessState[*string](ctx, func(ctx context.Context, state *string) error {
			query = *state
			return nil
		})
		res, err := chat_model.CM.Generate(ctx, []*schema.Message{schema.UserMessage(query)})
		if err != nil {
			return "", err
		}
		return res.Content, nil
	}))

	_ = g.AddLambdaNode(SQL, compose.InvokableLambda(func(ctx context.Context, _ string) (string, error) {
		sessionID, _ := ctx.Value("session_id").(string)
		var query string
		_ = compose.ProcessState[*string](ctx, func(ctx context.Context, state *string) error {
			query = *state
			return nil
		})

		// 显式传递 CheckPointID 给子图，并增加后缀以防与主图冲突
		// 同时透传 context (包含 Resume 信号)
		return sqlAgent.Invoke(ctx, query, compose.WithCheckPointID(sessionID+"_sql"))
	}))

	_ = g.AddBranch(Intention_Recognition, compose.NewGraphBranch(func(ctx context.Context, intent string) (string, error) {
		return intent, nil
	}, map[string]bool{Chat: true, SQL: true}))

	_ = g.AddEdge(compose.START, Intention_Recognition)
	_ = g.AddEdge(Chat, compose.END)
	_ = g.AddEdge(SQL, compose.END)

	return g.Compile(ctx,
		compose.WithGraphName("final_graph"),
		compose.WithCheckPointStore(store),
	)
}
