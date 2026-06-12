package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
	"gorm.io/datatypes"
)

func GetOutlines(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var outlines []model.Outline
	store.GetDB().Where("project_id = ?", project.ID).Order("act, chapter_num").Find(&outlines)
	c.JSON(http.StatusOK, outlines)
}

func CreateOutline(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req struct {
		Act        int             `json:"act"`
		ChapterNum int             `json:"chapter_num" binding:"required"`
		Summary    string          `json:"summary" binding:"required"`
		KeyEvents  json.RawMessage `json:"key_events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	keJSON := req.KeyEvents
	if keJSON == nil {
		keJSON = json.RawMessage("[]")
	}
	outline := model.Outline{
		ProjectID:  project.ID,
		Act:        req.Act,
		ChapterNum: req.ChapterNum,
		Summary:    req.Summary,
		KeyEvents:  datatypes.JSON(keJSON),
	}
	if err := store.CreateOutline(c.Request.Context(), &outline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create outline"})
		return
	}
	c.JSON(http.StatusCreated, outline)
}

func UpdateOutline(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var outline model.Outline
	if err := store.GetDB().Where("outlines.id = ? AND outlines.project_id IN (SELECT id FROM projects WHERE user_id = ?)", c.Param("id"), userID).First(&outline).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Outline not found"})
		return
	}
	var req struct {
		Act        int              `json:"act"`
		ChapterNum int              `json:"chapter_num"`
		Summary    string           `json:"summary"`
		KeyEvents  *json.RawMessage `json:"key_events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{
		"act":         req.Act,
		"chapter_num": req.ChapterNum,
		"summary":     req.Summary,
	}
	if req.KeyEvents != nil {
		keJSON := *req.KeyEvents
		if keJSON == nil {
			keJSON = json.RawMessage("[]")
		}
		updates["key_events"] = datatypes.JSON(keJSON)
	}
	if err := store.UpdateOutline(c.Request.Context(), outline.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update outline"})
		return
	}
	store.GetDB().Where("id = ?", outline.ID).First(&outline)
	c.JSON(http.StatusOK, outline)
}

func DeleteOutline(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var outline model.Outline
	if err := store.GetDB().Where("outlines.id = ? AND outlines.project_id IN (SELECT id FROM projects WHERE user_id = ?)", c.Param("id"), userID).First(&outline).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Outline not found"})
		return
	}
	store.GetDB().Delete(&outline)
	c.JSON(http.StatusOK, gin.H{"message": "Outline deleted"})
}
