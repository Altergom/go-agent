# Go-Agent

<p align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/cloudwego/eino?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Framework-Eino-blue?style=flat-square" alt="Framework">
  <img src="https://img.shields.io/badge/License-Apache--2.0-green?style=flat-square" alt="License">
  <img src="https://img.shields.io/badge/Production-Ready-brightgreen?style=flat-square" alt="Production Ready">
</p>

Go-Agent 是一个基于 [CloudWeGo-Eino](https://github.com/cloudwego/eino) 构建的**生产级 AI Agent 应用开发框架**。它不仅是技术演示，更是一套完整的工程化解决方案，深度整合了 RAG、MCP、HITL、智能分析等企业级能力，为 **AI 应用落地与面试实战** 提供最佳实践参考。

---

## 🌟 核心特性

### 🎯 智能编排引擎
- **意图识别路由**：基于 Eino Graph 的 DAG/Pregel 双模式编排，自动识别用户意图并分发至对应的专业子图
- **流程可视化**：支持复杂业务流程的图形化编排，状态可追溯、可恢复
- **条件分支控制**：内置 Branch 节点和 Lambda 表达式，实现灵活的业务逻辑判断

### 🔍 企业级 RAG 系统
- **混合检索架构**：Milvus 向量检索 + Elasticsearch 全文检索 + BM25/RRF 融合排序
- **多模态文档解析**：支持 PDF 中的文本、表格、图片混合提取与检索
- **智能缓存层**：基于 Redis 的检索结果缓存，大幅降低重复查询的延迟与成本

### 🤝 生产级人机协同 (HITL)
- **中断-恢复机制**：支持 SQL 执行、敏感操作等场景的人工审批，审批后无缝恢复执行
- **分布式状态持久化**：基于 Redis 的 CheckPoint 存储，服务重启后状态不丢失
- **会话管理**：多会话隔离，支持长时间异步审批场景

### 🔌 自建 MCP 工具服务器
- **完整 CRUD 能力**：不再局限于只读查询，支持 INSERT/UPDATE/DELETE 及 DDL 操作
- **多层安全防护**：SQL 白名单验证 + 危险模式拦截 + 强制 WHERE 子句 + 全链路审计日志
- **审计与合规**：所有数据库操作自动记录到审计表，满足企业安全与合规要求

### 📊 智能数据分析助手
- **自动洞察生成**：SQL 执行完成后，AnalystAgent 自动进行统计分析、趋势识别、异常检测
- **可视化配置输出**：基于数据特征智能推荐图表类型，生成开箱即用的 ECharts 配置
- **自然语言报告**：将数据分析结果转化为业务可读的文字报告

### 🔄 AI 自我进化闭环
- **无侵入式数据采集**：通过 Callbacks 自动记录用户交互轨迹
- **教师模型标注**：利用强模型对业务数据进行自动标注，构建高质量 SFT 语料
- **持续优化迭代**：支持模型微调与效果评估的完整工作流

### ⚡ 性能与成本优化
- **投机采样加速**：Draft Model + Target Model 协同，降低推理延迟
- **首字延迟优化 (TTFT)**：流式响应 + 并行处理，提升用户体验
- **多模型配置**：支持按场景选择不同成本/能力的模型组合

---

## 🏗️ 系统架构

Go-Agent 采用分层组件化架构，核心模块职责清晰、松耦合、高内聚：

```
┌─────────────────────────────────────────────────────────────┐
│                      API 接入层                              │
│   RESTful 路由 │ Session 管理 │ HITL 审批处理               │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                  编排引擎层 (Eino Graph)                     │
│   意图路由 │ RAG Chat │ SQL React │ Analyst 子图            │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┼───────────┐
         ▼           ▼           ▼
┌────────────┐ ┌──────────┐ ┌─────────────┐
│ RAG 模块   │ │ MCP 工具 │ │ 数据分析    │
│ 检索+召回  │ │ SQL执行  │ │ Analyst     │
└──────┬─────┘ └────┬─────┘ └──────┬──────┘
       │            │              │
┌──────▼────────────▼──────────────▼──────────────────────────┐
│                   基础设施层                                  │
│ Redis (状态/缓存) │ Milvus/ES (检索) │ MySQL (审计/业务)    │
│ LLM 工厂 (Ark/DeepSeek/Gemini/Qwen) │ Embedding 工厂       │
└──────────────────────────────────────────────────────────────┘
```

**核心设计理念**：
- **存储层分离**：状态、缓存、业务数据分别使用 Redis、向量库、关系型数据库
- **模型工厂抽象**：统一接口，支持多厂商模型的配置化切换与成本优化
- **子图可组合**：每个业务能力（RAG、SQL、分析）独立成图，可灵活组合编排

---

## 📂 项目结构

```bash
├── api/                    # 接入层：HTTP 路由、会话管理、HITL 审批接口
├── flow/                   # 编排层：业务流程图定义 (意图路由/RAG/SQL/Analyst)
├── model/                  # 模型工厂：多厂商 ChatModel 与 EmbeddingModel 适配
│   ├── chat_model/         # 对话模型：Ark, DeepSeek, Gemini, OpenAI, Qwen
│   └── embedding_model/    # 向量模型：含多模态 Embedding 支持
├── rag/                    # RAG 核心：文档索引、混合检索、召回融合
│   ├── rag_flow/           # 检索流程编排
│   └── rag_tools/          # 工具组件（Indexer, Retriever, DB 连接器）
├── tool/                   # 工具与工程化组件
│   ├── storage/            # Redis 持久化层 (CheckPoint/Session/Cache)
│   ├── sql_tools/          # MCP 客户端与 SQL 工具集成
│   ├── analyst_tools/      # 数据分析工具集 (统计/图表生成)
│   ├── document/           # 文档解析器 (含多模态 PDF 支持)
│   ├── sft/                # SFT 数据采集与标注
│   ├── memory/             # 对话记忆管理
│   └── trace/              # 链路追踪与日志
├── mcp_server/             # 自建 MCP 服务器 (CRUD + 安全审计)
│   ├── tools/              # 工具实现 (query/insert/update/delete)
│   └── security/           # 安全模块 (SQL 白名单 + 审计日志)
├── config/                 # 配置管理：环境变量加载与全局配置
├── test/                   # 测试套件：单元测试 + 性能基准测试
└── main.go                 # 服务入口：组件初始化与启动
```

---

## 🚀 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose（用于快速部署依赖服务）
- 至少一个 LLM 服务商的 API Key（Ark/OpenAI/DeepSeek/Gemini/Qwen）

### 一、克隆项目

```bash
git clone https://github.com/your-repo/go-agent.git
cd go-agent
```

### 二、部署基础设施

#### 1. 使用 Docker Compose 一键启动所有依赖服务

```bash
docker-compose up -d
```

该命令将自动启动：
- **Redis** (端口 6379)：用于状态持久化和缓存
- **Milvus** (端口 19530)：向量数据库
- **Elasticsearch** (端口 9200)：全文检索引擎
- **MySQL** (端口 3306)：审计日志存储

#### 2. 手动部署（可选）

如果你想单独部署某个服务：

**Redis**:
```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

**Elasticsearch**:
```bash
docker run -d --name elasticsearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
  elasticsearch:7.17.9
```

**Milvus**:
```bash
wget https://github.com/milvus-io/milvus/releases/download/v2.5.4/milvus-standalone-docker-compose.yml -O milvus-compose.yml
docker-compose -f milvus-compose.yml up -d
```

**MySQL** (用于审计日志):
```bash
docker run -d --name mysql-audit \
  -p 3307:3306 \
  -e MYSQL_ROOT_PASSWORD=audit_password \
  -e MYSQL_DATABASE=go_agent_audit \
  mysql:8
```

### 三、配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的配置：

```env
# ===== LLM 模型配置 =====
# Ark (豆包)
ARK_KEY=your_ark_api_key
ARK_BASE_URL=https://ark.cn-beijing.volces.com/api/v3

# OpenAI
OPENAI_API_KEY=your_openai_key
OPENAI_BASE_URL=https://api.openai.com/v1

# DeepSeek
DEEPSEEK_API_KEY=your_deepseek_key
DEEPSEEK_BASE_URL=https://api.deepseek.com

# Gemini
GEMINI_API_KEY=your_gemini_key

# Qwen (通义千问)
QWEN_API_KEY=your_qwen_key

# ===== Redis 配置 =====
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# ===== Milvus 配置 =====
MILVUS_ADDR=localhost:19530
MILVUS_COLLECTION_NAME=GoAgent

# ===== Elasticsearch 配置 =====
ES_ADDRESS=http://localhost:9200
ES_INDEX=go_agent_docs

# ===== MySQL 审计数据库 =====
AUDIT_MYSQL_HOST=localhost
AUDIT_MYSQL_PORT=3307
AUDIT_MYSQL_USER=root
AUDIT_MYSQL_PASSWORD=audit_password
AUDIT_MYSQL_DATABASE=go_agent_audit

# ===== MCP 服务器配置 =====
MCP_SERVER_PATH=./mcp_server/go-agent-mcp-server
MCP_WHITELIST_PATH=./mcp_server/whitelist.yaml
```

### 四、安装依赖并启动

```bash
# 安装 Go 依赖
go mod tidy

# (可选) 编译 MCP 服务器
cd mcp_server
go build -o go-agent-mcp-server
cd ..

# 启动主服务
go run main.go
```

### 五、访问应用

服务启动后，访问以下地址：

- **完整对话界面**: http://localhost:8080/final_graph.html
- **RAG 知识库管理**: http://localhost:8080/rag_knowledge.html
- **文档索引**: http://localhost:8080/rag_index.html
- **RAG 问答测试**: http://localhost:8080/rag_ask.html

---

## 📖 功能使用指南

### 1. RAG 知识库构建

#### 步骤 1: 上传文档

访问 `http://localhost:8080/rag_index.html`，上传 PDF/TXT/Markdown 文档。

支持的文档类型：
- 纯文本文档
- PDF（含文本、表格混合提取）
- 多模态 PDF（图片提取与向量化）

#### 步骤 2: 自动索引

系统会自动：
1. 解析文档内容
2. 切分为语义块
3. 生成 Embedding 向量
4. 同时存入 Milvus（向量索引）和 Elasticsearch（全文索引）

#### 步骤 3: 混合检索

在对话或 RAG 问答中，系统会：
- **向量检索**: 通过语义相似度召回相关文档
- **全文检索**: 通过关键词匹配召回精确文档
- **融合排序**: 使用 BM25 + RRF 算法融合两路结果

### 2. SQL 数据库对话

#### 自然语言查询

用户：`帮我查询销售额最高的前5个产品`

系统流程：
1. **意图识别** → 识别为 SQL 查询意图
2. **SQL 生成** → 调用 LLM 生成 SQL: `SELECT product_name, SUM(sales) FROM orders GROUP BY product_name ORDER BY SUM(sales) DESC LIMIT 5`
3. **HITL 审批** → 中断流程，等待用户审批 SQL
4. **执行与分析**:
   - 用户批准后，通过 MCP 工具执行 SQL
   - AnalystAgent 自动分析结果，生成统计报告
   - 推荐柱状图，输出 ECharts 配置
5. **返回结果** → 包含 SQL 结果 + 文字分析 + 图表配置

#### 数据修改操作

用户：`将产品ID为101的价格改为99.9`

系统流程：
1. 生成 SQL: `UPDATE products SET price = 99.9 WHERE id = 101`
2. **安全校验**:
   - MCP 服务器白名单验证
   - 检测是否包含 WHERE 子句（强制要求）
   - 审计日志记录
3. **HITL 审批** → 敏感操作，必须人工确认
4. **执行并记录** → 写入 `audit_log` 表

#### 危险操作拦截

用户：`清空整个订单表`

系统流程：
1. 生成 SQL: `TRUNCATE TABLE orders` 或 `DELETE FROM orders`
2. **拦截机制**:
   - 白名单检测到 `TRUNCATE` 或不含 WHERE 的 `DELETE`
   - 拒绝执行，返回错误提示
   - 审计日志记录拦截事件
3. **用户提示** → "该操作已被安全策略拦截，请联系管理员"

### 3. 人机协同审批 (HITL)

#### 中断机制

当系统需要人工介入时（如执行 SQL），会触发中断：

```json
{
  "status": "interrupted",
  "checkpoint_id": "ckpt_abc123",
  "session_id": "sess_xyz789",
  "pending_action": {
    "type": "sql_execution",
    "sql": "UPDATE products SET price = 99.9 WHERE id = 101",
    "reason": "需要审批数据修改操作"
  }
}
```

#### 审批与恢复

用户通过审批接口选择：

**批准执行**:
```bash
POST /api/approve
{
  "session_id": "sess_xyz789",
  "checkpoint_id": "ckpt_abc123",
  "action": "approve"
}
```

系统会：
1. 从 Redis 恢复 CheckPoint 状态
2. 继续执行剩余流程
3. 返回最终结果

**拒绝执行**:
```bash
POST /api/approve
{
  "session_id": "sess_xyz789",
  "checkpoint_id": "ckpt_abc123",
  "action": "reject",
  "reason": "SQL 语句有误"
}
```

系统会终止流程并返回拒绝原因。

### 4. 智能数据分析

#### 自动触发

当 SQL 执行成功后，AnalystAgent 会自动启动分析流程。

#### 分析能力

1. **统计计算**:
   - 均值、中位数、标准差
   - 最大值、最小值、四分位数
   - 数据分布特征

2. **趋势识别**:
   - 时间序列趋势（上升/下降/平稳）
   - 周期性模式
   - 异常值检测

3. **图表推荐**:
   - 数值型单列 → 推荐柱状图/折线图
   - 分类占比 → 推荐饼图
   - 时间序列 → 推荐折线图/面积图
   - 多维对比 → 推荐分组柱状图

4. **自然语言报告**:
```
数据分析结果：
共返回5条记录，销售额范围在 1.2万 至 8.5万 之间，平均值为 4.3万。
其中"产品A"表现最佳，销售额达到 8.5万，占比 42%。
建议重点关注"产品A"的库存补充。
```

#### 响应格式

```json
{
  "query": "查询销售额最高的5个产品",
  "answer": "已为您查询完成",
  "sql_result": "[[\"产品A\", 85000], [\"产品B\", 62000], ...]",
  "analysis": "数据分析结果：共返回5条记录...",
  "chart_config": {
    "type": "bar",
    "xAxis": { "data": ["产品A", "产品B", ...] },
    "yAxis": {},
    "series": [{ "type": "bar", "data": [85000, 62000, ...] }]
  },
  "status": "completed"
}
```

---

## 🔧 高级配置

### 模型选择策略

项目支持为不同场景配置不同的模型：

```go
// config/config.go

type ModelConfig struct {
    // 对话模型（高质量）
    ChatModel string // "ark:ep-xxxxx" | "deepseek-chat" | "gpt-4"
    
    // 快速模型（成本优化）
    FastModel string // "deepseek-chat" | "gpt-3.5-turbo"
    
    // Draft 模型（投机采样）
    DraftModel string // "qwen-turbo" | "deepseek-coder"
    
    // Embedding 模型
    EmbeddingModel string // "text-embedding-ada-002" | "ark-embedding"
}
```
配置方法为手动修改注册名(以意图识别模型为例)
```
// .env

INTENT_MODEL_TYPE=ark
```
如果存在调用同一提供商的不同模型清仿照`// config/config.go`和`// model/chat_model`的实现手动添加不同配置

### MCP 白名单定制

编辑 `mcp_server/whitelist.yaml`:

```yaml
allowed_operations:
  - SELECT
  - INSERT
  - UPDATE
  - DELETE
  - CREATE TABLE
  - ALTER TABLE

forbidden_patterns:
  - DROP DATABASE
  - DROP TABLE
  - TRUNCATE
  - LOAD_FILE
  - INTO OUTFILE
  - UNION.*SELECT  # 防止 SQL 注入

require_where_clause:
  - UPDATE
  - DELETE

audit_level:
  SELECT: info
  INSERT: warning
  UPDATE: warning
  DELETE: error
  CREATE: error
  ALTER: error
```

### 性能优化参数

```env
# 缓存 TTL
CACHE_RETRIEVAL_TTL=3600      # 检索缓存 1 小时
CACHE_EMBEDDING_TTL=86400     # Embedding 缓存 24 小时

# 检索参数
RAG_TOP_K=10                  # 召回文档数量
RAG_RERANK_TOP_K=5           # 重排序后保留数量

# 并发控制
MAX_CONCURRENT_REQUESTS=50    # 最大并发请求数
GRAPH_TIMEOUT=300             # Graph 执行超时时间（秒）
```

---

## 📊 监控与运维

### 审计日志查询

所有数据库操作都记录在 `audit_log` 表中：

```sql
-- 查看最近的 SQL 执行记录
SELECT * FROM audit_log ORDER BY timestamp DESC LIMIT 20;

-- 统计操作类型分布
SELECT operation_type, COUNT(*) as count 
FROM audit_log 
GROUP BY operation_type;

-- 查询失败的操作
SELECT * FROM audit_log WHERE status = 'failed';

-- 查询某个会话的所有操作
SELECT * FROM audit_log WHERE session_id = 'sess_xyz789';
```

### 性能指标

项目内置性能追踪，关键指标：

- **TTFT (Time To First Token)**: 首字延迟，目标 < 1s
- **检索延迟**: P99 < 500ms（缓存命中时）
- **缓存命中率**: 目标 > 60%
- **Graph 执行时间**: 完整流程 < 10s

### 日志级别

```env
LOG_LEVEL=info  # debug | info | warning | error
LOG_FORMAT=json # json | text
```


## 📖 技术深潜 (面向面试与实战)

本项目针对大模型面试中的常见场景题提供了参考实现：

### 1. Agent 如何解决一致性问题？

**痛点**: 多轮对话中状态混乱、工具调用结果丢失

**解决方案**:
- **类型安全的状态管理**: 使用 Eino 的强类型 State，编译期检查
- **CheckPoint 持久化**: 每次状态变更都写入 Redis，服务重启不丢失
- **会话隔离**: 每个用户会话独立管理，互不干扰

### 2. 如何处理长文本检索的精度问题？

**痛点**: 向量检索容易召回不相关文档，关键词检索容易遗漏

**解决方案**:
- **混合检索**: 向量检索 + 全文检索并行
- **融合排序**: BM25 + RRF 算法综合两路结果
- **Query Rewriting**: 在检索前改写查询，提升召回率
- **智能缓存**: 相似查询直接命中缓存，避免重复计算

### 3. 如何低成本提升模型能力？

**痛点**: 业务场景特殊，通用模型效果差，但标注数据成本高

**解决方案**:
- **自动数据采集**: Callbacks 机制无侵入式记录用户交互
- **教师模型标注**: 用 GPT-4/DeepSeek-V3 对业务数据自动打标
- **SFT 闭环**: 定期微调小模型，降低推理成本
- **效果评估**: A/B 测试验证微调效果

### 4. 如何优化高并发下的响应体验？

**痛点**: 强模型推理慢，用户等待时间长

**解决方案**:
- **投机采样**: 用快速小模型生成 Draft，强模型验证
- **流式响应**: Server-Sent Events (SSE) 实时返回 Token
- **并行处理**: 检索、Embedding、工具调用并行执行
- **多级缓存**: Redis 缓存 + 应用内存缓存

### 5. 如何保障 Agent 的安全性？

**痛点**: Agent 可能执行危险操作，如删库、泄露数据

**解决方案**:
- **HITL 机制**: 敏感操作必须人工审批
- **SQL 白名单**: 只允许配置的操作类型
- **危险模式拦截**: 正则 + AST 解析双重检测
- **全链路审计**: 所有操作记录可追溯
- **权限隔离**: 数据库账号最小权限原则

---

## 🤝 贡献指南

我们欢迎任何形式的贡献！

### 提交 Issue

- **Bug 报告**: 请详细描述复现步骤、环境信息、错误日志
- **功能请求**: 请说明使用场景、预期效果、优先级

### 提交 Pull Request

1. Fork 本仓库
2. 创建特性分支: `git checkout -b feature/amazing-feature`
3. 提交代码: `git commit -m 'Add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 提交 PR

**代码规范**:
- 遵循 [Effective Go](https://golang.org/doc/effective_go) 规范
- 添加必要的单元测试
- 更新相关文档

---

## 📚 相关文档(持续更新ing)

- **[依赖安装](./DEPENDENCIES.md)**: 所有依赖项的详细安装指南
- **[HITL 机制踩坑](./doc/HITL机制踩坑复盘.md)**: 人机协同实现的实战经验(所谓“你遇到过什么困难”)

---

## 📜 许可证

本项目基于 [Apache-2.0](./LICENSE) 协议开源。

---

## 📞 联系与支持

- **Email**: kele3325@gmail.com
- **Issues**: [GitHub Issues](https://github.com/altergom/go-agent/issues)

---

<p align="center">
  <strong>祝大家2026offer多多福满年❤️</strong><br>
</p>
