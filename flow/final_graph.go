package flow

import (
	"context"
	"go-agent/SQL/sql_flow"
	"go-agent/model/chat_model"

	"github.com/cloudwego/eino/compose"
)

const (
	Intention_Recognition = "Intention_Recognition"
	Chat                  = "Chat"
	SQL                   = "SQL"
)

func BuildFinalGraph() (compose.Runnable[string, string], error) {
	ctx := context.Background()

	sql, _ := sql_flow.BuildSQLGraph(ctx)

	g := compose.NewGraph[string, string]()

	//_ = g.AddLambdaNode(Intention_Recognition, compose.InvokableLambda())
	_ = g.AddChatModelNode(Chat, chat_model.CM)
	_ = g.AddLambdaNode(SQL, compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
		out, err := sql.Invoke(ctx, input)
		if err != nil {
			return "", err
		}

		return out, nil
	}))

	_ = g.AddEdge(compose.START, Intention_Recognition)
	_ = g.AddEdge(Intention_Recognition, Chat)
	_ = g.AddEdge(Intention_Recognition, SQL)
	_ = g.AddEdge(SQL, compose.END)

	r, err := g.Compile(ctx, compose.WithGraphName("final_graph"))
	if err != nil {
		return nil, err
	}

	return r, nil
}
