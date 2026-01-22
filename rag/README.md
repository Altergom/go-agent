# RAG (检索增强生成) 深度解析与教学

## 📖 RAG 基础流程
本项目严格遵循 RAG 的标准流水线设计，并利用 **Eino Graph** 进行了工业级编排：

1.  **数据摄入 (Ingestion)**
    *   **Loader**: 加载原始文档。
    *   **Splitter**: 智能分段，保证上下文语意完整。
    *   **Indexer**: 调用 Embedding 模型并存入 **Milvus** 向量数据库。
2.  **检索召回 (Retrieval)**
    *   **Retriever**: 基于向量相似度搜索相关的知识片段。
3.  **生成 (Generation)**
    *   将检索到的 Context 注入 Prompt，由大模型生成最终回答。

---

## 🛠️ 为什么选择 Eino Graph 编排？
在简单的 Demo 中，你可以使用顺序执行（Chain）。但在复杂的 RAG 场景下：
*   **分支逻辑**: 例如“如果检索不到内容，则直接转为通用对话”。
*   **并发执行**: 同时调用向量检索、关键词检索和工具调用。
*   **状态管理**: 在节点间流转复杂的中间变量。
**Graph (有向无环图)** 是处理这类复杂逻辑的最佳方案，它提供了比链式调用更高的灵活性和可维护性。

---

## ✨ 进阶优化功能 (面试加分点)

为了教学深度，本项目在基础 RAG 之上增加了以下优化：

### 1. 自动上下文采集 (SFT Ready)
在 `rag/compose` 编排中，我们通过回调机制自动捕获了：
*   用户原始提问。
*   **Retriever 检索到的原始片段** (存入 `Sample.Context`)。
*   模型的最终回答。
这为后续“分析模型为什么回答不好”以及“微调模型使其更忠实于背景”提供了原始语料。

### 2. TTFT (首字延迟) 推理加速
利用 **投机采样** 原理，在 RAG 回答阶段：
*   先由一个小模型根据检索内容生成“初稿”。
*   在用户阅读初稿时，大模型后台进行校对。
*   这种设计显著降低了用户感知的等待时间。

---

## 🗄️ 环境准备 (Milvus 启动)

RAG 强烈依赖向量数据库，请确保 Milvus 服务运行正常：

### Linux / Docker
```bash
curl -sfL https://raw.githubusercontent.com/milvus-io/milvus/master/scripts/standalone_embed.sh -o standalone_embed.sh
bash standalone_embed.sh start
```

### Windows (PowerShell)
```powershell
Invoke-WebRequest https://raw.githubusercontent.com/milvus-io/milvus/refs/heads/master/scripts/standalone_embed.bat -OutFile standalone.bat
.\standalone.bat start
```

---

## 📝 教学建议
学习本模块时，请务必阅读 `rag/compose/index.go`，理解如何将原子工具组合成一个具备决策能力的 Graph。尝试修改 Prompt 模板，观察模型对不同 Context 的敏感度。
