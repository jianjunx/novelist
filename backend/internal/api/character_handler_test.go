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
	"gorm.io/datatypes"
)

func setupCharacterRouter() *gin.Engine {
	r := setupGin()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	r.POST("/projects/:id/characters", CreateCharacter)
	r.PUT("/characters/:id", UpdateCharacter)
	return r
}

func TestUpdateCharacter_PreservesRelationshipsWhenOmitted(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupCharacterRouter()

	p := createTestProject(t, "人物项目")
	charID := uuid.New()
	rels := datatypes.JSON(`[{"target":"李四","type":"朋友"}]`)
	if err := store.GetDB().Create(&model.Character{
		ID: charID, ProjectID: p.ID, Name: "张三", Role: "主角",
		Relationships: rels,
	}).Error; err != nil {
		t.Fatalf("create character: %v", err)
	}

	body := `{"name":"张三","role":"主角","personality":"勇敢","background":"","appearance":""}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/characters/%s", charID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("UpdateCharacter() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var updated model.Character
	if err := store.GetDB().Where("id = ?", charID).First(&updated).Error; err != nil {
		t.Fatalf("reload character: %v", err)
	}
	if string(updated.Relationships) != string(rels) {
		t.Errorf("UpdateCharacter() relationships = %s, want %s", updated.Relationships, rels)
	}
	if updated.Personality != "勇敢" {
		t.Errorf("UpdateCharacter() personality = %q, want %q", updated.Personality, "勇敢")
	}
}

func TestUpdateCharacter_UpdatesRelationshipsWhenProvided(t *testing.T) {
	setupTestDBWithUUID(t)
	r := setupCharacterRouter()

	p := createTestProject(t, "人物项目2")
	charID := uuid.New()
	if err := store.GetDB().Create(&model.Character{
		ID: charID, ProjectID: p.ID, Name: "张三", Role: "主角",
		Relationships: datatypes.JSON(`[]`),
	}).Error; err != nil {
		t.Fatalf("create character: %v", err)
	}

	body := `{"name":"张三","role":"主角","personality":"","background":"","appearance":"","relationships":[{"target":"李四","type":"师徒"}]}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/characters/%s", charID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("UpdateCharacter() status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var updated model.Character
	store.GetDB().Where("id = ?", charID).First(&updated)
	var rels []map[string]string
	if err := json.Unmarshal(updated.Relationships, &rels); err != nil {
		t.Fatalf("unmarshal relationships: %v", err)
	}
	if len(rels) != 1 || rels[0]["target"] != "李四" {
		t.Errorf("UpdateCharacter() relationships = %v, want 李四", rels)
	}
}
