package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/api"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

func TestFeedbackHandler_MissingFields(t *testing.T) {
	p := pipeline.New(ai.NewMockProvider("ok"), &output.LocalWriter{}, "prompts/", "")
	h := api.NewFeedbackHandler(p)

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"missing topic_path", map[string]interface{}{"content_type": "teaching_notes", "observation": "test"}},
		{"missing content_type", map[string]interface{}{"topic_path": "math/01", "observation": "test"}},
		{"missing observation", map[string]interface{}{"topic_path": "math/01", "content_type": "teaching_notes"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestFeedbackHandler_MethodNotAllowed(t *testing.T) {
	p := pipeline.New(ai.NewMockProvider("ok"), &output.LocalWriter{}, "prompts/", "")
	h := api.NewFeedbackHandler(p)

	req := httptest.NewRequest(http.MethodGet, "/api/feedback", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestFeedbackHandler_InvalidJSON(t *testing.T) {
	p := pipeline.New(ai.NewMockProvider("ok"), &output.LocalWriter{}, "prompts/", "")
	h := api.NewFeedbackHandler(p)

	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestFeedbackHandler_ValidRequest verifies a well-formed request reaches the pipeline.
// The pipeline will fail (no real AI / repo), but the handler should return 500 not 400.
func TestFeedbackHandler_ValidRequest(t *testing.T) {
	// Use a mock provider that returns empty content — pipeline will error at generation.
	p := pipeline.New(ai.NewMockProvider(""), &output.LocalWriter{}, "testdata/", "")
	h := api.NewFeedbackHandler(p)

	payload := api.FeedbackRequest{
		TopicPath:   "mathematics/algebra/01",
		ContentType: "teaching_notes",
		Observation: "Students consistently confuse positive and negative exponents.",
		Source:      "pai-bot",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Use a context to avoid the pipeline blocking.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so the pipeline exits quickly
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	// Either 200 (unlikely without real deps) or 500 (pipeline error) is acceptable.
	// What's NOT acceptable is 400 (bad request) for a valid payload.
	if w.Code == http.StatusBadRequest {
		t.Errorf("valid request should not return 400, got body: %s", w.Body.String())
	}
}
