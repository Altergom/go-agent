package sft

import (
	"context"
	"io"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// SpeculativeEngine 投机采样引擎
// 实现原理：利用小模型快速生成首屏内容，
// 同时利用大模型在后台进行核验与修正，确保最终回答的质量。
type SpeculativeEngine struct {
	DraftModel  model.BaseChatModel // 小模型：负责响应速度
	TargetModel model.BaseChatModel // 大模型：负责最终质量
}

func NewSpeculativeEngine(draft, target model.BaseChatModel) *SpeculativeEngine {
	return &SpeculativeEngine{
		DraftModel:  draft,
		TargetModel: target,
	}
}

// SpeculativeStream 实现投机流式输出
// 该方法会立即返回一个通道，前端通过订阅此通道可以获得“抢跑”的体验
func (e *SpeculativeEngine) SpeculativeStream(ctx context.Context, msgs []*schema.Message) (<-chan string, error) {
	outCh := make(chan string, 100)

	go func() {
		defer close(outCh)

		startTime := time.Now()
		var fullDraft string

		// 小模型抢跑
		// 用户会在这里几乎瞬间看到第一个字
		draftReader, err := e.DraftModel.Stream(ctx, msgs)
		if err != nil {
			outCh <- "Error: Draft model unavailable"
			return
		}

		for {
			chunk, err := draftReader.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			// 立即推送给前端
			outCh <- chunk.Content
			fullDraft += chunk.Content
		}

		// 记录小模型完成时间，用于分析 TTFT 收益
		_ = time.Since(startTime).Milliseconds()

		// 大模型核验与修正
		// 我们将小模型的输出作为上下文传给大模型进行“校对”
		verifyMsgs := append(msgs, schema.AssistantMessage(fullDraft, []schema.ToolCall{}))
		verifyMsgs = append(verifyMsgs, schema.UserMessage("请检查上述回答。如果基本正确，请回复 'OK'；如果有误或可以改进，请直接输出修正后的完整答案。"))

		targetReader, err := e.TargetModel.Stream(ctx, verifyMsgs)
		if err != nil {
			return
		}

		isFirstCorrection := true
		for {
			chunk, err := targetReader.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			// 如果大模型返回 OK，说明小模型表现完美，无需修正
			if chunk.Content == "OK" && isFirstCorrection {
				return
			}

			// 如果大模型开始输出非 OK 内容，说明需要修正
			if isFirstCorrection {
				// 发送特殊协议头，通知前端清空并替换为大模型内容
				outCh <- "[CORRECTION_START]"
				isFirstCorrection = false
			}
			outCh <- chunk.Content
		}
	}()

	return outCh, nil
}
