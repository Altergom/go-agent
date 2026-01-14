package main

import (
	"context"
	"go-agent/config"
	"go-agent/rag"
	"log"
)

func main() {
	// 初始化config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("警告: 未找到 .env 文件")
	}
	config.Cfg = cfg

	ctx := context.Background()

	// 初始化模型

	// 初始化数据库
	db, err := rag.NewMilvus(ctx)
	if err != nil {
		log.Fatalf("Milvus init fail: %v", err)
	}
	rag.Milvus = db
	defer rag.Milvus.Close()

	// 初始化embedder
	emb, err := rag.NewEmbedding(ctx)
	if err != nil {
		log.Fatalf("embedder init fail: %v", err)
	}
	rag.Embedding = emb

	ind, err := rag.NewIndexer(ctx)
	if err != nil {
		log.Fatalf("indexer init fail: %v", err)
	}
	rag.Indexer = ind

	ret, err := rag.NewRetriever(ctx)
	if err != nil {
		log.Fatalf("retriever init fail: %v", err)
	}
	rag.Retriever = ret
}
