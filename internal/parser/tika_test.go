package parser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestTikaParser(t *testing.T) {
	// Mock Tika server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Extracted text from document"))
	}))
	defer server.Close()

	p := parser.NewTikaParser(server.URL)

	t.Run("supported-types", func(t *testing.T) {
		types := p.SupportedTypes()
		if len(types) < 5 {
			t.Errorf("SupportedTypes() should return many types, got %d", len(types))
		}
	})

	t.Run("extract-document", func(t *testing.T) {
		text, err := p.Extract(context.Background(), []byte("fake doc content"), "application/pdf")
		if err != nil {
			t.Fatalf("Extract() error = %v", err)
		}
		if text == "" {
			t.Error("Extract() returned empty text")
		}
	})

	t.Run("empty-input", func(t *testing.T) {
		_, err := p.Extract(context.Background(), nil, "application/pdf")
		if err == nil {
			t.Error("Extract() should error for empty input")
		}
	})
}

func TestTikaParserUnreachable(t *testing.T) {
	p := parser.NewTikaParser("http://localhost:1") // unreachable

	_, err := p.Extract(context.Background(), []byte("content"), "application/pdf")
	if err == nil {
		t.Error("Extract() should error when Tika is unreachable")
	}
}
