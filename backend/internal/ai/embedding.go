package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jj/novelist/internal/config"
)

type EmbeddingManager struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

var EmbeddingMgr *EmbeddingManager

func InitEmbeddingManager(cfg *config.Config) {
	apiKey := cfg.EmbeddingAPIKey
	baseURL := cfg.EmbeddingBaseURL

	if apiKey == "" {
		apiKey = cfg.OpenAIKey
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
	}
	if apiKey == "" {
		apiKey = cfg.DeepSeekKey
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1"
		}
	}

	if apiKey == "" {
		log.Println("Embedding manager not initialized: no API key")
		return
	}

	EmbeddingMgr = &EmbeddingManager{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   cfg.EmbeddingModel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	log.Printf("Embedding manager initialized (model: %s)", cfg.EmbeddingModel)
}

type embeddingRequest struct {
	Input interface{} `json:"input"`
	Model string      `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (m *EmbeddingManager) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}
	results, err := m.BatchGenerateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return results[0], nil
}

func (m *EmbeddingManager) BatchGenerateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty texts")
	}

	body, err := json.Marshal(embeddingRequest{
		Input: texts,
		Model: m.model,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result embeddingResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", result.Error.Message)
	}

	embeddings := make([][]float32, len(texts))
	for _, item := range result.Data {
		if item.Index >= 0 && item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}
	for i, emb := range embeddings {
		if emb == nil {
			return nil, fmt.Errorf("missing embedding at index %d", i)
		}
	}
	return embeddings, nil
}
