// Package api provides the HTTP handlers for the OSS Bot web portal backend.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

// FeedbackRequest is the payload sent by pai-bot to report observed learning patterns.
type FeedbackRequest struct {
	TopicPath   string `json:"topic_path"`   // Repo-relative topic path
	ContentType string `json:"content_type"` // "teaching_notes", "assessments", "examples"
	Observation string `json:"observation"`  // Natural language observation from pai-bot
	Source      string `json:"source"`       // Caller identifier (e.g. "pai-bot")
	Model       string `json:"model"`        // AI model that generated the observation
}

// FeedbackHandler handles POST /api/feedback requests from pai-bot.
// It runs the generation pipeline with provenance:ai-observed and creates a PR.
type FeedbackHandler struct {
	pipeline *pipeline.Pipeline
}

// NewFeedbackHandler creates a FeedbackHandler backed by the given pipeline.
func NewFeedbackHandler(p *pipeline.Pipeline) *FeedbackHandler {
	return &FeedbackHandler{pipeline: p}
}

func (h *FeedbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if req.TopicPath == "" || req.ContentType == "" || req.Observation == "" {
		http.Error(w, "topic_path, content_type, and observation are required", http.StatusBadRequest)
		return
	}

	source := req.Source
	if source == "" {
		source = "pai-bot"
	}

	result, err := h.pipeline.Execute(r.Context(), pipeline.Request{
		TopicPath:        req.TopicPath,
		ContributionType: req.ContentType,
		Content:          req.Observation,
		Mode:             pipeline.ModeCreatePR,
		Source:           source,
		Options:          map[string]string{"provenance": "ai-observed"},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("pipeline error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"pr_url":    result.PRUrl,
		"pr_number": result.PRNumber,
		"topic":     req.TopicPath,
	})
}
