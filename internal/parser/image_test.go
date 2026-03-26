package parser_test

import (
	"context"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestImageExtractor(t *testing.T) {
	// OCR-only extractor (no AI provider)
	p := parser.NewImageExtractor(nil, parser.ImageModeOCR)

	t.Run("supported-types", func(t *testing.T) {
		types := p.SupportedTypes()
		if len(types) < 2 {
			t.Errorf("SupportedTypes() should include png and jpeg, got %v", types)
		}
	})

	t.Run("empty-input", func(t *testing.T) {
		_, err := p.Extract(context.Background(), nil, "image/png")
		if err == nil {
			t.Error("Extract() should error for empty input")
		}
	})

	t.Run("non-image-type", func(t *testing.T) {
		_, err := p.Extract(context.Background(), []byte("not an image"), "application/pdf")
		if err == nil {
			t.Error("Extract() should error for non-image MIME type")
		}
	})
}

func TestImageExtractorVisionMode(t *testing.T) {
	mockProvider := ai.NewMockProvider("Topic: Algebra\nLearning Objective: Solve linear equations")
	p := parser.NewImageExtractor(mockProvider, parser.ImageModeVision)

	t.Run("vision-extracts-content", func(t *testing.T) {
		// Minimal valid PNG header
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		text, err := p.Extract(context.Background(), pngHeader, "image/png")
		if err != nil {
			t.Fatalf("Extract() error = %v", err)
		}
		if text == "" {
			t.Error("Extract() returned empty text from vision")
		}
	})

	t.Run("vision-requires-provider", func(t *testing.T) {
		noProvider := parser.NewImageExtractor(nil, parser.ImageModeVision)
		_, err := noProvider.Extract(context.Background(), []byte{0x89}, "image/png")
		if err == nil {
			t.Error("Extract() should error when AI provider is nil in vision mode")
		}
	})
}
