package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// TikaParser implements ContentExtractor using Apache Tika server.
// Used by the Bot and Web Portal for multi-format document extraction.
// Requires a running Tika instance (Docker sidecar).
type TikaParser struct {
	tikaURL string
	client  *http.Client
}

// NewTikaParser creates a parser that connects to an Apache Tika server.
// tikaURL is typically "http://tika:9998" in Docker or "http://localhost:9998" locally.
func NewTikaParser(tikaURL string) *TikaParser {
	return &TikaParser{
		tikaURL: tikaURL,
		client:  &http.Client{},
	}
}

func (p *TikaParser) Extract(ctx context.Context, input []byte, mimeType string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("empty input")
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", p.tikaURL+"/tika", bytes.NewReader(input))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "text/plain")
	if mimeType != "" {
		req.Header.Set("Content-Type", mimeType)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Tika server at %s: %w", p.tikaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Tika returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading Tika response: %w", err)
	}

	return string(body), nil
}

func (p *TikaParser) SupportedTypes() []string {
	return []string{
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",   // DOCX
		"application/vnd.openxmlformats-officedocument.presentationml.presentation", // PPTX
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",         // XLSX
		"text/plain",
		"text/html",
		"image/png",
		"image/jpeg",
		"application/rtf",
		"application/epub+zip",
	}
}
