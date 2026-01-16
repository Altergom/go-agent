package api

import (
	"context"
	"go-agent/model"
	"io"
	"net/http"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// ChatTestRequest 聊天测试请求结构
type ChatTestRequest struct {
	Question string            `json:"question" binding:"required"`
	History  []ChatTestMessage `json:"history,omitempty"`
}

// ChatTestMessage 聊天消息结构（用于前端传递）
type ChatTestMessage struct {
	Role    string `json:"role"` // "user" 或 "assistant"
	Content string `json:"content"`
}

// ChatTestResponse 聊天测试响应结构
type ChatTestResponse struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// ChatTest 测试聊天模型的常规输出
func ChatTest(c *gin.Context) {
	var req ChatTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 检查模型是否已初始化
	if model.CM == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ChatModel 未初始化"})
		return
	}

	ctx := context.Background()

	// 构建消息列表
	messages := make([]*schema.Message, 0)

	// 添加系统消息（可选）
	messages = append(messages, schema.SystemMessage("你是一个有用的AI助手。"))

	// 添加历史对话
	for _, msg := range req.History {
		if msg.Role == "user" {
			messages = append(messages, schema.UserMessage(msg.Content))
		} else if msg.Role == "assistant" {
			messages = append(messages, schema.AssistantMessage(msg.Content, []schema.ToolCall{}))
		}
	}

	// 添加当前问题
	messages = append(messages, schema.UserMessage(req.Question))

	// 调用模型的 Generate 方法
	response, err := model.CM.Generate(ctx, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate answer: " + err.Error()})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, ChatTestResponse{
		Question: req.Question,
		Answer:   response.Content,
	})
}

// ChatTestStream 测试聊天模型的流式输出
func ChatTestStream(c *gin.Context) {
	var req ChatTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 检查模型是否已初始化
	if model.CM == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ChatModel 未初始化"})
		return
	}

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	ctx := context.Background()

	// 构建消息列表
	messages := make([]*schema.Message, 0)

	// 添加系统消息（可选）
	messages = append(messages, schema.SystemMessage("你是一个有用的AI助手。"))

	// 添加历史对话
	for _, msg := range req.History {
		if msg.Role == "user" {
			messages = append(messages, schema.UserMessage(msg.Content))
		} else if msg.Role == "assistant" {
			messages = append(messages, schema.AssistantMessage(msg.Content, []schema.ToolCall{}))
		}
	}

	// 添加当前问题
	messages = append(messages, schema.UserMessage(req.Question))

	// 调用模型的 Stream 方法
	streamReader, err := model.CM.Stream(ctx, messages)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		flusher.Flush()
		return
	}

	// 发送开始事件
	c.SSEvent("message", gin.H{
		"type":    "start",
		"content": "",
	})
	flusher.Flush()

	// 读取流式数据
	for {
		msg, err := streamReader.Recv()

		if err != nil {
			if err == io.EOF {
				// 流结束
				c.SSEvent("message", gin.H{
					"type":    "end",
					"content": "",
				})
				flusher.Flush()
				return
			}
			// 发生错误
			c.SSEvent("error", gin.H{"error": err.Error()})
			flusher.Flush()
			return
		}

		// 发送接收到的增量内容
		if msg != nil && msg.Content != "" {
			c.SSEvent("message", gin.H{
				"type":    "data",
				"content": msg.Content,
			})
			flusher.Flush()
		}
	}
}
