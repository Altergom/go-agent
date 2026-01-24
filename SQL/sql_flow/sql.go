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

func BuildSQLGraph(ctx context.Context) (compose.Runnable[string, string], error) {
	g := compose.NewGraph[string, string]()

	_ = g.AddLambdaNode(SQL_Generator, compose.InvokableLambda(tools.SQLGenerate))
	_ = g.AddLambdaNode(SQL_Executor, compose.InvokableLambda(tools.SQLExecute))

	_ = g.AddEdge(compose.START, SQL_Generator)
	_ = g.AddEdge(SQL_Generator, SQL_Executor)
	_ = g.AddEdge(SQL_Executor, compose.END)

	r, err := g.Compile(ctx,
		compose.WithInterruptBeforeNodes([]string{SQL_Executor}),
		compose.WithCheckPointStore(NewInMemoryStore()),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

type inMemoryStore struct {
	data map[string][]byte
}

func (s *inMemoryStore) Get(ctx context.Context, key string) ([]byte, bool, error) {
	d, ok := s.data[key]
	return d, ok, nil
}

func (s *inMemoryStore) Set(ctx context.Context, key string, val []byte) error {
	s.data[key] = val
	return nil
}

func NewInMemoryStore() compose.CheckPointStore {
	return &inMemoryStore{data: make(map[string][]byte)}
}
