package output

import (
	"context"
	"fmt"
)

// GitHubWriter creates PRs via the GitHub API. Used by Bot and Web Portal.
// Placeholder — full implementation in Week 5 when internal/github package is built.
type GitHubWriter struct {
	RepoOwner string
	RepoName  string
}

// NewGitHubWriter creates a writer that creates PRs via the GitHub API.
func NewGitHubWriter(owner, repo string) *GitHubWriter {
	return &GitHubWriter{RepoOwner: owner, RepoName: repo}
}

func (w *GitHubWriter) WriteFiles(_ context.Context, _ string, _ map[string]string) error {
	return fmt.Errorf("GitHubWriter does not support local file writing — use LocalWriter")
}

func (w *GitHubWriter) CreatePR(_ context.Context, input PRInput) (*PROutput, error) {
	// Placeholder — will be implemented in Week 5 (Day 22) with internal/github package
	return nil, fmt.Errorf("GitHubWriter.CreatePR not yet implemented (Week 5)")
}
