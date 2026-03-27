package api

import (
	"fmt"
	"net/http"

	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

// NewRouter creates and returns the HTTP mux for the web portal API.
// It registers all API endpoints against the given pipeline.
func NewRouter(p *pipeline.Pipeline) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /api/feedback", NewFeedbackHandler(p))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	})

	return mux
}
