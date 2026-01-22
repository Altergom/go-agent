package sft

import (
	"context"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// SFTHandler 数据采集中间件
// 采集用户输入和模型输出数据作为样本为SFT做准备
type SFTHandler struct {
	callbacks.Handler
	AgentID string
}

func (h *SFTHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	// 只有 ChatModel 节点我们才记录输入
	if info.Component == components.ComponentOfChatModel {
		in := model.ConvCallbackInput(input)
		return context.WithValue(ctx, "sft_messages", in.Messages)
	}
	return ctx
}

func (h *SFTHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	// 核心过滤逻辑：只采集对话模型的最终输出
	if info.Component != components.ComponentOfChatModel {
		return ctx
	}

	out := model.ConvCallbackOutput(output)
	messages := ctx.Value("sft_messages").([]*schema.Message)

	go func() {
		sample := &Sample{
			ID:        uuid.New().String(),
			AgentID:   h.AgentID,
			NodeName:  info.Name,
			Component: string(info.Component),
			ModelType: info.Type,
			Messages:  append(messages, out.Message), // 存入完整上下文+模型回答
			Timestamp: time.Now().Unix(),
		}

		// 存储到数据仓库
		GetManager().SaveSample(sample)

		if err := Annotate(context.Background(), sample); err == nil {
			// 3. 更新已标注的样本
			GetManager().SaveSample(sample) // 再次调用 SaveSample 会追加或更新
		}
	}()

	return ctx
}
