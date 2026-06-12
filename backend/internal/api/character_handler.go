package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
	"gorm.io/datatypes"
)

type CreateCharacterRequest struct {
	Name          string          `json:"name" binding:"required"`
	Role          string          `json:"role"`
	Personality   string          `json:"personality"`
	Background    string          `json:"background"`
	Appearance    string          `json:"appearance"`
	Relationships json.RawMessage `json:"relationships"`
}

type UpdateCharacterRequest struct {
	Name          string           `json:"name" binding:"required"`
	Role          string           `json:"role"`
	Personality   string           `json:"personality"`
	Background    string           `json:"background"`
	Appearance    string           `json:"appearance"`
	Relationships *json.RawMessage `json:"relationships"`
}

func GetCharacters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var characters []model.Character
	store.GetDB().Where("project_id = ?", project.ID).Find(&characters)
	if characters == nil {
		characters = []model.Character{}
	}
	c.JSON(http.StatusOK, characters)
}

func CreateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	relJSON := req.Relationships
	if relJSON == nil {
		relJSON = json.RawMessage("[]")
	}
	character := model.Character{
		ProjectID:     project.ID,
		Name:          req.Name,
		Role:          req.Role,
		Personality:   req.Personality,
		Background:    req.Background,
		Appearance:    req.Appearance,
		Relationships: datatypes.JSON(relJSON),
	}
	if err := store.CreateCharacter(c.Request.Context(), &character); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create character"})
		return
	}
	c.JSON(http.StatusCreated, character)
}

func UpdateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var character model.Character
	if err := store.GetDB().Where("characters.id = ? AND characters.project_id IN (SELECT id FROM projects WHERE user_id = ?)", c.Param("id"), userID).First(&character).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"})
		return
	}
	var req UpdateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{
		"name":        req.Name,
		"role":        req.Role,
		"personality": req.Personality,
		"background":  req.Background,
		"appearance":  req.Appearance,
	}
	if req.Relationships != nil {
		relJSON := *req.Relationships
		if relJSON == nil {
			relJSON = json.RawMessage("[]")
		}
		updates["relationships"] = datatypes.JSON(relJSON)
	}
	if err := store.UpdateCharacter(c.Request.Context(), character.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update character"})
		return
	}
	store.GetDB().Where("id = ?", character.ID).First(&character)
	c.JSON(http.StatusOK, character)
}

func DeleteCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var character model.Character
	if err := store.GetDB().Where("characters.id = ? AND characters.project_id IN (SELECT id FROM projects WHERE user_id = ?)", c.Param("id"), userID).First(&character).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"})
		return
	}
	store.GetDB().Delete(&character)
	c.JSON(http.StatusOK, gin.H{"message": "Character deleted"})
}
