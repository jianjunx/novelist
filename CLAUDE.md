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

# 可选
DATABASE_URL=postgres://localhost:5432/novelist?sslmode=disable
SERVER_PORT=8080
DEEPSEEK_MODEL=deepseek-chat
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

- 5个Agent角色：Creator, Writer, Editor, Reader, Critic
- 系统提示在 `backend/internal/agent/prompts.go`
- Agent调用封装在 `backend/internal/agent/agent.go`
- 使用Eino框架的DeepSeek模型

### 记忆系统

- 三层记忆：长期（项目设定）、短期（近期章节）、工作（当前任务）
- 使用pgvector进行语义检索
- 在 `backend/internal/memory/memory.go` 实现

### 编排器

- 协调Agent调用和工作流
- 讨论工作流：并行调用Editor/Reader/Critic，汇总建议
- 在 `backend/internal/orchestrator/orchestrator.go` 实现

### API路由

- 认证路由：`/api/auth/*`
- 项目路由：`/api/projects/*`
- AI操作路由：`/api/creator/chat`, `/api/chapters/:id/*`
- WebSocket：`/ws`

## 数据库表

- `users` - 用户
- `projects` - 小说项目
- `characters` - 人物档案
- `world_settings` - 世界观设定
- `outlines` - 故事大纲
- `chapters` - 章节内容
- `discussions` - 讨论记录
- `conversations` - 对话记录
- `settings` - 用户设置

所有表使用UUID主键，支持pgvector向量嵌入（1536维）。

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
5. 前端开发服务器端口3000，后端8080
6. 前端通过Vite代理转发API请求到后端
