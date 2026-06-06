package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

type CreateProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
	StyleGuide  string `json:"style_guide"`
}

func GetProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var projects []model.Project
	if err := store.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	c.JSON(http.StatusOK, projects)
}

func CreateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	project := model.Project{
		UserID:      userID.(uuid.UUID),
		Title:       req.Title,
		Genre:       req.Genre,
		Description: req.Description,
		StyleGuide:  req.StyleGuide,
	}
	if err := store.GetDB().Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}
	c.JSON(http.StatusCreated, project)
}

func GetProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

func UpdateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	store.GetDB().Model(&project).Updates(map[string]interface{}{
		"title": req.Title, "genre": req.Genre,
		"description": req.Description, "style_guide": req.StyleGuide,
	})
	c.JSON(http.StatusOK, project)
}

func DeleteProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	result := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).Delete(&model.Project{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}
