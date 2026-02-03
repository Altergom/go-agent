package retriever

import (
	"context"
	"encoding/json"
	"go-agent/config"
	"go-agent/model/embedding_model"
	"go-agent/rag/rag_tools/db"
	"strconv"

	"github.com/cloudwego/eino-ext/components/retriever/es8"
	"github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

func initES() {
	registerRetriever("es", func(ctx context.Context) (retriever.Retriever, error) {
		var err error
		if db.ES == nil {
			if db.ES, err = db.NewES(); err != nil {
				return nil, err
			}
		}

		topK, _ := strconv.Atoi(config.Cfg.MilvusConf.TopK) // 复用 TopK 配置或新增 ES TopK

		emb, err := embedding_model.GetEmbeddingModel(context.Background(), config.Cfg.EmbeddingModelType)
		if err != nil {
			return nil, err
		}

		return es8.NewRetriever(ctx, &es8.RetrieverConfig{
			Client:    db.ES,
			Index:     config.Cfg.ESConf.Index,
			TopK:      topK,
			Embedding: emb,
			// 使用向量相似度搜索
			SearchMode: search_mode.SearchModeDenseVectorSimilarity(
				search_mode.DenseVectorSimilarityTypeCosineSimilarity,
				"content_vector",
			),
			ResultParser: func(ctx context.Context, hit types.Hit) (*schema.Document, error) {
				var src map[string]any
				if err := json.Unmarshal(hit.Source_, &src); err != nil {
					return nil, err
				}

				doc := &schema.Document{
					ID:       *hit.Id_,
					Content:  src["content"].(string),
					MetaData: make(map[string]any),
				}

				if meta, ok := src["metadata"].(map[string]any); ok {
					doc.MetaData = meta
				}

				if hit.Score_ != nil {
					doc.WithScore(float64(*hit.Score_))
				}

				return doc, nil
			},
		})
	})
}
