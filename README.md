# go-agent

基于 Eino 构建的数据分析/处理智能体示例项目，目标是尽可能覆盖常见 LLM 技术栈，作为入门与实践参考。

当前实现重点在 RAG 方向，后续计划扩展 Multi-Agent 与 HITL。

## 功能特性

- 基础对话与流式对话接口
- RAG 文档入库与召回答复
- Milvus 向量数据库集成
- 简单的前端/HTML 测试页面
- 支持 Ark / OpenAI ChatModel（Embedding 当前使用 Ark）

## 目录结构

- `api/`：HTTP 接口与路由
- `config/`：配置与环境变量加载
- `model/`：聊天模型封装
- `rag/`：RAG 相关工具与编排
- `tool/`：通用工具函数
- `front/`：前端相关（预留/实验）
- `template/`：模板资源（预留/实验）
- `*.html`：本地测试页面

## 快速开始

### 1) 环境准备

- Go 1.20+（建议）
- Milvus（本地或 Docker）

Milvus 启动参考：`rag/README.md`

### 2) 配置环境变量

在项目根目录创建 `.env`，示例：

```env
# 模型类型（ChatModel 支持 ark/openai）
CHAT_MODEL_TYPE=ark
EMBEDDING_MODEL_TYPE=ark

# Ark 配置
ARK_KEY=your_api_key
ARK_CHAT_MODEL=your_chat_model
ARK_EMBEDDING_MODEL=your_embedding_model

# OpenAI 配置（当 CHAT_MODEL_TYPE=openai 时生效）
OPENAI_KEY=your_api_key
OPENAI_CHAT_MODEL=gpt-4

# Milvus 配置
MILVUS_ADDR=localhost:27017
MILVUS_USERNAME=
MILVUS_PASSWORD=
MILVUS_COLLECTION=eino_collection
SIMILARITY_THRESHOLD=
MILVUS_TOPK=10
```

### 3) 启动服务

```bash
go run .
```

服务默认监听 `:8080`。

## 测试页面

该页面在分支：`retrieve-fix`生效

启动后可直接访问

- `http://localhost:8080/chat_test.html`：基础对话测试
- `http://localhost:8080/rag_index.html`：RAG 文档入库
- `http://localhost:8080/rag_ask.html`：RAG 召回答复

根路径 `/` 会自动跳转到可用的测试页面。

## API 简要说明

- `POST /api/chat/test`：常规对话
- `POST /api/chat/test/stream`：流式对话
- `POST /api/document/insert`：文档入库
- `POST /api/rag/ask`：RAG 问答
- `GET /api/milvus/collections`：列出集合
- `DELETE /api/milvus/collections/:name`：删除集合

RAG 说明

RAG 构建、Milvus 启动与相关说明见 `rag/README.md`。

## 许可

本项目采用 `LICENSE` 中声明的开源协议。
