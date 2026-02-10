package flow

import (
	"context"
	"fmt"
	"go-agent/model/chat_model"
	"go-agent/tool/analyst_tools"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	ParseDataNode      = "ParseData"
	AnalyzeDataNode    = "AnalyzeData"
	GenerateReportNode = "GenerateReport"
	GenerateChartNode  = "GenerateChart"
	MergeResultNode    = "MergeResult"
)

// AnalystState Graph状态
type AnalystState struct {
	SQLResult    string                    `json:"sql_result"`
	ParsedData   *analyst_tools.ParsedData `json:"parsed_data"`
	Statistics   *analyst_tools.Statistics `json:"statistics"`
	TextAnalysis string                    `json:"text_analysis"`
	ChartConfig  interface{}               `json:"chart_config"`
}

func init() {
	schema.Register[*AnalystState]()
}

// BuildAnalystGraph 构建数据分析流程图
func BuildAnalystGraph(ctx context.Context) (*compose.Graph[[]*schema.Message, *analyst_tools.AnalysisResult], error) {
	g := compose.NewGraph[[]*schema.Message, *analyst_tools.AnalysisResult](
		compose.WithGenLocalState(func(ctx context.Context) *AnalystState {
			return &AnalystState{}
		}),
	)

	cm, err := chat_model.GetChatModel(ctx, "analyst")
	if err != nil {
		return nil, err
	}

	// 解析数据节点
	_ = g.AddLambdaNode(ParseDataNode, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*analyst_tools.ParsedData, error) {
		// 从ToolCall结果中提取SQL结果
		if len(input) == 0 {
			return nil, fmt.Errorf("没有输入数据")
		}

		var sqlResult string
		for _, msg := range input {
			sqlResult = msg.Content
			break
		}

		if sqlResult == "" {
			return nil, fmt.Errorf("未找到SQL结果")
		}

		// 保存到状态
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			state.SQLResult = sqlResult
			return nil
		})

		// 解析数据
		data, err := analyst_tools.ParseSQLResult(sqlResult)
		if err != nil {
			return nil, err
		}

		// 保存到状态
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			state.ParsedData = data
			return nil
		})

		return data, nil
	}))

	// 分析数据节点（计算统计）
	_ = g.AddLambdaNode(AnalyzeDataNode, compose.InvokableLambda(func(ctx context.Context, input *analyst_tools.ParsedData) (*analyst_tools.Statistics, error) {
		stats, err := analyst_tools.ComputeStatistics(input)
		if err != nil {
			return nil, err
		}

		// 保存到状态
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			state.Statistics = stats
			return nil
		})

		return stats, nil
	}))

	// 生成文字报告节点
	_ = g.AddLambdaNode(GenerateReportNode, compose.InvokableLambda(func(ctx context.Context, input *analyst_tools.Statistics) (string, error) {
		var data *analyst_tools.ParsedData
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			data = state.ParsedData
			return nil
		})

		if data == nil {
			return "", fmt.Errorf("未找到解析数据")
		}

		report, err := analyst_tools.GenerateTextAnalysis(ctx, data, input, cm)
		if err != nil {
			return "", err
		}

		// 保存到状态
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			state.TextAnalysis = report
			return nil
		})

		return report, nil
	}))

	// 生成图表配置节点
	_ = g.AddLambdaNode(GenerateChartNode, compose.InvokableLambda(func(ctx context.Context, input *analyst_tools.Statistics) (interface{}, error) {
		var data *analyst_tools.ParsedData
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			data = state.ParsedData
			return nil
		})

		if data == nil {
			return nil, fmt.Errorf("未找到解析数据")
		}

		chart, err := analyst_tools.GenerateChartConfig(data)
		if err != nil {
			return nil, err
		}

		// 保存到状态
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, state *AnalystState) error {
			state.ChartConfig = chart
			return nil
		})

		return chart, nil
	}))

	// 合并结果节点
	_ = g.AddLambdaNode(MergeResultNode, compose.InvokableLambda(func(ctx context.Context, input interface{}) (*analyst_tools.AnalysisResult, error) {
		var state *AnalystState
		_ = compose.ProcessState[*AnalystState](ctx, func(ctx context.Context, s *AnalystState) error {
			state = s
			return nil
		})

		return &analyst_tools.AnalysisResult{
			TextAnalysis: state.TextAnalysis,
			ChartConfig:  state.ChartConfig,
			Statistics:   state.Statistics,
		}, nil
	}))

	// 构建节点连接
	_ = g.AddEdge(compose.START, ParseDataNode)
	_ = g.AddEdge(ParseDataNode, AnalyzeDataNode)
	_ = g.AddEdge(AnalyzeDataNode, GenerateReportNode)
	_ = g.AddEdge(AnalyzeDataNode, GenerateChartNode)
	_ = g.AddEdge(GenerateReportNode, MergeResultNode)
	_ = g.AddEdge(GenerateChartNode, MergeResultNode)
	_ = g.AddEdge(MergeResultNode, compose.END)

	return g, nil
}
