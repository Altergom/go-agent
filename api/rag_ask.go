package api

import (
	"go-agent/flow"
	"go-agent/model/chat_model"
	"go-agent/tool/memory"

	"github.com/gin-gonic/gin"
)

var memStore = memory.NewInMemoryStore() // 全局记忆存储

func RAGAsk(c *gin.Context) {
	var req struct {
		Query     string `json:"query"`
		SessionID string `json:"session_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.SessionID == "" {
		req.SessionID = "default_user"
	}

	ragRunner, err := flow.BuildRAGChatFlow(c.Request.Context(), memStore, chat_model.CM)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 执行
	answer, err := ragRunner.Invoke(c.Request.Context(), flow.RAGChatInput{
		Query:     req.Query,
		SessionID: req.SessionID,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"answer":  answer.Content,
	})
}
