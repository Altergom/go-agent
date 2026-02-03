package flow

import (
	"context"
	"go-agent/config"
	"go-agent/model/chat_model"
	"go-agent/rag/rag_flow"
	"go-agent/tool"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type SQLFlowState struct {
	History []*schema.Message `json:"history"`
}

const (
	SQL_Retrieve = "SQL_Retrieve"
	ToTplVar     = "ToTplVar"
	SQL_Tpl      = "SQL_Tpl"
	SQL_Model    = "SQL_Model"
	Approve      = "Approve"
	ToRefineVar  = "ToRefineVar"
)

func init() {
	schema.Register[*SQLFlowState]()
}

func BuildReactGraph(ctx context.Context) (*compose.Graph[[]*schema.Message, []*schema.Message], error) {
	g := compose.NewGraph[[]*schema.Message, []*schema.Message]()

	// RAG 检索：召回行业黑话、表结构信息等
	retriever, err := rag_flow.BuildRetrieverGraph(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddGraphNode(SQL_Retrieve, retriever)

	// 转换：[]*Document -> map[string]any (将检索结果包装为模板变量)
	_ = g.AddLambdaNode(ToTplVar, compose.InvokableLambda(func(ctx context.Context, input []*schema.Document) (map[string]any, error) {
		return nil, nil
	}))

	// SQL 模板节点
	sqlTemp := prompt.FromMessages(schema.FString,
		schema.SystemMessage(""),
		schema.UserMessage(""),
	)
	_ = g.AddChatTemplateNode(SQL_Tpl, sqlTemp)

	// SQL 生成模型 (ChatModel)
	chat, err := chat_model.GetChatModel(ctx, config.Cfg.ChatModelType)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(SQL_Model, chat)

	// 转换节点
	_ = g.AddLambdaNode(Trans_List, compose.InvokableLambda(tool.MsgToMsgs))

	// 用户审批节点 (Lambda + Interrupt)
	_ = g.AddLambdaNode(Approve, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output *schema.Message, err error) {
		return input, nil
	}))

	// 拒绝回流转换：*Message -> map[string]any (适配 SQL_Tpl)
	_ = g.AddLambdaNode(ToRefineVar, compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (output map[string]any, err error) {
		return map[string]any{"query": input.Content}, nil
	}))

	// 审批分支 (Branch)
	_ = g.AddBranch(Approve, compose.NewGraphBranch(func(ctx context.Context, input *schema.Message) (endNode string, err error) {
		return "", nil
	}, map[string]bool{
		ToRefineVar: true,
		Trans_List:  true,
	}))

	// 连线
	_ = g.AddEdge(compose.START, SQL_Retrieve)
	_ = g.AddEdge(SQL_Retrieve, ToTplVar)
	_ = g.AddEdge(ToTplVar, SQL_Tpl)
	_ = g.AddEdge(SQL_Tpl, SQL_Model)
	_ = g.AddEdge(SQL_Model, Approve)
	_ = g.AddEdge(ToRefineVar, SQL_Tpl)
	_ = g.AddEdge(Trans_List, compose.END)

	return g, nil
}
