# Eino HITL 机制踩坑复盘：从 Interrupt 到多轮人机交互的工程实践

> **背景**：在基于 Eino Graph 构建 SQL 生成 → 人工审批 → MCP 执行的 HITL 流程时，"拒绝后重新生成 SQL"的需求暴露了一系列 Interrupt 机制与会话状态管理的工程陷阱。本文完整还原问题发现、根因分析与修复过程。

---

## 1. 需求描述

期望的交互流程：

```
用户: "查找员工表所有数据"
AI:   生成 SQL → 展示给用户 → [批准] / [拒绝]
用户: 点击 [拒绝]
AI:   "请补充您的需求说明或表结构约束信息"
用户: "不需要使用 order by"
AI:   重新生成 SQL（带补充约束）→ 展示 → [批准] / [拒绝]
用户: 点击 [批准]
AI:   通过 MCP 执行 SQL → 返回结果
```

核心技术路径：`compose.Interrupt` → 前端展示 → `compose.ResumeWithData` → 继续执行。

---

## 2. 初始架构设计

### 2.1 图拓扑（React 子图）

```
START → SQL_Retrieve → ToTplVar → SQL_Tpl → SQL_Model → Approve
                                     ↑                      ↓
                                     └──── ToRefineVar ←────┘ (拒绝时回环)
                                                             ↓
                                                         Trans_List → END (批准时)
```

设计意图：`Approve` 节点调用 `compose.Interrupt` 中断图执行；用户拒绝后，通过图内分支回环到 `ToRefineVar → SQL_Tpl → SQL_Model → Approve`，实现"拒绝→修改→重新审批"的闭环。

### 2.2 API 层会话管理（初始版本）

```go
var interruptIDMap = make(map[string]string) // sessionID → interruptID

// 有 pending interrupt 就无条件 Resume
if id, ok := interruptIDMap[sessionID]; ok {
    invokeCtx = compose.ResumeWithData(invokeCtx, id, req.Query)
}
```

---

## 3. 问题现象

### 现象 1：拒绝后 SQL 模板收到 "用户需求：NO"

用户点击"拒绝"发送 `"NO"` 后，日志显示：

```
Name:DefaultChatTemplate  Variables:{"query":"NO","docs":"..."}
Name:GeminiChatModel       Messages:[...{"content":"用户需求：NO"}]
```

SQL 模型拿到的用户需求变成了字面量 `"NO"`，生成了无意义的 SQL。

### 现象 2：Stream 模式下链式 Interrupt 崩溃

即使 refine 回环"成功"生成了新 SQL，流程回到 Approve 节点时报错：

```
Failed to stream graph: [NodeRunError] concat stream reader fail:
stream reader is empty, concat fail --- node path: [React, Approve]
```

HTTP 状态码 500，前端展示错误信息。

### 现象 3：修复会话管理后，Refine 执行 input 为 nil

修复了问题 1 和 2 后，refine 作为新图执行时报错：

```
Failed to stream graph: [NodeRunError] input is nil --- node path: [React, Approve]
```

新执行直接跳到了 `Approve` 节点，跳过了前面所有节点。

---

## 4. 根因分析

### 4.1 Bug #1：API 层无条件 Resume —— 会话状态机缺失

```go
// ❌ 只判断"有没有中断"，不区分"批准"还是"拒绝"
if id, ok := interruptIDMap[sessionID]; ok {
    invokeCtx = compose.ResumeWithData(invokeCtx, id, req.Query)
}
```

**根因**：`interruptIDMap` 是一个 `map[string]string`，只能表达一种状态："有中断"。但实际业务需要三种状态：

| 状态 | 含义 | 应有的行为 |
|------|------|-----------|
| 等待审批 | 有 InterruptID | 区分批准/拒绝 |
| 等待补充信息 | 用户已拒绝 | 合并补充信息后重新执行 |
| 正常 | 无挂起任务 | 正常执行 |

用单一 map 表达多状态 → 拒绝和批准走了同一条路（都 Resume）→ `"NO"` 被当作 Resume Data 传入图执行。

**影响链**：API 无条件 Resume → Approve 收到 `data="NO"` → 分支路由到 ToRefineVar → `"NO"` 作为 query → SQL 模板收到 "用户需求：NO"。

### 4.2 Bug #2：Stream 模式不支持同一次执行中的链式 Interrupt

```
Approve(第1次 Interrupt) → Resume → ToRefineVar → SQL_Tpl → SQL_Model → Approve(第2次 Interrupt) → 💥
```

**根因**：Eino 的 `compose.Interrupt()` 通过返回 error 实现中断。在 Stream 模式下，每个节点需要产出一个 `StreamReader` 来做 concat 拼接。当 `Approve` 第二次调用 `Interrupt()` 时：

1. 节点返回 error（不是 StreamReader）
2. Stream 基础设施期望从该节点获得一个 StreamReader 来 concat
3. 拿到的是 nil → `concat stream reader fail: stream reader is empty`

**结论**：Eino 的 `compose.Interrupt` + `Stream` 模式**不支持同一次图执行中的链式中断**。图内回环设计（`Approve → Refine → Approve`）与框架运行时能力存在根本矛盾。

### 4.3 Bug #3：CheckPoint ID 未隔离 —— 旧状态污染新执行

修复前两个问题后，将"拒绝→补充→重新生成"改为 API 层多次独立图执行。但新执行仍然崩溃：

```go
// ❌ 新执行复用了旧的 checkpoint ID
reader, err := runnable.Stream(invokeCtx, req, compose.WithCheckPointID(sessionID))
```

**根因**：`compose.WithCheckPointID` 指定的 ID 用于在 `CheckPointStore` 中持久化图的执行状态。流程如下：

1. 第一次执行用 `CheckPointID = "default-session"` → 跑到 Approve → Interrupt → **checkpoint store 记录了"执行到 Approve 节点"的快照**
2. 拒绝 → API 层不 Resume ✅
3. Refine 新执行，仍用 `CheckPointID = "default-session"` → checkpoint store 发现该 ID 有旧状态 → **框架自动从 Approve 节点恢复执行** → 但没有 `ResumeWithData` 上下文 → Approve 节点 `input = nil` → 💥

**本质**：Checkpoint 机制是为 Resume 设计的状态恢复。如果新执行复用了被中断执行的 Checkpoint ID，框架会误认为是一次 Resume，从断点而非起点开始执行。

---

## 5. 修复方案

### 5.1 会话状态机升级

将 `map[string]string` 替换为结构化的会话上下文：

```go
type sessionContext struct {
    InterruptID   string // 中断 ID（等待审批时有值）
    CheckPointID  string // 本次图执行的 checkpoint ID
    OriginalQuery string // 原始用户查询（用于 refine 合并）
    WaitingRefine bool   // 是否正在等待用户补充信息
}

var sessionContextMap = make(map[string]*sessionContext)
```

API 层按三阶段处理：

- **阶段 A**（有 pending interrupt）：区分批准 → Resume / 拒绝 → 返回提示，进入 refine 模式
- **阶段 B**（refine 模式）：合并原始查询 + 补充信息，清除 refine 状态，继续到阶段 C
- **阶段 C**（正常执行）：构建图、生成新 CheckPointID、Stream 执行

### 5.2 架构调整：多次独立图执行替代图内回环

```
原设计（图内回环 —— 不可行）：
  Approve → ToRefineVar → SQL_Tpl → SQL_Model → Approve(第2次Interrupt) → 💥

新设计（API 层多次独立执行）：
  执行1: Intent → React → RAG → SQL → Approve(Interrupt) → 返回 SQL
  拒绝:  API 直接返回提示（不经过图）
  执行2: Intent → React → RAG(合并查询) → SQL → Approve(Interrupt) → 返回新 SQL
  批准:  Resume → MCP 执行 → 返回结果
```

每次图执行只有**一次 Interrupt**，从根本上规避链式 Interrupt 问题。

### 5.3 CheckPoint ID 隔离

```go
// 批准 Resume —— 复用旧 checkpoint ID（需要从断点恢复）
reader, err := runnable.Stream(invokeCtx, req, compose.WithCheckPointID(sc.CheckPointID))

// 正常执行 / Refine 执行 —— 生成新 checkpoint ID（确保干净状态）
checkPointID := fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano())
reader, err := runnable.Stream(invokeCtx, req, compose.WithCheckPointID(checkPointID))
```

在 interrupt 发生时保存 `checkPointID` 到会话上下文，确保 Resume 时能找回正确的断点状态。

---

## 6. 经验总结

| # | 踩坑点 | 教训 |
|---|--------|------|
| 1 | 用单一 map 管理多状态会话 | HITL 场景天然是状态机，用结构化类型显式建模每种状态和转换条件 |
| 2 | 在图内设计回环 Interrupt | `compose.Interrupt` + Stream 不支持链式中断；需要人工介入的多轮交互应拆分为多次独立图执行 |
| 3 | 复用 CheckPoint ID | Checkpoint 是为 Resume 设计的断点恢复机制；非 Resume 的新执行必须使用全新 ID，否则旧状态会污染新执行 |
| 4 | 先写逻辑后想状态 | 应先画出完整的状态转换图（等待审批 → 批准/拒绝 → 等待补充 → 重新执行），再写代码 |

### 核心原则

> **Interrupt 是"暂停-恢复"，不是"循环控制流"。** 当业务需要多轮人机交互时，每轮交互应该是一次独立的图执行，通过 API 层的会话状态机串联，而不是试图在单次图执行中用 Interrupt 实现循环。
