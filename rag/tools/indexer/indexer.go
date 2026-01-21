package indexer

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/indexer"
)

type IndexerFactory func(ctx context.Context) (indexer.Indexer, error)

var indexerRegistry = make(map[string]IndexerFactory)

func NewIndexer() {
	initMilvus()
	initES()
}

// registerIndexer 用于具体 Provider 在 init 时注册自己
func registerIndexer(name string, factory IndexerFactory) {
	indexerRegistry[name] = factory
}

func GetIndexer(ctx context.Context, name string) (indexer.Indexer, error) {
	create, ok := indexerRegistry[name]
	if !ok {
		return nil, fmt.Errorf("未注册的索引器类型: %s", name)
	}

	return create(ctx)
}
