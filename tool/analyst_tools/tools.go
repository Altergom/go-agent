package analyst_tools

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ParsedData 解析后的结构化数据
type ParsedData struct {
	Columns    []string                 `json:"columns"`     // 列名
	Rows       []map[string]interface{} `json:"rows"`        // 数据行
	RowCount   int                      `json:"row_count"`   // 行数
	SampleRows string                   `json:"sample_rows"` // 样本行（用于展示）
}

// ParseSQLResult 解析SQL结果字符串
func ParseSQLResult(sqlResult string) (*ParsedData, error) {
	// 尝试解析为JSON格式
	var jsonData []map[string]interface{}
	if err := json.Unmarshal([]byte(sqlResult), &jsonData); err == nil {
		return parseFromJSON(jsonData)
	}

	// 如果不是JSON，尝试解析为表格格式
	return parseFromTable(sqlResult)
}

// parseFromJSON 从JSON数组解析
func parseFromJSON(jsonData []map[string]interface{}) (*ParsedData, error) {
	if len(jsonData) == 0 {
		return &ParsedData{
			Columns:    []string{},
			Rows:       []map[string]interface{}{},
			RowCount:   0,
			SampleRows: "",
		}, nil
	}

	// 提取列名
	columns := make([]string, 0)
	for col := range jsonData[0] {
		columns = append(columns, col)
	}
	sort.Strings(columns)

	// 生成样本行（前3行）
	sampleRows := ""
	sampleCount := 3
	if len(jsonData) < sampleCount {
		sampleCount = len(jsonData)
	}
	for i := 0; i < sampleCount; i++ {
		row := jsonData[i]
		rowStr := make([]string, 0, len(columns))
		for _, col := range columns {
			rowStr = append(rowStr, fmt.Sprintf("%v", row[col]))
		}
		sampleRows += strings.Join(rowStr, " | ") + "\n"
	}

	return &ParsedData{
		Columns:    columns,
		Rows:       jsonData,
		RowCount:   len(jsonData),
		SampleRows: sampleRows,
	}, nil
}

// parseFromTable 从表格文本解析
func parseFromTable(tableText string) (*ParsedData, error) {
	lines := strings.Split(strings.TrimSpace(tableText), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("空表格数据")
	}

	// 第一行是列名
	columns := strings.Split(lines[0], "|")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}

	// 解析数据行
	rows := make([]map[string]interface{}, 0)
	sampleRows := ""
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		values := strings.Split(line, "|")
		if len(values) != len(columns) {
			continue
		}

		row := make(map[string]interface{})
		for j, col := range columns {
			val := strings.TrimSpace(values[j])
			// 尝试转换为数字
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				row[col] = num
			} else {
				row[col] = val
			}
		}
		rows = append(rows, row)

		// 记录前3行作为样本
		if i <= 3 {
			sampleRows += strings.Join(values, " | ") + "\n"
		}
	}

	return &ParsedData{
		Columns:    columns,
		Rows:       rows,
		RowCount:   len(rows),
		SampleRows: sampleRows,
	}, nil
}

// ComputeStatistics 计算统计数据
func ComputeStatistics(data *ParsedData) (*Statistics, error) {
	if data.RowCount == 0 {
		return &Statistics{
			Count: 0,
		}, nil
	}

	// 找第一个数值列
	var values []float64
	for _, col := range data.Columns {
		values = make([]float64, 0, data.RowCount)
		for _, row := range data.Rows {
			if val, ok := row[col]; ok {
				if num, ok := val.(float64); ok {
					values = append(values, num)
				} else if numStr, ok := val.(string); ok {
					if num, err := strconv.ParseFloat(numStr, 64); err == nil {
						values = append(values, num)
					}
				}
			}
		}
		// 如果找到数值列，停止搜索
		if len(values) > 0 {
			break
		}
	}

	if len(values) == 0 {
		return &Statistics{
			Count: data.RowCount,
		}, nil
	}

	// 排序
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	// 计算统计量
	stats := &Statistics{
		Count: len(values),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
	}

	// 均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	stats.Mean = sum / float64(len(values))

	// 中位数
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		stats.Median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		stats.Median = sorted[mid]
	}

	// 标准差
	variance := 0.0
	for _, v := range values {
		diff := v - stats.Mean
		variance += diff * diff
	}
	stats.StdDev = math.Sqrt(variance / float64(len(values)))

	// 四分位数
	q1 := sorted[len(sorted)/4]
	q3 := sorted[len(sorted)*3/4]
	stats.Quartiles = []float64{sorted[0], q1, stats.Median, q3, sorted[len(sorted)-1]}

	return stats, nil
}

// RecommendChartType 推荐图表类型
func RecommendChartType(data *ParsedData) string {
	if data.RowCount == 0 {
		return "table"
	}

	// 如果行数很少，使用表格
	if data.RowCount <= 5 {
		return "table"
	}

	// 如果有时间序列列，使用折线图
	for _, col := range data.Columns {
		colLower := strings.ToLower(col)
		if strings.Contains(colLower, "time") || strings.Contains(colLower, "date") {
			return "line"
		}
	}

	// 如果有分类列+数值列，使用柱状图
	hasCategory := false
	hasNumeric := false
	for _, row := range data.Rows {
		for _, val := range row {
			if _, ok := val.(string); ok {
				hasCategory = true
			}
			if _, ok := val.(float64); ok {
				hasNumeric = true
			}
		}
		if hasCategory && hasNumeric {
			return "bar"
		}
	}

	// 默认使用柱状图
	return "bar"
}

// GenerateBarChart 生成柱状图配置
func GenerateBarChart(data *ParsedData) (map[string]interface{}, error) {
	if len(data.Columns) < 2 {
		return GenerateTableChart(data)
	}

	xData := make([]interface{}, 0, data.RowCount)
	yData := make([]interface{}, 0, data.RowCount)

	xCol := data.Columns[0]
	yCol := data.Columns[1]

	for _, row := range data.Rows {
		xData = append(xData, row[xCol])
		yData = append(yData, row[yCol])
	}

	return map[string]interface{}{
		"title": map[string]interface{}{
			"text": "数据分析柱状图",
		},
		"tooltip": map[string]interface{}{},
		"xAxis": map[string]interface{}{
			"type": "category",
			"data": xData,
		},
		"yAxis": map[string]interface{}{
			"type": "value",
		},
		"series": []map[string]interface{}{
			{
				"name": yCol,
				"type": "bar",
				"data": yData,
			},
		},
	}, nil
}

// GenerateLineChart 生成折线图配置
func GenerateLineChart(data *ParsedData) (map[string]interface{}, error) {
	if len(data.Columns) < 2 {
		return GenerateTableChart(data)
	}

	xData := make([]interface{}, 0, data.RowCount)
	yData := make([]interface{}, 0, data.RowCount)

	xCol := data.Columns[0]
	yCol := data.Columns[1]

	for _, row := range data.Rows {
		xData = append(xData, row[xCol])
		yData = append(yData, row[yCol])
	}

	return map[string]interface{}{
		"title": map[string]interface{}{
			"text": "数据趋势折线图",
		},
		"tooltip": map[string]interface{}{
			"trigger": "axis",
		},
		"xAxis": map[string]interface{}{
			"type": "category",
			"data": xData,
		},
		"yAxis": map[string]interface{}{
			"type": "value",
		},
		"series": []map[string]interface{}{
			{
				"name": yCol,
				"type": "line",
				"data": yData,
			},
		},
	}, nil
}

// GeneratePieChart 生成饼图配置
func GeneratePieChart(data *ParsedData) (map[string]interface{}, error) {
	if len(data.Columns) < 2 {
		return GenerateTableChart(data)
	}

	seriesData := make([]map[string]interface{}, 0, data.RowCount)
	nameCol := data.Columns[0]
	valueCol := data.Columns[1]

	for _, row := range data.Rows {
		seriesData = append(seriesData, map[string]interface{}{
			"name":  row[nameCol],
			"value": row[valueCol],
		})
	}

	return map[string]interface{}{
		"title": map[string]interface{}{
			"text": "数据分布饼图",
		},
		"tooltip": map[string]interface{}{
			"trigger": "item",
		},
		"series": []map[string]interface{}{
			{
				"type": "pie",
				"data": seriesData,
			},
		},
	}, nil
}

// GenerateTableChart 生成表格配置
func GenerateTableChart(data *ParsedData) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":    "table",
		"columns": data.Columns,
		"rows":    data.Rows,
	}, nil
}
