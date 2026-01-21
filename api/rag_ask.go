package api

import (
	"context"
	"fmt"
	"go-agent/config"
	"go-agent/model/chat_model"
	"go-agent/rag/compose"
	"go-agent/tool/sft"
	"log"
	"net/http"
	"strconv"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	compose2 "github.com/cloudwego/eino/compose"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type RAGAskRequest struct {
	Query string `json:"query" binding:"required"`
}

type RAGAskResponse struct {
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
	Query          string  `json:"query,omitempty"`
	Answer         string  `json:"answer,omitempty"`
	RetrievedDocs  int     `json:"retrieved_docs,omitempty"`  // 检索到的文档数量
	MaxScore       float64 `json:"max_score,omitempty"`       // 最高相似度分数
	BelowThreshold bool    `json:"below_threshold,omitempty"` // 是否低于阈值
	Error          string  `json:"error,omitempty"`
}

// RAGAsk 处理 RAG 提问（从知识库检索并回答）
func RAGAsk(c *gin.Context) {

	// 获取用户问题
	var req RAGAskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RAGAskResponse{
			Success: false,
			Message: "获取用户问题失败",
			Error:   err.Error(),
		})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, RAGAskResponse{
			Success: false,
			Message: "用户问题不能为空",
		})
		return
	}

	// 获取基础 ctx
	baseCtx := c.Request.Context()

	// 初始化 SFT 处理器
	sftHandler := &sft.SFTHandler{AgentID: "rag_specialist"}

	// 构造统一的 SFT 上下文
	// 将用户原始问题存入，供 SFTHandler 最终汇总
	sftCtx := context.WithValue(baseCtx, "sft_input_msgs", []*schema.Message{schema.UserMessage(req.Query)})

	// 初始化 Callback 链路
	runInfo := &callbacks.RunInfo{
		Name:      "RAGFlow",
		Component: components.ComponentOfChatModel,
	}
	sftCtx = callbacks.InitCallbacks(sftCtx, runInfo, sftHandler)

	log.Printf("开始执行 RAG 检索，问题: %s", req.Query)

	// 构建检索图
	retrieverRunner, err := compose.BuildRetrieverGraph(sftCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RAGAskResponse{
			Success: false,
			Message: fmt.Sprintf("构建检索图失败: %v", err),
		})
		return
	}

	// 执行检索（输入 query string，输出 []*schema.Document）
	docs, err := retrieverRunner.Invoke(sftCtx, req.Query, compose2.WithCallbacks(sftHandler))
	if err != nil {
		c.JSON(http.StatusInternalServerError, RAGAskResponse{
			Success: false,
			Message: fmt.Sprintf("执行检索失败: %v", err),
			Error:   err.Error(),
		})
		return
	}

	log.Printf("检索成功，共找到 %d 个相关文档", len(docs))

	// 检查是否检索到文档
	if len(docs) == 0 {
		log.Printf("未检索到任何相关文档")
		c.JSON(http.StatusOK, RAGAskResponse{
			Success:        true,
			Message:        "知识库中未找到相关信息",
			Query:          req.Query,
			Answer:         "抱歉，知识库中不存在与您的问题相关的信息。",
			RetrievedDocs:  0,
			BelowThreshold: true,
		})
		return
	}

	// 打印检索到的文档（用于排查相似度问题）
	maxScore := 0.0
	for i, doc := range docs {
		score := doc.Score()
		// 兼容部分召回器把分数放在 metadata 的情况
		if score == 0 {
			if raw, ok := doc.MetaData["score"]; ok {
				if v, ok := raw.(float64); ok {
					score = v
				}
			}
			if raw, ok := doc.MetaData["distance"]; ok {
				if v, ok := raw.(float64); ok {
					score = 1 - v
				}
			}
		}
		if score > maxScore {
			maxScore = score
		}

		contentPreview := doc.Content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		log.Printf("召回文档[%d] ID=%s score=%.6f content=%s metadata=%v", i, doc.ID, score, contentPreview, doc.MetaData)
	}

	// 检查相似度阈值（使用最高相似度分数）
	similarityThreshold := 0.7 // 默认值
	if config.Cfg != nil && config.Cfg.MilvusConf.SimilarityThreshold != "" {
		if threshold, err := strconv.ParseFloat(config.Cfg.MilvusConf.SimilarityThreshold, 64); err == nil {
			similarityThreshold = threshold
		} else {
			log.Printf("警告: 相似度阈值配置无效，使用默认值 0.7: %v", err)
		}
	}

	log.Printf("最高相似度分数: %.4f, 阈值: %.4f", maxScore, similarityThreshold)

	// 7. 如果相似度低于阈值，返回提示信息
	if maxScore < similarityThreshold {
		log.Printf("相似度低于阈值，返回提示信息")
		c.JSON(http.StatusOK, RAGAskResponse{
			Success:        true,
			Message:        "检索到的文档相似度较低",
			Query:          req.Query,
			Answer:         "抱歉，知识库中不存在与您的问题高度相关的信息。",
			RetrievedDocs:  len(docs),
			MaxScore:       maxScore,
			BelowThreshold: true,
		})
		return
	}

	// 相似度达到阈值，格式化文档并生成回答
	var documentsText string
	for i, doc := range docs {
		score := doc.Score()
		if score == 0 {
			if raw, ok := doc.MetaData["score"]; ok {
				if v, ok := raw.(float64); ok {
					score = v
				}
			}
			if raw, ok := doc.MetaData["distance"]; ok {
				if v, ok := raw.(float64); ok {
					score = 1 - v
				}
			}
		}
		documentsText += fmt.Sprintf("文档 %d (相似度: %.4f):\n%s\n\n",
			i+1, score, doc.Content)
	}

	// 构建提示词并调用 ChatModel
	answer, err := generateRAGAnswer(sftCtx, req.Query, documentsText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RAGAskResponse{
			Success: false,
			Message: fmt.Sprintf("生成回答失败: %v", err),
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RAGAskResponse{
		Success:        true,
		Message:        "检索成功并生成回答",
		Query:          req.Query,
		Answer:         answer.Content,
		RetrievedDocs:  len(docs),
		MaxScore:       maxScore,
		BelowThreshold: false,
	})
}

// generateRAGAnswer 基于检索到的文档生成回答
func generateRAGAnswer(ctx context.Context, query, documentsText string) (*schema.Message, error) {
	// 检查模型是否已初始化
	if chat_model.CM == nil {
		return nil, fmt.Errorf("ChatModel 未初始化")
	}

	// 创建 ChatTemplate
	chatTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`你是一个有用的助手。请基于以下检索到的文档内容回答用户的问题。
如果文档中没有相关信息，请说明你不知道。

检索到的文档：
{documents}`),
		schema.UserMessage("{query}"),
	)

	// 格式化模板
	data := map[string]any{
		"query":     query,
		"documents": documentsText,
	}

	messages, err := chatTemplate.Format(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("格式化模板失败: %w", err)
	}

	// 调用 ChatModel 生成回答
	answer, err := chat_model.CM.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("生成回答失败: %w", err)
	}

	return answer, nil
}
