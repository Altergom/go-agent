package retriever

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/retriever"
)

type RetrieverFactory func(ctx context.Context) (retriever.Retriever, error)

var retrieverRegistry = make(map[string]RetrieverFactory)

func NewRetriever() {
	initMilvus()
	initES()
}

// registerRetriever 用于具体 Provider 在 init 时注册自己
func registerRetriever(name string, factory RetrieverFactory) {
	retrieverRegistry[name] = factory
}

func GetRetriever(ctx context.Context, name string) (retriever.Retriever, error) {
	create, ok := retrieverRegistry[name]
	if !ok {
		return nil, fmt.Errorf("未注册的索引器类型: %s", name)
	}

	return create(ctx)
}
