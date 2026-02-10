package main

import (
	"context"
	"go-agent/api"
	"go-agent/config"
	"go-agent/flow"
	"go-agent/model/chat_model"
	"go-agent/rag/rag_flow"
	"go-agent/rag/rag_tools/db"
	"go-agent/rag/rag_tools/indexer"
	"go-agent/rag/rag_tools/retriever"
	"go-agent/tool/document"
	"go-agent/tool/memory"
	"go-agent/tool/sql_tools"
	"go-agent/tool/storage"
	"go-agent/tool/trace"
	"log"

	"github.com/cloudwego/eino-ext/devops"
)

func main() {
	var err error
	ctx := context.Background()

	// 初始化config
	config.Cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatal("警告: 未找到 .env 文件")
	}

	err = devops.Init(ctx)
	if err != nil {
		log.Printf("[eino dev] init failed, err=%v", err)
		return
	}

	// 初始化Redis
	err = storage.InitRedis(ctx)
	if err != nil {
		log.Printf("警告: Redis 初始化失败，将使用内存模式: %v", err)
	}
	defer storage.CloseRedis()

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
	err = trace.NewLangSmith()
	if err != nil {
		log.Fatalf("langsmith init fail: %v", err)
	}

	// 初始化 CozeLoop
	closeCoze, err := trace.NewCozeLoop(ctx)
	if err != nil {
		log.Fatalf("cozeloop init fail: %v", err)
	}
	defer closeCoze()

	// 初始化MCP
	err = sql_tools.InitMCPTools(ctx)
	if err != nil {
		log.Fatalf("MCP tools init fail: %v", err)
	}
	log.Println("MCP 工具连接已建立")

	// 预编译索引图
	err = rag_flow.InitIndexingGraph(ctx)
	if err != nil {
		log.Fatalf("IndexingGraph init fail: %v", err)
	}
	log.Println("IndexingGraph 已编译缓存")

	// 预编译RAG对话图
	memStore := memory.NewMemoryStore()
	taskModel, err := chat_model.GetChatModel(ctx, config.Cfg.ChatModelType)
	if err != nil {
		log.Fatalf("task model init fail: %v", err)
	}
	err = flow.InitRAGChatFlow(ctx, memStore, taskModel)
	if err != nil {
		log.Fatalf("RAGChatFlow init fail: %v", err)
	}
	log.Println("RAGChatFlow 已编译缓存")

	// 预编译全局图（使用Redis缓存）
	checkPointStore := storage.NewRedisCheckPointStore()
	err = flow.InitFinalGraph(ctx, checkPointStore)
	if err != nil {
		log.Fatalf("FinalGraph init fail: %v", err)
	}
	log.Println("FinalGraph 已编译缓存")

	api.Run()
}
