package api

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20

	// 添加 CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静态文件服务 - 提供测试页面
	// 获取项目根目录（相对于当前工作目录）
	workDir, err := os.Getwd()
	if err != nil {
		log.Printf("获取工作目录失败: %v", err)
		workDir = "."
	}
	htmlPath := filepath.Join(workDir, "chat_test.html")
	ragIndexPath := filepath.Join(workDir, "rag_index.html")
	ragAskPath := filepath.Join(workDir, "rag_ask.html")
	hasRagAsk := false
	hasChatTest := false

	// 转换为绝对路径
	htmlPath, err = filepath.Abs(htmlPath)
	if err != nil {
		log.Printf("获取绝对路径失败: %v", err)
		htmlPath = filepath.Join(workDir, "chat_test.html")
	}
	ragIndexPath, err = filepath.Abs(ragIndexPath)
	if err != nil {
		log.Printf("获取绝对路径失败: %v", err)
		ragIndexPath = filepath.Join(workDir, "rag_index.html")
	}
	ragAskPath, err = filepath.Abs(ragAskPath)
	if err != nil {
		log.Printf("获取绝对路径失败: %v", err)
		ragAskPath = filepath.Join(workDir, "rag_ask.html")
	}

	// 检查文件是否存在
	if fileInfo, err := os.Stat(htmlPath); os.IsNotExist(err) {
		log.Printf("警告: 测试页面文件不存在: %s", htmlPath)
		r.GET("/chat_test.html", func(c *gin.Context) {
			c.String(404, "测试页面文件未找到: %s", htmlPath)
		})
	} else {
		log.Printf("找到测试页面文件: %s (大小: %d 字节)", htmlPath, fileInfo.Size())
		hasChatTest = true

		// 使用 File 方法直接提供文件
		r.GET("/chat_test.html", func(c *gin.Context) {
			c.File(htmlPath)
		})

		log.Printf("测试页面路由已注册: http://localhost:8080/chat_test.html")
	}

	// RAG 文档嵌入页面
	if fileInfo, err := os.Stat(ragIndexPath); os.IsNotExist(err) {
		log.Printf("警告: RAG 嵌入页面文件不存在: %s", ragIndexPath)
		r.GET("/rag_index.html", func(c *gin.Context) {
			c.String(404, "RAG 嵌入页面文件未找到: %s", ragIndexPath)
		})
	} else {
		log.Printf("找到 RAG 嵌入页面文件: %s (大小: %d 字节)", ragIndexPath, fileInfo.Size())
		r.GET("/rag_index.html", func(c *gin.Context) {
			c.File(ragIndexPath)
		})
		log.Printf("RAG 嵌入页面路由已注册: http://localhost:8080/rag_index.html")
	}

	// RAG 召回问答页面
	if fileInfo, err := os.Stat(ragAskPath); os.IsNotExist(err) {
		log.Printf("警告: RAG 问答页面文件不存在: %s", ragAskPath)
		r.GET("/rag_ask.html", func(c *gin.Context) {
			c.String(404, "RAG 问答页面文件未找到: %s", ragAskPath)
		})
	} else {
		log.Printf("找到 RAG 问答页面文件: %s (大小: %d 字节)", ragAskPath, fileInfo.Size())
		hasRagAsk = true
		r.GET("/rag_ask.html", func(c *gin.Context) {
			c.File(ragAskPath)
		})
		log.Printf("RAG 问答页面路由已注册: http://localhost:8080/rag_ask.html")
	}

	// 根路径优先指向 RAG 问答页面
	if hasRagAsk {
		r.GET("/", func(c *gin.Context) {
			c.Redirect(302, "/rag_ask.html")
		})
	} else if hasChatTest {
		r.GET("/", func(c *gin.Context) {
			c.Redirect(302, "/chat_test.html")
		})
	} else {
		r.GET("/", func(c *gin.Context) {
			c.String(200, "页面文件未找到。请确保 rag_ask.html 或 chat_test.html 在项目根目录: %s\n当前工作目录: %s", ragAskPath, workDir)
		})
	}

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "服务器运行正常",
		})
	})

	// 添加文档上传路由
	r.POST("/api/document/insert", InsertDocument)
	// RAG 文档嵌入（别名）
	r.POST("/api/rag/insert", InsertDocument)

	// 添加聊天测试路由
	r.POST("/api/chat/test", ChatGenerate)
	r.POST("/api/chat/test/stream", ChatStream)

	// RAG 召回问答
	r.POST("/api/rag/ask", RAGAsk)

	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("run fail: %v", err)
	}
}
