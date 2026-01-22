# Go-Agent: 大模型工程化实践与教学项目

[![Go Version](https://img.shields.io/github/go-mod/go-version/cloudwego/eino)](https://go.dev/)
[![Framework](https://img.shields.io/badge/Framework-Eino-blue)](https://github.com/cloudwego/eino)

## 📌 项目定位

本项目不仅是一个基于 **CloudWeGo-Eino** 构建的 AI Agent 示例，更是一份深度面向**大模型工程化面试与实战**的教学指南。我们旨在通过具体的代码实现，向开发者展示如何将 RAG、MCP、HITL等前沿技术落地于生产环境。

---

## 🏗️ 核心架构设计

项目采用组件化设计，严格遵循 Eino 的编排理念，将 AI 能力抽象为可插拔的节点：

### 1. 模型层 (Model & Embedding)

* **多模型工厂**: 统一封装了 Ark (豆包)、OpenAI、Qwen、DeepSeek 等主流模型。
* **配置驱动**: 通过环境变量实现模型热切换。

### 2. 闭环 SFT 系统 (Self-Evolution Loop)

这是本项目的核心亮点，实现了一套“自我进化”的数据工厂：

* **自动采集**: 通过 Eino Callbacks 机制，无侵入式拦截 ChatModel 的输入输出。
* **自动标注**: 引入“教师模型”对原始响应进行打分和修正，实现间接蒸馏。

### 3. 推理加速组件 (Inference Optimization)

* **投机采样 (Speculative Sampling)**: 模拟应用层投机采样逻辑，利用“学生模型”抢跑，“教师模型”核验。
* **TTFT 优化**: 侧重于首字响应延迟优化，通过流式差异化更新提升用户感知速度。

---

## 📂 目录结构说明

```bash
├── api/            # 接入层：RESTful 路由与业务逻辑转换
├── config/         # 配置层：环境变量加载与全局单例管理
├── model/          # 抽象层：ChatModel 与 EmbeddingModel 的工厂实现
│   ├── chat_model/ # 支持多厂商注册与按需实例化
├── rag/            # 核心层：检索增强生成（RAG）的完整编排
│   ├── tools/      # RAG 原子组件：Loader, Splitter, Indexer, Retriever
│   ├── compose/    # 编排逻辑：Graph 有向无环图的构建
├── tool/           # 工具层：SFT 采集、标注、投机采样逻辑
│   ├── sft/        # 包含自动标注器、存储管理与导出工具
├── main.go         # 入口：全局组件初始化与 Web 服务启动
```

---

## 🚀 快速开始

请参考 `.env.example` 配置你的 API Key，然后运行：

```bash
go run .
```

---

## 🎓 教学目标

我们希望通过本项目，让你掌握：

1. 如何设计一个高可扩展的模型工厂。
2. 如何利用 Eino 的 **Graph (DAG)** 编排复杂的多节点逻辑。
3. 如何构建自动化数据闭环，为模型微调积累高质量语料。
4. 如何在业务层实现推理加速策略，平衡成本、智力与速度。
