package flow

import (
	"context"
	"fmt"
	"go-agent/config"
	"go-agent/model/chat_model"
	"go-agent/tool"
	"go-agent/tool/sql_tools"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type FinalGraphRequest struct {
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
	SQL       string `json:"sql,omitempty"`  // 用于存储生成的 SQL
	Docs      string `json:"docs,omitempty"` // 用于存储检索到的表结构
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

var (
	cachedFinalGraph  compose.Runnable[FinalGraphRequest, []*schema.Message]
	finalGraphOnce    sync.Once
	finalGraphInitErr error
)

// InitFinalGraph 在应用启动时编译并缓存全局图
func InitFinalGraph(ctx context.Context, store compose.CheckPointStore) error {
	finalGraphOnce.Do(func() {
		cachedFinalGraph, finalGraphInitErr = buildFinalGraph(ctx, store)
	})
	return finalGraphInitErr
}

// GetFinalGraph 返回缓存的总控图
func GetFinalGraph() (compose.Runnable[FinalGraphRequest, []*schema.Message], error) {
	if cachedFinalGraph == nil {
		return nil, fmt.Errorf("FinalGraph 未初始化，请先调用 InitFinalGraph")
	}
	return cachedFinalGraph, nil
}

func buildFinalGraph(ctx context.Context, store compose.CheckPointStore) (compose.Runnable[FinalGraphRequest, []*schema.Message], error) {
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
	react, err := BuildReactGraph(ctx)
	if err != nil {
		return nil, fmt.Errorf("构建 React 子图失败: %w", err)
	}
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
		msg.Content = msg.Content[5:]
		return tool.MsgToSQLToolCall(ctx, msg)
	}))

	// MCP 执行节点
	tools, err := sql_tools.GetMCPTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 MCP 工具失败: %w", err)
	}
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

	return g.Compile(ctx, compose.WithCheckPointStore(store))
}
