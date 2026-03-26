package github

import "fmt"

// ContentsClient reads files and directories from a GitHub repository.
// Implementations: MockContentsClient (tests), GitHubContentsClient (production).
type ContentsClient interface {
	ReadFile(owner, repo, path, ref string) ([]byte, error)
	ListDir(owner, repo, path, ref string) ([]string, error)
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
