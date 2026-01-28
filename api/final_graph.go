package api

import (
	"context"
	"fmt"
	"go-agent/SQL/sql_flow"
	"go-agent/flow"
	"net/http"

	"github.com/cloudwego/eino/compose"
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

// FinalGraphInvoke 处理总控图的调用请求
func FinalGraphInvoke(c *gin.Context) {
	var req FinalGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 构建并编译总控图
	runnable, err := flow.BuildFinalGraph(globalStore)
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

	// 将 sessionID 注入 context，以便后续节点能获取并传递给子图
	var invokeCtx = context.WithValue(ctx, "session_id", sessionID)
	var input = req.Query
	// 检查是否是确认执行的指令
	if req.Query == "YES" || req.Query == "执行" || req.Query == "批准执行" {
		if id, ok := interruptIDMap[sessionID]; ok {
			fmt.Printf(">>> Resume detected: sessionID=%s, interruptID=%s\n", sessionID, id)
			invokeCtx = compose.Resume(invokeCtx, id)
		} else {
			fmt.Printf(">>> Warning: Resume command received but no interruptID found for sessionID=%s\n", sessionID)
		}
	}

	answer, err := runnable.Invoke(invokeCtx, input, compose.WithCheckPointID(sessionID))
	if err != nil {
		// 检查是否是中断错误
		if info, ok := compose.ExtractInterruptInfo(err); ok {
			// --- 核心修正：递归寻找嵌套在任何层级中的中断 ID ---
			var findID func(*compose.InterruptInfo) string
			findID = func(i *compose.InterruptInfo) string {
				if i == nil {
					return ""
				}
				if len(i.InterruptContexts) > 0 {
					return i.InterruptContexts[0].ID
				}
				for _, sub := range i.SubGraphs {
					if id := findID(sub); id != "" {
						return id
					}
				}
				return ""
			}

			targetID := findID(info)

			if targetID != "" {
				interruptIDMap[sessionID] = targetID
				fmt.Printf(">>> Interrupt Saved: SessionID=%s, InterruptID=%s\n", sessionID, targetID)
			} else {
				fmt.Printf(">>> Warning: No InterruptID found in info or any sub-graphs for session %s\n", sessionID)
			}

			displaySQL := "未知 SQL"
			if subInfo, ok := info.SubGraphs[flow.SQL]; ok {
				if state, ok := subInfo.State.(*flow.SQLFlowState); ok {
					displaySQL = state.CurrentSQL
				}
			}

			c.JSON(http.StatusOK, FinalGraphResponse{
				Query:     req.Query,
				Answer:    fmt.Sprintf("检测到 SQL 执行请求，请确认是否执行？\n\n```sql\n%s\n```", displaySQL),
				Status:    "need_approval",
				SessionID: sessionID,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invoke final graph: " + err.Error()})
		return
	}

	// 成功执行后，如果是 Resume 成功的，清除 interrupt ID
	delete(interruptIDMap, sessionID)

	c.JSON(http.StatusOK, FinalGraphResponse{
		Query:     req.Query,
		Answer:    answer,
		Status:    "success",
		SessionID: sessionID,
	})
}
