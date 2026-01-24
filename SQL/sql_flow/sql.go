package sql_flow

import (
	"context"
	"go-agent/SQL/tools"

	"github.com/cloudwego/eino/compose"
)

const (
	SQL_Generator = "SQL_Generator"
	SQL_Executor  = "SQL_Executor"
)

func BuildSQLGraph() (compose.Runnable[string, string], error) {
	ctx := context.Background()
	g := compose.NewGraph[string, string]()

	_ = g.AddLambdaNode(SQL_Generator, compose.InvokableLambda(tools.SQLGenerate))
	_ = g.AddLambdaNode(SQL_Executor, compose.InvokableLambda(tools.SQLExecute))

	_ = g.AddEdge(compose.START, SQL_Generator)
	_ = g.AddEdge(SQL_Generator, SQL_Executor)
	_ = g.AddEdge(SQL_Executor, compose.END)

	r, err := g.Compile(ctx, compose.WithInterruptBeforeNodes([]string{SQL_Executor}))
	if err != nil {
		return nil, err
	}

	return r, nil
}
