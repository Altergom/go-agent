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

	// 转换为绝对路径
	htmlPath, err = filepath.Abs(htmlPath)
	if err != nil {
		log.Printf("获取绝对路径失败: %v", err)
		htmlPath = filepath.Join(workDir, "chat_test.html")
	}

	// 检查文件是否存在
	if fileInfo, err := os.Stat(htmlPath); os.IsNotExist(err) {
		log.Printf("警告: 测试页面文件不存在: %s", htmlPath)
		// 添加一个提示路由
		r.GET("/", func(c *gin.Context) {
			c.String(200, "测试页面文件未找到。请确保 chat_test.html 文件在项目根目录: %s\n当前工作目录: %s", htmlPath, workDir)
		})
		r.GET("/chat_test.html", func(c *gin.Context) {
			c.String(404, "测试页面文件未找到: %s", htmlPath)
		})
	} else {
		log.Printf("找到测试页面文件: %s (大小: %d 字节)", htmlPath, fileInfo.Size())

		// 使用 File 方法直接提供文件
		r.GET("/chat_test.html", func(c *gin.Context) {
			c.File(htmlPath)
		})

		// 根路径重定向到测试页面
		r.GET("/", func(c *gin.Context) {
			c.Redirect(302, "/chat_test.html")
		})

		log.Printf("测试页面路由已注册: http://localhost:8080/chat_test.html")
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

	// 添加聊天测试路由
	r.POST("/api/chat/test", ChatTest)
	r.POST("/api/chat/test/stream", ChatTestStream)

	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("run fail: %v", err)
	}
}
