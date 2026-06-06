package memory

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

// Querier abstracts database queries so Memory can be tested without SQLite/CGO.
type Querier interface {
	GetProject(id uuid.UUID) (*model.Project, error)
	GetWorldSettings(projectID uuid.UUID) ([]model.WorldSetting, error)
	GetCharacters(projectID uuid.UUID) ([]model.Character, error)
	GetOutlines(projectID uuid.UUID) ([]model.Outline, error)
	GetChaptersBefore(projectID uuid.UUID, chapterNum int, limit int) ([]model.Chapter, error)
	GetChaptersWithEmbedding(projectID uuid.UUID, limit int) ([]model.Character, []model.WorldSetting, []model.Outline, []model.Chapter, error)
}

// StoreQuerier wraps the real GORM store for production use.
type StoreQuerier struct{}

func (s *StoreQuerier) GetProject(id uuid.UUID) (*model.Project, error) {
	var p model.Project
	if err := store.GetDB().Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *StoreQuerier) GetWorldSettings(projectID uuid.UUID) ([]model.WorldSetting, error) {
	var settings []model.WorldSetting
	store.GetDB().Where("project_id = ?", projectID).Find(&settings)
	return settings, nil
}

func (s *StoreQuerier) GetCharacters(projectID uuid.UUID) ([]model.Character, error) {
	var characters []model.Character
	store.GetDB().Where("project_id = ?", projectID).Find(&characters)
	return characters, nil
}

func (s *StoreQuerier) GetOutlines(projectID uuid.UUID) ([]model.Outline, error) {
	var outlines []model.Outline
	store.GetDB().Where("project_id = ?", projectID).Order("act, chapter_num").Find(&outlines)
	return outlines, nil
}

func (s *StoreQuerier) GetChaptersBefore(projectID uuid.UUID, chapterNum int, limit int) ([]model.Chapter, error) {
	var chapters []model.Chapter
	store.GetDB().Where("project_id = ? AND chapter_num < ?", projectID, chapterNum).
		Order("chapter_num DESC").Limit(limit).Find(&chapters)
	return chapters, nil
}

func (s *StoreQuerier) GetChaptersWithEmbedding(projectID uuid.UUID, limit int) ([]model.Character, []model.WorldSetting, []model.Outline, []model.Chapter, error) {
	embeddingStr := store.FormatVector(nil) // only used when embedding is non-nil
	_ = embeddingStr
	// This method is only called from SemanticSearch which needs pgvector.
	// In production, this would use vector distance ordering.
	// For non-embedding path, this is never called.
	return nil, nil, nil, nil, nil
}

type Memory struct {
	ProjectID uuid.UUID
	q         Querier
}

func NewMemory(projectID uuid.UUID, q ...Querier) *Memory {
	var querier Querier = &StoreQuerier{}
	if len(q) > 0 && q[0] != nil {
		querier = q[0]
	}
	return &Memory{ProjectID: projectID, q: querier}
}

// LoadLongTermMemory loads project info, world settings, outlines, and characters
func (m *Memory) LoadLongTermMemory(ctx context.Context) (string, error) {
	project, err := m.q.GetProject(m.ProjectID)
	if err != nil {
		return "", fmt.Errorf("project not found: %w", err)
	}

	settings, _ := m.q.GetWorldSettings(m.ProjectID)
	outlines, _ := m.q.GetOutlines(m.ProjectID)
	characters, _ := m.q.GetCharacters(m.ProjectID)

	result := fmt.Sprintf("## 项目信息\n标题: %s\n类型: %s\n风格: %s\n\n", project.Title, project.Genre, project.StyleGuide)

	result += "## 世界观设定\n"
	for _, s := range settings {
		result += fmt.Sprintf("- [%s] %s\n", s.Category, s.Content)
	}

	result += "\n## 人物档案\n"
	for _, c := range characters {
		result += fmt.Sprintf("- %s（%s）: %s, %s\n", c.Name, c.Role, c.Personality, c.Background)
	}

	result += "\n## 故事大纲\n"
	for _, o := range outlines {
		result += fmt.Sprintf("- 第%d章: %s\n", o.ChapterNum, o.Summary)
	}

	return result, nil
}

// LoadShortTermMemory loads recent chapters
func (m *Memory) LoadShortTermMemory(ctx context.Context, currentChapterNum int) (string, error) {
	chapters, _ := m.q.GetChaptersBefore(m.ProjectID, currentChapterNum, 5)

	result := "## 近期章节\n"
	for i := len(chapters) - 1; i >= 0; i-- {
		c := chapters[i]
		result += fmt.Sprintf("\n### 第%d章 %s\n%s\n", c.ChapterNum, c.Title, c.Content)
	}
	return result, nil
}

// SemanticSearch searches for related characters, settings, outlines, and chapters using pgvector
func (m *Memory) SemanticSearch(ctx context.Context, queryEmbedding []float32, limit int) (string, error) {
	if len(queryEmbedding) == 0 {
		return "", nil
	}

	embeddingStr := store.FormatVector(queryEmbedding)

	var characters []model.Character
	store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
		Order(fmt.Sprintf("embedding <-> '%s'", embeddingStr)).Limit(limit).Find(&characters)

	var settings []model.WorldSetting
	store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
		Order(fmt.Sprintf("embedding <-> '%s'", embeddingStr)).Limit(limit).Find(&settings)

	var outlines []model.Outline
	store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
		Order(fmt.Sprintf("embedding <-> '%s'", embeddingStr)).Limit(limit).Find(&outlines)

	var chapters []model.Chapter
	store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
		Order(fmt.Sprintf("embedding <-> '%s'", embeddingStr)).Limit(limit).Find(&chapters)

	if len(characters) == 0 && len(settings) == 0 && len(outlines) == 0 && len(chapters) == 0 {
		return "", nil
	}

	result := "## 语义相关记忆\n"
	for _, c := range characters {
		result += fmt.Sprintf("- 人物 %s（%s）: %s\n", c.Name, c.Role, c.Personality)
	}
	for _, s := range settings {
		result += fmt.Sprintf("- 设定 [%s] %s\n", s.Category, s.Content)
	}
	for _, o := range outlines {
		result += fmt.Sprintf("- 大纲 第%d章: %s\n", o.ChapterNum, o.Summary)
	}
	for _, ch := range chapters {
		content := ch.Content
		if len([]rune(content)) > 200 {
			content = string([]rune(content)[:200]) + "..."
		}
		result += fmt.Sprintf("- 章节 第%d章 %s: %s\n", ch.ChapterNum, ch.Title, content)
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

		// Semantic search: degrade gracefully if embedding unavailable
		if ai.EmbeddingMgr != nil {
			embedding, err := ai.EmbeddingMgr.GenerateEmbedding(ctx, workingMemory)
			if err == nil {
				semantic, err := m.SemanticSearch(ctx, embedding, 5)
				if err == nil && semantic != "" {
					fullContext += "\n" + semantic
				}
			}
		}
	}
	return fullContext, nil
}
