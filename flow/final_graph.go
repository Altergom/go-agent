package flow

import (
	"context"
	"fmt"
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

func BuildFinalGraph() (compose.Runnable[string, string], error) {
	ctx := context.Background()

	sqlAgent, err := BuildSQLReact(ctx, chat_model.CM, &compose.ToolsNode{})
	if err != nil {
		return nil, fmt.Errorf("build sql agent fail: %w", err)
	}

	g := compose.NewGraph[string, string](
		compose.WithGenLocalState(func(ctx context.Context) *string {
			return new(string)
		}),
	)

	_ = g.AddLambdaNode(Intention_Recognition, compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
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
		var query string
		_ = compose.ProcessState[*string](ctx, func(ctx context.Context, state *string) error {
			query = *state
			return nil
		})
		return sqlAgent.Invoke(ctx, query)
	}))

	_ = g.AddBranch(Intention_Recognition, compose.NewGraphBranch(func(ctx context.Context, intent string) (string, error) {
		return intent, nil
	}, map[string]bool{Chat: true, SQL: true}))

	_ = g.AddEdge(compose.START, Intention_Recognition)
	_ = g.AddEdge(Chat, compose.END)
	_ = g.AddEdge(SQL, compose.END)

	return g.Compile(ctx, compose.WithGraphName("final_graph"))
}
