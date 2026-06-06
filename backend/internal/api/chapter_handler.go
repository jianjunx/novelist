package api

import (
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

type CreateChapterRequest struct {
	OutlineID  *string `json:"outline_id"`
	ChapterNum int     `json:"chapter_num" binding:"required"`
	Title      string  `json:"title" binding:"required"`
	Content    string  `json:"content"`
}

func GetChapters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var chapters []model.Chapter
	store.GetDB().Where("project_id = ?", c.Param("id")).Order("chapter_num").Find(&chapters)
	c.JSON(http.StatusOK, chapters)
}

func CreateChapter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var project model.Project
	if err := store.GetDB().Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chapter := model.Chapter{
		ProjectID:  uuid.MustParse(c.Param("id")),
		ChapterNum: req.ChapterNum,
		Title:      req.Title,
		Content:    req.Content,
		WordCount:  utf8.RuneCountInString(req.Content),
	}
	if req.OutlineID != nil {
		oid := uuid.MustParse(*req.OutlineID)
		chapter.OutlineID = &oid
	}
	store.GetDB().Create(&chapter)
	c.JSON(http.StatusCreated, chapter)
}

func GetChapter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var chapter model.Chapter
	if err := store.GetDB().Joins("Project").Where("chapters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&chapter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	c.JSON(http.StatusOK, chapter)
}

func UpdateChapter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var chapter model.Chapter
	if err := store.GetDB().Joins("Project").Where("chapters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&chapter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
		updates["word_count"] = utf8.RuneCountInString(req.Content)
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	store.GetDB().Model(&chapter).Updates(updates)
	c.JSON(http.StatusOK, chapter)
}
