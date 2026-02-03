package indexer

import (
	"context"
	"go-agent/config"
	"go-agent/model/embedding_model"
	"go-agent/rag/rag_tools/db"

	"github.com/cloudwego/eino-ext/components/indexer/es8"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
)

func initES() {
	registerIndexer("es", func(ctx context.Context) (indexer.Indexer, error) {
		var err error
		if db.ES == nil {
			if db.ES, err = db.NewES(); err != nil {
				return nil, err
			}
		}

		emb, err := embedding_model.GetEmbeddingModel(context.Background(), config.Cfg.EmbeddingModelType)
		if err != nil {
			return nil, err
		}
		return es8.NewIndexer(ctx, &es8.IndexerConfig{
			Client:    db.ES,
			Index:     config.Cfg.ESConf.Index,
			Embedding: emb,
			DocumentToFields: func(ctx context.Context, doc *schema.Document) (map[string]es8.FieldValue, error) {
				// 定义文档如何映射到 ES 字段
				return map[string]es8.FieldValue{
					"content": {
						Value:    doc.Content,
						EmbedKey: "content_vector", // 将内容向量化并存入 content_vector 字段
					},
					"metadata": {
						Value: doc.MetaData,
					},
				}, nil
			},
		})
	})
}
