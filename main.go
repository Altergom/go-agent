package main

import (
	"context"
	"go-agent/api"
	"go-agent/config"
	"go-agent/model/chat_model"
	"go-agent/model/embedding_model"
	"go-agent/rag/tools"
	"go-agent/rag/tools/db"
	"go-agent/rag/tools/indexer"
	"go-agent/rag/tools/retriever"
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

	// 初始化数据库
	db.Milvus, err = db.NewMilvus(ctx)
	if err != nil {
		log.Fatalf("Milvus init fail: %v", err)
	}
	defer db.Milvus.Close()

	// 初始化模型
	chat_model.CM, err = chat_model.NewChatModel(ctx)
	if err != nil {
		log.Fatalf("ChatModel init fail: %v", err)
	}

	// 初始化嵌入模型
	embedding_model.Embedding, err = embedding_model.NewEmbeddingModel(ctx)
	if err != nil {
		log.Fatalf("embedder init fail: %v", err)
	}

	// 初始化检索器
	indexer.NewIndexer()

	// 初始化召回器
	retriever.NewRetriever()

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
