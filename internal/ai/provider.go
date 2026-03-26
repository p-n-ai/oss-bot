// Package ai provides a unified interface for AI content generation.
// This interface is shared with P&AI Bot for consistency.
package ai

import (
	"context"
	"fmt"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// CompletionRequest is the input to an AI completion.
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// CompletionResponse is the output from an AI completion.
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// StreamChunk represents a streaming response chunk.
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// ModelInfo describes an available model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MaxTokens   int    `json:"max_tokens"`
	Description string `json:"description"`
}

// Provider is the interface all AI providers must implement.
// This interface is shared with P&AI Bot.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	Models() []ModelInfo
}

// NewProvider creates a new AI provider based on the provider name.
func NewProvider(name, apiKey string) (Provider, error) {
	switch name {
	case "openai":
		return NewOpenAIProvider(apiKey)
	case "anthropic":
		return NewAnthropicProvider(apiKey)
	case "ollama":
		return NewOllamaProvider("")
	case "mock":
		return NewMockProvider(""), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s (supported: openai, anthropic, ollama)", name)
	}
}
