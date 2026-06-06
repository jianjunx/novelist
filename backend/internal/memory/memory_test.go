package memory

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// Check if CGO is available (needed for go-sqlite3)
	if err := exec.Command("gcc", "--version").Run(); err != nil {
		// No gcc available, skip all tests in this package
		// These tests require SQLite which needs CGO
		return
	}
	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	_ = db.Callback().Create().Before("gorm:create").Register("assign_uuid", func(tx *gorm.DB) {
		if tx.Statement.Schema == nil {
			return
		}
		pkField := tx.Statement.Schema.PrioritizedPrimaryField
		if pkField == nil {
			return
		}
		field := tx.Statement.Schema.LookUpField(pkField.Name)
		if field == nil {
			return
		}
		val, _ := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue)
		if id, ok := val.(uuid.UUID); ok && id == uuid.Nil {
			field.Set(tx.Statement.Context, tx.Statement.ReflectValue, uuid.New())
		}
	})

	if err := db.AutoMigrate(
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
		t.Fatalf("failed to migrate: %v", err)
	}
	store.DB = db
}

func createTestProject(t *testing.T, title, genre, style string) model.Project {
	t.Helper()
	p := model.Project{
		ID:         uuid.New(),
		ShortID:    model.GenerateShortID(),
		UserID:     uuid.New(),
		Title:      title,
		Genre:      genre,
		StyleGuide: style,
	}
	store.GetDB().Create(&p)
	return p
}

func TestLoadLongTermMemory_Success(t *testing.T) {
	setupTestDB(t)
	ai.EmbeddingMgr = nil

	project := createTestProject(t, "仙途", "玄幻", "古风")

	store.GetDB().Create(&model.WorldSetting{
		ID:        uuid.New(),
		ProjectID: project.ID,
		Category:  "地理",
		Content:   "大陆分为东西南北中五域",
	})

	store.GetDB().Create(&model.Character{
		ID:          uuid.New(),
		ProjectID:   project.ID,
		Name:        "李逍遥",
		Role:        "主角",
		Personality: "机智勇敢",
		Background:  "出身平凡",
	})

	store.GetDB().Create(&model.Outline{
		ID:         uuid.New(),
		ProjectID:  project.ID,
		Act:        1,
		ChapterNum: 1,
		Summary:    "少年踏上修仙之路",
	})

	mem := NewMemory(project.ID)
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
	setupTestDB(t)

	mem := NewMemory(uuid.New())
	_, err := mem.LoadLongTermMemory(context.Background())
	if err == nil {
		t.Error("LoadLongTermMemory() should error for nonexistent project")
	}
}

func TestLoadLongTermMemory_EmptyProject(t *testing.T) {
	setupTestDB(t)

	project := createTestProject(t, "空项目", "", "")
	mem := NewMemory(project.ID)
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
	setupTestDB(t)
	ai.EmbeddingMgr = nil

	project := createTestProject(t, "测试小说", "", "")

	for i := 1; i <= 6; i++ {
		store.GetDB().Create(&model.Chapter{
			ID:         uuid.New(),
			ProjectID:  project.ID,
			ChapterNum: i,
			Title:      "第" + string(rune('0'+i)) + "章",
			Content:    "这是第" + string(rune('0'+i)) + "章的内容",
		})
	}

	mem := NewMemory(project.ID)
	result, err := mem.LoadShortTermMemory(context.Background(), 6)
	if err != nil {
		t.Fatalf("LoadShortTermMemory() error: %v", err)
	}

	if !strings.Contains(result, "近期章节") {
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
	setupTestDB(t)

	project := createTestProject(t, "空项目", "", "")
	mem := NewMemory(project.ID)
	result, err := mem.LoadShortTermMemory(context.Background(), 1)
	if err != nil {
		t.Fatalf("LoadShortTermMemory() error: %v", err)
	}

	if !strings.Contains(result, "近期章节") {
		t.Error("LoadShortTermMemory() should still have header")
	}
}

func TestAssembleContext_WithoutEmbedding(t *testing.T) {
	setupTestDB(t)
	ai.EmbeddingMgr = nil

	project := createTestProject(t, "测试", "玄幻", "现代")

	store.GetDB().Create(&model.Character{
		ID:        uuid.New(),
		ProjectID: project.ID,
		Name:      "主角",
		Role:      "主角",
	})

	store.GetDB().Create(&model.Chapter{
		ID:         uuid.New(),
		ProjectID:  project.ID,
		ChapterNum: 1,
		Title:      "开篇",
		Content:    "故事开始了",
	})

	mem := NewMemory(project.ID)
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
	if !strings.Contains(result, "近期章节") {
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
	setupTestDB(t)
	ai.EmbeddingMgr = nil

	project := createTestProject(t, "测试", "", "")
	mem := NewMemory(project.ID)
	result, err := mem.AssembleContext(context.Background(), 1, "")
	if err != nil {
		t.Fatalf("AssembleContext() error: %v", err)
	}

	if strings.Contains(result, "当前任务上下文") {
		t.Error("AssembleContext() should not have working memory section when empty")
	}
}

func TestSemanticSearch_EmptyEmbedding(t *testing.T) {
	setupTestDB(t)

	project := createTestProject(t, "测试", "", "")
	mem := NewMemory(project.ID)
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
