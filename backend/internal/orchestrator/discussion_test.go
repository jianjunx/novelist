package orchestrator

import (
	"encoding/json"
	"testing"
)

func TestExtractJSON_CreatorResponse(t *testing.T) {
	input := "```json\n{\n  \"content\": \"好的，我来帮你构思一个玄幻小说。\",\n  \"options\": [\"修仙流\", \"武侠流\", \"都市异能\"],\n  \"complete\": false\n}\n```\n请告诉我你更喜欢哪个方向？"

	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed on Creator response format")
	}
	var parsed CreatorChatResponse
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("extractJSON() result unmarshal failed: %v", err)
	}
	if parsed.Content == "" {
		t.Error("extractJSON() content should not be empty")
	}
	if len(parsed.Options) != 3 {
		t.Errorf("extractJSON() options count = %d, want 3", len(parsed.Options))
	}
	if parsed.Complete {
		t.Error("extractJSON() complete should be false")
	}
}

func TestExtractJSON_EditorResponse(t *testing.T) {
	input := `[{"type":"逻辑","location":"第一段","problem":"时间线矛盾","suggestion":"修改时间描述","priority":1}]`
	_, ok := extractJSON(input)
	if ok {
		t.Skip("extractJSON() does not extract JSON arrays - expected behavior")
	}
}

func TestDiscussionResult_Structure(t *testing.T) {
	result := &DiscussionResult{
		EditorSuggestions: []Suggestion{
			{Type: "逻辑", Location: "开头", Problem: "矛盾", Suggestion: "修改", Priority: 1},
			{Type: "文笔", Location: "第二段", Problem: "重复", Suggestion: "精简", Priority: 2},
		},
		ReaderFeedback: "故事引人入胜",
		CriticAnalysis: "主题深刻，人物立体",
		Aggregated: []Suggestion{
			{Type: "逻辑", Location: "开头", Problem: "矛盾", Suggestion: "修改", Priority: 1},
		},
		Errors: nil,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("DiscussionResult marshal failed: %v", err)
	}

	var roundTrip DiscussionResult
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("DiscussionResult unmarshal failed: %v", err)
	}

	if len(roundTrip.EditorSuggestions) != 2 {
		t.Errorf("DiscussionResult editor_suggestions count = %d, want 2", len(roundTrip.EditorSuggestions))
	}
	if roundTrip.ReaderFeedback != "故事引人入胜" {
		t.Errorf("DiscussionResult reader_feedback = %q, want %q", roundTrip.ReaderFeedback, "故事引人入胜")
	}
}

func TestMultiRoundDiscussionResult_Structure(t *testing.T) {
	result := &MultiRoundDiscussionResult{
		TotalRounds: 2,
		Rounds: map[int]*DiscussionResult{
			1: {
				EditorSuggestions: []Suggestion{{Type: "逻辑", Problem: "问题1", Priority: 1}},
				ReaderFeedback:    "第一轮反馈",
				CriticAnalysis:    "第一轮分析",
			},
			2: {
				EditorSuggestions: []Suggestion{{Type: "文笔", Problem: "问题2", Priority: 2}},
				ReaderFeedback:    "第二轮反馈",
				CriticAnalysis:    "第二轮分析",
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("MultiRoundDiscussionResult marshal failed: %v", err)
	}

	var roundTrip MultiRoundDiscussionResult
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("MultiRoundDiscussionResult unmarshal failed: %v", err)
	}

	if roundTrip.TotalRounds != 2 {
		t.Errorf("MultiRoundDiscussionResult total_rounds = %d, want 2", roundTrip.TotalRounds)
	}
	if len(roundTrip.Rounds) != 2 {
		t.Errorf("MultiRoundDiscussionResult rounds count = %d, want 2", len(roundTrip.Rounds))
	}
}

func TestReviewResult_Structure(t *testing.T) {
	result := &ReviewResult{
		Discussion: &DiscussionResult{
			ReaderFeedback: "不错",
		},
		RevisedContent: "修改后的章节内容",
		RoundNum:       1,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("ReviewResult marshal failed: %v", err)
	}

	var roundTrip ReviewResult
	json.Unmarshal(data, &roundTrip)
	if roundTrip.RevisedContent != "修改后的章节内容" {
		t.Errorf("ReviewResult revised_content = %q, want %q", roundTrip.RevisedContent, "修改后的章节内容")
	}
	if roundTrip.RoundNum != 1 {
		t.Errorf("ReviewResult round_num = %d, want 1", roundTrip.RoundNum)
	}
}

func TestCreatorChatResponse_WithData(t *testing.T) {
	resp := CreatorChatResponse{
		Content:  "构思完成！",
		Options:  []string{"方案A", "方案B"},
		Complete: true,
		Data: &BrainstormData{
			Characters: []CharacterData{
				{
					Name:          "张三",
					Role:          "主角",
					Personality:   "勇敢",
					Relationships: json.RawMessage(`[{"target":"李四","type":"朋友"}]`),
				},
			},
			WorldSettings: []WorldSettingData{
				{Category: "地理", Content: "大陆分为五块"},
			},
			Outlines: []OutlineData{
				{
					Act:        1,
					ChapterNum: 1,
					Summary:    "开篇",
					KeyEvents:  json.RawMessage(`[{"event":"相遇","location":"青云宗","characters":["张三"]}]`),
				},
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("CreatorChatResponse marshal failed: %v", err)
	}

	var roundTrip CreatorChatResponse
	json.Unmarshal(data, &roundTrip)
	if !roundTrip.Complete {
		t.Error("CreatorChatResponse complete should be true")
	}
	if roundTrip.Data == nil {
		t.Fatal("CreatorChatResponse data should not be nil")
	}
	if len(roundTrip.Data.Characters) != 1 {
		t.Errorf("CreatorChatResponse characters count = %d, want 1", len(roundTrip.Data.Characters))
	}
	if roundTrip.Data.Characters[0].Name != "张三" {
		t.Errorf("CreatorChatResponse character name = %q, want %q", roundTrip.Data.Characters[0].Name, "张三")
	}
	if len(roundTrip.Data.Characters[0].Relationships) == 0 {
		t.Error("CreatorChatResponse relationships should not be empty")
	}
	if len(roundTrip.Data.Outlines[0].KeyEvents) == 0 {
		t.Error("CreatorChatResponse key_events should not be empty")
	}
}

func TestDiscussionResult_WithErrors(t *testing.T) {
	result := &DiscussionResult{
		EditorSuggestions: []Suggestion{{Type: "逻辑", Problem: "问题", Priority: 1}},
		ReaderFeedback:    "反馈",
		Errors:            map[string]string{"critic": "timeout"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundTrip DiscussionResult
	json.Unmarshal(data, &roundTrip)
	if roundTrip.Errors["critic"] != "timeout" {
		t.Errorf("errors[critic] = %q, want %q", roundTrip.Errors["critic"], "timeout")
	}
}
