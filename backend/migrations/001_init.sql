CREATE EXTENSION IF NOT EXISTS "vector";

-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 小说项目
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 故事大纲
CREATE TABLE outlines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 用户设置
CREATE TABLE settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
