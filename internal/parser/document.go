// Package parser handles input parsing (content extraction, natural language, commands).
package parser

import "context"

// ContentExtractor extracts text from various sources for the AI import pipeline.
// Implementations: PDFParser (CLI), TikaParser (server), URLFetcher, ImageExtractor.
type ContentExtractor interface {
	// Extract converts a source to plain text for AI processing.
	// input is the raw file bytes (for files/images) or nil (for URL fetcher).
	// mimeType hints at the format (e.g., "application/pdf", "image/png").
	Extract(ctx context.Context, input []byte, mimeType string) (string, error)

	// SupportedTypes returns the MIME types this extractor handles.
	SupportedTypes() []string
}

// ImageExtractionMode controls how images are processed.
type ImageExtractionMode int

const (
	// ImageModeAuto tries OCR first; if OCR returns low-confidence or sparse
	// text, falls back to AI Vision.
	ImageModeAuto ImageExtractionMode = iota
	// ImageModeOCR forces Tesseract/Tika OCR only (fast, no API cost).
	ImageModeOCR
	// ImageModeVision forces AI Vision via the AI provider (GPT-4o/Claude).
	// Best for handwriting, diagrams, flowcharts, whiteboard photos, complex layouts.
	ImageModeVision
)

// URLFetcher fetches and extracts text from web pages.
type URLFetcher interface {
	// Fetch retrieves a web page and returns its text content.
	// Handles static HTML and optionally renders JavaScript-heavy pages.
	Fetch(ctx context.Context, url string) (string, error)
}

// InputSource represents the three ways users can provide content.
type InputSource struct {
	Type      string              // "url", "text", "file"
	URL       string              // For URL input
	Text      string              // For text (copy-paste) input
	FileData  []byte              // For file upload input
	FileName  string              // Original filename (used to detect MIME type)
	MimeType  string              // MIME type of uploaded file
	ImageMode ImageExtractionMode // For images: Auto, OCR, or Vision
}
