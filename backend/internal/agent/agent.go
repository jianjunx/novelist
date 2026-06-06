package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/jj/novelist/internal/ai"
)

type AgentRole string

const (
	RoleCreator  AgentRole = "creator"
	RoleWriter   AgentRole = "writer"
	RoleEditor   AgentRole = "editor"
	RoleReader   AgentRole = "reader"
	RoleCritic   AgentRole = "critic"
	RoleReviser  AgentRole = "reviser"
)

func GetPrompt(role AgentRole) string {
	switch role {
	case RoleCreator:
		return CreatorPrompt
	case RoleWriter:
		return WriterPrompt
	case RoleEditor:
		return EditorPrompt
	case RoleReader:
		return ReaderPrompt
	case RoleCritic:
		return CriticPrompt
	case RoleReviser:
		return ReviserPrompt
	default:
		return ""
	}
}

func Chat(ctx context.Context, role AgentRole, messages []ai.Message) (string, error) {
	model, ok := ai.Manager.GetDefault()
	if !ok {
		return "", fmt.Errorf("no model available")
	}
	return ai.Chat(ctx, model, GetPrompt(role), messages)
}

func ChatStream(ctx context.Context, role AgentRole, messages []ai.Message) (*schema.StreamReader[*schema.Message], error) {
	model, ok := ai.Manager.GetDefault()
	if !ok {
		return nil, fmt.Errorf("no model available")
	}
	return ai.ChatStream(ctx, model, GetPrompt(role), messages)
}
