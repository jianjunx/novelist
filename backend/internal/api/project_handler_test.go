package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

func setupProjectRouter() *gin.Engine {
	r := setupGin()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Set("username", "testuser")
		c.Next()
	})
	r.GET("/projects", GetProjects)
	r.POST("/projects", CreateProject)
	r.GET("/projects/:id", GetProject)
	r.GET("/projects/:id/overview", GetProjectOverview)
	r.PUT("/projects/:id", UpdateProject)
	r.DELETE("/projects/:id", DeleteProject)
	return r
}

var testUserID = uuid.New()

func createTestProject(t *testing.T, title string) model.Project {
	t.Helper()
	p := model.Project{
		ID:      uuid.New(),
		ShortID: model.GenerateShortID(),
		UserID:  testUserID,
		Title:   title,
		Genre:   "玄幻",
	}
	if err := store.GetDB().Create(&p).Error; err != nil {
		t.Fatalf("createTestProject: %v", err)
	}
	return p
}

func TestCreateProject_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	body := `{"title":"我的小说","genre":"玄幻","description":"一个奇幻故事"}`
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("CreateProject() status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp model.Project
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Title != "我的小说" {
		t.Errorf("CreateProject() title = %q, want %q", resp.Title, "我的小说")
	}
	if resp.ShortID == "" {
		t.Error("CreateProject() short_id should not be empty")
	}
}

func TestCreateProject_MissingTitle(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	body := `{"genre":"玄幻"}`
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateProject() no title status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetProjects_Empty(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetProjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []ProjectWithStatus
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Errorf("GetProjects() returned %d projects, want 0", len(resp))
	}
}

func TestGetProjects_WithData(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	createTestProject(t, "项目一")
	createTestProject(t, "项目二")

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetProjects() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []ProjectWithStatus
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("GetProjects() returned %d projects, want 2", len(resp))
	}
}

func TestGetProject_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	p := createTestProject(t, "测试项目")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetProject() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestGetProject_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetProject() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetProject_ByUUID(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	p := createTestProject(t, "UUID项目")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s", p.ID.String()), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetProject() by UUID status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUpdateProject_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	p := createTestProject(t, "原标题")

	body := `{"title":"新标题","genre":"科幻","description":"新描述","style_guide":"新风格"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/projects/%s", p.ShortID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateProject() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	body := `{"title":"新标题"}`
	req := httptest.NewRequest(http.MethodPut, "/projects/nonexistent", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateProject() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDeleteProject_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	p := createTestProject(t, "待删除")

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/projects/%s", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DeleteProject() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Verify deleted
	var count int64
	store.GetDB().Model(&model.Project{}).Where("id = ?", p.ID).Count(&count)
	if count != 0 {
		t.Error("DeleteProject() project still exists in DB")
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	req := httptest.NewRequest(http.MethodDelete, "/projects/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("DeleteProject() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestProject_Authorization(t *testing.T) {
	setupTestDBWithUUID(t)

	// Create project as another user
	otherUserID := uuid.New()
	p := model.Project{
		ID:      uuid.New(),
		ShortID: model.GenerateShortID(),
		UserID:  otherUserID,
		Title:   "别人的项目",
	}
	store.GetDB().Create(&p)

	r := setupProjectRouter()

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetProject() other user's project status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetProjectOverview_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupProjectRouter()

	p := createTestProject(t, "仪表盘项目")

	outlineID := uuid.New()
	volumeID := uuid.New()
	store.GetDB().Create(&model.Volume{
		ID: volumeID, ProjectID: p.ID, VolumeNum: 1, Title: "第一篇",
	})
	store.GetDB().Create(&model.Outline{
		ID: outlineID, ProjectID: p.ID, VolumeID: &volumeID,
		Act: 1, ChapterNum: 1, Summary: "开篇",
	})
	chapterID := uuid.New()
	store.GetDB().Create(&model.Chapter{
		ID: chapterID, ProjectID: p.ID, OutlineID: &outlineID,
		ChapterNum: 1, Title: "第1章", Content: "正文", WordCount: 2,
	})
	store.GetDB().Create(&model.Character{
		ID: uuid.New(), ProjectID: p.ID, Name: "主角", Role: "主角",
	})
	store.GetDB().Create(&model.WorldSetting{
		ID: uuid.New(), ProjectID: p.ID, Category: "地点", Content: "青云宗",
	})

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/overview", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetProjectOverview() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp ProjectOverview
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("GetProjectOverview() unmarshal failed: %v", err)
	}
	if resp.Stats.ChapterCount != 1 {
		t.Errorf("GetProjectOverview() chapter_count = %d, want 1", resp.Stats.ChapterCount)
	}
	if resp.Stats.WrittenCount != 1 {
		t.Errorf("GetProjectOverview() written_count = %d, want 1", resp.Stats.WrittenCount)
	}
	if resp.Stats.TotalWords != 2 {
		t.Errorf("GetProjectOverview() total_words = %d, want 2", resp.Stats.TotalWords)
	}
	if len(resp.Outlines) != 1 || resp.Outlines[0].ChapterID == nil || *resp.Outlines[0].ChapterID != chapterID {
		t.Error("GetProjectOverview() outline chapter_id mapping incorrect")
	}
	if len(resp.Characters) != 1 {
		t.Errorf("GetProjectOverview() characters = %d, want 1", len(resp.Characters))
	}
}
