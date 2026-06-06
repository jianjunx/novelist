package orchestrator

import (
	"testing"

	"github.com/jj/novelist/internal/model"
)

func TestAggregateSuggestions_Dedup(t *testing.T) {
	input := []Suggestion{
		{Type: "逻辑", Problem: "重复问题", Suggestion: "修复", Priority: 1},
		{Type: "文笔", Problem: "重复问题", Suggestion: "另一种修复", Priority: 2},
		{Type: "节奏", Problem: "不同问题", Suggestion: "调整", Priority: 1},
	}
	result := aggregateSuggestions(input)
	if len(result) != 2 {
		t.Errorf("aggregateSuggestions() returned %d items, want 2", len(result))
	}
}

func TestAggregateSuggestions_SortByPriority(t *testing.T) {
	input := []Suggestion{
		{Type: "逻辑", Problem: "A", Suggestion: "fix", Priority: 3},
		{Type: "文笔", Problem: "B", Suggestion: "fix", Priority: 1},
		{Type: "节奏", Problem: "C", Suggestion: "fix", Priority: 2},
	}
	result := aggregateSuggestions(input)
	if len(result) != 3 {
		t.Fatalf("aggregateSuggestions() returned %d items, want 3", len(result))
	}
	if result[0].Priority != 1 {
		t.Errorf("First item priority = %d, want 1", result[0].Priority)
	}
	if result[1].Priority != 2 {
		t.Errorf("Second item priority = %d, want 2", result[1].Priority)
	}
	if result[2].Priority != 3 {
		t.Errorf("Third item priority = %d, want 3", result[2].Priority)
	}
}

func TestAggregateSuggestions_Empty(t *testing.T) {
	result := aggregateSuggestions(nil)
	if len(result) != 0 {
		t.Errorf("aggregateSuggestions(nil) returned %d items, want 0", len(result))
	}

	result = aggregateSuggestions([]Suggestion{})
	if len(result) != 0 {
		t.Errorf("aggregateSuggestions([]) returned %d items, want 0", len(result))
	}
}

func TestAggregateSuggestions_AllUnique(t *testing.T) {
	input := []Suggestion{
		{Type: "逻辑", Problem: "A", Suggestion: "fix", Priority: 1},
		{Type: "文笔", Problem: "B", Suggestion: "fix", Priority: 2},
	}
	result := aggregateSuggestions(input)
	if len(result) != 2 {
		t.Errorf("aggregateSuggestions() returned %d items, want 2", len(result))
	}
}

func TestFormatDiscussionSummary_Nil(t *testing.T) {
	result := formatDiscussionSummary(nil)
	if result != "" {
		t.Errorf("formatDiscussionSummary(nil) = %q, want empty", result)
	}
}

func TestFormatDiscussionSummary_WithSuggestions(t *testing.T) {
	dr := &DiscussionResult{
		EditorSuggestions: []Suggestion{
			{Type: "逻辑", Location: "第一段", Problem: "矛盾", Suggestion: "修改描述", Priority: 1},
		},
		ReaderFeedback: "开头吸引人",
		CriticAnalysis: "主题深刻",
	}
	result := formatDiscussionSummary(dr)
	if result == "" {
		t.Error("formatDiscussionSummary() returned empty for non-nil input")
	}
	// Check it contains key sections
	if !contains(result, "编辑建议") {
		t.Error("formatDiscussionSummary() missing editor section")
	}
	if !contains(result, "读者反馈") {
		t.Error("formatDiscussionSummary() missing reader section")
	}
	if !contains(result, "评论家分析") {
		t.Error("formatDiscussionSummary() missing critic section")
	}
}

func TestFormatDiscussionSummary_EmptyResult(t *testing.T) {
	dr := &DiscussionResult{}
	result := formatDiscussionSummary(dr)
	if result == "" {
		t.Error("formatDiscussionSummary() returned empty for empty DiscussionResult")
	}
}

func TestBuildWorkingMemory_WithTitle(t *testing.T) {
	chapter := &model.Chapter{
		ChapterNum: 1,
		Title:      "初遇",
		Content:    "这是一个故事的开始...",
	}
	result := buildWorkingMemory(chapter)
	if !contains(result, "第1章") {
		t.Error("buildWorkingMemory() missing chapter number")
	}
	if !contains(result, "初遇") {
		t.Error("buildWorkingMemory() missing chapter title")
	}
	if !contains(result, "这是一个故事的开始") {
		t.Error("buildWorkingMemory() missing chapter content")
	}
}

func TestBuildWorkingMemory_NoContent(t *testing.T) {
	chapter := &model.Chapter{
		ChapterNum: 2,
		Title:      "空白章节",
	}
	result := buildWorkingMemory(chapter)
	if !contains(result, "第2章") {
		t.Error("buildWorkingMemory() missing chapter number")
	}
	if !contains(result, "空白章节") {
		t.Error("buildWorkingMemory() missing chapter title")
	}
	// Should not contain "章节内容" section since content is empty
	if contains(result, "章节内容：") {
		t.Error("buildWorkingMemory() should not include content section when empty")
	}
}

func TestBuildWorkingMemory_LongContent(t *testing.T) {
	// Create content longer than 500 runes
	longContent := ""
	for i := 0; i < 600; i++ {
		longContent += "字"
	}
	chapter := &model.Chapter{
		ChapterNum: 3,
		Title:      "长章节",
		Content:    longContent,
	}
	result := buildWorkingMemory(chapter)
	// Should be truncated to 500 chars
	if contains(result, longContent) {
		t.Error("buildWorkingMemory() should truncate content longer than 500 runes")
	}
}

// helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
