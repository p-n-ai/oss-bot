package parser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestURLFetcher(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Mathematics Syllabus</h1><p>Topic 1: Algebra</p></body></html>`))
	}))
	defer server.Close()

	f := parser.NewURLFetcher()

	t.Run("fetch-html-page", func(t *testing.T) {
		text, err := f.Fetch(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if text == "" {
			t.Error("Fetch() returned empty text")
		}
	})

	t.Run("invalid-url", func(t *testing.T) {
		_, err := f.Fetch(context.Background(), "http://localhost:1/nonexistent")
		if err == nil {
			t.Error("Fetch() should error for unreachable URL")
		}
	})

	t.Run("empty-url", func(t *testing.T) {
		_, err := f.Fetch(context.Background(), "")
		if err == nil {
			t.Error("Fetch() should error for empty URL")
		}
	})
}
