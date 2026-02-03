package api

import (
	"go-agent/rag/rag_tools/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MilvusCollectionsResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message,omitempty"`
	Collections []string `json:"collections,omitempty"`
}

type MilvusDropCollectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// ListMilvusCollections 返回所有 Milvus 集合名称
func ListMilvusCollections(c *gin.Context) {
	if db.Milvus == nil {
		c.JSON(http.StatusInternalServerError, MilvusCollectionsResponse{
			Success: false,
			Message: "Milvus 客户端未初始化",
		})
		return
	}

	collections, err := db.Milvus.ListCollections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, MilvusCollectionsResponse{
			Success: false,
			Message: "获取集合失败: " + err.Error(),
		})
		return
	}

	names := make([]string, 0, len(collections))
	for _, collection := range collections {
		names = append(names, collection.Name)
	}

	c.JSON(http.StatusOK, MilvusCollectionsResponse{
		Success:     true,
		Collections: names,
	})
}

// DeleteMilvusCollection 删除指定 Milvus 集合
func DeleteMilvusCollection(c *gin.Context) {
	collectionName := c.Param("name")
	if collectionName == "" {
		c.JSON(http.StatusBadRequest, MilvusDropCollectionResponse{
			Success: false,
			Message: "集合名称不能为空",
		})
		return
	}

	if db.Milvus == nil {
		c.JSON(http.StatusInternalServerError, MilvusDropCollectionResponse{
			Success: false,
			Message: "Milvus 客户端未初始化",
		})
		return
	}

	_ = db.Milvus.ReleaseCollection(c.Request.Context(), collectionName)
	if err := db.Milvus.DropCollection(c.Request.Context(), collectionName); err != nil {
		c.JSON(http.StatusInternalServerError, MilvusDropCollectionResponse{
			Success: false,
			Message: "删除集合失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MilvusDropCollectionResponse{
		Success: true,
		Message: "删除成功",
	})
}
