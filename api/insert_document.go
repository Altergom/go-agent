package api

import (
	"context"
	"fmt"
	"go-agent/rag/compose"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/components/document"
	"github.com/gin-gonic/gin"
)

type InsertDocumentResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	DocumentIDs []string `json:"document_ids,omitempty"`
	ChunkCount  int      `json:"chunk_count,omitempty"`
}

// InsertDocument 处理文件上传并索引文档
func InsertDocument(c *gin.Context) {
	ctx := context.Background()

	// 1. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("获取上传文件失败: %v", err),
		})
		return
	}

	// 2. 验证文件大小（例如：限制 50MB）
	const maxFileSize = 50 << 20 // 50MB
	if file.Size > maxFileSize {
		c.JSON(400, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("文件大小超过限制 (最大 50MB), 当前: %.2f MB", float64(file.Size)/(1<<20)),
		})
		return
	}

	// 3. 创建临时目录保存文件
	tempDir := filepath.Join(os.TempDir(), "go-agent-uploads")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("创建临时目录失败: %v", err)
		c.JSON(500, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("创建临时目录失败: %v", err),
		})
		return
	}

	// 4. 生成唯一文件名
	timestamp := time.Now().UnixNano()
	fileName := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	tempFilePath := filepath.Join(tempDir, fileName)

	// 5. 保存上传的文件
	src, err := file.Open()
	if err != nil {
		c.JSON(500, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("打开上传文件失败: %v", err),
		})
		return
	}
	defer src.Close()

	dst, err := os.Create(tempFilePath)
	if err != nil {
		src.Close()
		c.JSON(500, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("创建临时文件失败: %v", err),
		})
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(tempFilePath) // 清理失败的文件
		c.JSON(500, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("保存文件失败: %v", err),
		})
		return
	}
	dst.Close()
	src.Close()

	// 6. 确保在处理完成后删除临时文件
	defer func() {
		if err := os.Remove(tempFilePath); err != nil {
			log.Printf("删除临时文件失败: %v, 文件路径: %s", err, tempFilePath)
		}
	}()

	// 7. 构建索引图
	indexingRunner, err := compose.BuildIndexingGraph(ctx)
	if err != nil {
		c.JSON(500, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("构建索引图失败: %v", err),
		})
		return
	}

	// 8. 创建文档源并执行索引
	docSource := document.Source{
		URI: tempFilePath,
	}

	// 9. 执行索引流程
	documentIDs, err := indexingRunner.Invoke(ctx, docSource)
	if err != nil {
		c.JSON(500, InsertDocumentResponse{
			Success: false,
			Message: fmt.Sprintf("索引文档失败: %v", err),
		})
		return
	}

	// 10. 返回成功响应
	c.JSON(200, InsertDocumentResponse{
		Success:     true,
		Message:     fmt.Sprintf("文档 '%s' 索引成功", file.Filename),
		DocumentIDs: documentIDs,
		ChunkCount:  len(documentIDs),
	})
}
