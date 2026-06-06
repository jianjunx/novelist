package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

func GetWorldSettings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var settings []model.WorldSetting
	store.GetDB().Where("project_id = ?", c.Param("id")).Find(&settings)
	c.JSON(http.StatusOK, settings)
}

func CreateWorldSetting(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
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
		ProjectID: uuid.MustParse(c.Param("id")),
		Category:  req.Category,
		Content:   req.Content,
	}
	if err := store.GetDB().Create(&setting).Error; err != nil {
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
	store.GetDB().Model(&setting).Updates(map[string]interface{}{
		"category": req.Category,
		"content":  req.Content,
	})
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
