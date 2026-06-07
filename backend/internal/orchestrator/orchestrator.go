package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
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

// buildWorkingMemory extracts keywords from chapter content and outline for semantic search
func buildWorkingMemory(chapter *model.Chapter) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("当前章节：第%d章 %s", chapter.ChapterNum, chapter.Title))

	db := store.GetDB()
	if db != nil {
		if chapter.OutlineID != nil {
			var outline model.Outline
			if db.Where("id = ?", *chapter.OutlineID).First(&outline).Error == nil {
				parts = append(parts, fmt.Sprintf("大纲要点：%s", outline.Summary))
			}
		} else {
			var outline model.Outline
			if db.Where("project_id = ? AND chapter_num = ?", chapter.ProjectID, chapter.ChapterNum).
				First(&outline).Error == nil {
				parts = append(parts, fmt.Sprintf("大纲要点：%s", outline.Summary))
			}
		}
	}

	if chapter.Content != "" {
		content := chapter.Content
		runes := []rune(content)
		if len(runes) > 500 {
			content = string(runes[:500])
		}
		parts = append(parts, fmt.Sprintf("章节内容：%s", content))
	}

	return strings.Join(parts, "\n")
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
	Title      string `json:"title"`
	Summary    string `json:"summary"`
}

func (o *Orchestrator) prepareCreatorMessages(ctx context.Context, projectID uuid.UUID, messages []ai.Message) []ai.Message {
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
	return messages
}

func (o *Orchestrator) parseCreatorResponse(ctx context.Context, projectID uuid.UUID, resp string) (*CreatorChatResponse, error) {
	if projectID != uuid.Nil {
		store.GetDB().Create(&model.Conversation{
			ProjectID: projectID, Role: "assistant", Content: resp,
		})
	}

	var result CreatorChatResponse
	if jsonStr, ok := extractJSON(resp); ok {
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			result = CreatorChatResponse{Content: resp}
		}
	} else {
		result = CreatorChatResponse{Content: resp}
	}

	if result.Complete && result.Data != nil && projectID != uuid.Nil {
		savedIDs, err := o.saveBrainstormData(ctx, projectID, result.Data)
		if err == nil {
			result.SavedIDs = savedIDs
		}
	}

	return &result, nil
}

// CreatorChat handles multi-round conversation with Creator Agent
func (o *Orchestrator) CreatorChat(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, messages []ai.Message) (*CreatorChatResponse, error) {
	messages = o.prepareCreatorMessages(ctx, projectID, messages)

	resp, err := agent.Chat(ctx, agent.RoleCreator, messages)
	if err != nil {
		return nil, err
	}

	return o.parseCreatorResponse(ctx, projectID, resp)
}

// CreatorChatStream starts a streaming Creator Agent conversation
func (o *Orchestrator) CreatorChatStream(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, messages []ai.Message) (*schema.StreamReader[*schema.Message], error) {
	messages = o.prepareCreatorMessages(ctx, projectID, messages)
	return agent.ChatStream(ctx, agent.RoleCreator, messages)
}

// FinalizeCreatorChat parses and persists the full streamed response
func (o *Orchestrator) FinalizeCreatorChat(ctx context.Context, projectID uuid.UUID, resp string) (*CreatorChatResponse, error) {
	return o.parseCreatorResponse(ctx, projectID, resp)
}

// saveBrainstormData persists brainstorm results to DB and creates chapters from outlines
func (o *Orchestrator) saveBrainstormData(ctx context.Context, projectID uuid.UUID, data *BrainstormData) (*SavedIDs, error) {
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
		if err := store.CreateCharacter(ctx, &char); err == nil {
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
		if err := store.CreateWorldSetting(ctx, &setting); err == nil {
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
		if err := store.CreateOutline(ctx, &outline); err == nil {
			saved.OutlineIDs = append(saved.OutlineIDs, outline.ID)

			// Create a chapter for each outline entry
			chapter := model.Chapter{
				ProjectID:  projectID,
				OutlineID:  &outline.ID,
				ChapterNum: o.ChapterNum,
				Title:      fmt.Sprintf("第%d章", o.ChapterNum),
				Status:     "draft",
			}
			if err := store.CreateChapter(ctx, &chapter); err == nil {
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

	// Load existing context (project info + world settings + characters + outlines)
	mem := memory.NewMemory(projectID)
	contextStr, err := mem.LoadLongTermMemory(ctx)
	if err != nil {
		contextStr = ""
	}

	// Get existing outlines to infer act and count
	var existingOutlines []model.Outline
	db.Where("project_id = ?", projectID).Order("act, chapter_num").Find(&existingOutlines)
	existingCount := len(existingOutlines)

	// Infer start act: find max act, check if current act is full (>=3 chapters)
	maxAct := 1
	currentActCount := 0
	for _, o := range existingOutlines {
		if o.Act > maxAct {
			maxAct = o.Act
			currentActCount = 1
		} else if o.Act == maxAct {
			currentActCount++
		}
	}
	startAct := maxAct
	if currentActCount >= 3 {
		startAct = maxAct + 1
	}

	// Get last chapter content summary for context
	var lastChapter model.Chapter
	db.Where("project_id = ?", projectID).Order("chapter_num DESC").First(&lastChapter)
	recentSummary := ""
	if lastChapter.ID != uuid.Nil && lastChapter.Content != "" {
		content := lastChapter.Content
		if len([]rune(content)) > 200 {
			content = string([]rune(content)[:200]) + "..."
		}
		recentSummary = fmt.Sprintf("\n\n最近成稿内容（第%d章前200字）：\n%s", lastChapter.ChapterNum, content)
	}

	// Get project info
	var project model.Project
	db.Where("id = ?", projectID).First(&project)

	prompt := fmt.Sprintf(`你正在为小说《%s》扩写后续章节大纲。

已有 %d 章大纲：
%s%s

请为接下来的内容生成 3-5 个新章节大纲，继续故事发展。
章节编号从 %d 开始，归属第 %d 幕。

必须以JSON格式输出：
{"outlines": [{"act": %d, "chapter_num": %d, "title": "章节标题（4-8字）", "summary": "章节概要"}]}`,
		project.Title, existingCount, contextStr, recentSummary,
		existingCount+1, startAct,
		startAct, existingCount+1)

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

	// Save new outlines and create chapters, collecting errors
	saved := &SavedIDs{}
	var saveErrors []string

	for i, o := range result.Outlines {
		// Fallback title if AI didn't provide one
		title := o.Title
		if title == "" {
			title = fmt.Sprintf("第%d章", o.ChapterNum)
		}

		outline := model.Outline{
			ProjectID:  projectID,
			Act:        o.Act,
			ChapterNum: o.ChapterNum,
			Summary:    o.Summary,
			Status:     "draft",
		}
		if err := store.CreateOutline(ctx, &outline); err != nil {
			saveErrors = append(saveErrors, fmt.Sprintf("第%d条大纲保存失败: %v", i+1, err))
			continue
		}
		saved.OutlineIDs = append(saved.OutlineIDs, outline.ID)

		chapter := model.Chapter{
			ProjectID:  projectID,
			OutlineID:  &outline.ID,
			ChapterNum: o.ChapterNum,
			Title:      title,
			Status:     "draft",
		}
		if err := store.CreateChapter(ctx, &chapter); err != nil {
			saveErrors = append(saveErrors, fmt.Sprintf("第%d章创建失败: %v", o.ChapterNum, err))
			continue
		}
		saved.ChapterIDs = append(saved.ChapterIDs, chapter.ID)
	}
	saved.ChapterCount = len(saved.ChapterIDs)

	if len(saveErrors) > 0 {
		return saved, fmt.Errorf("部分保存失败:\n%s", strings.Join(saveErrors, "\n"))
	}

	return saved, nil
}

// GenerateChapter generates chapter content using Writer Agent
func (o *Orchestrator) GenerateChapter(ctx context.Context, chapterID uuid.UUID) (string, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return "", fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	workingMemory := buildWorkingMemory(&chapter)
	contextStr, err := mem.AssembleContext(ctx, chapter.ChapterNum, workingMemory)
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
	workingMemory := buildWorkingMemory(&chapter)
	if currentContent != "" {
		workingMemory += "\n续写起点：" + currentContent
	}
	contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, workingMemory)

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
	EditorSuggestions []Suggestion      `json:"editor_suggestions"`
	ReaderFeedback    string            `json:"reader_feedback"`
	CriticAnalysis    string            `json:"critic_analysis"`
	Aggregated        []Suggestion      `json:"aggregated"`
	Errors            map[string]string `json:"errors,omitempty"`
}

// MultiRoundDiscussionResult holds discussion results keyed by round number
type MultiRoundDiscussionResult struct {
	TotalRounds int                        `json:"total_rounds"`
	Rounds      map[int]*DiscussionResult  `json:"rounds"`
}

// StartDiscussion starts the discussion workflow with all review agents
func (o *Orchestrator) StartDiscussion(ctx context.Context, chapterID uuid.UUID) (*DiscussionResult, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return nil, fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	workingMemory := buildWorkingMemory(&chapter)
	contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, workingMemory)

	reviewPrompt := fmt.Sprintf("请审查以下章节：\n\n%s\n\n章节内容：\n%s", contextStr, chapter.Content)
	messages := []ai.Message{{Role: "user", Content: reviewPrompt}}

	return o.runReviewAgents(ctx, chapterID, 1, messages), nil
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
	store.UpdateChapter(ctx, chapterID, map[string]interface{}{
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
	store.UpdateChapter(ctx, chapterID, map[string]interface{}{
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
	discussion, err := o.StartDiscussionWithRound(ctx, chapterID, maxRound+1, nil)
	if err != nil {
		return nil, err
	}

	// Revise
	revised, err := o.reviseChapter(ctx, chapterID, chapter.Content, discussion)
	if err != nil {
		return &ReviewResult{Discussion: discussion, RevisedContent: chapter.Content, RoundNum: maxRound + 1}, nil
	}

	// Save revised content
	store.UpdateChapter(ctx, chapterID, map[string]interface{}{
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

func formatDiscussionSummary(result *DiscussionResult) string {
	if result == nil {
		return ""
	}
	summary := "## 上一轮讨论汇总\n"
	summary += "### 编辑建议\n"
	for _, s := range result.EditorSuggestions {
		summary += fmt.Sprintf("- [%s] %s：%s（建议：%s）\n", s.Type, s.Location, s.Problem, s.Suggestion)
	}
	summary += "\n### 读者反馈\n" + result.ReaderFeedback
	summary += "\n### 评论家分析\n" + result.CriticAnalysis
	return summary
}

// StartDiscussionWithRound starts discussion with a specific round number.
// When previous is provided, its summary is included in the prompt for deeper review.
func (o *Orchestrator) StartDiscussionWithRound(ctx context.Context, chapterID uuid.UUID, roundNum int, previous *DiscussionResult) (*DiscussionResult, error) {
	var chapter model.Chapter
	if err := store.GetDB().Where("id = ?", chapterID).First(&chapter).Error; err != nil {
		return nil, fmt.Errorf("chapter not found: %w", err)
	}

	mem := memory.NewMemory(chapter.ProjectID)
	workingMemory := buildWorkingMemory(&chapter)
	contextStr, _ := mem.AssembleContext(ctx, chapter.ChapterNum, workingMemory)

	var reviewPrompt string
	if previous != nil {
		reviewPrompt = fmt.Sprintf(
			"请审查以下章节。\n\n以下是上一轮讨论的反馈，请在此基础上进行更深层次的审查，关注尚未解决的问题和新发现的问题：\n\n%s\n\n%s\n\n章节内容：\n%s",
			formatDiscussionSummary(previous), contextStr, chapter.Content,
		)
	} else {
		reviewPrompt = fmt.Sprintf("请审查以下章节：\n\n%s\n\n章节内容：\n%s", contextStr, chapter.Content)
	}
	messages := []ai.Message{{Role: "user", Content: reviewPrompt}}

	return o.runReviewAgents(ctx, chapterID, roundNum, messages), nil
}

// runReviewAgents runs Editor, Reader, and Critic in parallel.
// Individual agent failures are recorded in result.Errors and do not abort the discussion.
func (o *Orchestrator) runReviewAgents(ctx context.Context, chapterID uuid.UUID, roundNum int, messages []ai.Message) *DiscussionResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := &DiscussionResult{
		Errors: make(map[string]string),
	}

	recordError := func(role string, err error) {
		mu.Lock()
		result.Errors[role] = err.Error()
		mu.Unlock()
		slog.Warn("discussion agent failed",
			"agent", role,
			"chapter_id", chapterID,
			"round", roundNum,
			"error", err,
		)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		resp, err := agent.Chat(ctx, agent.RoleEditor, messages)
		if err != nil {
			recordError("editor", err)
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
			recordError("reader", err)
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
			recordError("critic", err)
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
	if len(result.Errors) == 0 {
		result.Errors = nil
	}

	return result
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
