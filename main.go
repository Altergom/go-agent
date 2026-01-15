package main

import (
	"context"
	"go-agent/api"
	"go-agent/config"
	"go-agent/rag/tools"
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
	db, err := tools.NewMilvus(ctx)
	if err != nil {
		log.Fatalf("Milvus init fail: %v", err)
	}
	tools.Milvus = db
	defer tools.Milvus.Close()

	// 初始化embedder
	emb, err := tools.NewEmbedding(ctx)
	if err != nil {
		log.Fatalf("embedder init fail: %v", err)
	}
	tools.Embedding = emb

	ind, err := tools.NewIndexer(ctx)
	if err != nil {
		log.Fatalf("indexer init fail: %v", err)
	}
	tools.Indexer = ind

	ret, err := tools.NewRetriever(ctx)
	if err != nil {
		log.Fatalf("retriever init fail: %v", err)
	}
	tools.Retriever = ret

	parser, err := tools.NewParser(ctx)
	if err != nil {
		log.Fatalf("parser init fail: %v", err)
	}
	tools.Parser = parser

	api.Run()
}
