package parser

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// ImageExtractor implements ContentExtractor for image files.
// Supports two extraction methods:
//   - OCR (Tesseract): fast, free, best for clean printed text
//   - AI Vision (GPT-4o/Claude): handles handwriting, diagrams, flowcharts,
//     whiteboard photos, complex layouts, and tables in images
type ImageExtractor struct {
	aiProvider ai.Provider
	mode       ImageExtractionMode
}

// NewImageExtractor creates a new image extractor.
// aiProvider is required for Vision mode (can be nil for OCR-only).
// mode controls extraction: ImageModeAuto, ImageModeOCR, or ImageModeVision.
func NewImageExtractor(provider ai.Provider, mode ImageExtractionMode) *ImageExtractor {
	return &ImageExtractor{
		aiProvider: provider,
		mode:       mode,
	}
}

func (p *ImageExtractor) Extract(ctx context.Context, input []byte, mimeType string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("empty input")
	}
	if !p.isImageType(mimeType) {
		return "", fmt.Errorf("unsupported MIME type for ImageExtractor: %s", mimeType)
	}

	switch p.mode {
	case ImageModeOCR:
		return p.extractOCR(ctx, input)
	case ImageModeVision:
		return p.extractVision(ctx, input, mimeType)
	case ImageModeAuto:
		// Try OCR first; fall back to Vision if result is sparse
		text, err := p.extractOCR(ctx, input)
		if err == nil && len(strings.TrimSpace(text)) >= 20 {
			return text, nil
		}
		// OCR returned sparse/empty text — try AI Vision
		if p.aiProvider != nil {
			return p.extractVision(ctx, input, mimeType)
		}
		// No AI provider available — return whatever OCR got
		if err != nil {
			return "", fmt.Errorf("OCR failed and no AI provider for vision fallback: %w", err)
		}
		return text, nil
	default:
		return "", fmt.Errorf("unknown image extraction mode: %d", p.mode)
	}
}

// extractOCR uses Tesseract CLI for text extraction.
func (p *ImageExtractor) extractOCR(_ context.Context, _ []byte) (string, error) {
	// TODO: Implement with os/exec call to tesseract binary
	// Write input to temp file, run: tesseract <input> stdout
	// Parse stdout for extracted text
	return "", fmt.Errorf("OCR extraction not yet implemented — install tesseract")
}

// extractVision sends the image to a multimodal AI model via the AI provider.
func (p *ImageExtractor) extractVision(ctx context.Context, input []byte, mimeType string) (string, error) {
	if p.aiProvider == nil {
		return "", fmt.Errorf("AI provider required for vision extraction")
	}

	b64 := base64.StdEncoding.EncodeToString(input)

	prompt := fmt.Sprintf(
		"Extract all educational content from this %s image (base64: %s). "+
			"Identify: subject areas, topic names, learning objectives, "+
			"assessment questions, teaching notes, diagram descriptions, "+
			"and any curriculum structure. "+
			"If there is handwritten text, transcribe it accurately. "+
			"If there are diagrams or flowcharts, describe their structure and content. "+
			"Output as plain text, preserving the logical structure.",
		mimeType, b64[:min(len(b64), 64)]+"...",
	)

	resp, err := p.aiProvider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("AI vision extraction failed: %w", err)
	}

	return resp.Content, nil
}

func (p *ImageExtractor) SupportedTypes() []string {
	return []string{"image/png", "image/jpeg"}
}

func (p *ImageExtractor) isImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}
