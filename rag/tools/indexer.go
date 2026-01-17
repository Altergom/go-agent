package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"go-agent/config"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// Indexer 检索器 把向量化的结果塞入milvus
var Indexer *milvus.Indexer

type floatSchema struct {
	ID       string    `json:"id" milvus:"name:id"`
	Content  string    `json:"content" milvus:"name:content"`
	Vector   []float32 `json:"vector" milvus:"name:vector"`
	Metadata []byte    `json:"metadata" milvus:"name:metadata"`
}

func NewIndexer(ctx context.Context) (*milvus.Indexer, error) {
	dim, err := getEmbeddingDim(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("embedding dim: %d", dim)

	indexer, err := milvus.NewIndexer(ctx, buildIndexerConfig(dim))
	if err != nil {
		// 自动处理 schema 不匹配：删除旧集合并重建
		if strings.Contains(err.Error(), "collection schema not match") {
			log.Printf("collection schema 不匹配，准备删除旧集合并重建: %s", config.Cfg.MilvusConf.CollectionName)
			_ = Milvus.ReleaseCollection(ctx, config.Cfg.MilvusConf.CollectionName)
			if dropErr := Milvus.DropCollection(ctx, config.Cfg.MilvusConf.CollectionName); dropErr != nil {
				return nil, fmt.Errorf("drop collection failed: %w", dropErr)
			}
			if waitErr := waitCollectionDropped(ctx, config.Cfg.MilvusConf.CollectionName, 15*time.Second); waitErr != nil {
				// 兜底：切换新集合名，避免启动失败
				newName := fmt.Sprintf("%s_%d", config.Cfg.MilvusConf.CollectionName, time.Now().Unix())
				log.Printf("旧集合仍存在，改用新集合: %s", newName)
				config.Cfg.MilvusConf.CollectionName = newName
			}
			indexer, err = milvus.NewIndexer(ctx, buildIndexerConfig(dim))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return indexer, nil
}

func buildIndexerConfig(dim int) *milvus.IndexerConfig {
	return &milvus.IndexerConfig{
		Client:     Milvus,
		Embedding:  Embedding,
		Collection: config.Cfg.MilvusConf.CollectionName,
		MetricType: milvus.COSINE,
		Fields: []*entity.Field{
			entity.NewField().
				WithName("id").
				WithDescription("document id").
				WithIsPrimaryKey(true).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(255),
			entity.NewField().
				WithName("vector").
				WithDescription("document vector").
				WithIsPrimaryKey(false).
				WithDataType(entity.FieldTypeFloatVector).
				WithDim(int64(dim)),
			entity.NewField().
				WithName("content").
				WithDescription("document content").
				WithIsPrimaryKey(false).
				WithDataType(entity.FieldTypeVarChar).
				WithMaxLength(65535),
			entity.NewField().
				WithName("metadata").
				WithDescription("document metadata").
				WithIsPrimaryKey(false).
				WithDataType(entity.FieldTypeJSON),
		},
		DocumentConverter: func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
			if len(vectors) != len(docs) {
				return nil, fmt.Errorf("vector size mismatch, docs=%d vectors=%d", len(docs), len(vectors))
			}

			rows := make([]interface{}, 0, len(docs))
			for i, doc := range docs {
				metadata, err := json.Marshal(doc.MetaData)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal metadata: %w", err)
				}
				vec := make([]float32, len(vectors[i]))
				for j, v := range vectors[i] {
					vec[j] = float32(v)
				}
				rows = append(rows, &floatSchema{
					ID:       doc.ID,
					Content:  doc.Content,
					Vector:   vec,
					Metadata: metadata,
				})
			}

			return rows, nil
		},
	}
}

func waitCollectionDropped(ctx context.Context, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		exists, err := Milvus.HasCollection(ctx, name)
		if err != nil {
			return fmt.Errorf("check collection failed: %w", err)
		}
		if !exists {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("collection still exists after drop: %s", name)
}

func getEmbeddingDim(ctx context.Context) (int, error) {
	if Embedding == nil {
		return 0, fmt.Errorf("embedding not initialized")
	}
	vecs, err := Embedding.EmbedStrings(ctx, []string{"dim"})
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding dim: %w", err)
	}
	if len(vecs) != 1 || len(vecs[0]) == 0 {
		return 0, fmt.Errorf("invalid embedding dim result")
	}
	return len(vecs[0]), nil
}
