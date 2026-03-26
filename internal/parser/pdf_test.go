package parser_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestPDFParser(t *testing.T) {
	p := parser.NewPDFParser()

	t.Run("supported-types", func(t *testing.T) {
		types := p.SupportedTypes()
		if len(types) != 1 || types[0] != "application/pdf" {
			t.Errorf("SupportedTypes() = %v, want [application/pdf]", types)
		}
	})

	t.Run("wrong-mime-type", func(t *testing.T) {
		_, err := p.Extract(context.Background(), []byte("not a pdf"), "text/plain")
		if err == nil {
			t.Error("Extract() should error for non-PDF MIME type")
		}
	})

	t.Run("invalid-pdf-bytes", func(t *testing.T) {
		// Valid MIME type but not real PDF content — should error at pdf.NewReader
		_, err := p.Extract(context.Background(), []byte("not a pdf"), "application/pdf")
		if err == nil {
			t.Error("Extract() should error for invalid PDF bytes")
		}
	})

	t.Run("empty-input", func(t *testing.T) {
		_, err := p.Extract(context.Background(), nil, "application/pdf")
		if err == nil {
			t.Error("Extract() should error for empty input")
		}
	})
}

func TestExtractPDFText(t *testing.T) {
	t.Run("non-existent-file", func(t *testing.T) {
		_, err := parser.ExtractPDFText("/nonexistent/file.pdf")
		if err == nil {
			t.Error("ExtractPDFText() should error for non-existent file")
		}
	})

	t.Run("non-pdf-file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "test.txt")
		os.WriteFile(f, []byte("not a pdf"), 0o644)
		_, err := parser.ExtractPDFText(f)
		if err == nil {
			t.Error("ExtractPDFText() should error for non-PDF file")
		}
	})
}
