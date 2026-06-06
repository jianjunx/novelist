package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

func GetWorldSettings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var settings []model.WorldSetting
	store.GetDB().Where("project_id = ?", project.ID).Find(&settings)
	c.JSON(http.StatusOK, settings)
}

func CreateWorldSetting(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req struct {
		Category string `json:"category" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	setting := model.WorldSetting{
		ProjectID: project.ID,
		Category:  req.Category,
		Content:   req.Content,
	}
	if err := store.CreateWorldSetting(c.Request.Context(), &setting); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create world setting"})
		return
	}
	c.JSON(http.StatusCreated, setting)
}

func UpdateWorldSetting(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var setting model.WorldSetting
	if err := store.GetDB().Joins("Project").Where("world_settings.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}
	var req struct {
		Category string `json:"category"`
		Content  string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := store.UpdateWorldSetting(c.Request.Context(), setting.ID, map[string]interface{}{
		"category": req.Category,
		"content":  req.Content,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update world setting"})
		return
	}
	c.JSON(http.StatusOK, setting)
}

func DeleteWorldSetting(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var setting model.WorldSetting
	if err := store.GetDB().Joins("Project").Where("world_settings.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "World setting not found"})
		return
	}
	store.GetDB().Delete(&setting)
	c.JSON(http.StatusOK, gin.H{"message": "World setting deleted"})
}
