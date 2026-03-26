package parser

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFParser implements ContentExtractor using Go-native PDF extraction.
// Used by the CLI for standalone operation without external dependencies.
type PDFParser struct{}

// NewPDFParser creates a new Go-native PDF parser.
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Extract(_ context.Context, input []byte, mimeType string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("empty input")
	}
	if mimeType != "" && mimeType != "application/pdf" {
		return "", fmt.Errorf("unsupported MIME type for PDFParser: %s (only application/pdf supported)", mimeType)
	}

	r, err := pdf.NewReader(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		return "", fmt.Errorf("opening PDF: %w", err)
	}

	plain, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extracting PDF text: %w", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(plain); err != nil {
		return "", fmt.Errorf("reading PDF text: %w", err)
	}

	return buf.String(), nil
}

func (p *PDFParser) SupportedTypes() []string {
	return []string{"application/pdf"}
}

// ExtractPDFText is a convenience function for CLI file-based extraction.
func ExtractPDFText(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".pdf" {
		return "", fmt.Errorf("not a PDF file: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	parser := NewPDFParser()
	return parser.Extract(context.Background(), data, "application/pdf")
}
