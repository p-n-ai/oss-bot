package ai

import (
	"context"
	"time"
)

// MockProvider is a test double for AI providers.
type MockProvider struct {
	Response string
	Err      error
	Delay    time.Duration
}

// NewMockProvider creates a mock provider that returns the given response.
func NewMockProvider(response string) *MockProvider {
	return &MockProvider{Response: response}
}

// NewMockProviderWithDelay creates a mock provider that introduces a delay before responding.
// Used to test concurrency limits.
func NewMockProviderWithDelay(response string, delay time.Duration) *MockProvider {
	return &MockProvider{Response: response, Delay: delay}
}

// NewMockProviderWithError creates a mock provider that always returns the given error.
func NewMockProviderWithError(err error) *MockProvider {
	return &MockProvider{Err: err}
}

func (m *MockProvider) Complete(ctx context.Context, _ CompletionRequest) (CompletionResponse, error) {
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return CompletionResponse{}, ctx.Err()
		}
	}
	if m.Err != nil {
		return CompletionResponse{}, m.Err
	}
	return CompletionResponse{
		Content:      m.Response,
		Model:        "mock",
		InputTokens:  10,
		OutputTokens: len(m.Response),
	}, nil
}

func (m *MockProvider) StreamComplete(ctx context.Context, _ CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		if m.Delay > 0 {
			select {
			case <-time.After(m.Delay):
			case <-ctx.Done():
				ch <- StreamChunk{Error: ctx.Err()}
				return
			}
		}
		if m.Err != nil {
			ch <- StreamChunk{Error: m.Err}
			return
		}
		ch <- StreamChunk{Content: m.Response, Done: true}
	}()
	return ch, nil
}

func (m *MockProvider) Models() []ModelInfo {
	return []ModelInfo{
		{ID: "mock", Name: "Mock Model", MaxTokens: 4096, Description: "Test mock"},
	}
}
