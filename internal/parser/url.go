package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPURLFetcher implements URLFetcher using Go net/http.
// Fetches web pages and extracts visible text content.
type HTTPURLFetcher struct {
	client *http.Client
}

// NewURLFetcher creates a new URL fetcher with sensible defaults.
func NewURLFetcher() *HTTPURLFetcher {
	return &HTTPURLFetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *HTTPURLFetcher) Fetch(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "oss-bot/1.0 (curriculum importer)")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("URL returned status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if strings.Contains(contentType, "text/html") {
		return extractTextFromHTML(string(body)), nil
	}

	return string(body), nil
}

// extractTextFromHTML strips HTML tags and returns visible text content.
// Skips script, style, and other non-content elements.
func extractTextFromHTML(html string) string {
	var sb strings.Builder
	inTag := false
	inSkip := false // inside script/style/nav/footer/header
	tagBuf := strings.Builder{}

	i := 0
	for i < len(html) {
		ch := html[i]
		switch {
		case ch == '<':
			inTag = true
			tagBuf.Reset()
			i++
		case ch == '>' && inTag:
			inTag = false
			tag := strings.ToLower(strings.TrimSpace(tagBuf.String()))
			// Check for block-level skip tags
			tagName := strings.Fields(tag)
			if len(tagName) > 0 {
				name := strings.TrimPrefix(tagName[0], "/")
				switch name {
				case "script", "style", "nav", "footer", "header":
					inSkip = strings.HasPrefix(tag, "/") == false && !strings.HasPrefix(tag, "/")
					// closing tag clears skip
					if strings.HasPrefix(tagName[0], "/") {
						inSkip = false
					} else {
						inSkip = true
					}
				}
			}
			i++
		case inTag:
			tagBuf.WriteByte(ch)
			i++
		default:
			if !inSkip {
				sb.WriteByte(ch)
			}
			i++
		}
	}

	// Collapse whitespace
	text := sb.String()
	lines := strings.Split(text, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, " ")
}
