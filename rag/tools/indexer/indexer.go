package indexer

import (
	"context"
	"fmt"
	"go-agent/config"

	"github.com/cloudwego/eino/components/indexer"
)

type IndexerFactory func(ctx context.Context) (indexer.Indexer, error)

var indexerRegistry = make(map[string]IndexerFactory)

// Indexer 全局索引器接口
var Indexer indexer.Indexer

// NewIndexer 根据配置查找并创建对应的索引器实例
func NewIndexer(ctx context.Context) (indexer.Indexer, error) {
	initMilvus()
	initES()
	dbType := config.Cfg.VectorDBType
	create, ok := indexerRegistry[dbType]
	if !ok {
		return nil, fmt.Errorf("未注册的索引器类型: %s", dbType)
	}

	return create(ctx)
}

// registerIndexer 用于具体 Provider 在 init 时注册自己
func registerIndexer(name string, factory IndexerFactory) {
	indexerRegistry[name] = factory
}
