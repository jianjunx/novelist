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

func GetAvailableModels(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var settings model.Setting
	if err := store.GetDB().Where("user_id = ?", userID).First(&settings).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"models": defaultModels()})
		return
	}

	var models []gin.H
	if settings.DeepSeekKey != "" {
		models = append(models,
			gin.H{"value": "deepseek-chat", "label": "DeepSeek Chat", "provider": "deepseek"},
			gin.H{"value": "deepseek-reasoner", "label": "DeepSeek Reasoner", "provider": "deepseek"},
		)
	}
	if settings.ClaudeKey != "" {
		models = append(models,
			gin.H{"value": "claude-sonnet-4-20250514", "label": "Claude Sonnet 4", "provider": "claude"},
			gin.H{"value": "claude-haiku-4-20250414", "label": "Claude Haiku 4", "provider": "claude"},
		)
	}
	if settings.OpenAIKey != "" {
		models = append(models,
			gin.H{"value": "gpt-4o", "label": "GPT-4o", "provider": "openai"},
			gin.H{"value": "gpt-4o-mini", "label": "GPT-4o Mini", "provider": "openai"},
		)
	}
	if settings.LocalModelURL != "" {
		models = append(models,
			gin.H{"value": "local-model", "label": "本地模型", "provider": "local"},
		)
	}
	if len(models) == 0 {
		models = defaultModels()
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func defaultModels() []gin.H {
	return []gin.H{
		{"value": "deepseek-chat", "label": "DeepSeek Chat", "provider": "deepseek"},
		{"value": "deepseek-reasoner", "label": "DeepSeek Reasoner", "provider": "deepseek"},
	}
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
