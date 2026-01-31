package api

import (
	"context"
	"fmt"
	"go-agent/SQL/sql_flow"
	"go-agent/flow"
	"io"
	"net/http"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type FinalGraphRequest struct {
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
}

type FinalGraphResponse struct {
	Query     string `json:"query"`
	Answer    string `json:"answer"`
	Status    string `json:"status"`
	SessionID string `json:"session_id,omitempty"`
}

var globalStore = sql_flow.NewInMemoryStore()
var interruptIDMap = make(map[string]string)

// FinalGraphInvoke 处理总控图的调用请求，支持流式输出
func FinalGraphInvoke(c *gin.Context) {
	var req FinalGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 构建并编译总控图
	runnable, err := flow.BuildFinalGraph(c, globalStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build final graph: " + err.Error()})
		return
	}

	// 调用总控图
	ctx := c.Request.Context()
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "default-session"
	}

	fmt.Printf(">>> FinalGraphInvoke: sessionID=%s, query=%s\n", sessionID, req.Query)

	var invokeCtx = context.WithValue(ctx, "session_id", sessionID)
	var input = req.Query

	// 检查是否有挂起的中断
	if id, ok := interruptIDMap[sessionID]; ok {
		// 只有当输入是确认指令时才 Resume
		if req.Query == "YES" || req.Query == "执行" || req.Query == "批准执行" {
			fmt.Printf(">>> Resume detected: sessionID=%s, interruptID=%s\n", sessionID, id)
			invokeCtx = compose.Resume(invokeCtx, id)
		}
	}

	// 核心修改：使用 Stream 而非 Invoke 获取流式读取器
	// 输入包装为 Message 数组对象
	reader, err := runnable.Stream(invokeCtx, []*schema.Message{schema.UserMessage(input)}, compose.WithCheckPointID(sessionID))
	if err != nil {
		// 处理中断逻辑（与之前一致）
		if info, ok := compose.ExtractInterruptInfo(err); ok {
			interruptID := info.InterruptContexts[0].ID
			interruptIDMap[sessionID] = interruptID // 保存 ID 用于后续恢复

			sql := info.InterruptContexts[0].Info.(string)
			c.JSON(http.StatusOK, gin.H{
				"status":       "need_approval",
				"answer":       fmt.Sprintf("检测到 SQL 执行请求，请确认是否执行？\n\n\n%s\n```", sql),
				"session_id":   sessionID,
				"interrupt_id": interruptID,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream graph: " + err.Error()})
		return
	}
	defer reader.Close()

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 成功执行后，如果是从 Resume 恢复的，清理 map
	delete(interruptIDMap, sessionID)

	// 使用 c.Stream 迭代读取流式数据
	c.Stream(func(w io.Writer) bool {
		chunk, err := reader.Recv()
		if err != nil {
			if err == io.EOF {
				return false // 流结束
			}
			fmt.Printf(">>> Stream Recv Error: %v\n", err)
			return false
		}

		// 因为总控图输出类型是 []*schema.Message，我们需要遍历发送
		for _, msg := range chunk {
			if msg.Content != "" {
				c.SSEvent("message", msg.Content)
			}
		}
		return true
	})
}
