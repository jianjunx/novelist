package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/orchestrator"
	"github.com/jj/novelist/internal/store"
)

var orch = orchestrator.NewOrchestrator()

// CreatorChat handles multi-round conversation with Creator Agent
func CreatorChat(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		ProjectID string       `json:"project_id"`
		Messages  []ai.Message `json:"messages" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var projectID uuid.UUID
	if req.ProjectID != "" {
		projectID = uuid.MustParse(req.ProjectID)
	}

	resp, err := orch.CreatorChat(c.Request.Context(), userID.(uuid.UUID), projectID, req.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// GenerateChapter generates chapter content
func GenerateChapter(c *gin.Context) {
	chapterID := c.Param("id")
	resp, err := orch.GenerateChapter(c.Request.Context(), uuid.MustParse(chapterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	store.GetDB().Model(&model.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"content":    resp,
		"word_count": len([]rune(resp)),
	})

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// ContinueWriting continues writing from current content
func ContinueWriting(c *gin.Context) {
	chapterID := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := orch.ContinueWriting(c.Request.Context(), uuid.MustParse(chapterID), req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// PolishContent polishes selected content
func PolishContent(c *gin.Context) {
	chapterID := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := orch.PolishContent(c.Request.Context(), uuid.MustParse(chapterID), req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// StartDiscussion starts the discussion workflow
func StartDiscussion(c *gin.Context) {
	chapterID := c.Param("id")
	result, err := orch.StartDiscussion(c.Request.Context(), uuid.MustParse(chapterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
