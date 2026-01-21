package algorithm

import "math"

/**
 * BM25 (Best Matching 25) 算法实现。
 *
 * 背景：
 * BM25 是搜索引擎中衡量关键词与文档相关性的核心算法，被广泛应用于 Lucene、Elasticsearch 等系统中。
 * 它是 TF-IDF 算法的增强版，解决了 TF-IDF 中词频（TF）增长过快导致评分失真的问题。
 *
 * 核心组件：
 * 1. TF (Term Frequency - 词频): 关键词在文档中出现的次数。BM25 引入了饱和度机制，即词频增加到一定程度后，对总分的贡献趋于平缓。
 * 2. IDF (Inverse Document Frequency - 逆文档频率): 衡量关键词的稀有程度。越稀有的词，权重越高。
 * 3. Document Length Normalization (文档长度归一化): 较短文档中出现关键词的权重通常高于长文档。
 *
 * 公式参数：
 * - k1: 控制词频饱和度的参数。通常取值范围为 [1.2, 2.0]。k1 越大，词频对得分的影响越持久。
 * - b: 控制文档长度归一化程度的参数。取值范围为 [0, 1]。b=1 表示完全归一化，b=0 表示不考虑长度。通常取 0.75。
 */

type bm25 struct {
	docs     [][]string         // 原始文档集（已分词）
	avgdl    float64            // 所有文档的平均长度
	k1       float64            // 饱和度调节因子
	b        float64            // 长度归一化调节因子
	idf      map[string]float64 // 词项的逆文档频率映射
	docCount int                // 总文档数
}

// NewBM25 初始化 BM25 算法实例
func NewBM25(docs [][]string) *bm25 {
	bm := &bm25{
		docs:     docs,
		k1:       1.5,
		b:        0.75,
		docCount: len(docs),
		idf:      make(map[string]float64),
	}
	bm.calculateStats()

	return bm
}

// calculateStats 预计算全局统计信息：IDF 和平均文档长度
func (bm *bm25) calculateStats() {
	var totalLen int
	docFreq := make(map[string]int) // 词项出现的文档频率

	for _, doc := range bm.docs {
		totalLen += len(doc)
		// 统计每个词在多少个文档中出现过
		uniqueWords := make(map[string]bool)
		for _, word := range doc {
			uniqueWords[word] = true
		}
		for word := range uniqueWords {
			docFreq[word]++
		}
	}

	// 计算平均文档长度 (Average Document Length)
	bm.avgdl = float64(totalLen) / float64(bm.docCount)

	// 计算每个词的 IDF
	// 采用常用的 Lucene/ES 变体公式：log(1 + (N - n + 0.5) / (n + 0.5))
	for word, freq := range docFreq {
		bm.idf[word] = math.Log(1 + (float64(bm.docCount-freq)+0.5)/(float64(freq)+0.5))
	}
}

// Score 计算查询语句(query)与目标文档(doc)之间的相关性得分
// query: 查询词列表
// doc: 目标文档词列表
func (bm *bm25) Score(query, doc []string) float64 {
	var totalScore float64
	docLen := float64(len(doc))

	// 统计当前文档中的词频 (TF)
	tfMap := make(map[string]int)
	for _, word := range doc {
		tfMap[word]++
	}

	for _, qWord := range query {
		tf := float64(tfMap[qWord])
		if tf == 0 {
			continue
		}

		// 获取该词的全局 IDF，如果词从未在索引中出现，则忽略
		idf, exists := bm.idf[qWord]
		if !exists {
			continue
		}

		// BM25 核心公式部分：
		// 分子: tf * (k1 + 1)
		// 分母: tf + k1 * (1 - b + b * L / avgdl)
		// 其中 L 是文档长度，avgdl 是平均文档长度
		numerator := tf * (bm.k1 + 1)
		denominator := tf + bm.k1*(1-bm.b+bm.b*docLen/bm.avgdl)

		totalScore += idf * (numerator / denominator)
	}

	return totalScore
}
