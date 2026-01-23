package api

import (
	"encoding/json"
	"fmt"
	"go-agent/flow"
	"go-agent/model/chat_model"
	"io"
	"log"

	"github.com/cloudwego/eino-ext/callbacks/langsmith"
	"github.com/gin-gonic/gin"
)

func RAGChatStream(c *gin.Context) {
	var req struct {
		Query     string `json:"query"`
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	if req.SessionID == "" {
		req.SessionID = "default_user"
	}

	ctx := langsmith.SetTrace(c.Request.Context(),
		langsmith.WithSessionName("GoAgent-RAG-Stream"),
		langsmith.AddTag("session:"+req.SessionID),
	)

	ragRunner, err := flow.BuildRAGChatFlow(ctx, memStore, chat_model.CM)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 使用 Stream 模式调用图
	stream, err := ragRunner.Stream(ctx, flow.RAGChatInput{
		Query:     req.Query,
		SessionID: req.SessionID,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	// 设置响应头为 SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	c.Stream(func(w io.Writer) bool {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return false
			}
			log.Printf("stream recv error: %v", err)
			return false
		}

		// 发送消息内容
		data, _ := json.Marshal(gin.H{
			"content": msg.Content,
		})
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		return true
	})
}
