package algorithm

import (
	"context"
	"sort"

	"github.com/cloudwego/eino/schema"
)

/**
 * RRF (Reciprocal Rank Fusion) 倒数排名融合算法。
 *
 * 背景：
 * 在检索系统中，我们经常使用多种检索方式（如向量检索和关键词检索）。
 * 向量检索返回的分数通常是余弦相似度或欧氏距离，而关键词检索（如 ES 的 BM25）返回的是相关性得分。
 * 这两种分数的量纲（Scale）完全不同，无法直接相加或比较。
 *
 * 原理：
 * RRF 算法不依赖于原始分数的绝对值，而是依赖于文档在各路检索结果中的“排名（Rank）”。
 * 其核心公式为：score = Σ (1 / (k + rank))
 * 其中 rank 是文档在某一路检索中的排名（从 1 开始），k 是一个平滑常数（通常取 60）。
 *
 * 优点：
 * 1. 无需对不同系统的分数进行归一化（Normalization）。
 * 2. 能够有效提升在多路检索中都排名靠前的文档权重。
 * 3. 实现简单且效果稳定。
 */

// RRFFusion 实现了 RRF 融合算法。
// inputs: 二维切片，第一层是召回的路径（如 inputs[0] 是 Milvus，inputs[1] 是 ES），第二层是每路召回的文档列表。
func RRFFusion(ctx context.Context, inputs [][]*schema.Document) ([]*schema.Document, error) {
	// k 是 RRF 算法中的平滑常数，默认取 60。
	// 较大的 k 值会减小高排名文档之间的得分差距。
	const k = 60

	// 用于存储每个文档 ID 对应的 RRF 累计得分
	docScores := make(map[string]float64)
	// 用于存储文档 ID 对应的文档对象，用于最后还原
	docMap := make(map[string]*schema.Document)

	// 遍历每一路召回的结果
	for _, docs := range inputs {
		// 遍历单路召回结果中的文档及其排名
		for rank, doc := range docs {
			if doc.ID == "" {
				continue
			}

			// 计算当前文档在当前路径下的 RRF 得分
			// rank 是从 0 开始的索引，物理排名需要 rank + 1
			score := 1.0 / float64(k+rank+1)

			// 累加得分
			docScores[doc.ID] += score

			// 如果文档在多路中出现，保留一份文档元数据即可
			if _, exists := docMap[doc.ID]; !exists {
				docMap[doc.ID] = doc
			}
		}
	}

	// 构造最终结果列表
	finalDocs := make([]*schema.Document, 0, len(docMap))
	for id, totalScore := range docScores {
		doc := docMap[id]
		// 将 RRF 得分回填到文档对象的 Score 属性中
		doc.WithScore(totalScore)
		finalDocs = append(finalDocs, doc)
	}

	// 按照 RRF 得分从高到低进行降序排序
	sort.Slice(finalDocs, func(i, j int) bool {
		return finalDocs[i].Score() > finalDocs[j].Score()
	})

	return finalDocs, nil
}
