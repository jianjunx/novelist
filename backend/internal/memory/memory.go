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
	embeddingStr := formatVector(queryEmbedding)

	var characters []model.Character
	store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
		Order(fmt.Sprintf("embedding <-> '%s'", embeddingStr)).Limit(limit).Find(&characters)

	var settings []model.WorldSetting
	store.GetDB().Where("project_id = ? AND embedding IS NOT NULL", m.ProjectID).
		Order(fmt.Sprintf("embedding <-> '%s'", embeddingStr)).Limit(limit).Find(&settings)

	result := "## 语义相关设定\n"
	for _, c := range characters {
		result += fmt.Sprintf("- 人物 %s: %s\n", c.Name, c.Personality)
	}
	for _, s := range settings {
		result += fmt.Sprintf("- [%s] %s\n", s.Category, s.Content)
	}
	return result, nil
}

// formatVector formats a []float32 as a PostgreSQL vector literal string.
func formatVector(v []float32) string {
	s := "["
	for i, f := range v {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%f", f)
	}
	s += "]"
	return s
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
