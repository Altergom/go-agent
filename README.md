# Go-Agent

<p align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/cloudwego/eino?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Framework-Eino-blue?style=flat-square" alt="Framework">
  <img src="https://img.shields.io/badge/License-Apache--2.0-green?style=flat-square" alt="License">
</p>

Go-Agent 是一个基于 [CloudWeGo-Eino](https://github.com/cloudwego/eino) 构建的高性能、工程化 AI Agent 实战项目。它不仅是一个演示 Demo，更是一套深度面向 **大模型工程化面试与生产实践** 的参考架构，旨在展示如何将 RAG、MCP、HITL 等前沿技术真正落地。


---

## 🌟 核心特性

- 🚀 **意图驱动编排**：基于 Eino Graph (DAG) 实现复杂的意图识别与任务分发流。
- 🔍 **全栈 RAG 系统**：集成从文档加载、切分、向量化到多路召回的完整检索增强生成链路。
- 🤝 **人机协同 (HITL)**：深度实现“中断-审批-恢复”机制，支持复杂的 SQL 执行人工确认流程。
- 🔌 **MCP 协议集成**：接入 Model Context Protocol (MCP)，实现 Agent 与外部数据库、工具的标准化交互。
- 🔄 **自动 SFT 数据闭环**：无侵入式采集业务轨迹，利用“教师模型”自动标注，构建自我进化的模型微调语料库。
- ⚡ **推理性能优化**：内置投机采样 (Speculative Sampling) 及 TTFT (首字延迟) 优化策略，平衡成本、智力与响应速度。

---

## 🏗️ 系统架构

项目采用组件化设计，将 AI 能力抽象为可插拔的节点，核心逻辑结构如下：

*   **模型工厂层**：统一封装 Ark (豆包)、OpenAI、DeepSeek 等主流模型，支持配置化热切换。
*   **编排层 (Eino Graph)**：利用有向无环图管理长链路对话、意图路由及工具调用。
*   **数据工厂层**：通过 Callbacks 拦截器实现自动化的数据采集与“间接蒸馏”标注。

---

## 📂 项目结构

```bash
├── api/            # 接入层：RESTful 路由、Session 管理及 HITL 逻辑处理
├── flow/           # 编排层：核心业务逻辑图 (Intention, SQL, Chat) 的构建
├── model/          # 抽象层：多厂商 ChatModel 与 EmbeddingModel 工厂实现
├── rag/            # 核心层：RAG 完整组件与编排流 (Index & Retriever)
├── SQL/            # 工具层：MCP 工具集成与 SQL 生成、执行逻辑
├── tool/           # 工程化：SFT 采集标注、投机采样、推理加速组件
├── config/         # 配置层：全局单例与环境变量加载
└── main.go         # 入口：服务启动与组件初始化
```

---

## 🚀 快速开始

### 1. 环境配置
克隆项目并配置环境变量：

```bash
git clone https://github.com/your-repo/go-agent.git
cd go-agent
cp .env.example .env
```

### 2. 基础组件准备 (RAG 依赖)
本项目 RAG 部分依赖 Milvus (向量数据库) 和 Elasticsearch (全文搜索)，推荐使用 Docker 快速部署：

#### Elasticsearch 安装
```bash
docker run -d --name elasticsearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  elasticsearch:7.17.9
```

#### Milvus 安装 (Standalone)
```bash
# 1. 下载 docker-compose 配置文件
wget https://github.com/milvus-io/milvus/releases/download/v2.5.4/milvus-standalone-docker-compose.yml -O docker-compose.yml

# 2. 启动服务
docker-compose up -d
```
*详细文档请参考：[Milvus 官方安装指南](https://milvus.io/docs/zh/install_standalone-docker-compose.md)*

### 3. 填写配置
在 `.env` 文件中配置您的模型服务商信息及基础组件地址：

```env
# 模型配置
ARK_KEY=your_ark_key
# ... 其他厂商配置

# Milvus 配置
MILVUS_ADDR=localhost:19530
MILVUS_COLLECTION_NAME=GoAgent

# Elasticsearch 配置
ES_ADDRESS=http://localhost:9200
ES_INDEX=go_agent_docs
```

### 4. 启动服务
```bash
go mod tidy
go run .
```
服务启动后，访问 `http://localhost:8080/final_graph.html` 体验完整功能。

---

## 📖 技术深潜 (面向面试与实战)

本项目针对大模型工程化中的常见痛点提供了参考实现：

1.  **Agent 如何解决一致性？**：通过 Eino 的状态管理与类型检查。
2.  **如何处理长文本检索的精度？**：通过混合路由与重写 (Rewriter) 机制。
3.  **如何低成本提升模型能力？**：通过本项目内置的 SFT 自动闭环采集业务数据。
4.  **如何优化高并发下的响应体验？**：通过投机采样模拟应用层加速。

---

## 📜 许可证

本项目基于 [Apache-2.0](./LICENSE) 协议开源。

---

## 🤝 联系与贡献

*   **提交 Issue**：欢迎提交 Bug 或 Feature Request。
*   **联系作者**：`kele3325@gmail.com`您的建议对本项目非常重要。

---
<p align="center">Built with ❤️ by the Go-Agent Community</p>
