package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/orchestrator"
	"github.com/jj/novelist/internal/store"
)

var orch = orchestrator.NewOrchestrator()

// CreatorChat handles multi-round conversation with Creator Agent
func CreatorChat(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		ProjectID string       `json:"project_id"`
		Messages  []ai.Message `json:"messages" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var projectID uuid.UUID
	if req.ProjectID != "" {
		if project, err := findProjectByParam(req.ProjectID, userID); err == nil {
			projectID = project.ID
		}
	}

	if strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
		creatorChatStream(c, userID.(uuid.UUID), projectID, req.Messages)
		return
	}

	resp, err := orch.CreatorChat(c.Request.Context(), userID.(uuid.UUID), projectID, req.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func creatorChatStream(c *gin.Context, userID uuid.UUID, projectID uuid.UUID, messages []ai.Message) {
	stream, err := orch.CreatorChatStream(c.Request.Context(), userID, projectID, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	var fullResp strings.Builder
	flusher, _ := c.Writer.(http.Flusher)

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				writeSSE(c, flusher, gin.H{"error": err.Error()})
			}
			break
		}
		if msg.Content == "" {
			continue
		}
		fullResp.WriteString(msg.Content)
		writeSSE(c, flusher, gin.H{"content": msg.Content})
	}

	result, err := orch.FinalizeCreatorChat(c.Request.Context(), projectID, fullResp.String())
	if err != nil {
		writeSSE(c, flusher, gin.H{"error": err.Error()})
	} else {
		writeSSE(c, flusher, gin.H{
			"final":      true,
			"content":    result.Content,
			"options":    result.Options,
			"complete":   result.Complete,
			"data":       result.Data,
			"saved_ids":  result.SavedIDs,
		})
	}

	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	if flusher != nil {
		flusher.Flush()
	}
}

func writeSSE(c *gin.Context, flusher http.Flusher, data interface{}) {
	payload, _ := json.Marshal(data)
	fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
	if flusher != nil {
		flusher.Flush()
	}
}

// GenerateChapter generates chapter content
func GenerateChapter(c *gin.Context) {
	chapterID := c.Param("id")
	resp, err := orch.GenerateChapter(c.Request.Context(), uuid.MustParse(chapterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	store.UpdateChapter(c.Request.Context(), uuid.MustParse(chapterID), map[string]interface{}{
		"content":    resp,
		"word_count": len([]rune(resp)),
	})

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// ContinueWriting continues writing from current content
func ContinueWriting(c *gin.Context) {
	chapterID := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := orch.ContinueWriting(c.Request.Context(), uuid.MustParse(chapterID), req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// PolishContent polishes selected content
func PolishContent(c *gin.Context) {
	chapterID := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := orch.PolishContent(c.Request.Context(), uuid.MustParse(chapterID), req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": resp})
}

// StartDiscussion starts the discussion workflow with user-configured round count
func StartDiscussion(c *gin.Context) {
	userID, _ := c.Get("user_id")
	chapterID := uuid.MustParse(c.Param("id"))

	discussionRounds := 1
	if settings, err := store.GetSettings(c.Request.Context(), userID.(uuid.UUID)); err == nil && settings.DiscussionRounds > 0 {
		discussionRounds = settings.DiscussionRounds
	}

	multiResult := &orchestrator.MultiRoundDiscussionResult{
		TotalRounds: discussionRounds,
		Rounds:      make(map[int]*orchestrator.DiscussionResult),
	}

	var previous *orchestrator.DiscussionResult
	for round := 1; round <= discussionRounds; round++ {
		result, err := orch.StartDiscussionWithRound(c.Request.Context(), chapterID, round, previous)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		multiResult.Rounds[round] = result
		previous = result
	}

	c.JSON(http.StatusOK, multiResult)
}

// GenerateAndReview generates content, runs one review round, and revises
func GenerateAndReview(c *gin.Context) {
	chapterID := c.Param("id")
	result, err := orch.GenerateAndReview(c.Request.Context(), uuid.MustParse(chapterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ReviewAndRevise runs a new review round on existing content and revises
func ReviewAndRevise(c *gin.Context) {
	chapterID := c.Param("id")
	result, err := orch.ReviewAndRevise(c.Request.Context(), uuid.MustParse(chapterID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ExpandOutlines generates additional chapter outlines for a project
func ExpandOutlines(c *gin.Context) {
	userID, _ := c.Get("user_id")
	project, err := findProjectByParam(c.Param("id"), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Accept optional volume_id
	var req struct {
		VolumeID *string `json:"volume_id"`
	}
	c.ShouldBindJSON(&req) // ignore error, body is optional

	var volumeUUID *uuid.UUID
	if req.VolumeID != nil {
		vid := uuid.MustParse(*req.VolumeID)
		volumeUUID = &vid
	}

	result, err := orch.ExpandOutlines(c.Request.Context(), project.ID, volumeUUID)
	if err != nil && result == nil {
		// Complete failure — no partial result
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check volume completion status
	volumeComplete := false
	if result.VolumeID != nil {
		var outlines []model.Outline
		store.GetDB().Where("project_id = ? AND volume_id = ?", project.ID, *result.VolumeID).Find(&outlines)
		actCounts := map[int]int{}
		for _, o := range outlines {
			actCounts[o.Act]++
		}
		volumeComplete = actCounts[1] >= 2 && actCounts[2] >= 2 && actCounts[3] >= 2
	}

	resp := gin.H{
		"outline_ids":     result.OutlineIDs,
		"chapter_ids":     result.ChapterIDs,
		"chapter_count":   result.ChapterCount,
		"volume_complete": volumeComplete,
	}
	if err != nil {
		// Partial failure — include saved result + error info
		resp["partial_error"] = err.Error()
	}
	c.JSON(http.StatusOK, resp)
}
