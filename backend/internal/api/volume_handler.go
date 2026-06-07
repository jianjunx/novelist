package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

type CreateVolumeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func GetVolumes(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	volumes, err := store.GetVolumesByProject(c.Request.Context(), project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch volumes"})
		return
	}
	c.JSON(http.StatusOK, volumes)
}

func CreateVolume(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Determine next volume_num
	volumes, _ := store.GetVolumesByProject(c.Request.Context(), project.ID)
	nextNum := 1
	if len(volumes) > 0 {
		nextNum = volumes[len(volumes)-1].VolumeNum + 1
	}

	var req CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = CreateVolumeRequest{}
	}

	title := req.Title
	if title == "" {
		title = defaultVolumeTitle(nextNum)
	}

	volume := model.Volume{
		ProjectID:   project.ID,
		VolumeNum:   nextNum,
		Title:       title,
		Description: req.Description,
	}
	if err := store.CreateVolume(c.Request.Context(), &volume); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create volume"})
		return
	}
	c.JSON(http.StatusCreated, volume)
}

func UpdateVolume(c *gin.Context) {
	var volume model.Volume
	if err := store.GetDB().Where("id = ?", c.Param("id")).First(&volume).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Volume not found"})
		return
	}
	// Verify ownership
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", volume.ProjectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var req CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if err := store.UpdateVolume(c.Request.Context(), volume.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update volume"})
		return
	}
	c.JSON(http.StatusOK, volume)
}

func DeleteVolume(c *gin.Context) {
	var volume model.Volume
	if err := store.GetDB().Where("id = ?", c.Param("id")).First(&volume).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Volume not found"})
		return
	}
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", volume.ProjectID, userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if err := store.DeleteVolume(c.Request.Context(), volume.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete volume"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Volume deleted"})
}

func defaultVolumeTitle(n int) string {
	titles := []string{"第一篇", "第二篇", "第三篇", "第四篇", "第五篇", "第六篇", "第七篇", "第八篇", "第九篇", "第十篇"}
	if n >= 1 && n <= len(titles) {
		return titles[n-1]
	}
	return fmt.Sprintf("第%d篇", n)
}
