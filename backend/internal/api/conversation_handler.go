package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

func GetConversations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var conversations []model.Conversation
	store.GetDB().Where("project_id = ?", project.ID).Order("created_at ASC").Find(&conversations)
	c.JSON(http.StatusOK, conversations)
}
