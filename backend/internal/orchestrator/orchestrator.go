package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jj/novelist/internal/agent"
	"github.com/jj/novelist/internal/ai"
	"github.com/jj/novelist/internal/memory"
	"github.com/jj/novelist/internal/model"
	"github.com/jj/novelist/internal/store"
)

type Orchestrator struct{}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

// extractJSON attempts to extract a JSON object from a string that may contain
// markdown code blocks, explanatory text, or other non-JSON content.
func extractJSON(s string) (string, bool) {
	s = strings.TrimSpace(s)

	// Try direct parse first (pure JSON)
	if json.Valid([]byte(s)) {
		return s, true
	}

	// Try to extract from markdown code block: ```json ... ``` or ``` ... ```
	re := regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(\\{.*?\\})\\s*\\n?```")
	if matches := re.FindStringSubmatch(s); len(matches) > 1 {
		candidate := strings.TrimSpace(matches[1])
		if json.Valid([]byte(candidate)) {
			return candidate, true
		}
	}

	// Try to find the first { ... } block in the text
	depth := 0
	start := -1
	for i, ch := range s {
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && start >= 0 {
				candidate := s[start : i+1]
				if json.Valid([]byte(candidate)) {
					return candidate, true
				}
			}
		}
	}

	return "", false
}

// CreatorChatResponse represents the structured response from Creator agent
type CreatorChatResponse struct {
	Content  string          `json:"content"`
	Options  []string        `json:"options,omitempty"`
	Complete bool            `json:"complete,omitempty"`
	Data     *BrainstormData `json:"data,omitempty"`
	SavedIDs *SavedIDs       `json:"saved_ids,omitempty"`
}

// SavedIDs holds the IDs of saved brainstorm data and created chapters
type SavedIDs struct {
	CharacterIDs    []uuid.UUID `json:"character_ids,omitempty"`
	WorldSettingIDs []uuid.UUID `json:"world_setting_ids,omitempty"`
	OutlineIDs      []uuid.UUID `json:"outline_ids,omitempty"`
	ChapterIDs      []uuid.UUID `json:"chapter_ids,omitempty"`
	ChapterCount    int         `json:"chapter_count,omitempty"`
}

// BrainstormData represents the structured brainstorming results
type BrainstormData struct {
	Characters    []CharacterData    `json:"characters,omitempty"`
	WorldSettings []WorldSettingData `json:"worldSettings,omitempty"`
	Outlines      []OutlineData      `json:"outlines,omitempty"`
}

type CharacterData struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Personality string `json:"personality"`
	Background  string `json:"background"`
	Appearance  string `json:"appearance"`
}

type WorldSettingData struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}

type OutlineData struct {
	Act        int    `json:"act"`
	ChapterNum int    `json:"chapter_num"`
	Summary    string `json:"summary"`
}

// CreatorChat handles multi-round conversation with Creator Agent
func (o *Orchestrator) CreatorChat(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, messages []ai.Message) (*CreatorChatResponse, error) {
	if projectID != uuid.Nil {
		store.GetDB().Create(&model.Conversation{
			ProjectID: projectID, Role: "user", Content: messages[len(messages)-1].Content,
		})
	}

	var contextStr string
	if projectID != uuid.Nil {
		mem := memory.NewMemory(projectID)
		longTerm, err := mem.LoadLongTermMemory(ctx)
		if err == nil {
			contextStr = longTerm
		}
	}

	if contextStr != "" {
		messages = append([]ai.Message{{Role: "system", Content: "当前项目上下文：\n" + contextStr}}, messages...)
	}

	resp, err := agent.Chat(ctx, agent.RoleCreator, messages)
	if err != nil {
		return nil, err
	}

	if projectID != uuid.Nil {
		store.GetDB().Create(&model.Conversation{
			ProjectID: projectID, Role: "assistant", Content: resp,
		})
	}

	// Parse structured response with robust JSON extraction
	var result CreatorChatResponse
	if jsonStr, ok := extractJSON(resp); ok {
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			result = CreatorChatResponse{Content: resp}
		}
	} else {
		result = CreatorChatResponse{Content: resp}
	}

	// Auto-save brainstorm data when complete and projectID is valid
	if result.Complete && result.Data != nil && projectID != uuid.Nil {
		savedIDs, err := o.saveBrainstormData(ctx, projectID, result.Data)
		if err == nil {
			result.SavedIDs = savedIDs
		}
	}

	return &result, nil
}

// saveBrainstormData persists brainstorm results to DB and creates chapters from outlines
func (o *Orchestrator) saveBrainstormData(ctx context.Context, projectID uuid.UUID, data *BrainstormData) (*SavedIDs, error) {
	db := store.GetDB()
	saved := &SavedIDs{}

	// Save characters
	for _, c := range data.Characters {
		char := model.Character{
			ProjectID:  projectID,
			Name:       c.Name,
			Role:       c.Role,
			Personality: c.Personality,
			Background: c.Background,
			Appearance: c.Appearance,
		}
		if err := db.Create(&char).Error; err == nil {
			saved.CharacterIDs = append(saved.CharacterIDs, char.ID)
		}
	}

	// Save world settings
	for _, ws := range data.WorldSettings {
		setting := model.WorldSetting{
			ProjectID: projectID,
			Category:  ws.Category,
			Content:   ws.Content,
		}
		if err := db.Create(&setting).Error; err == nil {
			saved.WorldSettingIDs = append(saved.WorldSettingIDs, setting.ID)
		}
	}

	// Save outlines and create corresponding chapters
	for _, o := range data.Outlines {
		outline := model.Outline{
			ProjectID:  projectID,
			Act:        o.Act,
			ChapterNum: o.ChapterNum,
			Summary:    o.Summary,
			Status:     "draft",
		}
		if err := db.Create(&outline).Error; err == nil {
			saved.OutlineIDs = append(saved.OutlineIDs, outline.ID)

			// Create a chapter for each outline entry
			chapter := model.Chapter{
				ProjectID:  projectID,
				OutlineID:  &outline.ID,
				ChapterNum: o.ChapterNum,
				Title:      fmt.Sprintf("第%d章", o.ChapterNum),
				Status:     "draft",
			}
			if err := db.Create(&chapter).Error; err == nil {
				saved.ChapterIDs = append(saved.ChapterIDs, chapter.ID)
			}
		}
	}

	saved.ChapterCount = len(saved.ChapterIDs)
	return saved, nil
}

// ExpandOutlines generates additional chapter outlines for a project
func (o *Orchestrator) ExpandOutlines(ctx context.Context, projectID uuid.UUID) (*SavedIDs, error) {
	db := store.GetDB()

	// Load existing context
	mem := memory.NewMemory(projectID)
	contextStr, err := mem.LoadLongTermMemory(ctx)
	if err != nil {
		contextStr = ""
	}

	// Get existing chapter count
	var existingChapters []model.Chapter
	db.Where("project_id = ?", projectID).Order("chapter_num").Find(&existingChapters)
	existingCount := len(existingChapters)

	// Get project info
	var project model.Project
	db.Where("id = ?", projectID).First(&project)

	prompt := fmt.Sprintf(`你正在为小说《%s》扩写后续章节大纲。

已有 %d 章大纲：
%s

请为接下来的内容生成 3-5 个新章节大纲，继续故事发展。章节编号从 %d 开始。

必须以JSON格式输出：
{"outlines": [{"act": %d, "chapter_num": %d, "summary": "章节概要"}]}`,
		project.Title, existingCount, contextStr,
		existingCount+1, 2, existingCount+1)

	messages := []ai.Message{{Role: "user", Content: prompt}}
	resp, err := agent.Chat(ctx, agent.RoleCreator, messages)
	if err != nil {
		return nil, err
	}

	// Parse response
	jsonStr, ok := extractJSON(resp)
	if !ok {
		return nil, fmt.Errorf("failed to parse AI response")
	}

	var result struct {
		Outlines []OutlineData `json:"outlines"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse outlines: %w", err)
	}

	// Save new outlines and create chapters
	saved := &SavedIDs{}
	for _, o := range result.Outlines {
		outline := model.Outline{
			ProjectID:  projectID,
			Act:        o.Act,
			ChapterNum: o.ChapterNum,
			Summary:    o.Summary,
			Status:     "draft",
		}
		if err := db.Create(&outline).Error; err == nil {
			saved.OutlineIDs = append(saved.OutlineIDs, outline.ID)

			chapter := model.Chapter{
				ProjectID:  projectID,
				OutlineID:  &outline.ID,
				ChapterNum: o.ChapterNum,
				Title:      fmt.Sprintf("第%d章", o.ChapterNum),
				Status:     "draft",
			}
			if err := db.Create(&chapter).Error; err == nil {
				saved.ChapterIDs = append(saved.ChapterIDs, chapter.ID)
			}
		}
	}
	saved.ChapterCount = len(saved.ChapterIDs)

	return saved, nil
}

// GenerateChapter generates chapter content using Writer Agent
func (o *Orchestrator) GenerateChapter(ctx context.Context, chapterID uuid.UUID) (string, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return "", fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	contextStr, err := mem.AssembleContext(ctx, chapter.ChapterNum, "")
	if err != nil {
		return "", err
	}

	messages := []ai.Message{
		{Role: "user", Content: fmt.Sprintf("请生成第%d章内容。\n\n%s\n\n章节标题：%s", chapter.ChapterNum, contextStr, chapter.Title)},
	}

	return agent.Chat(ctx, agent.RoleWriter, messages)
}

// ContinueWriting continues writing from current content
func (o *Orchestrator) ContinueWriting(ctx context.Context, chapterID uuid.UUID, currentContent string) (string, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return "", fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, "")

	messages := []ai.Message{
		{Role: "user", Content: fmt.Sprintf("请续写以下内容（500-1000字）：\n\n%s\n\n上下文：\n%s", currentContent, contextStr)},
	}

	return agent.Chat(ctx, agent.RoleWriter, messages)
}

// PolishContent polishes selected content
func (o *Orchestrator) PolishContent(ctx context.Context, chapterID uuid.UUID, content string) (string, error) {
	messages := []ai.Message{
		{Role: "user", Content: fmt.Sprintf("请润色以下文字，保持原意，提升表达质量：\n\n%s", content)},
	}
	return agent.Chat(ctx, agent.RoleWriter, messages)
}

// Suggestion represents a structured suggestion from an agent
type Suggestion struct {
	Type       string `json:"type"`
	Location   string `json:"location"`
	Problem    string `json:"problem"`
	Suggestion string `json:"suggestion"`
	Priority   int    `json:"priority"`
}

// DiscussionResult represents the aggregated discussion result
type DiscussionResult struct {
	EditorSuggestions []Suggestion `json:"editor_suggestions"`
	ReaderFeedback    string       `json:"reader_feedback"`
	CriticAnalysis    string       `json:"critic_analysis"`
	Aggregated        []Suggestion `json:"aggregated"`
}

// StartDiscussion starts the discussion workflow with all review agents
func (o *Orchestrator) StartDiscussion(ctx context.Context, chapterID uuid.UUID) (*DiscussionResult, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return nil, fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, "")

	reviewPrompt := fmt.Sprintf("请审查以下章节：\n\n%s\n\n章节内容：\n%s", contextStr, chapter.Content)
	messages := []ai.Message{{Role: "user", Content: reviewPrompt}}

	var wg sync.WaitGroup
	var mu sync.Mutex
	result := &DiscussionResult{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleEditor, messages)
		if err != nil {
			return
		}
		var suggestions []Suggestion
		json.Unmarshal([]byte(resp), &suggestions)
		mu.Lock()
		result.EditorSuggestions = suggestions
		mu.Unlock()
		for _, s := range suggestions {
			store.GetDB().Create(&model.Discussion{
				ChapterID: chapterID, RoundNum: 1, AgentRole: "editor",
				Content: s.Problem, SuggestionType: s.Type, Priority: s.Priority,
			})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleReader, messages)
		if err != nil {
			return
		}
		mu.Lock()
		result.ReaderFeedback = resp
		mu.Unlock()
		store.GetDB().Create(&model.Discussion{
			ChapterID: chapterID, RoundNum: 1, AgentRole: "reader", Content: resp,
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleCritic, messages)
		if err != nil {
			return
		}
		mu.Lock()
		result.CriticAnalysis = resp
		mu.Unlock()
		store.GetDB().Create(&model.Discussion{
			ChapterID: chapterID, RoundNum: 1, AgentRole: "critic", Content: resp,
		})
	}()

	wg.Wait()

	result.Aggregated = aggregateSuggestions(result.EditorSuggestions)

	return result, nil
}

// ReviewResult holds the combined result of a review round
type ReviewResult struct {
	Discussion   *DiscussionResult `json:"discussion"`
	RevisedContent string          `json:"revised_content"`
	RoundNum       int             `json:"round_num"`
}

// GenerateAndReview generates chapter content, runs one review round, and revises
func (o *Orchestrator) GenerateAndReview(ctx context.Context, chapterID uuid.UUID) (*ReviewResult, error) {
	// Step 1: Generate content
	content, err := o.GenerateChapter(ctx, chapterID)
	if err != nil {
		return nil, fmt.Errorf("generate failed: %w", err)
	}

	// Save generated content
	store.GetDB().Model(&model.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"content":    content,
		"word_count": len([]rune(content)),
	})

	// Step 2: Run review
	discussion, err := o.StartDiscussion(ctx, chapterID)
	if err != nil {
		// Return generated content even if review fails
		return &ReviewResult{RevisedContent: content, RoundNum: 0}, nil
	}

	// Step 3: Revise based on feedback
	revised, err := o.reviseChapter(ctx, chapterID, content, discussion)
	if err != nil {
		return &ReviewResult{Discussion: discussion, RevisedContent: content, RoundNum: 1}, nil
	}

	// Save revised content
	store.GetDB().Model(&model.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"content":    revised,
		"word_count": len([]rune(revised)),
	})

	return &ReviewResult{Discussion: discussion, RevisedContent: revised, RoundNum: 1}, nil
}

// ReviewAndRevise runs a new review round on existing content and revises
func (o *Orchestrator) ReviewAndRevise(ctx context.Context, chapterID uuid.UUID) (*ReviewResult, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return nil, fmt.Errorf("chapter not found: %w", err)
	}

	// Count existing rounds to determine next round number
	var maxRound int
	store.GetDB().Model(&model.Discussion{}).Where("chapter_id = ?", chapterID).Select("COALESCE(MAX(round_num), 0)").Scan(&maxRound)

	// Run review
	discussion, err := o.StartDiscussionWithRound(ctx, chapterID, maxRound+1)
	if err != nil {
		return nil, err
	}

	// Revise
	revised, err := o.reviseChapter(ctx, chapterID, chapter.Content, discussion)
	if err != nil {
		return &ReviewResult{Discussion: discussion, RevisedContent: chapter.Content, RoundNum: maxRound + 1}, nil
	}

	// Save revised content
	store.GetDB().Model(&model.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"content":    revised,
		"word_count": len([]rune(revised)),
	})

	return &ReviewResult{Discussion: discussion, RevisedContent: revised, RoundNum: maxRound + 1}, nil
}

// reviseChapter uses the Reviser agent to revise content based on feedback
func (o *Orchestrator) reviseChapter(ctx context.Context, chapterID uuid.UUID, content string, discussion *DiscussionResult) (string, error) {
	// Build feedback summary
	feedback := "## 编辑建议\n"
	for _, s := range discussion.EditorSuggestions {
		feedback += fmt.Sprintf("- [%s] %s：%s（建议：%s）\n", s.Type, s.Location, s.Problem, s.Suggestion)
	}
	feedback += "\n## 读者反馈\n" + discussion.ReaderFeedback
	feedback += "\n## 评论家分析\n" + discussion.CriticAnalysis

	prompt := fmt.Sprintf("请根据以下反馈修改章节内容。\n\n%s\n\n原文：\n%s", feedback, content)
	messages := []ai.Message{{Role: "user", Content: prompt}}

	revised, err := agent.Chat(ctx, agent.RoleReviser, messages)
	if err != nil {
		return "", err
	}

	// Save discussion revision record
	store.GetDB().Create(&model.Discussion{
		ChapterID: chapterID, RoundNum: 0, AgentRole: "reviser",
		Content: "内容已根据反馈修改",
	})

	return revised, nil
}

// StartDiscussionWithRound starts discussion with a specific round number
func (o *Orchestrator) StartDiscussionWithRound(ctx context.Context, chapterID uuid.UUID, roundNum int) (*DiscussionResult, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return nil, fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, "")

	reviewPrompt := fmt.Sprintf("请审查以下章节：\n\n%s\n\n章节内容：\n%s", contextStr, chapter.Content)
	messages := []ai.Message{{Role: "user", Content: reviewPrompt}}

	var wg sync.WaitGroup
	var mu sync.Mutex
	result := &DiscussionResult{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleEditor, messages)
		if err != nil {
			return
		}
		var suggestions []Suggestion
		json.Unmarshal([]byte(resp), &suggestions)
		mu.Lock()
		result.EditorSuggestions = suggestions
		mu.Unlock()
		for _, s := range suggestions {
			store.GetDB().Create(&model.Discussion{
				ChapterID: chapterID, RoundNum: roundNum, AgentRole: "editor",
				Content: s.Problem, SuggestionType: s.Type, Priority: s.Priority,
			})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleReader, messages)
		if err != nil {
			return
		}
		mu.Lock()
		result.ReaderFeedback = resp
		mu.Unlock()
		store.GetDB().Create(&model.Discussion{
			ChapterID: chapterID, RoundNum: roundNum, AgentRole: "reader", Content: resp,
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleCritic, messages)
		if err != nil {
			return
		}
		mu.Lock()
		result.CriticAnalysis = resp
		mu.Unlock()
		store.GetDB().Create(&model.Discussion{
			ChapterID: chapterID, RoundNum: roundNum, AgentRole: "critic", Content: resp,
		})
	}()

	wg.Wait()

	result.Aggregated = aggregateSuggestions(result.EditorSuggestions)

	return result, nil
}

// aggregateSuggestions deduplicates and sorts suggestions by priority
func aggregateSuggestions(suggestions []Suggestion) []Suggestion {
	seen := make(map[string]bool)
	var result []Suggestion
	for _, s := range suggestions {
		if !seen[s.Problem] {
			seen[s.Problem] = true
			result = append(result, s)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority < result[j].Priority
	})
	return result
}
