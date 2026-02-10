package analyst_tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// AnalysisResult 分析结果
type AnalysisResult struct {
	TextAnalysis string      `json:"text_analysis"` // 文字分析报告
	ChartConfig  interface{} `json:"chart_config"`  // ECharts配置
	Statistics   *Statistics `json:"statistics"`    // 统计数据
}

// Statistics 统计数据
type Statistics struct {
	Mean      float64   `json:"mean"`      // 均值
	Median    float64   `json:"median"`    // 中位数
	StdDev    float64   `json:"std_dev"`   // 标准差
	Min       float64   `json:"min"`       // 最小值
	Max       float64   `json:"max"`       // 最大值
	Count     int       `json:"count"`     // 总数
	Quartiles []float64 `json:"quartiles"` // 四分位数
}

// GenerateTextAnalysis 生成文字分析报告
func GenerateTextAnalysis(ctx context.Context, data *ParsedData, stats *Statistics, cm model.BaseChatModel) (string, error) {
	// 构造提示词
	prompt := fmt.Sprintf(`你是一位专业的数据分析师。请基于以下数据和统计信息，生成一份简洁的分析报告。

数据行数: %d
列名: %v
统计信息:
- 均值: %.2f
- 中位数: %.2f
- 标准差: %.2f
- 范围: [%.2f, %.2f]

数据样例:
%s

请生成一份200字以内的分析报告，包含：
1. 数据总体概况
2. 关键指标解读
3. 发现的趋势或异常
`,
		data.RowCount,
		data.Columns,
		stats.Mean,
		stats.Median,
		stats.StdDev,
		stats.Min,
		stats.Max,
		data.SampleRows,
	)

	// 调用LLM生成分析
	messages := []*schema.Message{
		{Role: "user", Content: prompt},
	}

	response, err := cm.Generate(ctx, messages)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// GenerateChartConfig 生成ECharts配置
func GenerateChartConfig(data *ParsedData) (interface{}, error) {
	// 推荐图表类型
	chartType := RecommendChartType(data)

	// 根据图表类型生成配置
	switch chartType {
	case "bar":
		return GenerateBarChart(data)
	case "line":
		return GenerateLineChart(data)
	case "pie":
		return GeneratePieChart(data)
	case "table":
		return GenerateTableChart(data)
	default:
		return GenerateTableChart(data)
	}
}
