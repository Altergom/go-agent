package main

import (
	"context"
	"go-agent/api"
	"go-agent/config"
	"go-agent/rag/rag_tools/db"
	"go-agent/rag/rag_tools/indexer"
	"go-agent/rag/rag_tools/retriever"
	"go-agent/tool/document"
	"go-agent/tool/trace"
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

	// 初始化检索器
	indexer.NewIndexer()

	// 初始化召回器
	retriever.NewRetriever()

	// 初始化解析器
	document.Parser, err = document.NewParser(ctx)
	if err != nil {
		log.Fatalf("parser init fail: %v", err)
	}

	// 初始化载入器
	document.Loader, err = document.NewLoader(ctx)
	if err != nil {
		log.Fatalf("loader init fail: %v", err)
	}

	// 初始化切分器
	document.Splitter, err = document.NewSplitter(ctx)
	if err != nil {
		log.Fatalf("splitter init fail: %v", err)
	}

	// 初始化langsmith
	//err = trace.NewLangSmith()
	//if err != nil {
	//	log.Fatalf("langsmith init fail: %v", err)
	//}

	// 初始化 CozeLoop
	closeCoze, err := trace.NewCozeLoop(ctx)
	if err != nil {
		log.Fatalf("cozeloop init fail: %v", err)
	}
	defer closeCoze()

	api.Run()
}
