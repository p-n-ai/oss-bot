package github

import (
	"context"
	"fmt"
)

// ContentsClient reads files and directories from a GitHub repository.
// Implementations: MockContentsClient (tests), GitHubContentsClient (production).
type ContentsClient interface {
	ReadFile(owner, repo, path, ref string) ([]byte, error)
	ListDir(owner, repo, path, ref string) ([]string, error)
}

// GitHubContentsClient is the production ContentsClient using stdlib net/http via Client.
type GitHubContentsClient struct {
	token string
}

// NewGitHubContentsClient creates a ContentsClient authenticated with the given token.
func NewGitHubContentsClient(token string) *GitHubContentsClient {
	return &GitHubContentsClient{token: token}
}

// ReadFile fetches and decodes a file from GitHub. owner/repo/path/ref from the interface.
func (c *GitHubContentsClient) ReadFile(owner, repo, path, ref string) ([]byte, error) {
	client := NewClient(c.token, owner, repo)
	return client.ReadFile(context.Background(), path, ref)
}

// ListDir returns the paths of entries in a GitHub repository directory.
func (c *GitHubContentsClient) ListDir(owner, repo, path, ref string) ([]string, error) {
	client := NewClient(c.token, owner, repo)
	return client.ListDir(context.Background(), path, ref)
}

// GitHubContentsReader adapts ContentsClient to the pipeline.ContentReader interface.
// It fixes the owner and repo so callers only need to supply path and ref.
type GitHubContentsReader struct {
	Client ContentsClient
	Owner  string
	Repo   string
}

// ReadFile implements pipeline.ContentReader.
func (r *GitHubContentsReader) ReadFile(path, ref string) ([]byte, error) {
	return r.Client.ReadFile(r.Owner, r.Repo, path, ref)
}

// MockContentsClient is an in-memory ContentsClient for use in tests.
type MockContentsClient struct {
	Files map[string][]byte
	Dirs  map[string][]string
	Err   error
}

// ReadFile returns file content from the mock's Files map, or Err if set.
func (m *MockContentsClient) ReadFile(_, _, path, _ string) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

// ListDir returns directory entries from the mock's Dirs map, or Err if set.
// Returns nil (not an error) if the path is not in Dirs.
func (m *MockContentsClient) ListDir(_, _, path, _ string) ([]string, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Dirs[path], nil
}
