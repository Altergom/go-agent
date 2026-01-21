package compose

import (
	"context"
	"go-agent/config"
	"go-agent/rag/tools/retriever"
	"sort"
	"strconv"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BuildRetrieverGraph 仅负责检索，输入 query，输出文档列表
func BuildRetrieverGraph(ctx context.Context) (compose.Runnable[string, []*schema.Document], error) {
	const (
		MilvusRetriever = "MilvusRetriever"
		ESRetriever     = "ESRetriever"
		Reranker        = "Reranker"
	)

	g := compose.NewGraph[string, []*schema.Document]()

	// 构建召回节点
	milvus, err := retriever.NewRetriever(ctx, "milvus")
	if err != nil {
		return nil, err
	}
	es, err := retriever.NewRetriever(ctx, "es")
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(MilvusRetriever, milvus)
	_ = g.AddRetrieverNode(ESRetriever, es)
	_ = g.AddLambdaNode(Reranker, compose.InvokableLambda(RRF))

	// 构建节点指向
	_ = g.AddEdge(compose.START, MilvusRetriever)
	_ = g.AddEdge(compose.START, ESRetriever)
	_ = g.AddEdge(MilvusRetriever, Reranker)
	_ = g.AddEdge(ESRetriever, Reranker)
	_ = g.AddEdge(Reranker, compose.END)

	r, err := g.Compile(
		ctx,
		compose.WithGraphName("RAGRetriever"),
		compose.WithNodeTriggerMode(compose.AnyPredecessor),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// RRF 混合检索重排算法
// 具体讲解见algorithm/rrf.go
func RRF(ctx context.Context, inputs [][]*schema.Document) ([]*schema.Document, error) {
	const k = 60
	docScores := make(map[string]float64)
	// docMap ID-对象映射
	docMap := make(map[string]*schema.Document)

	// 遍历每一路召回的结果 (例如 inputs[0] 是 Milvus, inputs[1] 是 ES)
	for _, docs := range inputs {
		for rank, doc := range docs {
			id := doc.ID
			if id == "" {
				continue
			}

			score := 1.0 / float64(k+rank+1)
			docScores[id] += score

			// 如果 ID 重复，保留评分较高的那一版元数据
			if _, exists := docMap[id]; !exists {
				docMap[id] = doc
			}
		}
	}

	results := make([]*schema.Document, 0, len(docMap))
	for id, score := range docScores {
		doc := docMap[id]
		doc.WithScore(score)
		results = append(results, doc)
	}

	// 按照 RRF 分数从高到低排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})

	topk, _ := strconv.Atoi(config.Cfg.MilvusConf.TopK)
	return results[:topk], nil
}
