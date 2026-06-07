# CLAUDE.md - 项目开发指南

## 项目概述

Novelist是一个多Agent协作的AI小说创作平台，使用Go后端 + React前端 + PostgreSQL数据库。

## 技术栈

- **后端**: Go 1.25, Gin, Eino (AI框架), GORM, PostgreSQL + pgvector
- **前端**: React, TypeScript, Vite, Tailwind CSS, TipTap, Zustand
- **认证**: JWT (HS256, 72小时过期)
- **实时通信**: WebSocket (gorilla/websocket)

## 开发命令

### 后端

```bash
# 启动后端
cd backend && go run cmd/server/main.go

# 构建
cd backend && go build ./...

# 运行测试
cd backend && go test ./...

# 格式化代码
cd backend && go fmt ./...

# 静态分析
cd backend && go vet ./...
```

### 前端

```bash
# 启动开发服务器
cd frontend && npm run dev

# 构建
cd frontend && npm run build

# TypeScript类型检查
cd frontend && npx tsc --noEmit

# 代码检查
cd frontend && npm run lint
```

### 数据库

```bash
# 创建数据库
psql -d postgres -c "CREATE DATABASE novelist;"

# 执行迁移
psql -d novelist -f backend/migrations/001_init.sql

# 连接数据库
psql -d novelist
```

## 环境变量

```bash
# 必需
JWT_SECRET=your-secret-key-here
DEEPSEEK_API_KEY=your-deepseek-api-key

# 数据库（两种方式二选一）
DATABASE_URL=postgres://localhost:5432/novelist?sslmode=disable
# 或者分开指定：
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=novelist

# 可选
SERVER_PORT=8090
DEEPSEEK_MODEL=deepseek-chat

# Embedding（语义检索需要）
EMBEDDING_API_KEY=your-embedding-api-key
EMBEDDING_MODEL=text-embedding-3-small
EMBEDDING_BASE_URL=https://api.openai.com/v1
```

## 项目结构

```
novelist/
├── backend/
│   ├── cmd/server/main.go       # 入口
│   ├── internal/
│   │   ├── api/                 # HTTP处理器
│   │   ├── auth/                # JWT认证
│   │   ├── agent/               # Agent定义和系统提示
│   │   ├── ai/                  # Eino模型管理
│   │   ├── memory/              # 记忆系统
│   │   ├── orchestrator/        # 编排器和工作流
│   │   ├── model/               # 数据模型
│   │   └── store/               # 数据库连接
│   └── migrations/              # SQL迁移
├── frontend/
│   └── src/
│       ├── api/client.ts        # Axios客户端
│       ├── stores/              # Zustand状态管理
│       ├── pages/               # 页面组件
│       └── components/          # 通用组件
└── docs/
```

## 架构要点

### Agent系统

- 6个Agent角色：Creator, Writer, Editor, Reader, Critic, Reviser
  - Creator：小说构思（题材→风格→人物→世界观→大纲）
  - Writer：章节生成（含反AI味指南）
  - Editor：审查逻辑、文笔、节奏、风格、自然度
  - Reader：第一人称读者视角反馈
  - Critic：主题、人物、叙事的文学分析
  - Reviser：根据 Editor/Reader/Critic 反馈修订章节
- 系统提示在 `backend/internal/agent/prompts.go`
- Agent调用封装在 `backend/internal/agent/agent.go`
- 使用Eino框架的DeepSeek模型

### 记忆系统

- 三层记忆：长期（项目设定+人物+世界观+大纲）、短期（近5章）、工作（当前任务上下文）
- 使用pgvector进行语义检索（characters/world_settings/outlines/chapters四类）
- embedding 生成通过 OpenAI 兼容 API（text-embedding-3-small，1536维）
- 在 `backend/internal/memory/memory.go` 实现

### 编排器

- 协调Agent调用和工作流
- 讨论工作流：并行调用Editor/Reader/Critic，汇总建议
- 在 `backend/internal/orchestrator/orchestrator.go` 实现

### API路由

- 认证路由：`/api/auth/*`
- 项目路由：`/api/projects/*`
- AI操作路由：`/api/creator/chat`, `/api/chapters/:id/*`
- WebSocket：`/api/ws`

## 数据库表

- `users` - 用户
- `projects` - 小说项目（含 `short_id` 短ID字段）
- `characters` - 人物档案（含 `embedding vector(1536)`）
- `world_settings` - 世界观设定
- `outlines` - 故事大纲
- `chapters` - 章节内容
- `discussions` - 讨论记录
- `conversations` - 对话记录
- `settings` - 用户设置（含 API Key 配置、讨论轮数、模型选择等）

所有表使用UUID主键，支持pgvector向量嵌入（1536维）。
迁移文件：`001_init.sql`（基础表）、`002_add_short_id.sql`（项目短ID）。

## 前端状态管理

- `authStore` - 用户认证状态
- `projectStore` - 项目列表和当前项目
- `agentStore` - Agent对话状态
- `discussionStore` - 讨论结果状态

## 注意事项

1. Go版本要求1.25+（gin v1.10.0和eino-ext需要）
2. PostgreSQL需要安装pgvector扩展
3. JWT_SECRET必须设置，不能使用默认值
4. DeepSeek API Key必须设置才能使用AI功能
5. 前端开发服务器端口3000，后端8080（vite代理转发）
6. 前端通过Vite代理转发API请求到后端
