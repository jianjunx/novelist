package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
	"gorm.io/datatypes"
)

type CreateCharacterRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Role          string                 `json:"role"`
	Personality   string                 `json:"personality"`
	Background    string                 `json:"background"`
	Appearance    string                 `json:"appearance"`
	Relationships map[string]interface{} `json:"relationships"`
}

func GetCharacters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var characters []model.Character
	store.GetDB().Where("project_id = ?", c.Param("id")).Find(&characters)
	c.JSON(http.StatusOK, characters)
}

func CreateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	relJSON, _ := json.Marshal(req.Relationships)
	character := model.Character{
		ProjectID:     uuid.MustParse(c.Param("id")),
		Name:          req.Name,
		Role:          req.Role,
		Personality:   req.Personality,
		Background:    req.Background,
		Appearance:    req.Appearance,
		Relationships: datatypes.JSON(relJSON),
	}
	if err := store.GetDB().Create(&character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create character"})
		return
	}
	c.JSON(http.StatusCreated, character)
}

func UpdateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var character model.Character
	if err := store.GetDB().Joins("Project").Where("characters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&character).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"})
		return
	}
	var req CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	relJSON, _ := json.Marshal(req.Relationships)
	store.GetDB().Model(&character).Updates(map[string]interface{}{
		"name":          req.Name,
		"role":          req.Role,
		"personality":   req.Personality,
		"background":    req.Background,
		"appearance":    req.Appearance,
		"relationships": datatypes.JSON(relJSON),
	})
	c.JSON(http.StatusOK, character)
}
