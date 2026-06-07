package memory

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/model"
)

func TestLoadLongTermMemory_Success(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "仙途",
		Genre:      "玄幻",
		StyleGuide: "古风",
	}
	q.addProject(project)

	q.addWorldSetting(model.WorldSetting{
		ID:        uuid.New(),
		ProjectID: project.ID,
		Category:  "地理",
		Content:   "大陆分为东西南北中五域",
	})

	q.addCharacter(model.Character{
		ID:          uuid.New(),
		ProjectID:   project.ID,
		Name:        "李逍遥",
		Role:        "主角",
		Personality: "机智勇敢",
		Background:  "出身平凡",
	})

	q.addOutline(model.Outline{
		ID:         uuid.New(),
		ProjectID:  project.ID,
		Act:        1,
		ChapterNum: 1,
		Summary:    "少年踏上修仙之路",
	})

	ai.EmbeddingMgr = nil
	mem := NewMemory(project.ID, q)
	result, err := mem.LoadLongTermMemory(context.Background())
	if err != nil {
		t.Fatalf("LoadLongTermMemory() error: %v", err)
	}

	if !strings.Contains(result, "仙途") {
		t.Error("LoadLongTermMemory() missing project title")
	}
	if !strings.Contains(result, "玄幻") {
		t.Error("LoadLongTermMemory() missing genre")
	}
	if !strings.Contains(result, "古风") {
		t.Error("LoadLongTermMemory() missing style guide")
	}
	if !strings.Contains(result, "地理") {
		t.Error("LoadLongTermMemory() missing world setting category")
	}
	if !strings.Contains(result, "大陆分为东西南北中五域") {
		t.Error("LoadLongTermMemory() missing world setting content")
	}
	if !strings.Contains(result, "李逍遥") {
		t.Error("LoadLongTermMemory() missing character name")
	}
	if !strings.Contains(result, "主角") {
		t.Error("LoadLongTermMemory() missing character role")
	}
	if !strings.Contains(result, "少年踏上修仙之路") {
		t.Error("LoadLongTermMemory() missing outline summary")
	}
}

func TestLoadLongTermMemory_ProjectNotFound(t *testing.T) {
	q := newMockQuerier()
	mem := NewMemory(uuid.New(), q)
	_, err := mem.LoadLongTermMemory(context.Background())
	if err == nil {
		t.Error("LoadLongTermMemory() should error for nonexistent project")
	}
}

func TestLoadLongTermMemory_EmptyProject(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "空项目",
		Genre:      "",
		StyleGuide: "",
	}
	q.addProject(project)

	mem := NewMemory(project.ID, q)
	result, err := mem.LoadLongTermMemory(context.Background())
	if err != nil {
		t.Fatalf("LoadLongTermMemory() error: %v", err)
	}

	if !strings.Contains(result, "空项目") {
		t.Error("LoadLongTermMemory() should still contain project title")
	}
	if !strings.Contains(result, "世界观设定") {
		t.Error("LoadLongTermMemory() should have world settings section header")
	}
	if !strings.Contains(result, "人物档案") {
		t.Error("LoadLongTermMemory() should have characters section header")
	}
	if !strings.Contains(result, "故事大纲") {
		t.Error("LoadLongTermMemory() should have outlines section header")
	}
}

func TestLoadShortTermMemory_Success(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "测试小说",
	}
	q.addProject(project)

	for i := 1; i <= 6; i++ {
		q.addChapter(model.Chapter{
			ID:         uuid.New(),
			ProjectID:  project.ID,
			ChapterNum: i,
			Title:      "第" + string(rune('0'+i)) + "章",
			Content:    "这是第" + string(rune('0'+i)) + "章的内容",
		})
	}

	ai.EmbeddingMgr = nil
	mem := NewMemory(project.ID, q)
	result, err := mem.LoadShortTermMemory(context.Background(), 6)
	if err != nil {
		t.Fatalf("LoadShortTermMemory() error: %v", err)
	}

	if !strings.Contains(result, "本篇前文") {
		t.Error("LoadShortTermMemory() missing header")
	}
	if !strings.Contains(result, "第1章") {
		t.Error("LoadShortTermMemory() missing chapter 1")
	}
	if !strings.Contains(result, "第5章") {
		t.Error("LoadShortTermMemory() missing chapter 5")
	}
	if strings.Contains(result, "第6章") {
		t.Error("LoadShortTermMemory() should not include current chapter")
	}
}

func TestLoadShortTermMemory_NoChapters(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "空项目",
	}
	q.addProject(project)

	mem := NewMemory(project.ID, q)
	result, err := mem.LoadShortTermMemory(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadShortTermMemory() error: %v", err)
	}

	if result != "" {
		t.Error("LoadShortTermMemory() should return empty when no chapters")
	}
}

func TestAssembleContext_WithoutEmbedding(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "测试",
		Genre:      "玄幻",
		StyleGuide: "现代",
	}
	q.addProject(project)

	q.addCharacter(model.Character{
		ID:        uuid.New(),
		ProjectID: project.ID,
		Name:      "主角",
		Role:      "主角",
	})

	q.addChapter(model.Chapter{
		ID:         uuid.New(),
		ProjectID:  project.ID,
		ChapterNum: 1,
		Title:      "开篇",
		Content:    "故事开始了",
	})

	ai.EmbeddingMgr = nil
	mem := NewMemory(project.ID, q)
	result, err := mem.AssembleContext(context.Background(), 2, "当前写作任务：第二章")
	if err != nil {
		t.Fatalf("AssembleContext() error: %v", err)
	}

	if !strings.Contains(result, "测试") {
		t.Error("AssembleContext() missing project info")
	}
	if !strings.Contains(result, "主角") {
		t.Error("AssembleContext() missing character info")
	}
	if !strings.Contains(result, "本篇前文") {
		t.Error("AssembleContext() missing short-term memory")
	}
	if !strings.Contains(result, "当前写作任务") {
		t.Error("AssembleContext() missing working memory")
	}
	if !strings.Contains(result, "当前任务上下文") {
		t.Error("AssembleContext() missing working memory section header")
	}
}

func TestAssembleContext_EmptyWorkingMemory(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "测试",
	}
	q.addProject(project)

	ai.EmbeddingMgr = nil
	mem := NewMemory(project.ID, q)
	result, err := mem.AssembleContext(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("AssembleContext() error: %v", err)
	}

	if strings.Contains(result, "当前任务上下文") {
		t.Error("AssembleContext() should not have working memory section when empty")
	}
}

func TestSemanticSearch_EmptyEmbedding(t *testing.T) {
	q := newMockQuerier()

	project := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      "测试",
	}
	q.addProject(project)

	mem := NewMemory(project.ID, q)

	result, err := mem.SemanticSearch(context.Background(), nil, 5)
	if err != nil {
		t.Fatalf("SemanticSearch() error: %v", err)
	}
	if result != "" {
		t.Errorf("SemanticSearch() with nil embedding should return empty, got %q", result)
	}

	result, err = mem.SemanticSearch(context.Background(), []float32{}, 5)
	if err != nil {
		t.Fatalf("SemanticSearch() error: %v", err)
	}
	if result != "" {
		t.Errorf("SemanticSearch() with empty embedding should return empty, got %q", result)
	}
}

// SemanticSearch with non-empty embeddings requires pgvector (PostgreSQL extension).
// These tests are skipped in unit tests without a real database.
func TestSemanticSearch_RequiresDatabase(t *testing.T) {
	t.Skip("SemanticSearch with non-empty embedding requires pgvector; tested in integration tests")
}
