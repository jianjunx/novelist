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
