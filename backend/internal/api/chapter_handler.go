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

// ChapterWithOutline extends Chapter with outline summary and generation availability
type ChapterWithOutline struct {
	model.Chapter
	OutlineSummary string `json:"outline_summary"`
	CanGenerate    bool   `json:"can_generate"`
}

func GetChapters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var chapters []model.Chapter
	store.GetDB().Where("project_id = ?", project.ID).Order("chapter_num").Find(&chapters)

	// Load outlines for this project
	var outlines []model.Outline
	store.GetDB().Where("project_id = ?", project.ID).Find(&outlines)
	outlineMap := make(map[uuid.UUID]string)
	for _, o := range outlines {
		outlineMap[o.ID] = o.Summary
	}

	// Build response with outline_summary and can_generate
	var results []ChapterWithOutline
	for i, ch := range chapters {
		cwo := ChapterWithOutline{
			Chapter:     ch,
			CanGenerate: false,
		}
		if ch.OutlineID != nil {
			cwo.OutlineSummary = outlineMap[*ch.OutlineID]
		}
		// First chapter can always generate; others need previous chapter to have content
		if i == 0 {
			cwo.CanGenerate = ch.Content == ""
		} else if ch.Content == "" && chapters[i-1].Content != "" {
			cwo.CanGenerate = true
		}
		results = append(results, cwo)
	}
	c.JSON(http.StatusOK, results)
}

func CreateChapter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chapter := model.Chapter{
		ProjectID:  project.ID,
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

func DeleteChapter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var chapter model.Chapter
	if err := store.GetDB().Joins("Project").Where("chapters.id = ? AND projects.user_id = ?", c.Param("id"), userID).First(&chapter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chapter not found"})
		return
	}
	store.GetDB().Delete(&chapter)
	c.JSON(http.StatusOK, gin.H{"message": "Chapter deleted"})
}
