package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

func GetSettings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var settings model.Setting
	if err := store.GetDB().Where("user_id = ?", userID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func UpdateSettings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var settings model.Setting
	if err := store.GetDB().Where("user_id = ?", userID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"})
		return
	}
	var req struct {
		DefaultModel     string `json:"default_model"`
		DeepSeekKey      string `json:"deepseek_key"`
		ClaudeKey        string `json:"claude_key"`
		OpenAIKey        string `json:"openai_key"`
		LocalModelURL    string `json:"local_model_url"`
		DefaultWordCount int    `json:"default_word_count"`
		DiscussionRounds int    `json:"discussion_rounds"`
		LanguageStyle    string `json:"language_style"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	store.GetDB().Model(&settings).Updates(map[string]interface{}{
		"default_model":     req.DefaultModel,
		"deepseek_key":      req.DeepSeekKey,
		"claude_key":        req.ClaudeKey,
		"openai_key":        req.OpenAIKey,
		"local_model_url":   req.LocalModelURL,
		"default_word_count": req.DefaultWordCount,
		"discussion_rounds": req.DiscussionRounds,
		"language_style":    req.LanguageStyle,
	})
	c.JSON(http.StatusOK, settings)
}
