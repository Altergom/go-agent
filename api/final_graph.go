package api

import (
	"go-agent/flow"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FinalGraphRequest 总控图请求结构
type FinalGraphRequest struct {
	Query string `json:"query" binding:"required"`
}

// FinalGraphResponse 总控图响应结构
type FinalGraphResponse struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
}

// FinalGraphInvoke 处理总控图的调用请求
func FinalGraphInvoke(c *gin.Context) {
	var req FinalGraphRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 构建并编译总控图
	runnable, err := flow.BuildFinalGraph()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build final graph: " + err.Error()})
		return
	}

	// 调用总控图
	ctx := c.Request.Context()
	answer, err := runnable.Invoke(ctx, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invoke final graph: " + err.Error()})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, FinalGraphResponse{
		Query:  req.Query,
		Answer: answer,
	})
}
