package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

// OpenAIProvider implements the Provider interface for OpenAI.
type OpenAIProvider struct {
	apiKey       string
	baseURL      string
	defaultModel string // overrides "gpt-4o" default when set (used by ReasoningProvider)
	client       *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(apiKey string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required (set OSS_AI_API_KEY)")
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{},
	}, nil
}

// maxRetries is the number of retry attempts for transient errors.
const maxRetries = 3

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		if p.defaultModel != "" {
			model = p.defaultModel
		} else {
			model = "gpt-4o"
		}
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	body := map[string]interface{}{
		"model":      model,
		"messages":   req.Messages,
		"max_tokens": maxTokens,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshaling request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 5 * time.Second
			slog.Warn("retrying after transient error",
				"attempt", attempt, "backoff", backoff, "error", lastErr)
			select {
			case <-ctx.Done():
				return CompletionResponse{}, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := p.doRequest(ctx, jsonBody)
		if err != nil {
			if isTransientError(err) {
				lastErr = err
				continue
			}
			return CompletionResponse{}, err
		}
		return resp, nil
	}
	return CompletionResponse{}, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// doRequest performs a single HTTP request to the chat completions endpoint.
func (p *OpenAIProvider) doRequest(ctx context.Context, jsonBody []byte) (CompletionResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return CompletionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
		return CompletionResponse{}, &transientHTTPError{
			statusCode: resp.StatusCode,
			body:       string(respBody),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return CompletionResponse{}, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return CompletionResponse{}, fmt.Errorf("parsing response: %w", err)
	}

	if len(result.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("API returned no choices")
	}

	return CompletionResponse{
		Content:      result.Choices[0].Message.Content,
		Model:        result.Model,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
	}, nil
}

// transientHTTPError represents a retryable HTTP status (5xx, 429).
type transientHTTPError struct {
	statusCode int
	body       string
}

func (e *transientHTTPError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.statusCode, e.body)
}

// isTransientError returns true for errors that are worth retrying:
// connection resets, timeouts, DNS failures, and HTTP 5xx/429.
func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	// HTTP 5xx or 429
	var httpErr *transientHTTPError
	if errors.As(err, &httpErr) {
		return true
	}
	// Connection reset by peer
	if errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	// Network-level errors (timeout, DNS, connection refused)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	// Wrapped errors — check the error string as a fallback for connection resets
	// that may not unwrap cleanly through all layers.
	errStr := err.Error()
	for _, substr := range []string{"connection reset by peer", "broken pipe", "EOF", "connection refused"} {
		if strings.Contains(errStr, substr) {
			return true
		}
	}
	return false
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
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

func (p *OpenAIProvider) Models() []ModelInfo {
	return []ModelInfo{
		{ID: "gpt-4o", Name: "GPT-4o", MaxTokens: 128000, Description: "Most capable general model"},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", MaxTokens: 128000, Description: "Fast and affordable"},
	}
}
