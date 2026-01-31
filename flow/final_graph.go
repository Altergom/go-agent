package flow

import (
	"context"
	"fmt"
	"go-agent/SQL/sql_tools"
	"go-agent/model/chat_model"
	"go-agent/tool"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	ToVar        = "ToVar"
	Intent_Tpl   = "Intent_Tpl"
	Intent_Model = "Intent_Model"
	IntentBranch = "IntentBranch"
	React        = "React"
	Chat         = "Chat"
	ToToolCall   = "ToToolCall"
	MCP          = "MCP"
	ResultClean  = "ResultClean"
)

func BuildFinalGraph(ctx context.Context, store compose.CheckPointStore) (compose.Runnable[[]*schema.Message, []*schema.Message], error) {
	g := compose.NewGraph[[]*schema.Message, []*schema.Message]()

	// 类型转换：[]*Message -> map[string]any
	_ = g.AddLambdaNode(ToVar, compose.InvokableLambda(tool.MsgToMap))

	intentTemp := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个意图识别专家。请分析用户输入，如果是关于数据库查询、数据统计、报表需求，回答 'SQL'；否则回答 'Chat'。"),
		schema.UserMessage("{query}"),
	)
	// 意图识别模板
	_ = g.AddChatTemplateNode(Intent_Tpl, intentTemp)

	// 意图识别模型
	_ = g.AddChatModelNode(Intent_Model, chat_model.CM)

	//  React 子图
	react, _ := BuildReactGraph(ctx)
	_ = g.AddGraphNode(React, react)

	// 聊天路径
	_ = g.AddChatModelNode(Chat, chat_model.CM)

	// 意图分支
	_ = g.AddBranch(Intent_Model, compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (endNode string, err error) {
		return "", nil
	}, map[string]bool{
		React: true,
		Chat:  true,
	}))

	// 类型转换：[]*Message -> *Message
	_ = g.AddLambdaNode(ToToolCall, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (*schema.Message, error) {
		msg, err := tool.MsgsToMsg(ctx, input)
		if err != nil {
			return nil, err
		}
		return tool.MsgToSQLToolCall(ctx, msg)
	}))

	// MCP 执行节点
	tools, _ := sql_tools.GetMCPTool(ctx)
	mcpTool, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: tools,
	})
	if err != nil {
		return nil, err
	}
	_ = g.AddToolsNode(MCP, mcpTool)

	// 结果清理与格式化
	_ = g.AddLambdaNode(ResultClean, compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (output []*schema.Message, err error) {
		return input, nil
	}))

	// 连线
	_ = g.AddEdge(compose.START, ToVar)
	_ = g.AddEdge(ToVar, Intent_Tpl)
	_ = g.AddEdge(Intent_Tpl, Intent_Model)
	_ = g.AddEdge(Intent_Model, IntentBranch)

	_ = g.AddEdge(IntentBranch, React)
	_ = g.AddEdge(IntentBranch, Chat)

	_ = g.AddEdge(React, ToToolCall)
	_ = g.AddEdge(ToToolCall, MCP)
	_ = g.AddEdge(MCP, ResultClean)
	_ = g.AddEdge(ResultClean, compose.END)

	_ = g.AddEdge(Chat, compose.END)

	return g.Compile(ctx,
		compose.WithCheckPointStore(store),
		compose.WithInterruptBeforeNodes([]string{
			fmt.Sprintf("%s/%s", React, Approve),
		}))
}
