# 多Agent协作AI小说创作平台 - 实施计划

> **致自动化工作者：** 必须使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 技能来逐任务实施本计划。步骤使用复选框（`- [ ]`）语法进行跟踪。

**目标：** 构建一个支持多Agent协作的AI小说创作平台，让个人创作者能够与AI角色（编辑、读者、评论家）协作改进作品。

**架构：** 中心编排器控制工作流，Eino框架实现Agent层，DeepSeek作为主模型，PostgreSQL + pgvector存储，React前端实时展示讨论过程。

**技术栈：** Go (Gin + Eino) / React (Vite + TipTap + Zustand) / PostgreSQL + pgvector / WebSocket

---

## 第一阶段：项目初始化与基础设施

### Task 1: 初始化Go后端项目

**文件：**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/config/config.go`

- [ ] **Step 1: 初始化Go模块**

```bash
cd /home/jj/projects/novelist
mkdir -p backend/cmd/server backend/internal/config
cd backend
go mod init github.com/jj/novelist
```

- [ ] **Step 2: 安装核心依赖**

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get github.com/gorilla/websocket
go get github.com/google/uuid
go get golang.org/x/crypto
go get github.com/cloudwego/eino
go get github.com/cloudwego/eino-ext/components/model/deepseek
```

- [ ] **Step 3: 创建配置模块**

```go
// backend/internal/config/config.go
package config

import "os"

type Config struct {
    DatabaseURL string
    JWTSecret   string
    ServerPort  string
    DeepSeekKey string
    DeepSeekModel string
}

func Load() *Config {
    return &Config{
        DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost:5432/novelist?sslmode=disable"),
        JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
        ServerPort:    getEnv("SERVER_PORT", "8080"),
        DeepSeekKey:   getEnv("DEEPSEEK_API_KEY", ""),
        DeepSeekModel: getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

- [ ] **Step 4: 创建主入口**

```go
// backend/cmd/server/main.go
package main

import (
    "log"
    "github.com/jj/novelist/internal/config"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()
    _ = cfg

    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    log.Printf("Server starting on port %s", cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

- [ ] **Step 5: 验证服务器启动**

```bash
cd /home/jj/projects/novelist/backend
go run cmd/server/main.go &
curl http://localhost:8080/health
# 预期输出: {"status":"ok"}
kill %1
```

- [ ] **Step 6: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: initialize Go backend with Gin server and Eino dependencies"
```

---

### Task 2: 设置PostgreSQL与pgvector

**文件：**
- Create: `backend/migrations/001_init.sql`
- Create: `backend/internal/model/models.go`

- [ ] **Step 1: 创建数据库迁移文件**

```sql
-- backend/migrations/001_init.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 小说项目
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    genre TEXT,
    description TEXT,
    style_guide TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 人物档案
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    role TEXT,
    personality TEXT,
    background TEXT,
    appearance TEXT,
    relationships JSONB,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 世界观设定
CREATE TABLE world_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 故事大纲
CREATE TABLE outlines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    act INT,
    chapter_num INT,
    summary TEXT,
    key_events JSONB,
    status TEXT DEFAULT 'draft',
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 章节
CREATE TABLE chapters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    outline_id UUID REFERENCES outlines(id),
    chapter_num INT NOT NULL,
    title TEXT,
    content TEXT,
    word_count INT DEFAULT 0,
    status TEXT DEFAULT 'draft',
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 讨论记录
CREATE TABLE discussions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chapter_id UUID REFERENCES chapters(id) ON DELETE CASCADE,
    round_num INT NOT NULL,
    agent_role TEXT NOT NULL,
    content TEXT NOT NULL,
    suggestion_type TEXT,
    priority INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 对话记录（构思Agent多轮对话）
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 用户设置
CREATE TABLE settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    default_model TEXT DEFAULT 'deepseek-chat',
    deepseek_key TEXT,
    claude_key TEXT,
    openai_key TEXT,
    local_model_url TEXT,
    agent_model_config JSONB,
    default_word_count INT DEFAULT 800,
    discussion_rounds INT DEFAULT 1,
    language_style TEXT DEFAULT '现代中文',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 索引
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_characters_project_id ON characters(project_id);
CREATE INDEX idx_world_settings_project_id ON world_settings(project_id);
CREATE INDEX idx_outlines_project_id ON outlines(project_id);
CREATE INDEX idx_chapters_project_id ON chapters(project_id);
CREATE INDEX idx_chapters_outline_id ON chapters(outline_id);
CREATE INDEX idx_discussions_chapter_id ON discussions(chapter_id);
CREATE INDEX idx_conversations_project_id ON conversations(project_id);
```

- [ ] **Step 2: 执行迁移**

```bash
psql -d postgres -c "CREATE DATABASE novelist;"
psql -d novelist -f /home/jj/projects/novelist/backend/migrations/001_init.sql
```

- [ ] **Step 3: 创建Go数据模型**

```go
// backend/internal/model/models.go
package model

import (
    "time"
    "github.com/google/uuid"
    "github.com/lib/pq"
)

type User struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Username     string    `gorm:"uniqueIndex;not null" json:"username"`
    PasswordHash string    `gorm:"not null" json:"-"`
    CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    UserID      uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
    Title       string    `gorm:"not null" json:"title"`
    Genre       string    `json:"genre"`
    Description string    `json:"description"`
    StyleGuide  string    `json:"style_guide"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Character struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ProjectID     uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
    Name          string    `gorm:"not null" json:"name"`
    Role          string    `json:"role"`
    Personality   string    `json:"personality"`
    Background    string    `json:"background"`
    Appearance    string    `json:"appearance"`
    Relationships pq.Jsonb  `gorm:"type:jsonb" json:"relationships"`
    Embedding     []float32 `gorm:"type:vector(1536)" json:"-"`
    CreatedAt     time.Time `json:"created_at"`
}

type WorldSetting struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ProjectID uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
    Category  string    `gorm:"not null" json:"category"`
    Content   string    `gorm:"not null" json:"content"`
    Embedding []float32 `gorm:"type:vector(1536)" json:"-"`
    CreatedAt time.Time `json:"created_at"`
}

type Outline struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ProjectID  uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
    Act        int       `json:"act"`
    ChapterNum int       `json:"chapter_num"`
    Summary    string    `json:"summary"`
    KeyEvents  pq.Jsonb  `gorm:"type:jsonb" json:"key_events"`
    Status     string    `gorm:"default:draft" json:"status"`
    Embedding  []float32 `gorm:"type:vector(1536)" json:"-"`
    CreatedAt  time.Time `json:"created_at"`
}

type Chapter struct {
    ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ProjectID  uuid.UUID  `gorm:"type:uuid;index" json:"project_id"`
    OutlineID  *uuid.UUID `gorm:"type:uuid" json:"outline_id"`
    ChapterNum int        `gorm:"not null" json:"chapter_num"`
    Title      string     `json:"title"`
    Content    string     `json:"content"`
    WordCount  int        `gorm:"default:0" json:"word_count"`
    Status     string     `gorm:"default:draft" json:"status"`
    Embedding  []float32  `gorm:"type:vector(1536)" json:"-"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
}

type Discussion struct {
    ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ChapterID      uuid.UUID `gorm:"type:uuid;index" json:"chapter_id"`
    RoundNum       int       `gorm:"not null" json:"round_num"`
    AgentRole      string    `gorm:"not null" json:"agent_role"`
    Content        string    `gorm:"not null" json:"content"`
    SuggestionType string    `json:"suggestion_type"`
    Priority       int       `gorm:"default:0" json:"priority"`
    CreatedAt      time.Time `json:"created_at"`
}

type Conversation struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ProjectID uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
    Role      string    `gorm:"not null" json:"role"`
    Content   string    `gorm:"not null" json:"content"`
    CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
    ID               uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    UserID           uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"user_id"`
    DefaultModel     string    `gorm:"default:deepseek-chat" json:"default_model"`
    DeepSeekKey      string    `json:"deepseek_key"`
    ClaudeKey        string    `json:"claude_key"`
    OpenAIKey        string    `json:"openai_key"`
    LocalModelURL    string    `json:"local_model_url"`
    AgentModelConfig pq.Jsonb  `gorm:"type:jsonb" json:"agent_model_config"`
    DefaultWordCount int       `gorm:"default:800" json:"default_word_count"`
    DiscussionRounds int       `gorm:"default:1" json:"discussion_rounds"`
    LanguageStyle    string    `gorm:"default:现代中文" json:"language_style"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

- [ ] **Step 4: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/migrations/ backend/internal/model/
git commit -m "feat: add PostgreSQL schema with pgvector and Go models"
```

---

### Task 3: 初始化React前端项目

**文件：**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/src/main.tsx`
- Create: `frontend/src/App.tsx`

- [ ] **Step 1: 创建Vite项目**

```bash
cd /home/jj/projects/novelist
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install
```

- [ ] **Step 2: 安装依赖**

```bash
npm install zustand @tiptap/react @tiptap/starter-kit @tiptap/extension-placeholder
npm install react-router-dom axios
npm install -D tailwindcss @tailwindcss/vite
```

- [ ] **Step 3: 配置Vite**

```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
})
```

- [ ] **Step 4: 创建主入口和App**

```tsx
// frontend/src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
)
```

```tsx
// frontend/src/App.tsx
import { Routes, Route } from 'react-router-dom'

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <Routes>
        <Route path="/" element={<div>Home</div>} />
        <Route path="/login" element={<div>Login</div>} />
      </Routes>
    </div>
  )
}

export default App
```

- [ ] **Step 5: 验证前端启动**

```bash
cd /home/jj/projects/novelist/frontend
npm run dev
```

- [ ] **Step 6: 提交**

```bash
cd /home/jj/projects/novelist
git add frontend/
git commit -m "feat: initialize React frontend with Vite"
```

---

## 第二阶段：后端核心功能

### Task 4: 数据库连接与GORM配置

**文件：**
- Create: `backend/internal/store/database.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 创建数据库连接模块**

```go
// backend/internal/store/database.go
package store

import (
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/jj/novelist/internal/model"
)

var DB *gorm.DB

func InitDB(dsn string) {
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    if err := DB.AutoMigrate(
        &model.User{},
        &model.Project{},
        &model.Character{},
        &model.WorldSetting{},
        &model.Outline{},
        &model.Chapter{},
        &model.Discussion{},
        &model.Conversation{},
        &model.Setting{},
    ); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    log.Println("Database connected and migrated")
}

func GetDB() *gorm.DB {
    return DB
}
```

- [ ] **Step 2: 更新main.go使用数据库**

```go
// backend/cmd/server/main.go
package main

import (
    "log"
    "github.com/jj/novelist/internal/config"
    "github.com/jj/novelist/internal/store"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()
    store.InitDB(cfg.DatabaseURL)

    r := gin.Default()
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    log.Printf("Server starting on port %s", cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

- [ ] **Step 3: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: add GORM database connection and auto-migration"
```

---

### Task 5: JWT认证系统

**文件：**
- Create: `backend/internal/auth/jwt.go`
- Create: `backend/internal/auth/middleware.go`
- Create: `backend/internal/api/auth_handler.go`
- Create: `backend/internal/api/router.go`

- [ ] **Step 1: 创建JWT工具**

```go
// backend/internal/auth/jwt.go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

var jwtSecret []byte

func SetSecret(secret string) {
    jwtSecret = []byte(secret)
}

type Claims struct {
    UserID   uuid.UUID `json:"user_id"`
    Username string    `json:"username"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, username string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, jwt.ErrSignatureInvalid
}
```

- [ ] **Step 2: 创建认证中间件**

```go
// backend/internal/auth/middleware.go
package auth

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try Authorization header first
        tokenString := ""
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
        }
        // Fallback to query parameter (for WebSocket)
        if tokenString == "" {
            tokenString = c.Query("token")
        }
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
            c.Abort()
            return
        }

        claims, err := ParseToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}
```

- [ ] **Step 3: 创建认证API处理器**

```go
// backend/internal/api/auth_handler.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/jj/novelist/internal/auth"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
    "golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var existingUser model.User
    if err := store.GetDB().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
        return
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    user := model.User{Username: req.Username, PasswordHash: string(hashedPassword)}
    if err := store.GetDB().Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    // Create default settings for user
    settings := model.Setting{UserID: user.ID}
    store.GetDB().Create(&settings)

    token, err := auth.GenerateToken(user.ID, user.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "token": token,
        "user":  gin.H{"id": user.ID, "username": user.Username},
    })
}

func Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user model.User
    if err := store.GetDB().Where("username = ?", req.Username).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := auth.GenerateToken(user.ID, user.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "token": token,
        "user":  gin.H{"id": user.ID, "username": user.Username},
    })
}

func GetMe(c *gin.Context) {
    userID, _ := c.Get("user_id")
    username, _ := c.Get("username")
    c.JSON(http.StatusOK, gin.H{"user_id": userID, "username": username})
}
```

- [ ] **Step 4: 创建路由**

```go
// backend/internal/api/router.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/jj/novelist/internal/auth"
)

func SetupRouter(r *gin.Engine) {
    api := r.Group("/api")

    authGroup := api.Group("/auth")
    {
        authGroup.POST("/register", Register)
        authGroup.POST("/login", Login)
        authGroup.GET("/me", auth.AuthMiddleware(), GetMe)
    }

    protected := api.Group("")
    protected.Use(auth.AuthMiddleware())
    {
        // Routes will be added in subsequent tasks
        _ = protected
    }
}
```

- [ ] **Step 5: 更新main.go使用路由**

```go
// backend/cmd/server/main.go
package main

import (
    "log"
    "github.com/jj/novelist/internal/config"
    "github.com/jj/novelist/internal/store"
    "github.com/jj/novelist/internal/auth"
    "github.com/jj/novelist/internal/api"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()
    auth.SetSecret(cfg.JWTSecret)
    store.InitDB(cfg.DatabaseURL)

    r := gin.Default()
    api.SetupRouter(r)
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    log.Printf("Server starting on port %s", cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
```

- [ ] **Step 6: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement JWT authentication with WebSocket token support"
```

---

### Task 6: 项目与章节CRUD API

**文件：**
- Create: `backend/internal/api/project_handler.go`
- Create: `backend/internal/api/chapter_handler.go`
- Modify: `backend/internal/api/router.go`

- [ ] **Step 1: 创建项目处理器**

```go
// backend/internal/api/project_handler.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
)

type CreateProjectRequest struct {
    Title       string `json:"title" binding:"required"`
    Genre       string `json:"genre"`
    Description string `json:"description"`
    StyleGuide  string `json:"style_guide"`
}

func GetProjects(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var projects []model.Project
    if err := store.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&projects).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
        return
    }
    c.JSON(http.StatusOK, projects)
}

func CreateProject(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var req CreateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    project := model.Project{
        UserID: userID.(uuid.UUID), Title: req.Title,
        Genre: req.Genre, Description: req.Description, StyleGuide: req.StyleGuide,
    }
    if err := store.GetDB().Create(&project).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
        return
    }
    c.JSON(http.StatusCreated, project)
}

func GetProject(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }
    c.JSON(http.StatusOK, project)
}

func UpdateProject(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }
    var req CreateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    store.GetDB().Model(&project).Updates(map[string]interface{}{
        "title": req.Title, "genre": req.Genre,
        "description": req.Description, "style_guide": req.StyleGuide,
    })
    c.JSON(http.StatusOK, project)
}

func DeleteProject(c *gin.Context) {
    userID, _ := c.Get("user_id")
    result := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).Delete(&model.Project{})
    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}
```

- [ ] **Step 2: 创建章节处理器**

```go
// backend/internal/api/chapter_handler.go
package api

import (
    "net/http"
    "unicode/utf8"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
)

type CreateChapterRequest struct {
    OutlineID  *string `json:"outline_id"`
    ChapterNum int     `json:"chapter_num" binding:"required"`
    Title      string  `json:"title" binding:"required"`
    Content    string  `json:"content"`
}

func GetChapters(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }
    var chapters []model.Chapter
    store.GetDB().Where("project_id = ?", c.Param("id")).Order("chapter_num").Find(&chapters)
    c.JSON(http.StatusOK, chapters)
}

func CreateChapter(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
        return
    }
    var req CreateChapterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    chapter := model.Chapter{
        ProjectID: uuid.MustParse(c.Param("id")), ChapterNum: req.ChapterNum,
        Title: req.Title, Content: req.Content, WordCount: utf8.RuneCountInString(req.Content),
    }
    if req.OutlineID != nil {
        oid := uuid.MustParse(*req.OutlineID)
        chapter.OutlineID = &oid
    }
    store.GetDB().Create(&chapter)
    c.JSON(http.StatusCreated, chapter)
}

func GetChapter(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var chapter model.Chapter
    if err := store.GetDB().Joins("Project").Where("chapters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&chapter).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
        return
    }
    c.JSON(http.StatusOK, chapter)
}

func UpdateChapter(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var chapter model.Chapter
    if err := store.GetDB().Joins("Project").Where("chapters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&chapter).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
        return
    }
    var req struct {
        Title   string `json:"title"`
        Content string `json:"content"`
        Status  string `json:"status"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    updates := map[string]interface{}{}
    if req.Title != "" { updates["title"] = req.Title }
    if req.Content != "" { updates["content"] = req.Content; updates["word_count"] = utf8.RuneCountInString(req.Content) }
    if req.Status != "" { updates["status"] = req.Status }
    store.GetDB().Model(&chapter).Updates(updates)
    c.JSON(http.StatusOK, chapter)
}
```

- [ ] **Step 3: 更新路由**

```go
// backend/internal/api/router.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/jj/novelist/internal/auth"
)

func SetupRouter(r *gin.Engine) {
    api := r.Group("/api")

    authGroup := api.Group("/auth")
    {
        authGroup.POST("/register", Register)
        authGroup.POST("/login", Login)
        authGroup.GET("/me", auth.AuthMiddleware(), GetMe)
    }

    protected := api.Group("")
    protected.Use(auth.AuthMiddleware())
    {
        projects := protected.Group("/projects")
        {
            projects.GET("", GetProjects)
            projects.POST("", CreateProject)
            projects.GET("/:id", GetProject)
            projects.PUT("/:id", UpdateProject)
            projects.DELETE("/:id", DeleteProject)
            projects.GET("/:id/chapters", GetChapters)
            projects.POST("/:id/chapters", CreateChapter)
        }
        protected.GET("/chapters/:id", GetChapter)
        protected.PUT("/chapters/:id", UpdateChapter)
    }
}
```

- [ ] **Step 4: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement project and chapter CRUD APIs"
```

---

### Task 7: 人物、世界观、大纲、设置API

**文件：**
- Create: `backend/internal/api/character_handler.go`
- Create: `backend/internal/api/world_setting_handler.go`
- Create: `backend/internal/api/outline_handler.go`
- Create: `backend/internal/api/settings_handler.go`
- Modify: `backend/internal/api/router.go`

- [ ] **Step 1: 创建人物处理器**

```go
// backend/internal/api/character_handler.go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
    "github.com/lib/pq"
)

type CreateCharacterRequest struct {
    Name          string                 `json:"name" binding:"required"`
    Role          string                 `json:"role"`
    Personality   string                 `json:"personality"`
    Background    string                 `json:"background"`
    Appearance    string                 `json:"appearance"`
    Relationships map[string]interface{} `json:"relationships"`
}

func GetCharacters(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"}); return
    }
    var characters []model.Character
    store.GetDB().Where("project_id = ?", c.Param("id")).Find(&characters)
    c.JSON(http.StatusOK, characters)
}

func CreateCharacter(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"}); return
    }
    var req CreateCharacterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    relJSON, _ := json.Marshal(req.Relationships)
    character := model.Character{
        ProjectID: uuid.MustParse(c.Param("id")), Name: req.Name, Role: req.Role,
        Personality: req.Personality, Background: req.Background, Appearance: req.Appearance,
        Relationships: pq.Jsonb{RawMessage: relJSON},
    }
    store.GetDB().Create(&character)
    c.JSON(http.StatusCreated, character)
}

func UpdateCharacter(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var character model.Character
    if err := store.GetDB().Joins("Project").Where("characters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&character).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"}); return
    }
    var req CreateCharacterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    relJSON, _ := json.Marshal(req.Relationships)
    store.GetDB().Model(&character).Updates(map[string]interface{}{
        "name": req.Name, "role": req.Role, "personality": req.Personality,
        "background": req.Background, "appearance": req.Appearance,
        "relationships": pq.Jsonb{RawMessage: relJSON},
    })
    c.JSON(http.StatusOK, character)
}
```

- [ ] **Step 2: 创建世界观设定处理器**

```go
// backend/internal/api/world_setting_handler.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
)

func GetWorldSettings(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"}); return
    }
    var settings []model.WorldSetting
    store.GetDB().Where("project_id = ?", c.Param("id")).Find(&settings)
    c.JSON(http.StatusOK, settings)
}

func CreateWorldSetting(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"}); return
    }
    var req struct { Category string `json:"category" binding:"required"`; Content string `json:"content" binding:"required"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    setting := model.WorldSetting{ProjectID: uuid.MustParse(c.Param("id")), Category: req.Category, Content: req.Content}
    store.GetDB().Create(&setting)
    c.JSON(http.StatusCreated, setting)
}

func UpdateWorldSetting(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var setting model.WorldSetting
    if err := store.GetDB().Joins("Project").Where("world_settings.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&setting).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"}); return
    }
    var req struct { Category string `json:"category"`; Content string `json:"content"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    store.GetDB().Model(&setting).Updates(map[string]interface{}{"category": req.Category, "content": req.Content})
    c.JSON(http.StatusOK, setting)
}
```

- [ ] **Step 3: 创建大纲处理器**

```go
// backend/internal/api/outline_handler.go
package api

import (
    "encoding/json"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
    "github.com/lib/pq"
)

func GetOutlines(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"}); return
    }
    var outlines []model.Outline
    store.GetDB().Where("project_id = ?", c.Param("id")).Order("act, chapter_num").Find(&outlines)
    c.JSON(http.StatusOK, outlines)
}

func CreateOutline(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var project model.Project
    if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"}); return
    }
    var req struct {
        Act int `json:"act"`; ChapterNum int `json:"chapter_num" binding:"required"`
        Summary string `json:"summary" binding:"required"`; KeyEvents map[string]interface{} `json:"key_events"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    keJSON, _ := json.Marshal(req.KeyEvents)
    outline := model.Outline{
        ProjectID: uuid.MustParse(c.Param("id")), Act: req.Act, ChapterNum: req.ChapterNum,
        Summary: req.Summary, KeyEvents: pq.Jsonb{RawMessage: keJSON},
    }
    store.GetDB().Create(&outline)
    c.JSON(http.StatusCreated, outline)
}

func UpdateOutline(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var outline model.Outline
    if err := store.GetDB().Joins("Project").Where("outlines.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&outline).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Outline not found"}); return
    }
    var req struct {
        Act int `json:"act"`; ChapterNum int `json:"chapter_num"`
        Summary string `json:"summary"`; KeyEvents map[string]interface{} `json:"key_events"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    keJSON, _ := json.Marshal(req.KeyEvents)
    store.GetDB().Model(&outline).Updates(map[string]interface{}{
        "act": req.Act, "chapter_num": req.ChapterNum, "summary": req.Summary,
        "key_events": pq.Jsonb{RawMessage: keJSON},
    })
    c.JSON(http.StatusOK, outline)
}
```

- [ ] **Step 4: 创建设置处理器**

```go
// backend/internal/api/settings_handler.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
)

func GetSettings(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var settings model.Setting
    if err := store.GetDB().Where("user_id = ?", userID).First(&settings).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"}); return
    }
    c.JSON(http.StatusOK, settings)
}

func UpdateSettings(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var settings model.Setting
    if err := store.GetDB().Where("user_id = ?", userID).First(&settings).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"}); return
    }
    var req struct {
        DefaultModel     string `json:"default_model"`
        DeepSeekKey      string `json:"deepseek_key"`
        ClaudeKey        string `json:"claude_key"`
        OpenAIKey        string `json:"openai_key"`
        LocalModelURL    string `json:"local_model_url"`
        DefaultWordCount int    `json:"default_word_count"`
        DiscussionRounds int    `json:"discussion_rounds"`
        LanguageStyle    string `json:"language_style"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }
    store.GetDB().Model(&settings).Updates(map[string]interface{}{
        "default_model": req.DefaultModel, "deepseek_key": req.DeepSeekKey,
        "claude_key": req.ClaudeKey, "openai_key": req.OpenAIKey,
        "local_model_url": req.LocalModelURL, "default_word_count": req.DefaultWordCount,
        "discussion_rounds": req.DiscussionRounds, "language_style": req.LanguageStyle,
    })
    c.JSON(http.StatusOK, settings)
}
```

- [ ] **Step 5: 更新路由**

```go
// backend/internal/api/router.go - 在protected组内添加
projects.GET("/:id/characters", GetCharacters)
projects.POST("/:id/characters", CreateCharacter)
projects.GET("/:id/world-settings", GetWorldSettings)
projects.POST("/:id/world-settings", CreateWorldSetting)
projects.GET("/:id/outlines", GetOutlines)
projects.POST("/:id/outlines", CreateOutline)

protected.PUT("/characters/:id", UpdateCharacter)
protected.PUT("/world-settings/:id", UpdateWorldSetting)
protected.PUT("/outlines/:id", UpdateOutline)
protected.GET("/settings", GetSettings)
protected.PUT("/settings", UpdateSettings)
```

- [ ] **Step 6: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement character, world setting, outline, and settings APIs"
```

---

## 第三阶段：AI/Agent层 (Eino框架)

### Task 8: Eino DeepSeek模型集成

**文件：**
- Create: `backend/internal/ai/model.go`

- [ ] **Step 1: 创建Eino模型管理器**

```go
// backend/internal/ai/model.go
package ai

import (
    "context"
    "log"
    "sync"
    "github.com/cloudwego/eino-ext/components/model/deepseek"
    "github.com/cloudwego/eino/schema"
    "github.com/jj/novelist/internal/config"
)

type ModelManager struct {
    models map[string]*deepseek.ChatModel
    mu     sync.RWMutex
}

var Manager *ModelManager

func InitModelManager(cfg *config.Config) {
    Manager = &ModelManager{models: make(map[string]*deepseek.ChatModel)}

    if cfg.DeepSeekKey != "" {
        cm, err := deepseek.NewChatModel(context.Background(), &deepseek.ChatModelConfig{
            APIKey:      cfg.DeepSeekKey,
            Model:       cfg.DeepSeekModel,
            MaxTokens:   4096,
            Temperature: 0.7,
        })
        if err != nil {
            log.Printf("Failed to init DeepSeek model: %v", err)
        } else {
            Manager.models["deepseek"] = cm
            log.Println("DeepSeek model initialized")
        }
    }
}

func (m *ModelManager) GetModel(name string) (*deepseek.ChatModel, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    model, ok := m.models[name]
    return model, ok
}

func (m *ModelManager) GetDefault() (*deepseek.ChatModel, bool) {
    return m.GetModel("deepseek")
}

// Chat sends a chat request and returns the response
func Chat(ctx context.Context, model *deepseek.ChatModel, systemPrompt string, messages []Message) (string, error) {
    einoMessages := []*schema.Message{
        {Role: schema.System, Content: systemPrompt},
    }
    for _, msg := range messages {
        role := schema.User
        if msg.Role == "assistant" {
            role = schema.Assistant
        }
        einoMessages = append(einoMessages, &schema.Message{Role: role, Content: msg.Content})
    }

    resp, err := model.Generate(ctx, einoMessages)
    if err != nil {
        return "", err
    }
    return resp.Content, nil
}

// ChatStream sends a chat request and returns a stream
func ChatStream(ctx context.Context, model *deepseek.ChatModel, systemPrompt string, messages []Message) (*schema.StreamReader[*schema.Message], error) {
    einoMessages := []*schema.Message{
        {Role: schema.System, Content: systemPrompt},
    }
    for _, msg := range messages {
        role := schema.User
        if msg.Role == "assistant" {
            role = schema.Assistant
        }
        einoMessages = append(einoMessages, &schema.Message{Role: role, Content: msg.Content})
    }

    return model.Stream(ctx, einoMessages)
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

- [ ] **Step 2: 更新main.go初始化模型**

```go
// backend/cmd/server/main.go - 在main()中添加
ai.InitModelManager(cfg)
```

- [ ] **Step 3: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: integrate Eino DeepSeek chat model"
```

---

### Task 9: Agent系统提示与定义

**文件：**
- Create: `backend/internal/agent/prompts.go`
- Create: `backend/internal/agent/agent.go`

- [ ] **Step 1: 创建Agent系统提示**

```go
// backend/internal/agent/prompts.go
package agent

const CreatorPrompt = `你是一个专业的小说构思助手。你的任务是帮助作者构建小说的世界观、人物设定和故事大纲。

工作流程：
1. 了解作者想写什么类型的小说（玄幻、都市、科幻、言情等）
2. 询问风格偏好（轻松、热血、悬疑、文艺等）
3. 了解核心冲突和主题
4. 逐步构建人物设定
5. 设计世界观
6. 生成故事大纲

要求：
- 主动提问，引导作者表达创意
- 提供专业的建议和参考
- 尊重作者的创意方向
- 输出结构化内容（大纲、人物、世界观）`

const WriterPrompt = `你是一个专业的小说写作助手。根据大纲和设定生成高质量的章节内容。

写作准则：
1. 保持人物性格一致
2. 遵循世界观设定
3. 注意情节逻辑
4. 对话自然生动
5. 描写细腻但不冗长
6. 节奏张弛有度

去AI味要求：
- 避免过于工整的句式
- 使用多样化的表达方式
- 对话要口语化、有个性
- 避免重复的过渡词
- 加入细节和不完美，让文字更有人味`

const EditorPrompt = `你是一个专业的小说编辑。审查章节内容，提供详细的修改建议。

审查维度：
1. 逻辑一致性：情节是否合理，有无矛盾
2. 文笔质量：表达是否准确、生动
3. 节奏把控：张弛是否得当
4. 风格统一：是否符合整体风格
5. 自然度：是否有AI痕迹，是否有人味

输出格式（JSON数组）：
[{"type":"逻辑|文笔|节奏|风格|自然度","location":"具体位置","problem":"问题描述","suggestion":"修改建议","priority":1-3}]

优先级：1=高，2=中，3=低`

const ReaderPrompt = `你是一个普通读者，正在阅读这部小说。从读者角度给出真实的阅读体验反馈。

反馈内容：
1. 吸引力：开头是否抓人
2. 节奏感：是否拖沓
3. 代入感：是否能代入人物
4. 困惑点：是否有看不懂的地方
5. 惊喜点：是否有意料之外的情节

用第一人称表达，说真话，不要客套。`

const CriticPrompt = `你是一个文学评论家，擅长分析小说的艺术价值和风格特点。

分析维度：
1. 主题深度
2. 人物塑造
3. 叙事技巧
4. 语言风格
5. 文学价值

提供专业分析，引用具体例子，给出建设性建议。`
```

- [ ] **Step 2: 创建Agent调用封装**

```go
// backend/internal/agent/agent.go
package agent

import (
    "context"
    "github.com/jj/novelist/internal/ai"
)

type AgentRole string

const (
    RoleCreator AgentRole = "creator"
    RoleWriter  AgentRole = "writer"
    RoleEditor  AgentRole = "editor"
    RoleReader  AgentRole = "reader"
    RoleCritic  AgentRole = "critic"
)

func GetPrompt(role AgentRole) string {
    switch role {
    case RoleCreator: return CreatorPrompt
    case RoleWriter: return WriterPrompt
    case RoleEditor: return EditorPrompt
    case RoleReader: return ReaderPrompt
    case RoleCritic: return CriticPrompt
    default: return ""
    }
}

func Chat(ctx context.Context, role AgentRole, messages []ai.Message) (string, error) {
    model, ok := ai.Manager.GetDefault()
    if !ok {
        return "", fmt.Errorf("no model available")
    }
    return ai.Chat(ctx, model, GetPrompt(role), messages)
}

func ChatStream(ctx context.Context, role AgentRole, messages []ai.Message) (*schema.StreamReader[*schema.Message], error) {
    model, ok := ai.Manager.GetDefault()
    if !ok {
        return nil, fmt.Errorf("no model available")
    }
    return ai.ChatStream(ctx, model, GetPrompt(role), messages)
}
```

- [ ] **Step 3: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement agent system prompts and Eino-based agent calls"
```

---

### Task 10: 记忆系统（含pgvector语义检索）

**文件：**
- Create: `backend/internal/memory/memory.go`

- [ ] **Step 1: 创建记忆系统**

```go
// backend/internal/memory/memory.go
package memory

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
)

type Memory struct {
    ProjectID uuid.UUID
}

func NewMemory(projectID uuid.UUID) *Memory {
    return &Memory{ProjectID: projectID}
}

// LoadLongTermMemory loads project info, world settings, outlines, and characters
func (m *Memory) LoadLongTermMemory(ctx context.Context) (string, error) {
    var project model.Project
    if err := store.GetDB().Where("id = ?", m.ProjectID).First(&project).Error; err != nil {
        return "", fmt.Errorf("project not found: %w", err)
    }

    var settings []model.WorldSetting
    store.GetDB().Where("project_id = ?", m.ProjectID).Find(&settings)

    var outlines []model.Outline
    store.GetDB().Where("project_id = ?", m.ProjectID).Order("act, chapter_num").Find(&outlines)

    var characters []model.Character
    store.GetDB().Where("project_id = ?", m.ProjectID).Find(&characters)

    ctx := fmt.Sprintf("## 项目信息\n标题: %s\n类型: %s\n风格: %s\n\n", project.Title, project.Genre, project.StyleGuide)

    ctx += "## 世界观设定\n"
    for _, s := range settings {
        ctx += fmt.Sprintf("- [%s] %s\n", s.Category, s.Content)
    }

    ctx += "\n## 人物档案\n"
    for _, c := range characters {
        ctx += fmt.Sprintf("- %s（%s）: %s, %s\n", c.Name, c.Role, c.Personality, c.Background)
    }

    ctx += "\n## 故事大纲\n"
    for _, o := range outlines {
        ctx += fmt.Sprintf("- 第%d章: %s\n", o.ChapterNum, o.Summary)
    }

    return ctx, nil
}

// LoadShortTermMemory loads recent chapters
func (m *Memory) LoadShortTermMemory(ctx context.Context, currentChapterNum int) (string, error) {
    var chapters []model.Chapter
    store.GetDB().Where("project_id = ? AND chapter_num < ?", m.ProjectID, currentChapterNum).
        Order("chapter_num DESC").Limit(5).Find(&chapters)

    result := "## 近期章节\n"
    for i := len(chapters) - 1; i >= 0; i-- {
        c := chapters[i]
        result += fmt.Sprintf("\n### 第%d章 %s\n%s\n", c.ChapterNum, c.Title, c.Content)
    }
    return result, nil
}

// SemanticSearch searches for related characters and settings using pgvector
func (m *Memory) SemanticSearch(ctx context.Context, queryEmbedding []float32, limit int) (string, error) {
    // Search related characters
    var characters []model.Character
    store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
        Order("embedding <-> ?", queryEmbedding).Limit(limit).Find(&characters)

    // Search related world settings
    var settings []model.WorldSetting
    store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
        Order("embedding <-> ?", queryEmbedding).Limit(limit).Find(&settings)

    result := "## 语义相关设定\n"
    for _, c := range characters {
        result += fmt.Sprintf("- 人物 %s: %s\n", c.Name, c.Personality)
    }
    for _, s := range settings {
        result += fmt.Sprintf("- [%s] %s\n", s.Category, s.Content)
    }
    return result, nil
}

// AssembleContext builds full context for an agent call
func (m *Memory) AssembleContext(ctx context.Context, currentChapterNum int, workingMemory string) (string, error) {
    longTerm, err := m.LoadLongTermMemory(ctx)
    if err != nil {
        return "", err
    }

    shortTerm, err := m.LoadShortTermMemory(ctx, currentChapterNum)
    if err != nil {
        return "", err
    }

    fullContext := longTerm + "\n" + shortTerm
    if workingMemory != "" {
        fullContext += "\n## 当前任务上下文\n" + workingMemory
    }
    return fullContext, nil
}
```

- [ ] **Step 2: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement memory system with pgvector semantic search"
```

---

### Task 11: 编排器与工作流引擎

**文件：**
- Create: `backend/internal/orchestrator/orchestrator.go`

- [ ] **Step 1: 创建编排器**

```go
// backend/internal/orchestrator/orchestrator.go
package orchestrator

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/agent"
    "github.com/jj/novelist/internal/ai"
    "github.com/jj/novelist/internal/memory"
    "github.com/jj/novelist/internal/model"
    "github.com/jj/novelist/internal/store"
)

type Orchestrator struct{}

func NewOrchestrator() *Orchestrator {
    return &Orchestrator{}
}

// CreatorChat handles multi-round conversation with Creator Agent
func (o *Orchestrator) CreatorChat(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, messages []ai.Message) (string, error) {
    // Save user message to conversations
    if projectID != uuid.Nil {
        store.GetDB().Create(&model.Conversation{
            ProjectID: projectID, Role: "user", Content: messages[len(messages)-1].Content,
        })
    }

    // Load project context if exists
    var contextStr string
    if projectID != uuid.Nil {
        mem := memory.NewMemory(projectID)
        longTerm, err := mem.LoadLongTermMemory(ctx)
        if err == nil {
            contextStr = longTerm
        }
    }

    // Prepend context
    if contextStr != "" {
        messages = append([]ai.Message{{Role: "system", Content: "当前项目上下文：\n" + contextStr}}, messages...)
    }

    resp, err := agent.Chat(ctx, agent.RoleCreator, messages)
    if err != nil {
        return "", err
    }

    // Save agent response
    if projectID != uuid.Nil {
        store.GetDB().Create(&model.Conversation{
            ProjectID: projectID, Role: "assistant", Content: resp,
        })
    }

    return resp, nil
}

// GenerateChapter generates chapter content using Writer Agent
func (o *Orchestrator) GenerateChapter(ctx context.Context, chapterID uuid.UUID) (string, error) {
    var chapter model.Chapter
    if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
        return "", fmt.Errorf("chapter not found: %w", err)
    }

    mem := memory.NewMemory(chapter.ProjectID)
    contextStr, err := mem.AssembleContext(ctx, chapter.ChapterNum, "")
    if err != nil {
        return "", err
    }

    messages := []ai.Message{
        {Role: "user", Content: fmt.Sprintf("请生成第%d章内容。\n\n%s\n\n章节标题：%s", chapter.ChapterNum, contextStr, chapter.Title)},
    }

    return agent.Chat(ctx, agent.RoleWriter, messages)
}

// ContinueWriting continues writing from current content
func (o *Orchestrator) ContinueWriting(ctx context.Context, chapterID uuid.UUID, currentContent string) (string, error) {
    var chapter model.Chapter
    if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
        return "", fmt.Errorf("chapter not found: %w", err)
    }

    mem := memory.NewMemory(chapter.ProjectID)
    contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, "")

    messages := []ai.Message{
        {Role: "user", Content: fmt.Sprintf("请续写以下内容（500-1000字）：\n\n%s\n\n上下文：\n%s", currentContent, contextStr)},
    }

    return agent.Chat(ctx, agent.RoleWriter, messages)
}

// PolishContent polishes selected content
func (o *Orchestrator) PolishContent(ctx context.Context, chapterID uuid.UUID, content string) (string, error) {
    messages := []ai.Message{
        {Role: "user", Content: fmt.Sprintf("请润色以下文字，保持原意，提升表达质量：\n\n%s", content)},
    }
    return agent.Chat(ctx, agent.RoleWriter, messages)
}

// Suggestion represents a structured suggestion from an agent
type Suggestion struct {
    Type     string `json:"type"`
    Location string `json:"location"`
    Problem  string `json:"problem"`
    Suggestion string `json:"suggestion"`
    Priority int    `json:"priority"`
}

// DiscussionResult represents the aggregated discussion result
type DiscussionResult struct {
    EditorSuggestions  []Suggestion `json:"editor_suggestions"`
    ReaderFeedback     string       `json:"reader_feedback"`
    CriticAnalysis     string       `json:"critic_analysis"`
    Aggregated         []Suggestion `json:"aggregated"`
}

// StartDiscussion starts the discussion workflow with all review agents
func (o *Orchestrator) StartDiscussion(ctx context.Context, chapterID uuid.UUID) (*DiscussionResult, error) {
    var chapter model.Chapter
    if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
        return nil, fmt.Errorf("chapter not found: %w", err)
    }

    mem := memory.NewMemory(chapter.ProjectID)
    contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, "")

    reviewPrompt := fmt.Sprintf("请审查以下章节：\n\n%s\n\n章节内容：\n%s", contextStr, chapter.Content)
    messages := []ai.Message{{Role: "user", Content: reviewPrompt}}

    // Run review agents in parallel
    var wg sync.WaitGroup
    var mu sync.Mutex
    result := &DiscussionResult{}

    // Editor Agent
    wg.Add(1)
    go func() {
        defer wg.Done()
        resp, err := agent.Chat(ctx, agent.RoleEditor, messages)
        if err != nil {
            return
        }
        // Parse suggestions from JSON
        var suggestions []Suggestion
        json.Unmarshal([]byte(resp), &suggestions)
        mu.Lock()
        result.EditorSuggestions = suggestions
        mu.Unlock()

        // Save to DB
        for _, s := range suggestions {
            store.GetDB().Create(&model.Discussion{
                ChapterID: chapterID, RoundNum: 1, AgentRole: "editor",
                Content: s.Problem, SuggestionType: s.Type, Priority: s.Priority,
            })
        }
    }()

    // Reader Agent
    wg.Add(1)
    go func() {
        defer wg.Done()
        resp, err := agent.Chat(ctx, agent.RoleReader, messages)
        if err != nil {
            return
        }
        mu.Lock()
        result.ReaderFeedback = resp
        mu.Unlock()

        store.GetDB().Create(&model.Discussion{
            ChapterID: chapterID, RoundNum: 1, AgentRole: "reader", Content: resp,
        })
    }()

    // Critic Agent
    wg.Add(1)
    go func() {
        defer wg.Done()
        resp, err := agent.Chat(ctx, agent.RoleCritic, messages)
        if err != nil {
            return
        }
        mu.Lock()
        result.CriticAnalysis = resp
        mu.Unlock()

        store.GetDB().Create(&model.Discussion{
            ChapterID: chapterID, RoundNum: 1, AgentRole: "critic", Content: resp,
        })
    }()

    wg.Wait()

    // Aggregate and deduplicate suggestions
    result.Aggregated = aggregateSuggestions(result.EditorSuggestions)

    return result, nil
}

// aggregateSuggestions deduplicates and sorts suggestions by priority
func aggregateSuggestions(suggestions []Suggestion) []Suggestion {
    // Simple deduplication by problem text
    seen := make(map[string]bool)
    var result []Suggestion
    for _, s := range suggestions {
        if !seen[s.Problem] {
            seen[s.Problem] = true
            result = append(result, s)
        }
    }
    // Sort by priority (1 = highest)
    sort.Slice(result, func(i, j int) bool {
        return result[i].Priority < result[j].Priority
    })
    return result
}
```

- [ ] **Step 2: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement orchestrator with discussion workflow and suggestion aggregation"
```

---

### Task 12: AI操作API端点

**文件：**
- Create: `backend/internal/api/ai_handler.go`
- Modify: `backend/internal/api/router.go`

- [ ] **Step 1: 创建AI操作处理器**

```go
// backend/internal/api/ai_handler.go
package api

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jj/novelist/internal/orchestrator"
    "github.com/jj/novelist/internal/store"
    "github.com/jj/novelist/internal/model"
)

var orch = orchestrator.NewOrchestrator()

// CreatorChat handles multi-round conversation with Creator Agent
func CreatorChat(c *gin.Context) {
    userID, _ := c.Get("user_id")
    var req struct {
        ProjectID string            `json:"project_id"`
        Messages  []orchestrator.Message `json:"messages" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }

    var projectID uuid.UUID
    if req.ProjectID != "" {
        projectID = uuid.MustParse(req.ProjectID)
    }

    resp, err := orch.CreatorChat(c.Request.Context(), userID.(uuid.UUID), projectID, req.Messages)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }

    c.JSON(http.StatusOK, gin.H{"content": resp})
}

// GenerateChapter generates chapter content
func GenerateChapter(c *gin.Context) {
    chapterID := c.Param("id")
    resp, err := orch.GenerateChapter(c.Request.Context(), uuid.MustParse(chapterID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }

    // Update chapter content
    store.GetDB().Model(&model.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
        "content": resp, "word_count": len([]rune(resp)),
    })

    c.JSON(http.StatusOK, gin.H{"content": resp})
}

// ContinueWriting continues writing from current content
func ContinueWriting(c *gin.Context) {
    chapterID := c.Param("id")
    var req struct { Content string `json:"content" binding:"required"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }

    resp, err := orch.ContinueWriting(c.Request.Context(), uuid.MustParse(chapterID), req.Content)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }

    c.JSON(http.StatusOK, gin.H{"content": resp})
}

// PolishContent polishes selected content
func PolishContent(c *gin.Context) {
    chapterID := c.Param("id")
    var req struct { Content string `json:"content" binding:"required"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return
    }

    resp, err := orch.PolishContent(c.Request.Context(), uuid.MustParse(chapterID), req.Content)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }

    c.JSON(http.StatusOK, gin.H{"content": resp})
}

// StartDiscussion starts the discussion workflow
func StartDiscussion(c *gin.Context) {
    chapterID := c.Param("id")
    result, err := orch.StartDiscussion(c.Request.Context(), uuid.MustParse(chapterID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return
    }

    c.JSON(http.StatusOK, result)
}
```

- [ ] **Step 2: 更新路由**

```go
// backend/internal/api/router.go - 在protected组内添加
protected.POST("/creator/chat", CreatorChat)
protected.POST("/chapters/:id/generate", GenerateChapter)
protected.POST("/chapters/:id/continue", ContinueWriting)
protected.POST("/chapters/:id/polish", PolishContent)
protected.POST("/chapters/:id/discuss", StartDiscussion)
```

- [ ] **Step 3: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/
git commit -m "feat: implement AI operation API endpoints"
```

---

## 第四阶段：前端核心功能

### Task 13: 认证与项目管理前端

**文件：**
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/stores/authStore.ts`
- Create: `frontend/src/stores/projectStore.ts`
- Create: `frontend/src/pages/Login.tsx`
- Create: `frontend/src/pages/Dashboard.tsx`
- Create: `frontend/src/components/ProjectCard.tsx`

- [ ] **Step 1: 创建API客户端和状态管理**

```typescript
// frontend/src/api/client.ts
import axios from 'axios'
const api = axios.create({ baseURL: '/api' })
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})
api.interceptors.response.use(
  (r) => r,
  (e) => { if (e.response?.status === 401) { localStorage.removeItem('token'); window.location.href = '/login' }; return Promise.reject(e) }
)
export default api
```

```typescript
// frontend/src/stores/authStore.ts
import { create } from 'zustand'
import api from '../api/client'

interface User { id: string; username: string }
interface AuthState {
  user: User | null; token: string | null; isLoading: boolean; error: string | null
  login: (u: string, p: string) => Promise<void>
  register: (u: string, p: string) => Promise<void>
  logout: () => void; checkAuth: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null, token: localStorage.getItem('token'), isLoading: false, error: null,
  login: async (u, p) => {
    set({ isLoading: true, error: null })
    try {
      const { data } = await api.post('/auth/login', { username: u, password: p })
      localStorage.setItem('token', data.token); set({ user: data.user, token: data.token, isLoading: false })
    } catch (e: any) { set({ error: e.response?.data?.error || 'Login failed', isLoading: false }); throw e }
  },
  register: async (u, p) => {
    set({ isLoading: true, error: null })
    try {
      const { data } = await api.post('/auth/register', { username: u, password: p })
      localStorage.setItem('token', data.token); set({ user: data.user, token: data.token, isLoading: false })
    } catch (e: any) { set({ error: e.response?.data?.error || 'Registration failed', isLoading: false }); throw e }
  },
  logout: () => { localStorage.removeItem('token'); set({ user: null, token: null }) },
  checkAuth: async () => {
    const token = localStorage.getItem('token')
    if (!token) { set({ user: null }); return }
    try { const { data } = await api.get('/auth/me'); set({ user: data }) }
    catch { localStorage.removeItem('token'); set({ user: null, token: null }) }
  },
}))
```

```typescript
// frontend/src/stores/projectStore.ts
import { create } from 'zustand'
import api from '../api/client'

interface Project { id: string; title: string; genre: string; description: string; style_guide: string; created_at: string; updated_at: string }
interface ProjectState {
  projects: Project[]; currentProject: Project | null; isLoading: boolean
  fetchProjects: () => Promise<void>; createProject: (d: Partial<Project>) => Promise<Project>
  setCurrentProject: (p: Project | null) => void
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  projects: [], currentProject: null, isLoading: false,
  fetchProjects: async () => { set({ isLoading: true }); const { data } = await api.get('/projects'); set({ projects: data, isLoading: false }) },
  createProject: async (d) => { const { data } = await api.post('/projects', d); set({ projects: [data, ...get().projects] }); return data },
  setCurrentProject: (p) => set({ currentProject: p }),
}))
```

- [ ] **Step 2: 创建登录页面**

```tsx
// frontend/src/pages/Login.tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

export default function Login() {
  const [isLogin, setIsLogin] = useState(true)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const { login, register, isLoading, error } = useAuthStore()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try { isLogin ? await login(username, password) : await register(username, password); navigate('/') } catch {}
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow">
        <div><h1 className="text-3xl font-bold text-center">Novelist</h1><p className="text-center text-gray-500 mt-2">AI小说创作平台</p></div>
        <div className="flex space-x-4 justify-center">
          <button className={`px-4 py-2 ${isLogin ? 'border-b-2 border-blue-500 text-blue-600' : 'text-gray-500'}`} onClick={() => setIsLogin(true)}>登录</button>
          <button className={`px-4 py-2 ${!isLogin ? 'border-b-2 border-blue-500 text-blue-600' : 'text-gray-500'}`} onClick={() => setIsLogin(false)}>注册</button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && <div className="bg-red-50 text-red-500 p-3 rounded">{error}</div>}
          <div><label className="block text-sm font-medium text-gray-700">用户名</label><input type="text" value={username} onChange={(e) => setUsername(e.target.value)} className="mt-1 block w-full px-3 py-2 border rounded-md" required minLength={3} /></div>
          <div><label className="block text-sm font-medium text-gray-700">密码</label><input type="password" value={password} onChange={(e) => setPassword(e.target.value)} className="mt-1 block w-full px-3 py-2 border rounded-md" required minLength={6} /></div>
          <button type="submit" disabled={isLoading} className="w-full py-2 px-4 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50">{isLoading ? '处理中...' : isLogin ? '登录' : '注册'}</button>
        </form>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: 创建仪表板和项目卡片**

```tsx
// frontend/src/components/ProjectCard.tsx
import { useNavigate } from 'react-router-dom'
export default function ProjectCard({ project }: { project: any }) {
  const navigate = useNavigate()
  return (
    <div className="bg-white p-6 rounded-lg shadow hover:shadow-md cursor-pointer" onClick={() => navigate(`/projects/${project.id}`)}>
      <h3 className="text-xl font-semibold mb-2">{project.title}</h3>
      {project.genre && <span className="inline-block bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded mb-2">{project.genre}</span>}
      <p className="text-gray-600 text-sm line-clamp-3">{project.description}</p>
    </div>
  )
}
```

```tsx
// frontend/src/pages/Dashboard.tsx
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useProjectStore } from '../stores/projectStore'
import { useAuthStore } from '../stores/authStore'
import ProjectCard from '../components/ProjectCard'

export default function Dashboard() {
  const { projects, fetchProjects, createProject, isLoading } = useProjectStore()
  const { user, logout } = useAuthStore()
  const navigate = useNavigate()
  const [showCreate, setShowCreate] = useState(false)
  const [newTitle, setNewTitle] = useState('')

  useEffect(() => { fetchProjects() }, [])

  const handleCreate = async () => {
    if (!newTitle.trim()) return
    const project = await createProject({ title: newTitle })
    setShowCreate(false); setNewTitle(''); navigate(`/projects/${project.id}/creator`)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow"><div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
        <h1 className="text-2xl font-bold">Novelist</h1>
        <div className="flex items-center space-x-4"><span className="text-gray-600">{user?.username}</span><button onClick={logout} className="text-gray-500">退出</button></div>
      </div></header>
      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="flex justify-between items-center mb-8"><h2 className="text-xl font-semibold">我的项目</h2><button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-blue-600 text-white rounded-md">新建项目</button></div>
        {showCreate && <div className="bg-white p-6 rounded-lg shadow mb-8"><input value={newTitle} onChange={(e) => setNewTitle(e.target.value)} placeholder="项目标题" className="w-full px-3 py-2 border rounded-md mb-4" /><button onClick={handleCreate} className="px-4 py-2 bg-blue-600 text-white rounded-md">创建并开始构思</button></div>}
        {isLoading ? <div className="text-center py-12">加载中...</div> : <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">{projects.map((p) => <ProjectCard key={p.id} project={p} />)}</div>}
      </main>
    </div>
  )
}
```

- [ ] **Step 4: 更新App路由**

```tsx
// frontend/src/App.tsx
import { Routes, Route, Navigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuthStore } from './stores/authStore'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuthStore()
  return token ? <>{children}</> : <Navigate to="/login" />
}

export default function App() {
  const { checkAuth } = useAuthStore()
  useEffect(() => { checkAuth() }, [])
  return (
    <div className="min-h-screen bg-gray-50">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<PrivateRoute><Dashboard /></PrivateRoute>} />
      </Routes>
    </div>
  )
}
```

- [ ] **Step 5: 提交**

```bash
cd /home/jj/projects/novelist
git add frontend/
git commit -m "feat: implement auth and project management frontend"
```

---

### Task 14: 构思对话页面

**文件：**
- Create: `frontend/src/stores/agentStore.ts`
- Create: `frontend/src/pages/Creator.tsx`

- [ ] **Step 1: 创建Agent状态管理**

```typescript
// frontend/src/stores/agentStore.ts
import { create } from 'zustand'
import api from '../api/client'

interface Message { role: 'user' | 'agent'; content: string; agent?: string }
interface AgentState {
  messages: Message[]; isStreaming: boolean; streamContent: string
  sendMessage: (projectId: string, content: string) => Promise<void>
  clearMessages: () => void
}

export const useAgentStore = create<AgentState>((set, get) => ({
  messages: [], isStreaming: false, streamContent: '',
  sendMessage: async (projectId, content) => {
    const userMessage: Message = { role: 'user', content }
    const allMessages = [...get().messages, userMessage]
    set({ messages: allMessages, isStreaming: true, streamContent: '' })

    try {
      const { data } = await api.post('/creator/chat', {
        project_id: projectId,
        messages: allMessages.map(m => ({ role: m.role === 'agent' ? 'assistant' : 'user', content: m.content })),
      })
      const agentMessage: Message = { role: 'agent', content: data.content, agent: 'creator' }
      set({ messages: [...allMessages, agentMessage], isStreaming: false })
    } catch { set({ isStreaming: false }) }
  },
  clearMessages: () => set({ messages: [] }),
}))
```

- [ ] **Step 2: 创建构思对话页面**

```tsx
// frontend/src/pages/Creator.tsx
import { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAgentStore } from '../stores/agentStore'
import { useProjectStore } from '../stores/projectStore'

export default function Creator() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const { messages, isStreaming, sendMessage, clearMessages } = useAgentStore()
  const { currentProject, fetchProjects } = useProjectStore()
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => { if (projectId) fetchProjects(); return () => clearMessages() }, [projectId])
  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [messages])

  const handleSend = async () => {
    if (!input.trim() || !projectId) return
    const msg = input; setInput(''); await sendMessage(projectId, msg)
  }

  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <header className="bg-white shadow px-4 py-3"><div className="max-w-4xl mx-auto flex justify-between items-center">
        <h1 className="text-xl font-semibold">构思对话 - {currentProject?.title || '新项目'}</h1>
        <button onClick={() => navigate('/')} className="text-gray-500">返回</button>
      </div></header>
      <div className="flex-1 max-w-4xl mx-auto w-full p-4 flex flex-col">
        <div className="flex-1 overflow-y-auto mb-4 space-y-4">
          {messages.length === 0 && <div className="text-center text-gray-500 py-12"><p className="text-lg mb-2">你好！请告诉我你想写什么类型的小说？</p><p>有什么初步的想法或灵感吗？</p></div>}
          {messages.map((msg, i) => (
            <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
              <div className={`max-w-[80%] p-4 rounded-lg ${msg.role === 'user' ? 'bg-blue-600 text-white' : 'bg-white shadow'}`}>
                <div className="whitespace-pre-wrap">{msg.content}</div>
              </div>
            </div>
          ))}
          {isStreaming && <div className="flex justify-start"><div className="max-w-[80%] p-4 rounded-lg bg-white shadow"><div className="whitespace-pre-wrap">思考中...</div></div></div>}
          <div ref={messagesEndRef} />
        </div>
        <div className="bg-white rounded-lg shadow p-4">
          <div className="flex space-x-4">
            <textarea value={input} onChange={(e) => setInput(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() } }} placeholder="输入你的想法..." className="flex-1 px-3 py-2 border rounded-md resize-none" rows={2} />
            <button onClick={handleSend} disabled={isStreaming || !input.trim()} className="px-6 py-2 bg-blue-600 text-white rounded-md disabled:opacity-50 self-end">发送</button>
          </div>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: 更新路由**

```tsx
// frontend/src/App.tsx - 添加路由
<Route path="/projects/:projectId/creator" element={<PrivateRoute><Creator /></PrivateRoute>} />
```

- [ ] **Step 4: 提交**

```bash
cd /home/jj/projects/novelist
git add frontend/
git commit -m "feat: implement creator conversation page"
```

---

### Task 15: WebSocket与讨论面板

**文件：**
- Create: `backend/internal/api/websocket.go`
- Create: `frontend/src/hooks/useWebSocket.ts`
- Create: `frontend/src/stores/discussionStore.ts`
- Create: `frontend/src/components/DiscussionPanel.tsx`

- [ ] **Step 1: 创建WebSocket处理器**

```go
// backend/internal/api/websocket.go
package api

import (
    "log"
    "net/http"
    "sync"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/google/uuid"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type WSManager struct {
    clients map[uuid.UUID]map[*websocket.Conn]bool
    mu      sync.RWMutex
}

var WS = &WSManager{clients: make(map[uuid.UUID]map[*websocket.Conn]bool)}

func (m *WSManager) Add(userID uuid.UUID, conn *websocket.Conn) {
    m.mu.Lock(); defer m.mu.Unlock()
    if m.clients[userID] == nil { m.clients[userID] = make(map[*websocket.Conn]bool) }
    m.clients[userID][conn] = true
}

func (m *WSManager) Remove(userID uuid.UUID, conn *websocket.Conn) {
    m.mu.Lock(); defer m.mu.Unlock()
    if m.clients[userID] != nil { delete(m.clients[userID], conn) }
}

func (m *WSManager) Send(userID uuid.UUID, msg interface{}) {
    m.mu.RLock(); defer m.mu.RUnlock()
    for conn := range m.clients[userID] {
        if err := conn.WriteJSON(msg); err != nil { log.Printf("WS send error: %v", err) }
    }
}

func HandleWebSocket(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists { c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"}); return }

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil { return }
    defer conn.Close()

    uid := userID.(uuid.UUID)
    WS.Add(uid, conn)
    defer WS.Remove(uid, conn)

    for { _, _, err := conn.ReadMessage(); if err != nil { break } }
}
```

- [ ] **Step 2: 创建WebSocket Hook**

```typescript
// frontend/src/hooks/useWebSocket.ts
import { useEffect, useRef, useCallback } from 'react'
import { useAuthStore } from '../stores/authStore'

export function useWebSocket(onMessage?: (data: any) => void) {
  const ws = useRef<WebSocket | null>(null)
  const { token } = useAuthStore()

  useEffect(() => {
    if (!token) return
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws.current = new WebSocket(`${protocol}//${window.location.host}/ws?token=${token}`)
    ws.current.onmessage = (e) => { try { onMessage?.(JSON.parse(e.data)) } catch {} }
    return () => { ws.current?.close() }
  }, [token, onMessage])

  const send = useCallback((data: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) ws.current.send(JSON.stringify(data))
  }, [])

  return { send }
}
```

- [ ] **Step 3: 创建讨论状态管理和面板**

```typescript
// frontend/src/stores/discussionStore.ts
import { create } from 'zustand'
import api from '../api/client'

interface Suggestion { type: string; location: string; problem: string; suggestion: string; priority: number }
interface DiscussionResult {
  editor_suggestions: Suggestion[]; reader_feedback: string; critic_analysis: string; aggregated: Suggestion[]
}
interface DiscussionState {
  result: DiscussionResult | null; isDiscussing: boolean
  startDiscussion: (chapterId: string) => Promise<void>
  clearResult: () => void
}

export const useDiscussionStore = create<DiscussionState>((set) => ({
  result: null, isDiscussing: false,
  startDiscussion: async (chapterId) => {
    set({ isDiscussing: true, result: null })
    try {
      const { data } = await api.post(`/chapters/${chapterId}/discuss`)
      set({ result: data, isDiscussing: false })
    } catch { set({ isDiscussing: false }) }
  },
  clearResult: () => set({ result: null }),
}))
```

```tsx
// frontend/src/components/DiscussionPanel.tsx
import { useDiscussionStore } from '../stores/discussionStore'

export default function DiscussionPanel({ chapterId, onClose }: { chapterId: string; onClose: () => void }) {
  const { result, isDiscussing, startDiscussion } = useDiscussionStore()

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[80vh] overflow-y-auto p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">审稿讨论</h2>
          <button onClick={onClose} className="text-gray-500">关闭</button>
        </div>

        {!result && !isDiscussing && (
          <button onClick={() => startDiscussion(chapterId)} className="px-4 py-2 bg-blue-600 text-white rounded-md">开始审稿</button>
        )}

        {isDiscussing && <div className="text-center py-8">正在审稿中，请稍候...</div>}

        {result && (
          <div className="space-y-6">
            <div className="bg-gray-50 p-4 rounded-lg">
              <h3 className="font-semibold mb-2">编辑建议</h3>
              {result.aggregated?.map((s, i) => (
                <div key={i} className="mb-2 p-2 bg-white rounded">
                  <span className="text-xs px-2 py-1 bg-blue-100 text-blue-800 rounded">{s.type}</span>
                  <span className="text-xs px-2 py-1 ml-2 bg-gray-100 rounded">优先级 {s.priority}</span>
                  <p className="mt-1 text-sm">{s.problem}</p>
                  <p className="text-sm text-gray-600">建议：{s.suggestion}</p>
                </div>
              ))}
            </div>
            <div className="bg-gray-50 p-4 rounded-lg"><h3 className="font-semibold mb-2">读者反馈</h3><p className="text-sm whitespace-pre-wrap">{result.reader_feedback}</p></div>
            <div className="bg-gray-50 p-4 rounded-lg"><h3 className="font-semibold mb-2">评论家分析</h3><p className="text-sm whitespace-pre-wrap">{result.critic_analysis}</p></div>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: 更新路由注册WebSocket**

```go
// backend/internal/api/router.go - 在protected组内添加
protected.GET("/ws", HandleWebSocket)
```

- [ ] **Step 5: 提交**

```bash
cd /home/jj/projects/novelist
git add backend/ frontend/
git commit -m "feat: implement WebSocket and discussion panel"
```

---

## 验证计划

1. **后端API**：启动Go服务器，验证REST端点
2. **认证流程**：注册 → 登录 → JWT访问
3. **前端**：启动开发服务器，验证页面渲染
4. **构思对话**：与Agent多轮对话，验证大纲生成
5. **讨论工作流**：开始审稿 → 验证3个Agent响应 → 验证建议汇总
6. **记忆系统**：验证上下文组装包含相关设定
7. **WebSocket**：验证实时连接正常

---

## 后续任务（可选）

- **章节编辑器**：TipTap集成、AI写作面板
- **设置页面**：模型配置、写作偏好
- **人物/世界观管理页面**：卡片式布局、AI辅助生成
- **导出功能**：导出为TXT/EPUB格式
