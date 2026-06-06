package orchestrator

import (
	"testing"
)

func TestExtractJSON_PureJSON(t *testing.T) {
	input := `{"content": "hello", "options": ["a", "b"]}`
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for pure JSON")
	}
	if result != input {
		t.Errorf("extractJSON() = %q, want %q", result, input)
	}
}

func TestExtractJSON_MarkdownCodeBlock(t *testing.T) {
	input := "Here is the result:\n```json\n{\"content\": \"hello\"}\n```\nDone."
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for markdown code block")
	}
	if result != `{"content": "hello"}` {
		t.Errorf("extractJSON() = %q, want %q", result, `{"content": "hello"}`)
	}
}

func TestExtractJSON_MarkdownCodeBlockNoLang(t *testing.T) {
	input := "Some text\n```\n{\"key\": \"value\"}\n```"
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for code block without language")
	}
	if result != `{"key": "value"}` {
		t.Errorf("extractJSON() = %q, want %q", result, `{"key": "value"}`)
	}
}

func TestExtractJSON_MixedText(t *testing.T) {
	input := "The AI says: {\"status\": \"ok\"} and more text"
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for mixed text")
	}
	if result != `{"status": "ok"}` {
		t.Errorf("extractJSON() = %q, want %q", result, `{"status": "ok"}`)
	}
}

func TestExtractJSON_NestedJSON(t *testing.T) {
	input := `{"data": {"characters": [{"name": "Alice"}]}}`
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for nested JSON")
	}
	if result != input {
		t.Errorf("extractJSON() = %q, want %q", result, input)
	}
}

func TestExtractJSON_InvalidJSON(t *testing.T) {
	input := "This is just plain text with no JSON"
	_, ok := extractJSON(input)
	if ok {
		t.Error("extractJSON() should fail for plain text")
	}
}

func TestExtractJSON_EmptyInput(t *testing.T) {
	_, ok := extractJSON("")
	if ok {
		t.Error("extractJSON() should fail for empty input")
	}
}

func TestExtractJSON_IncompleteJSON(t *testing.T) {
	input := `{"content": "hello"`
	_, ok := extractJSON(input)
	if ok {
		t.Error("extractJSON() should fail for incomplete JSON")
	}
}

func TestExtractJSON_WhitespacePadding(t *testing.T) {
	input := "   \n  {\"key\": \"value\"}  \n  "
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for whitespace-padded JSON")
	}
	if result != `{"key": "value"}` {
		t.Errorf("extractJSON() = %q, want %q", result, `{"key": "value"}`)
	}
}

func TestExtractJSON_MultipleJSONBlocks(t *testing.T) {
	// Should extract the first valid one
	input := `First: {"a": 1} then {"b": 2}`
	result, ok := extractJSON(input)
	if !ok {
		t.Fatal("extractJSON() failed for multiple JSON blocks")
	}
	if result != `{"a": 1}` {
		t.Errorf("extractJSON() = %q, want %q", result, `{"a": 1}`)
	}
}
