package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAnthropicProvider_StopReason verifies that the Anthropic response parser
// reads stop_reason and exposes it on CompletionResponse, so callers can detect
// max_tokens truncation instead of silently accepting incomplete output.
func TestAnthropicProvider_StopReason(t *testing.T) {
	cases := []struct {
		name      string
		raw       string
		wantNorm  string
	}{
		{"end_turn -> stop", "end_turn", "stop"},
		{"max_tokens -> max_tokens", "max_tokens", "max_tokens"},
		{"stop_sequence -> stop", "stop_sequence", "stop"},
		{"unknown -> empty", "something_new", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := map[string]interface{}{
					"content":     []map[string]string{{"text": "hi"}},
					"model":       "claude-test",
					"stop_reason": tc.raw,
					"usage":       map[string]int{"input_tokens": 1, "output_tokens": 2},
				}
				_ = json.NewEncoder(w).Encode(body)
			}))
			defer srv.Close()

			p := &AnthropicProvider{apiKey: "test-key", baseURL: srv.URL, client: srv.Client()}
			resp, err := p.Complete(context.Background(), CompletionRequest{
				Messages: []Message{{Role: "user", Content: "hi"}},
			})
			if err != nil {
				t.Fatalf("Complete() error = %v", err)
			}
			if resp.StopReason != tc.wantNorm {
				t.Errorf("StopReason = %q, want %q", resp.StopReason, tc.wantNorm)
			}
		})
	}
}

// TestOpenAIProvider_FinishReason verifies that the OpenAI response parser
// reads finish_reason and normalizes "length" to "max_tokens" so the import
// pipeline can fail loudly on truncation.
func TestOpenAIProvider_FinishReason(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		wantNorm string
	}{
		{"stop -> stop", "stop", "stop"},
		{"length -> max_tokens", "length", "max_tokens"},
		{"tool_calls -> stop", "tool_calls", "stop"},
		{"unknown -> empty", "filtered", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := map[string]interface{}{
					"choices": []map[string]interface{}{{
						"message":       map[string]string{"content": "hi"},
						"finish_reason": tc.raw,
					}},
					"model": "gpt-test",
					"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 2},
				}
				_ = json.NewEncoder(w).Encode(body)
			}))
			defer srv.Close()

			p := &OpenAIProvider{apiKey: "test-key", baseURL: srv.URL, client: srv.Client()}
			resp, err := p.Complete(context.Background(), CompletionRequest{
				Messages: []Message{{Role: "user", Content: "hi"}},
			})
			if err != nil {
				t.Fatalf("Complete() error = %v", err)
			}
			if resp.StopReason != tc.wantNorm {
				t.Errorf("StopReason = %q, want %q", resp.StopReason, tc.wantNorm)
			}
		})
	}
}
