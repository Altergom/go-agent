package sql_flow

import (
	"context"
	"go-agent/SQL/sql_tools"
	"go-agent/model/chat_model"
	"go-agent/rag/rag_flow"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	SQL_Generator = "SQL_Generator"
	SQL_Executor  = "SQL_Executor"
	Retriever     = "Retriever"
	Rewriter      = "Rewriter"
)

const RewritePrompt = `参考以下背景摘要和最近对话，把用户提问和检索后的结果相结合，重写出更符合业务场景DDL命名规范的生成SQL语句的提示词。
背景摘要: %s
用户提问: %s
召回结果: %s
重写后的生成语句（直接输出语句）: `

func init() {
	schema.RegisterName[*sql_tools.SQLState]("SQLState")
}

func BuildSQLGraph(ctx context.Context) (compose.Runnable[string, string], error) {
	g := compose.NewGraph[string, string](
		compose.WithGenLocalState(func(ctx context.Context) *sql_tools.SQLState {
			return &sql_tools.SQLState{}
		}),
	)

	retriever, err := rag_flow.BuildRetrieverGraph(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(SQL_Generator, compose.InvokableLambda(sql_tools.SQLGenerate))
	_ = g.AddLambdaNode(SQL_Executor, compose.InvokableLambda(sql_tools.SQLExecute))
	_ = g.AddLambdaNode(Retriever, compose.InvokableLambda(func(ctx context.Context, input string) (output []*schema.Document, err error) {
		output, err = retriever.Invoke(ctx, input)
		if err != nil {
			return nil, err
		}

		return output, nil
	}))
	// 再加一个重写节点，根据召回的规则重写传来的意图
	_ = g.AddLambdaNode(Retriever, compose.InvokableLambda(func(ctx context.Context, input []*schema.Document) (output string, err error) {
		cm, _ := chat_model.GetChatModel(ctx, "rewriter")
		newPrompt, err := sql_tools.Rewrite(ctx, RewritePrompt, input, "", cm)
		if err != nil {
			return "", err
		}

		return newPrompt, nil
	}))

	_ = g.AddEdge(compose.START, Retriever)
	_ = g.AddEdge(Retriever, Rewriter)
	_ = g.AddEdge(Rewriter, SQL_Generator)
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

// TODO 这里可以换成redis
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
