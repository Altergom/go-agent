package algorithm

import "math"

type bm25 struct {
	docs     [][]string         // 文档
	avgdl    float64            // 文档平均长度
	k1       float64            // 影响速度调节因子 1.2-2.0
	b        float64            // 惩罚调节因子 0.75
	idf      map[string]float64 // 逆文档频率
	docCount int                // 文档数量
}

func newBM25(docs [][]string) *bm25 {
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

// calculateStats 计算IDF和平均长度
// 计算每个词的IDF(逆文档频率)
// IDF是计算一个词在各个文档出现的次数，目的是判断某一个词是否是常用词(如：“的” “是”)，从而找到稀有的关键词
// 计算方法为 对(文档总数/出现过的文档数)取对数
func (bm *bm25) calculateStats() {
	var totalLen int
	docFreq := make(map[string]int)

	for _, doc := range bm.docs {
		totalLen += len(doc)
		// 统计词频
		uniqueWords := make(map[string]bool)
		for _, word := range doc {
			uniqueWords[word] = true
		}
		for word := range uniqueWords {
			docFreq[word]++
		}
	}

	bm.avgdl = float64(totalLen) / float64(bm.docCount)

	for word, freq := range docFreq {
		bm.idf[word] = math.Log(float64(bm.docCount-freq)+0.5) / (float64(freq) + 0.5)
	}
}

// score 计算关键词在文档的得分
// tf(词频)：某个关键词在文档出现次数
// k1：调节影响效果的因子，k1越大这个关键字的得分就越高
// b： 调节文档长度影响效果的因子，b越大得分受长度影响越严重
func (bm *bm25) score(query, doc []string) float64 {
	var score float64
	docLen := float64(len(doc))

	// 统计词频
	tfMap := make(map[string]int)
	for _, word := range doc {
		tfMap[word]++
	}

	for _, qWord := range query {
		tf := float64(tfMap[qWord])
		if tf == 0 {
			continue
		}
		idf := bm.idf[qWord]
		numerator := tf * (bm.k1 + 1)
		denominator := tf + bm.k1*(1-bm.b*docLen/bm.avgdl)

		score += idf * (numerator / denominator)
	}

	return score
}
