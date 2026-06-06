package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
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

// CreatorChatResponse represents the structured response from Creator agent
type CreatorChatResponse struct {
	Content  string   `json:"content"`
	Options  []string `json:"options,omitempty"`
	Complete bool     `json:"complete,omitempty"`
	Data     *BrainstormData `json:"data,omitempty"`
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

	// Parse structured response
	var result CreatorChatResponse
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		// If parsing fails, return as plain text
		result = CreatorChatResponse{Content: resp}
	}

	return &result, nil
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
