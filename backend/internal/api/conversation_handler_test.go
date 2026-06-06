package api

import (
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

func setupConversationRouter() *gin.Engine {
	r := setupGin()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Set("username", "testuser")
		c.Next()
	})
	r.GET("/projects/:id/conversations", GetConversations)
	return r
}

func createTestConversation(t *testing.T, projectID uuid.UUID, role, content string) model.Conversation {
	t.Helper()
	c := model.Conversation{
		ID:        uuid.New(),
		ProjectID: projectID,
		Role:      role,
		Content:   content,
	}
	if err := store.GetDB().Create(&c).Error; err != nil {
		t.Fatalf("createTestConversation: %v", err)
	}
	return c
}

func TestGetConversations_Empty(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupConversationRouter()
	p := createTestProject(t, "测试项目")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/conversations", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetConversations() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []model.Conversation
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Errorf("GetConversations() returned %d conversations, want 0", len(resp))
	}
}

func TestGetConversations_WithData(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupConversationRouter()
	p := createTestProject(t, "测试项目")

	createTestConversation(t, p.ID, "user", "你好")
	createTestConversation(t, p.ID, "assistant", "你好！有什么可以帮助你的？")
	createTestConversation(t, p.ID, "user", "帮我写小说")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/conversations", p.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetConversations() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []model.Conversation
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 3 {
		t.Errorf("GetConversations() returned %d conversations, want 3", len(resp))
	}
	// Verify order (ASC by created_at)
	if len(resp) >= 3 && resp[0].Role != "user" {
		t.Errorf("GetConversations() first role = %q, want %q", resp[0].Role, "user")
	}
}

func TestGetConversations_ProjectNotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupConversationRouter()

	req := httptest.NewRequest(http.MethodGet, "/projects/nonexistent/conversations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetConversations() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetConversations_OnlyProjectConversations(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupConversationRouter()

	p1 := createTestProject(t, "项目一")
	p2 := createTestProject(t, "项目二")

	createTestConversation(t, p1.ID, "user", "项目一的消息")
	createTestConversation(t, p2.ID, "user", "项目二的消息")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/conversations", p1.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp []model.Conversation
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Errorf("GetConversations() returned %d conversations for project 1, want 1", len(resp))
	}
	if len(resp) >= 1 && resp[0].Content != "项目一的消息" {
		t.Errorf("GetConversations() content = %q, want %q", resp[0].Content, "项目一的消息")
	}
}

func TestGetConversations_Authorization(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupConversationRouter()

	otherUserID := uuid.New()
	otherProject := model.Project{
		ID:      uuid.New(),
		ShortID: model.GenerateShortID(),
		UserID:  otherUserID,
		Title:   "别人的项目",
	}
	store.GetDB().Create(&otherProject)
	createTestConversation(t, otherProject.ID, "user", "秘密消息")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/projects/%s/conversations", otherProject.ShortID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetConversations() other user's project status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
