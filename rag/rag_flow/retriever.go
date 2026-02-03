package rag_flow

import (
	"context"
	"go-agent/config"
	"go-agent/rag/rag_tools/retriever"
	"go-agent/tool"
	"sort"
	"strconv"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	MilvusRetriever = "MilvusRetriever"
	ESRetriever     = "ESRetriever"
	Reranker        = "Reranker"
	Trans_String    = "Trans_String"
)

// BuildRetrieverGraph 仅负责检索，输入 query，输出文档列表
func BuildRetrieverGraph(ctx context.Context) (*compose.Graph[[]*schema.Message, []*schema.Document], error) {
	g := compose.NewGraph[[]*schema.Message, []*schema.Document]()

	// 构建召回节点
	milvus, err := retriever.GetRetriever(ctx, "milvus")
	if err != nil {
		return nil, err
	}
	es, err := retriever.GetRetriever(ctx, "es")
	if err != nil {
		return nil, err
	}

	_ = g.AddRetrieverNode(MilvusRetriever, milvus, compose.WithOutputKey("milvus_retriever"))
	_ = g.AddRetrieverNode(ESRetriever, es, compose.WithOutputKey("es_retriever"))

	_ = g.AddLambdaNode(Trans_String, compose.InvokableLambda(tool.MsgsToQuery))

	_ = g.AddLambdaNode(Reranker, compose.InvokableLambda(func(ctx context.Context, input map[string]any) ([]*schema.Document, error) {
		// RRF 混合检索重排算法
		// 具体讲解见algorithm/rrf.go
		const k = 60
		docScores := make(map[string]float64)
		docMap := make(map[string]*schema.Document)

		for _, val := range input {
			docs, ok := val.([]*schema.Document)
			if !ok {
				continue
			}

			for rank, doc := range docs {
				id := doc.ID
				if id == "" {
					continue
				}

				// 标准 RRF 公式: 1 / (k + rank)
				score := 1.0 / float64(k+rank+1)
				docScores[id] += score

				// 如果文档在多路中重复出现，保留分数较高的原始对象
				if oldDoc, exists := docMap[id]; !exists || doc.Score() > oldDoc.Score() {
					docMap[id] = doc
				}
			}
		}

		// 将结果汇总并排序
		results := make([]*schema.Document, 0, len(docMap))
		for id, score := range docScores {
			doc := docMap[id]
			doc.WithScore(score)
			results = append(results, doc)
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].Score() > results[j].Score()
		})

		topk := 10
		if config.Cfg != nil && config.Cfg.MilvusConf.TopK != "" {
			if val, err := strconv.Atoi(config.Cfg.MilvusConf.TopK); err == nil {
				topk = val
			}
		}

		if len(results) > topk {
			results = results[:topk]
		}

		return results, nil
	}))

	// 构建节点指向
	_ = g.AddEdge(compose.START, Trans_String)
	_ = g.AddEdge(Trans_String, MilvusRetriever)
	_ = g.AddEdge(Trans_String, ESRetriever)
	_ = g.AddEdge(MilvusRetriever, Reranker)
	_ = g.AddEdge(ESRetriever, Reranker)
	_ = g.AddEdge(Reranker, compose.END)

	return g, nil
}
