// Package output provides writers for generated content.
package output

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Writer abstracts where generated content is written.
type Writer interface {
	// WriteFiles writes files to the local filesystem (CLI).
	WriteFiles(ctx context.Context, baseDir string, files map[string]string) error

	// CreatePR creates a GitHub PR with the given files (Bot, Web Portal).
	CreatePR(ctx context.Context, input PRInput) (*PROutput, error)
}

// PRInput holds the data needed to create a PR.
type PRInput struct {
	Files        map[string]string // filepath -> content
	TopicPath    string
	ContentType  string
	Quality      int
	Source       string // "cli", "bot", "web"
	MergeDetails string // Optional summary of content merge for PR description.
}

// PROutput holds the result of creating a PR.
type PROutput struct {
	URL    string
	Number int
	Branch string
}

// LocalWriter writes files to the local filesystem. Used by CLI.
type LocalWriter struct{}

func (w *LocalWriter) WriteFiles(_ context.Context, baseDir string, files map[string]string) error {
	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

func (w *LocalWriter) CreatePR(_ context.Context, _ PRInput) (*PROutput, error) {
	return nil, fmt.Errorf("LocalWriter does not support PR creation — use GitHubWriter")
}
