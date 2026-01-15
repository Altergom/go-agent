package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20

	// 添加文档上传路由
	r.POST("/api/document/insert", InsertDocument)
	err := r.Run("8080")
	if err != nil {
		log.Fatalf("run fail: %v", err)
	}
}
