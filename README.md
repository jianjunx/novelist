# Novelist - 多Agent协作AI小说创作平台

一个支持多Agent协作的AI小说创作平台，让个人创作者能够与AI角色（编辑、读者、评论家）协作改进作品。

## 功能特性

- **多Agent协作**：5个专业Agent（构思、写作、编辑、读者、评论家）协同工作
- **智能构思**：与构思Agent多轮对话，逐步构建世界观、人物、大纲
- **AI写作**：根据大纲和设定生成章节内容，支持续写、润色
- **审稿讨论**：编辑、读者、评论家三个Agent并行审稿，提供建设性建议
- **记忆系统**：三层记忆架构（长期/短期/工作），保持上下文连贯
- **实时推送**：WebSocket支持，实时显示Agent讨论进度

## 技术栈

### 后端
- **语言**：Go 1.25
- **Web框架**：Gin
- **AI框架**：Eino（ByteDance）
- **数据库**：PostgreSQL + pgvector
- **ORM**：GORM
- **WebSocket**：gorilla/websocket

### 前端
- **框架**：React + TypeScript
- **构建工具**：Vite
- **UI**：Tailwind CSS
- **富文本编辑器**：TipTap
- **状态管理**：Zustand
- **HTTP客户端**：Axios

## 快速开始

### 前置要求

- Go 1.25+
- Node.js 18+
- PostgreSQL 14+（带pgvector扩展）
- DeepSeek API Key

### 1. 克隆仓库

```bash
git clone git@github.com:jianjunx/novelist.git
cd novelist
```

### 2. 设置数据库

```bash
# 创建数据库
psql -d postgres -c "CREATE DATABASE novelist;"

# 执行迁移
psql -d novelist -f backend/migrations/001_init.sql
```

### 3. 配置环境变量

```bash
export JWT_SECRET="your-secret-key-here"
export DEEPSEEK_API_KEY="your-deepseek-api-key"
export DATABASE_URL="postgres://localhost:5432/novelist?sslmode=disable"
```

### 4. 启动后端

```bash
cd backend
go run cmd/server/main.go
```

后端将在 http://localhost:8080 启动

### 5. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端将在 http://localhost:3000 启动

## API端点

### 认证
- `POST /api/auth/register` - 注册
- `POST /api/auth/login` - 登录
- `GET /api/auth/me` - 获取当前用户

### 项目
- `GET /api/projects` - 列出项目
- `POST /api/projects` - 创建项目
- `GET /api/projects/:id` - 获取项目详情
- `PUT /api/projects/:id` - 更新项目
- `DELETE /api/projects/:id` - 删除项目

### 人物
- `GET /api/projects/:id/characters` - 列出人物
- `POST /api/projects/:id/characters` - 创建人物
- `PUT /api/characters/:id` - 更新人物
- `DELETE /api/characters/:id` - 删除人物

### 世界观设定
- `GET /api/projects/:id/world-settings` - 列出设定
- `POST /api/projects/:id/world-settings` - 创建设定
- `PUT /api/world-settings/:id` - 更新设定
- `DELETE /api/world-settings/:id` - 删除设定

### 大纲
- `GET /api/projects/:id/outlines` - 列出大纲
- `POST /api/projects/:id/outlines` - 创建大纲
- `PUT /api/outlines/:id` - 更新大纲
- `DELETE /api/outlines/:id` - 删除大纲

### 章节
- `GET /api/projects/:id/chapters` - 列出章节
- `POST /api/projects/:id/chapters` - 创建章节
- `GET /api/chapters/:id` - 获取章节内容
- `PUT /api/chapters/:id` - 更新章节
- `DELETE /api/chapters/:id` - 删除章节

### AI操作
- `POST /api/creator/chat` - 构思Agent多轮对话
- `POST /api/chapters/:id/generate` - 生成章节内容
- `POST /api/chapters/:id/continue` - 续写章节
- `POST /api/chapters/:id/polish` - 润色选中内容
- `POST /api/chapters/:id/discuss` - 开始讨论工作流

### 设置
- `GET /api/settings` - 获取用户设置
- `PUT /api/settings` - 更新设置

### WebSocket
- `GET /ws` - 实时通信端点

## 项目结构

```
novelist/
├── backend/
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── api/             # HTTP处理器
│   │   ├── auth/            # JWT认证
│   │   ├── agent/           # Agent定义
│   │   ├── ai/              # Eino模型管理
│   │   ├── memory/          # 记忆系统
│   │   ├── orchestrator/    # 编排器
│   │   ├── model/           # 数据模型
│   │   └── store/           # 数据库访问
│   ├── migrations/          # 数据库迁移
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/             # API客户端
│   │   ├── stores/          # Zustand状态管理
│   │   ├── pages/           # 页面组件
│   │   ├── components/      # 通用组件
│   │   └── hooks/           # 自定义Hook
│   └── package.json
└── docs/
    ├── superpowers/
    │   ├── specs/           # 设计文档
    │   └── plans/           # 实施计划
    └── README.md
```

## Agent角色

| Agent | 职责 |
|-------|------|
| **构思Agent** | 世界观构建、人物设计、故事大纲 |
| **写作Agent** | 根据大纲和设定生成章节内容 |
| **编辑Agent** | 审查逻辑一致性、文笔质量、节奏、自然度 |
| **读者Agent** | 模拟读者视角，反馈阅读体验 |
| **评论家Agent** | 文学性分析、风格建议 |

## 记忆系统

三层记忆架构，确保Agent输出连贯：

- **长期记忆**：世界观设定、主线大纲、核心人物档案
- **短期记忆**：近期章节、相关设定（通过pgvector语义检索）
- **工作记忆**：当前讨论建议、用户最近修改

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License
