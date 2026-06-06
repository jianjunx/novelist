package api

import (
	"fmt"
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

// ProjectWithStatus extends Project with computed status fields
type ProjectWithStatus struct {
	model.Project
	Brainstormed   bool       `json:"brainstormed"`
	HasChapters    bool       `json:"has_chapters"`
	HasContent     bool       `json:"has_content"`
	FirstChapterID *uuid.UUID `json:"first_chapter_id"`
}

// findProjectByParam looks up a project by short_id, falling back to UUID
func findProjectByParam(id string, userID interface{}) (*model.Project, error) {
	var project model.Project
	db := store.GetDB()
	// Try short_id first
	if err := db.Where("short_id = ? AND user_id = ?", id, userID).First(&project).Error; err == nil {
		return &project, nil
	}
	// Fallback to UUID
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&project).Error; err == nil {
		return &project, nil
	}
	return nil, fmt.Errorf("project not found")
}

const projectStatusSQL = `
	SELECT p.*,
		EXISTS(SELECT 1 FROM outlines WHERE project_id = p.id) AS brainstormed,
		EXISTS(SELECT 1 FROM chapters WHERE project_id = p.id) AS has_chapters,
		EXISTS(SELECT 1 FROM chapters WHERE project_id = p.id AND content != '') AS has_content,
		(SELECT id FROM chapters WHERE project_id = p.id ORDER BY chapter_num ASC LIMIT 1) AS first_chapter_id
	FROM projects p
`

func GetProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var results []ProjectWithStatus
	err := store.GetDB().Raw(projectStatusSQL+"WHERE p.user_id = ? ORDER BY p.created_at DESC", userID).Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	c.JSON(http.StatusOK, results)
}

func CreateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	project := model.Project{
		ShortID:     model.GenerateShortID(),
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
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var result ProjectWithStatus
	err = store.GetDB().Raw(projectStatusSQL+"WHERE p.id = ? AND p.user_id = ?", project.ID, userID).Scan(&result).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func UpdateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
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
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	result := store.GetDB().Delete(project)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}
