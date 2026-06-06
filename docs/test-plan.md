# Novelist 测试用例计划

> 项目当前测试覆盖率为 0。按层分优先级，从底层往上补。

---

## 测试策略

| 层级 | 测试类型 | 框架 | 说明 |
|------|---------|------|------|
| 后端纯函数 | 单元测试 | `go test` | config解析、JWT、JSON提取、embedding格式化 |
| Store层 | 集成测试 | `go test` + testcontainers/postgres | 数据库CRUD、embedding读写 |
| API层 | 集成测试 | `httptest` + mock store | 路由、认证、请求/响应格式 |
| Orchestrator | 单元测试 | `go test` + mock agent | 工作流编排、多轮讨论、错误处理 |
| Agent | 单元测试 | `go test` | prompt构建、消息格式化 |
| Memory | 单元测试 | `go test` | 上下文组装、语义检索逻辑 |

---

## Phase 1 — 基础设施（无外部依赖）

### 1.1 config 解析测试
**文件**: `backend/internal/config/config_test.go`
- `DATABASE_URL` 解析（正常格式、缺字段、无效URL）
- `DB_*` 分字段拼接
- `DATABASE_URL` 优先于 `DB_*`
- 缺少必填字段时的 fallback 默认值
- 环境变量为空时的行为

### 1.2 JWT 测试
**文件**: `backend/internal/auth/jwt_test.go`
- GenerateToken：正常生成、包含正确 claims
- ParseToken：有效token、过期token、篡改token、空token
- 72小时过期验证

### 1.3 JSON 提取测试
**文件**: `backend/internal/orchestrator/extract_test.go`（如果 extractJSON 是独立函数）
- 正常 JSON 块提取
- LLM 输出中混杂文本时的提取
- 无效 JSON 降级处理
- 空输入

### 1.4 Embedding 向量格式化
**文件**: `backend/internal/store/embedding_test.go`
- `formatEmbeddingVector`：float32 slice → PostgreSQL vector 字符串
- 空 slice 处理
- 维度验证（1536维）

### 1.5 Agent Prompt 构建
**文件**: `backend/internal/agent/agent_test.go`
- 各角色 prompt 模板不为空
- 包含关键指令词（反AI味、第一人称等）
- 消息格式化正确

---

## Phase 2 — Store 层（需要数据库）

### 2.1 测试数据库基础设施
**文件**: `backend/internal/store/testutil_test.go`
- 用 SQLite in-memory 或 testcontainers 搭建测试DB
- 自动迁移测试表
- 提供 `setupTestDB()` / `teardownTestDB()` 工具函数

### 2.2 Project CRUD
**文件**: `backend/internal/store/project_test.go`
- 创建项目（含 short_id 生成）
- 查询项目列表（按用户过滤）
- 更新项目
- 删除项目（级联检查）

### 2.3 Chapter CRUD
**文件**: `backend/internal/store/chapter_test.go`
- 创建章节（embedding 自动生成或跳过）
- 按项目查询章节列表
- 更新章节内容
- 章节不存在时的错误处理

### 2.4 Character / WorldSetting / Outline CRUD
**文件**: `backend/internal/store/character_test.go` 等
- 基本 CRUD
- embedding 字段写入（mock embedding manager 时）

### 2.5 Settings CRUD
**文件**: `backend/internal/store/settings_test.go`
- 创建/读取/更新用户设置
- 多个用户设置隔离

### 2.6 Discussion / Conversation CRUD
**文件**: `backend/internal/store/discussion_test.go`
- 讨论记录写入（含 round_num、agent_role）
- 对话记录按项目查询

---

## Phase 3 — API 层（httptest）

### 3.1 测试 HTTP 基础设施
**文件**: `backend/internal/api/testutil_test.go`
- 用 `httptest.NewServer` + mock store
- 生成测试 JWT token
- 提供 `makeRequest()` 工具函数

### 3.2 Auth API
**文件**: `backend/internal/api/auth_test.go`
- POST /api/auth/register：正常注册、重复邮箱、缺字段
- POST /api/auth/login：正常登录、密码错误、用户不存在
- GET /api/auth/me：有效token、无效token、过期token

### 3.3 Project API
**文件**: `backend/internal/api/project_test.go`
- CRUD 全流程
- 未认证访问 401
- 访问他人项目 403

### 3.4 Chapter API
**文件**: `backend/internal/api/chapter_test.go`
- 章节 CRUD
- GET /api/chapters/:id 加载内容
- PUT /api/chapters/:id 保存

### 3.5 Settings API
**文件**: `backend/internal/api/settings_test.go`
- GET /api/settings：读取设置
- PUT /api/settings：更新设置
- 未认证 401

### 3.6 Conversation API
**文件**: `backend/internal/api/conversation_test.go`
- GET /api/projects/:id/conversations：查询对话历史
- 空对话返回空数组
- 不存在的项目返回 404

---

## Phase 4 — Orchestrator（mock Agent）

### 4.1 编排器基础
**文件**: `backend/internal/orchestrator/orchestrator_test.go`
- `buildWorkingMemory`：从章节标题/大纲提取关键词
- `aggregateSuggestions`：建议去重、排序
- `formatDiscussionSummary`：格式化上一轮结果

### 4.2 讨论工作流
**文件**: `backend/internal/orchestrator/discussion_test.go`
- 单轮讨论：3个 agent 并行执行
- 多轮讨论：结果串联传递
- 部分 agent 失败：错误记录在 `errors` 字段，不阻断流程
- 全部失败：返回空结果 + errors

### 4.3 生成+审查流水线
**文件**: `backend/internal/orchestrator/pipeline_test.go`
- `GenerateAndReview`：生成→讨论→修订
- `ReviewAndRevise`：审查→修订

---

## Phase 5 — Memory（可选，视时间）

### 5.1 上下文组装
**文件**: `backend/internal/memory/memory_test.go`
- `AssembleContext`：长期+短期+工作记忆拼接
- 空记忆时的行为
- workingMemory 非空时触发语义检索

---

## 前端测试（Phase 6，可选）

### 6.1 Store 单元测试
- `authStore`：login/logout/checkAuth 状态转换
- `projectStore`：项目列表加载、章节操作
- `agentStore`：消息发送、流式接收

### 6.2 组件冒烟测试
- `Login.tsx`：渲染、表单提交
- `Dashboard.tsx`：项目列表渲染
- `Creator.tsx`：消息列表渲染

框架：Vitest + React Testing Library

---

## 执行顺序

```
Phase 1 (P0): 基础设施测试 → 立即可写，无外部依赖
Phase 2 (P1): Store层测试 → 需要测试数据库
Phase 3 (P1): API层测试 → 需要 mock store
Phase 4 (P2): Orchestrator测试 → 需要 mock agent
Phase 5 (P3): Memory测试 → 可选
Phase 6 (P3): 前端测试 → 可选
```

**预估工作量**：Phase 1 约 0.5天，Phase 2-3 约 1.5天，Phase 4 约 1天，Phase 5-6 约 1天

## 要求

1. 每个测试文件独立可运行：`go test ./internal/config/...`
2. 整体运行：`go test ./...`
3. 测试覆盖目标：Phase 1-4 完成后 ≥ 60%
4. mock 外部依赖（AI API、embedding API），不调真实接口
5. 测试数据用工厂函数生成，不硬编码 UUID
