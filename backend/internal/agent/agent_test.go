package agent

import (
	"strings"
	"testing"
)

func TestGetPrompt_AllRoles(t *testing.T) {
	roles := []struct {
		role     AgentRole
		name     string
		keywords []string
	}{
		{RoleCreator, "Creator", []string{"构思", "JSON", "complete", "data"}},
		{RoleWriter, "Writer", []string{"大纲", "去AI味", "写作准则"}},
		{RoleEditor, "Editor", []string{"逻辑一致性", "文笔质量", "JSON", "priority"}},
		{RoleReader, "Reader", []string{"读者", "吸引力", "第一人称"}},
		{RoleCritic, "Critic", []string{"文学评论", "主题深度", "叙事技巧"}},
		{RoleReviser, "Reviser", []string{"反馈", "修改", "不要大幅重写"}},
	}

	for _, tc := range roles {
		t.Run(tc.name, func(t *testing.T) {
			prompt := GetPrompt(tc.role)
			if prompt == "" {
				t.Errorf("GetPrompt(%s) returned empty string", tc.role)
			}
			for _, kw := range tc.keywords {
				if !strings.Contains(prompt, kw) {
					t.Errorf("GetPrompt(%s) missing keyword %q", tc.role, kw)
				}
			}
		})
	}
}

func TestGetPrompt_UnknownRole(t *testing.T) {
	prompt := GetPrompt("nonexistent")
	if prompt != "" {
		t.Errorf("GetPrompt(unknown) = %q, want empty", prompt)
	}
}

func TestPromptConstants_NotEmpty(t *testing.T) {
	prompts := map[string]string{
		"CreatorPrompt":  CreatorPrompt,
		"WriterPrompt":   WriterPrompt,
		"EditorPrompt":   EditorPrompt,
		"ReaderPrompt":   ReaderPrompt,
		"CriticPrompt":   CriticPrompt,
		"ReviserPrompt":  ReviserPrompt,
	}
	for name, p := range prompts {
		if p == "" {
			t.Errorf("%s is empty", name)
		}
	}
}

func TestWriterPrompt_DeAIGuidelines(t *testing.T) {
	prompt := WriterPrompt
	// Verify it contains anti-AI-ification guidelines
	if !strings.Contains(prompt, "去AI味") {
		t.Error("WriterPrompt should contain de-AI guidelines")
	}
	if !strings.Contains(prompt, "口语化") {
		t.Error("WriterPrompt should mention colloquial dialogue")
	}
}

func TestEditorPrompt_JSONFormat(t *testing.T) {
	prompt := EditorPrompt
	if !strings.Contains(prompt, "JSON") {
		t.Error("EditorPrompt should specify JSON output format")
	}
	if !strings.Contains(prompt, "priority") {
		t.Error("EditorPrompt should mention priority field")
	}
}

func TestReaderPrompt_FirstPerson(t *testing.T) {
	prompt := ReaderPrompt
	if !strings.Contains(prompt, "第一人称") {
		t.Error("ReaderPrompt should require first-person perspective")
	}
}

func TestRoleConstants(t *testing.T) {
	roles := map[AgentRole]string{
		RoleCreator: "creator",
		RoleWriter:  "writer",
		RoleEditor:  "editor",
		RoleReader:  "reader",
		RoleCritic:  "critic",
		RoleReviser: "reviser",
	}
	for role, expected := range roles {
		if string(role) != expected {
			t.Errorf("Role %v = %q, want %q", role, string(role), expected)
		}
	}
}

func TestGetPrompt_RoundTrip(t *testing.T) {
	// Verify GetPrompt returns the correct constant for each role
	tests := []struct {
		role     AgentRole
		expected string
	}{
		{RoleCreator, CreatorPrompt},
		{RoleWriter, WriterPrompt},
		{RoleEditor, EditorPrompt},
		{RoleReader, ReaderPrompt},
		{RoleCritic, CriticPrompt},
		{RoleReviser, ReviserPrompt},
	}
	for _, tc := range tests {
		if GetPrompt(tc.role) != tc.expected {
			t.Errorf("GetPrompt(%s) doesn't match the corresponding constant", tc.role)
		}
	}
}
