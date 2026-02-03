package flow

import (
	"context"
	"fmt"
	"go-agent/config"
	"go-agent/model/chat_model"
	"go-agent/tool"
	"go-agent/tool/sql_tools"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type FinalGraphRequest struct {
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
}

const (
	Trans_List   = "Trans_List"
	Intent_Model = "Intent_Model"
	React        = "React"
	Chat         = "Chat"
	ChatToEnd    = "ChatToEnd"
	ToToolCall   = "ToToolCall"
	MCP          = "MCP"
)

func init() {
	schema.Register[*FinalGraphRequest]()
}

func BuildFinalGraph(ctx context.Context, store compose.CheckPointStore) (compose.Runnable[FinalGraphRequest, []*schema.Message], error) {
	g := compose.NewGraph[FinalGraphRequest, []*schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *FinalGraphRequest {
			return &FinalGraphRequest{}
		}),
	)

	// 意图识别模型
	_ = g.AddLambdaNode(Intent_Model, compose.InvokableLambda(func(ctx context.Context, input FinalGraphRequest) (*schema.Message, error) {
		_ = compose.ProcessState[*FinalGraphRequest](ctx, func(ctx context.Context, state *FinalGraphRequest) error {
			*state = input
			return nil
		})

		intentTemp := prompt.FromMessages(schema.FString,
			schema.SystemMessage("你是一个意图识别专家。请分析用户输入，如果是关于数据库查询、数据统计、报表需求，回答 'SQL'；否则回答 'Chat'。"),
			schema.UserMessage("{query}"),
		)
		cm, _ := chat_model.GetChatModel(context.Background(), config.Cfg.IntentModelType)
		output, err := intentTemp.Format(ctx, map[string]any{
			"query": input.Query,
		})
		if err != nil {
			return nil, err
		}
		return cm.Generate(ctx, output)
	}))
	//  React 子图
	react, _ := BuildReactGraph(ctx)
	_ = g.AddGraphNode(React, react, compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *FinalGraphRequest) ([]*schema.Message, error) {
		return []*schema.Message{schema.UserMessage(state.Query)}, nil
	}))

	// 聊天路径
	chat, err := chat_model.GetChatModel(ctx, config.Cfg.ChatModelType)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(Chat, chat, compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *FinalGraphRequest) ([]*schema.Message, error) {
		return []*schema.Message{schema.UserMessage(state.Query)}, nil
	}))

	_ = g.AddLambdaNode(ChatToEnd, compose.InvokableLambda(tool.MsgToMsgs))

	// 转换节点
	_ = g.AddLambdaNode(Trans_List, compose.InvokableLambda(tool.MsgToMsgs))

	// 意图分支
	_ = g.AddBranch(Trans_List, compose.NewGraphBranch(func(ctx context.Context, input []*schema.Message) (endNode string, err error) {
		content := strings.ToUpper(input[len(input)-1].Content)
		if strings.Contains(content, "SQL") {
			return React, nil
		}
		return Chat, nil
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

	// 连线
	_ = g.AddEdge(compose.START, Intent_Model)
	_ = g.AddEdge(Intent_Model, Trans_List)

	_ = g.AddEdge(React, ToToolCall)
	_ = g.AddEdge(ToToolCall, MCP)
	_ = g.AddEdge(MCP, compose.END)

	_ = g.AddEdge(Chat, ChatToEnd)
	_ = g.AddEdge(ChatToEnd, compose.END)

	return g.Compile(ctx,
		compose.WithCheckPointStore(store),
		compose.WithInterruptBeforeNodes([]string{
			fmt.Sprintf("%s/%s", React, Approve),
		}))
}
