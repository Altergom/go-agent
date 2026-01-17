package main

import (
	"context"
	"go-agent/api"
	"go-agent/config"
	"go-agent/model"
	"go-agent/rag/tools"
	"log"
)

func main() {
	var err error
	ctx := context.Background()

	// 初始化config
	config.Cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatal("警告: 未找到 .env 文件")
	}

	// 初始化模型
	model.CM, err = model.NewChatModel(ctx)
	if err != nil {
		log.Fatalf("ChatModel init fail: %v", err)
	}

	// 初始化数据库
	tools.Milvus, err = tools.NewMilvus(ctx)
	if err != nil {
		log.Fatalf("Milvus init fail: %v", err)
	}
	defer tools.Milvus.Close()

	// 初始化嵌入模型
	tools.Embedding, err = tools.NewEmbedding(ctx)
	if err != nil {
		log.Fatalf("embedder init fail: %v", err)
	}

	// 初始化检索器
	tools.Indexer, err = tools.NewIndexer(ctx)
	if err != nil {
		log.Fatalf("indexer init fail: %v", err)
	}

	// 初始化召回器
	tools.Retriever, err = tools.NewRetriever(ctx)
	if err != nil {
		log.Fatalf("retriever init fail: %v", err)
	}

	// 初始化解析器
	tools.Parser, err = tools.NewParser(ctx)
	if err != nil {
		log.Fatalf("parser init fail: %v", err)
	}

	// 初始化载入器
	tools.Loader, err = tools.NewLoader(ctx)
	if err != nil {
		log.Fatalf("loader init fail: %v", err)
	}

	// 初始化切分器
	tools.Splitter, err = tools.NewSplitter(ctx)
	if err != nil {
		log.Fatalf("splitter init fail: %v", err)
	}

	api.Run()
}
