package ai

import (
	"context"
	"log"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino/schema"
	"github.com/jj/novelist/internal/config"
)

type ModelManager struct {
	models map[string]*deepseek.ChatModel
	mu     sync.RWMutex
}

var Manager *ModelManager

func InitModelManager(cfg *config.Config) {
	Manager = &ModelManager{models: make(map[string]*deepseek.ChatModel)}

	if cfg.DeepSeekKey != "" {
		cm, err := deepseek.NewChatModel(context.Background(), &deepseek.ChatModelConfig{
			APIKey:      cfg.DeepSeekKey,
			Model:       cfg.DeepSeekModel,
			MaxTokens:   4096,
			Temperature: 0.7,
		})
		if err != nil {
			log.Printf("Failed to init DeepSeek model: %v", err)
		} else {
			Manager.models["deepseek"] = cm
			log.Println("DeepSeek model initialized")
		}
	}
}

func (m *ModelManager) GetModel(name string) (*deepseek.ChatModel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model, ok := m.models[name]
	return model, ok
}

func (m *ModelManager) GetDefault() (*deepseek.ChatModel, bool) {
	return m.GetModel("deepseek")
}

// Chat sends a chat request and returns the response
func Chat(ctx context.Context, model *deepseek.ChatModel, systemPrompt string, messages []Message) (string, error) {
	einoMessages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
	}
	for _, msg := range messages {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		}
		einoMessages = append(einoMessages, &schema.Message{Role: role, Content: msg.Content})
	}

	resp, err := model.Generate(ctx, einoMessages)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// ChatStream sends a chat request and returns a stream
func ChatStream(ctx context.Context, model *deepseek.ChatModel, systemPrompt string, messages []Message) (*schema.StreamReader[*schema.Message], error) {
	einoMessages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
	}
	for _, msg := range messages {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		}
		einoMessages = append(einoMessages, &schema.Message{Role: role, Content: msg.Content})
	}

	return model.Stream(ctx, einoMessages)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
