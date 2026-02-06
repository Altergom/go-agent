package api

import (
	"context"
	"fmt"
	"go-agent/flow"
	"go-agent/tool/memory"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type FinalGraphResponse struct {
	Query     string `json:"query"`
	Answer    string `json:"answer"`
	Status    string `json:"status"`
	SessionID string `json:"session_id,omitempty"`
}

type sessionContext struct {
	InterruptID   string
	CheckPointID  string
	OriginalQuery string
	WaitingRefine bool
}

var sessionContextMap = make(map[string]*sessionContext)
var store = memory.NewInMemoryStore()

// FinalGraphInvoke 处理总控图的调用请求，支持流式输出
func FinalGraphInvoke(c *gin.Context) {
	var req flow.FinalGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "default-session"
	}

	ctx := c.Request.Context()

	fmt.Printf(">>> FinalGraphInvoke: sessionID=%s, query=%s\n", sessionID, req.Query)

	invokeCtx := context.WithValue(ctx, "session_id", sessionID)

	// 第一次判断：被打断进行批准拒绝
	if sc, ok := sessionContextMap[sessionID]; ok && sc.InterruptID != "" {
		upper := strings.ToUpper(strings.TrimSpace(req.Query))
		if upper == "YES" || upper == "执行" || upper == "批准执行" {
			// 如果是批准 使用保存的CheckPointID恢复
			fmt.Printf(">>> Approve: sessionID=%s, interruptID=%s\n", sessionID, sc.InterruptID)
			invokeCtx = compose.ResumeWithData(invokeCtx, sc.InterruptID, req.Query)

			runnable, err := flow.BuildFinalGraph(c, store)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build graph: " + err.Error()})
				return
			}

			reader, err := runnable.Stream(invokeCtx, req, compose.WithCheckPointID(sc.CheckPointID))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream graph: " + err.Error()})
				return
			}
			defer reader.Close()

			delete(sessionContextMap, sessionID)
			streamResponse(c, reader)
			return
		}

		// 如果是拒绝 返回补充信息提示进入refine
		fmt.Printf(">>> Reject: sessionID=%s\n", sessionID)
		sessionContextMap[sessionID] = &sessionContext{
			OriginalQuery: sc.OriginalQuery,
			WaitingRefine: true,
		}
		c.JSON(http.StatusOK, gin.H{
			"status":     "need_refinement",
			"answer":     "SQL已拒绝。请补充您的需求说明或表结构约束信息，我将根据您的补充重新生成SQL。",
			"session_id": sessionID,
		})
		return
	}

	// 处于refine时 合并用户补充信息
	if sc, ok := sessionContextMap[sessionID]; ok && sc.WaitingRefine {
		fmt.Printf(">>> Refine: sessionID=%s, original=%s, supplement=%s\n",
			sessionID, sc.OriginalQuery, req.Query)
		req.Query = fmt.Sprintf("%s（补充约束：%s）", sc.OriginalQuery, req.Query)
		delete(sessionContextMap, sessionID)
	}

	// 没有打断时
	runnable, err := flow.BuildFinalGraph(c, store)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build graph: " + err.Error()})
		return
	}

	// 每次执行使用不同的CheckPointID 防止记忆污染
	checkPointID := fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano())

	reader, err := runnable.Stream(invokeCtx, req, compose.WithCheckPointID(checkPointID))
	if err != nil {
		// 处理中断
		if info, ok := compose.ExtractInterruptInfo(err); ok {
			interruptID := info.InterruptContexts[0].ID
			sql := info.InterruptContexts[0].Info.(string)

			// 提取原始提问
			originalQuery := req.Query
			if idx := strings.Index(originalQuery, "（补充约束："); idx > 0 {
				originalQuery = originalQuery[:idx]
			}

			// 保存会话上下文
			sessionContextMap[sessionID] = &sessionContext{
				InterruptID:   interruptID,
				CheckPointID:  checkPointID,
				OriginalQuery: originalQuery,
			}

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

	delete(sessionContextMap, sessionID)
	streamResponse(c, reader)
}

func streamResponse(c *gin.Context, reader *schema.StreamReader[[]*schema.Message]) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 立即刷新响应头，防止前端超时
	c.Writer.WriteHeaderNow()
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		chunk, err := reader.Recv()
		if err != nil {
			if err == io.EOF {
				c.SSEvent("done", "EOF")
				return false
			}
			fmt.Printf(">>> Stream Recv Error: %v\n", err)
			c.SSEvent("error", err.Error())
			return false
		}
		for _, msg := range chunk {
			if msg.Content != "" {
				c.SSEvent("message", msg.Content)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
		return true
	})
}
