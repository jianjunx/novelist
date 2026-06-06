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

func setupChapterRouter() *gin.Engine {
	r := setupGin()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Set("username", "testuser")
		c.Next()
	})
	r.GET("/projects/:id/chapters", GetChapters)
	r.POST("/projects/:id/chapters", CreateChapter)
	r.GET("/chapters/:id", GetChapter)
	r.PUT("/chapters/:id", UpdateChapter)
	r.DELETE("/chapters/:id", DeleteChapter)
	return r
}

func createTestChapter(t *testing.T, projectID uuid.UUID, chapterNum int, title, content string) model.Chapter {
	t.Helper()
	ch := model.Chapter{
		ID:         uuid.New(),
		ProjectID:  projectID,
		ChapterNum: chapterNum,
		Title:      title,
		Content:    content,
		WordCount:  len([]rune(content)),
	}
	if err := store.GetDB().Create(&ch).Error; err != nil {
		t.Fatalf("createTestChapter: %v", err)
	}
	return ch
}

func TestGetChapters_Empty(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/chapters", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetChapters() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []ChapterWithOutline
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Errorf("GetChapters() returned %d chapters, want 0", len(resp))
	}
}

func TestGetChapters_WithData(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")

	createTestChapter(t, p.ID, 1, "第一章", "内容一")
	createTestChapter(t, p.ID, 2, "第二章", "内容二")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/chapters", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetChapters() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []ChapterWithOutline
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("GetChapters() returned %d chapters, want 2", len(resp))
	}
}

func TestGetChapters_CanGenerate(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")

	// First chapter without content can generate
	createTestChapter(t, p.ID, 1, "第一章", "")
	// Second chapter can't generate because first has no content
	createTestChapter(t, p.ID, 2, "第二章", "")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/chapters", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp []ChapterWithOutline
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp) != 2 {
		t.Fatalf("GetChapters() returned %d chapters, want 2", len(resp))
	}
	if !resp[0].CanGenerate {
		t.Error("first chapter with no content should be able to generate")
	}
	if resp[1].CanGenerate {
		t.Error("second chapter should not be able to generate when first has no content")
	}
}

func TestGetChapters_CanGenerate_SecondChapter(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")

	// First chapter with content
	createTestChapter(t, p.ID, 1, "第一章", "有内容")
	// Second chapter without content - should be able to generate
	createTestChapter(t, p.ID, 2, "第二章", "")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/chapters", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp []ChapterWithOutline
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp) != 2 {
		t.Fatalf("GetChapters() returned %d chapters, want 2", len(resp))
	}
	// First chapter has content, so can't generate
	if resp[0].CanGenerate {
		t.Error("first chapter with content should not be able to generate")
	}
	// Second chapter has no content and previous has content
	if !resp[1].CanGenerate {
		t.Error("second chapter should be able to generate when first has content")
	}
}

func TestGetChapters_ProjectNotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()

	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent/chapters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetChapters() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCreateChapter_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")

	body := `{"chapter_num":1,"title":"第一章","content":"这是内容"}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/projects/%s/chapters", p.ShortID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("CreateChapter() status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp model.Chapter
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Title != "第一章" {
		t.Errorf("CreateChapter() title = %q, want %q", resp.Title, "第一章")
	}
	if resp.WordCount != 4 {
		t.Errorf("CreateChapter() word_count = %d, want 4", resp.WordCount)
	}
}

func TestCreateChapter_MissingFields(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")

	body := `{"chapter_num":1}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/projects/%s/chapters", p.ShortID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateChapter() missing title status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCreateChapter_ProjectNotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()

	body := `{"chapter_num":1,"title":"第一章"}`
	req := httptest.NewRequest(http.MethodPost, "/projects/nonexistent/chapters", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("CreateChapter() project not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetChapter_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")
	ch := createTestChapter(t, p.ID, 1, "第一章", "内容")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/chapters/%s", ch.ID.String()), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetChapter() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestGetChapter_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/chapters/%s", uuid.New().String()), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetChapter() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdateChapter_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")
	ch := createTestChapter(t, p.ID, 1, "原标题", "原内容")

	body := `{"title":"新标题","content":"新内容在这里","status":"revised"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/chapters/%s", ch.ID.String()), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateChapter() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUpdateChapter_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()

	body := `{"title":"新标题"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/chapters/%s", uuid.New().String()), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateChapter() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestDeleteChapter_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()
	p := createTestProject(t, "测试项目")
	ch := createTestChapter(t, p.ID, 1, "待删除", "内容")

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/chapters/%s", ch.ID.String()), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("DeleteChapter() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var count int64
	store.GetDB().Model(&model.Chapter{}).Where("id = ?", ch.ID).Count(&count)
	if count != 0 {
		t.Error("DeleteChapter() chapter still exists in DB")
	}
}

func TestDeleteChapter_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/chapters/%s", uuid.New().String()), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("DeleteChapter() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestChapter_Authorization(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupChapterRouter()

	// Create chapter for another user
	otherUserID := uuid.New()
	otherProject := model.Project{
		ID:      uuid.New(),
		ShortID: model.GenerateShortID(),
		UserID:  otherUserID,
		Title:   "别人的项目",
	}
	store.GetDB().Create(&otherProject)
	ch := createTestChapter(t, otherProject.ID, 1, "别人的章节", "内容")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/chapters/%s", ch.ID.String()), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetChapter() other user's chapter status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
