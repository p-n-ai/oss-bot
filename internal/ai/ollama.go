package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// OllamaProvider implements the Provider interface for Ollama (local LLM).
type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider(baseURL string) (*OllamaProvider, error) {
	if baseURL == "" {
		baseURL = os.Getenv("OSS_AI_OLLAMA_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		baseURL: baseURL,
		client:  &http.Client{},
	}, nil
}

func (p *OllamaProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = "llama3"
	}

	// Convert messages to Ollama format
	var messages []map[string]string
	for _, m := range req.Messages {
		messages = append(messages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	body := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return CompletionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("Ollama API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return CompletionResponse{}, fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Model            string `json:"model"`
		PromptEvalCount  int    `json:"prompt_eval_count"`
		EvalCount        int    `json:"eval_count"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return CompletionResponse{}, fmt.Errorf("parsing response: %w", err)
	}

	return CompletionResponse{
		Content:      result.Message.Content,
		Model:        result.Model,
		InputTokens:  result.PromptEvalCount,
		OutputTokens: result.EvalCount,
	}, nil
}

func (p *OllamaProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	// Fall back to non-streaming for now
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		resp, err := p.Complete(ctx, req)
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		ch <- StreamChunk{Content: resp.Content, Done: true}
	}()
	return ch, nil
}

func (p *OllamaProvider) Models() []ModelInfo {
	return []ModelInfo{
		{ID: "llama3", Name: "Llama 3", MaxTokens: 8192, Description: "Free, self-hosted"},
		{ID: "llama3:70b", Name: "Llama 3 70B", MaxTokens: 8192, Description: "Larger self-hosted model"},
	}
}
