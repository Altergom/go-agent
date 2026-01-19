package tools

import (
	"context"
	"encoding/json"
	"go-agent/config"
	"strconv"

	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// Retriever 召回器 召回miLvus存储结果(就是MySQL的select)
var Retriever *milvus.Retriever

func NewRetriever(ctx context.Context) (*milvus.Retriever, error) {
	topK, err := strconv.Atoi(config.Cfg.MilvusConf.TopK)
	if err != nil || topK <= 0 {
		topK = 10
	}
	sp, _ := entity.NewIndexAUTOINDEXSearchParam(1)
	ret, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:       Milvus,
		Embedding:    Embedding,
		TopK:         topK,
		Collection:   config.Cfg.MilvusConf.CollectionName,
		VectorField:  "vector",
		OutputFields: []string{"id", "content", "metadata"},
		MetricType:   entity.COSINE,
		Sp:           sp,
		VectorConverter: func(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
			vecs := make([]entity.Vector, 0, len(vectors))
			for _, v := range vectors {
				v32 := make([]float32, len(v))
				for i, val := range v {
					v32[i] = float32(val)
				}
				vecs = append(vecs, entity.FloatVector(v32))
			}
			return vecs, nil
		},
		DocumentConverter: func(ctx context.Context, result client.SearchResult) ([]*schema.Document, error) {
			docs := make([]*schema.Document, result.IDs.Len())
			for i := range docs {
				docs[i] = &schema.Document{MetaData: map[string]any{}}
			}

			for _, field := range result.Fields {
				switch field.Name() {
				case "id":
					for i := range docs {
						id, err := result.IDs.GetAsString(i)
						if err != nil {
							return nil, err
						}
						docs[i].ID = id
					}
				case "content":
					for i := range docs {
						content, err := field.GetAsString(i)
						if err != nil {
							return nil, err
						}
						docs[i].Content = content
					}
				case "metadata":
					for i := range docs {
						raw, err := field.Get(i)
						if err != nil {
							return nil, err
						}
						if b, ok := raw.([]byte); ok {
							_ = json.Unmarshal(b, &docs[i].MetaData)
						}
					}
				}
			}

			// 写入相似度分数（Milvus 返回的是 distance）
			for i := range docs {
				if i < len(result.Scores) {
					distance := float64(result.Scores[i])
					docs[i].MetaData["distance"] = distance
					docs[i].WithScore(1 - distance)
				}
			}

			return docs, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}
