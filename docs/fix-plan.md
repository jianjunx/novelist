# Novelist 设计-实现对齐修复方案

> 基于 CLAUDE.md 设计文档与代码实现的审计结果，按优先级排列。

---

## P0 - 核心功能阻断（必须先修）

### 1. WebSocket 路径三方不一致
**现状**：文档写 `/ws`，后端注册 `/api/ws`，前端连 `/ws`  
**修复**：统一为 `/api/ws`（后端已有JWT认证，合理）
- `frontend/src/hooks/useWebSocket.ts`：连接路径改为 `/api/ws`
- `CLAUDE.md`：更新为 `/api/ws`

### 2. pgvector 语义检索名存实亡
**现状**：表有 `vector(1536)` 字段，`SemanticSearch` 方法写了，但没有 embedding 生成逻辑，从未被调用  
**修复**：补全 embedding 写入链路
- 新增 `embedding.go`：调用 DeepSeek/OpenAI embedding API 生成向量
- `store` 层 CRUD 操作中，在写入 characters/world_settings/outlines/chapters 时同步生成并存储 embedding
- `orchestrator` 的 `AssembleContext` 中接入 `SemanticSearch`，用当前章节内容做相似性检索补充上下文

### 3. ChapterEditor 没接后端
**现状**：不加载章节内容、不保存、不调续写/润色API  
**修复**：
- `ChapterEditor.tsx` 的 `useEffect` 中调 `GET /api/chapters/:id` 加载内容
- 保存按钮调 `PUT /api/chapters/:id`
- 续写/润色按钮接入 `/continue` 和 `/polish` API
- 编辑器内容变化同步到 `projectStore`

---

## P1 - 功能不完整

### 4. 工作记忆层是空壳
**现状**：`AssembleContext` 的 `workingMemory` 恒传 `""`  
**修复**：
- 编排器在调用 Agent 前，将当前任务上下文（如：正在写第N章、当前大纲要点、用户反馈）传入 `workingMemory`
- 可以是一个结构体或 map，至少包含 `current_chapter_number`、`current_outline_point`、`user_feedback`

### 5. Reviser Agent 没记录
**现状**：代码有 `RoleReviser` 和 `ReviserPrompt`，CLAUDE.md 没写  
**修复**：CLAUDE.md 补充第6个角色描述

### 6. `discussion_rounds` 配置没接入
**现状**：settings 表存了 `discussion_rounds`，编排器写死 `round_num=1`  
**修复**：
- `orchestrator.go` 的 `StartDiscussion` 从 settings 读取 `discussion_rounds`
- 支持多轮讨论，每轮汇总后反馈给下一轮

### 7. Creator 刷新后对话清空
**现状**：conversations 表存了数据，但 Creator 页面不恢复  
**修复**：
- 后端补充 `GET /api/projects/:id/conversations` 查询接口（或复用已有但未暴露的）
- `agentStore` 在进入 Creator 页面时加载历史对话
- `Creator.tsx` 渲染历史消息

---

## P2 - 开发环境 & 文档

### 8. 端口不一致
**现状**：文档写 8080，`vite.config.ts` 代理 8090  
**修复**：
- 统一为一个端口（建议 8080）
- 更新 `vite.config.ts` 的 proxy target
- 或者更新 CLAUDE.md 为 8090

### 9. 环境变量格式不一致
**现状**：文档写 `DATABASE_URL`，代码用 `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME`  
**修复**：
- `config.go` 同时支持两种方式：优先读 `DATABASE_URL`，fallback 到 `DB_*` 分字段
- CLAUDE.md 补充 `DB_*` 变量说明

### 10. `short_id` 和迁移 002 文档缺失
**修复**：CLAUDE.md 补充 `projects.short_id` 说明和 `002_add_short_id.sql` 迁移

### 11. Settings 前后端断层
**现状**：后端有完整 API，前端没页面  
**修复**：
- 新增 `Settings.tsx` 页面：API Key 配置、讨论轮数、模型选择
- 新增 `settingsStore.ts`
- 路由 `/settings`

---

## P3 - 锦上添花

### 12. 流式响应没落地
**现状**：`ChatStream` 实现了，前端没调用  
**修复**：Creator 对话改用 SSE 流式，提升体验

### 13. conversations 只存不读
**修复**：补充查询 API + 前端恢复（已包含在 #7）

### 14. 评审错误被静默吞掉
**修复**：goroutine 中 err 时记录日志 + 返回部分结果给调用方

---

## 执行顺序建议

```
Phase 1 (P0):  #1 WebSocket → #3 ChapterEditor → #2 pgvector
Phase 2 (P1):  #7 Creator恢复 → #4 工作记忆 → #6 多轮讨论 → #5 文档补Reviser
Phase 3 (P2):  #9 环境变量 → #8 端口 → #10 文档补short_id → #11 Settings页面
Phase 4 (P3):  #12 流式 → #14 错误处理
```

**预估工作量**：P0 约 2-3天，P1 约 2天，P2 约 1天，P3 约 1天
