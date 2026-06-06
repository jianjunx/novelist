package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

func setupSettingsRouter() *gin.Engine {
	r := setupGin()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Set("username", "testuser")
		c.Next()
	})
	r.GET("/settings", GetSettings)
	r.PUT("/settings", UpdateSettings)
	return r
}

func createTestSettings(t *testing.T, userID uuid.UUID) model.Setting {
	t.Helper()
	s := model.Setting{
		ID:               uuid.New(),
		UserID:           userID,
		DefaultModel:     "deepseek-chat",
		DefaultWordCount: 800,
		DiscussionRounds: 1,
		LanguageStyle:    "现代中文",
	}
	if err := store.GetDB().Create(&s).Error; err != nil {
		t.Fatalf("createTestSettings: %v", err)
	}
	return s
}

func TestGetSettings_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupSettingsRouter()
	createTestSettings(t, testUserID)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetSettings() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp model.Setting
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.DefaultModel != "deepseek-chat" {
		t.Errorf("GetSettings() default_model = %q, want %q", resp.DefaultModel, "deepseek-chat")
	}
	if resp.DefaultWordCount != 800 {
		t.Errorf("GetSettings() default_word_count = %d, want 800", resp.DefaultWordCount)
	}
	if resp.LanguageStyle != "现代中文" {
		t.Errorf("GetSettings() language_style = %q, want %q", resp.LanguageStyle, "现代中文")
	}
}

func TestGetSettings_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupSettingsRouter()

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GetSettings() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdateSettings_Success(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupSettingsRouter()
	createTestSettings(t, testUserID)

	body := `{"default_model":"gpt-4","deepseek_key":"sk-test","default_word_count":1000,"discussion_rounds":3,"language_style":"古典中文"}`
	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateSettings() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	// Verify the update persisted
	var settings model.Setting
	store.GetDB().Where("user_id = ?", testUserID).First(&settings)
	if settings.DefaultModel != "gpt-4" {
		t.Errorf("UpdateSettings() default_model = %q, want %q", settings.DefaultModel, "gpt-4")
	}
	if settings.DeepSeekKey != "sk-test" {
		t.Errorf("UpdateSettings() deepseek_key = %q, want %q", settings.DeepSeekKey, "sk-test")
	}
	if settings.DefaultWordCount != 1000 {
		t.Errorf("UpdateSettings() default_word_count = %d, want 1000", settings.DefaultWordCount)
	}
	if settings.DiscussionRounds != 3 {
		t.Errorf("UpdateSettings() discussion_rounds = %d, want 3", settings.DiscussionRounds)
	}
	if settings.LanguageStyle != "古典中文" {
		t.Errorf("UpdateSettings() language_style = %q, want %q", settings.LanguageStyle, "古典中文")
	}
}

func TestUpdateSettings_NotFound(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupSettingsRouter()

	body := `{"default_model":"gpt-4"}`
	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateSettings() not found status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestUpdateSettings_PartialUpdate(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupSettingsRouter()
	createTestSettings(t, testUserID)

	// Only update language_style
	body := `{"language_style":"文言文"}`
	req := httptest.NewRequest(http.MethodPut, "/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateSettings() partial status = %d, want %d", w.Code, http.StatusOK)
	}

	var settings model.Setting
	store.GetDB().Where("user_id = ?", testUserID).First(&settings)
	if settings.LanguageStyle != "文言文" {
		t.Errorf("UpdateSettings() language_style = %q, want %q", settings.LanguageStyle, "文言文")
	}
}
