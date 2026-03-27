package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client is a minimal GitHub REST API client using stdlib net/http.
// Authentication is via a Bearer token (GitHub App installation access token).
type Client struct {
	token      string
	owner      string
	repo       string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a Client for the given repo using an installation access token.
func NewClient(token, owner, repo string) *Client {
	return &Client{
		token:      token,
		owner:      owner,
		repo:       repo,
		httpClient: http.DefaultClient,
		baseURL:    "https://api.github.com",
	}
}

// GetRef returns the SHA of a git ref (e.g. "heads/main").
// GET /repos/{owner}/{repo}/git/ref/{ref}
func (c *Client) GetRef(ctx context.Context, ref string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/ref/%s", c.baseURL, c.owner, c.repo, ref)
	var result struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := c.do(ctx, http.MethodGet, url, nil, &result); err != nil {
		return "", err
	}
	return result.Object.SHA, nil
}

// CreateRef creates a new branch pointing at the given SHA.
// POST /repos/{owner}/{repo}/git/refs
func (c *Client) CreateRef(ctx context.Context, ref, sha string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs", c.baseURL, c.owner, c.repo)
	body := map[string]string{"ref": ref, "sha": sha}
	return c.do(ctx, http.MethodPost, url, body, nil)
}

// PutContents creates or updates a single file in the repo on the given branch.
// If the file already exists on the branch (inherited from main), its sha is
// fetched automatically and included in the request — required by the GitHub API.
// PUT /repos/{owner}/{repo}/contents/{path}
func (c *Client) PutContents(ctx context.Context, path, message, content, branch string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.baseURL, c.owner, c.repo, path)
	body := map[string]interface{}{
		"message": message,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
		"branch":  branch,
	}
	// GitHub requires the existing file's sha when updating. Fetch it if present.
	if sha, err := c.fileSHA(ctx, path, branch); err == nil {
		body["sha"] = sha
	}
	return c.do(ctx, http.MethodPut, url, body, nil)
}

// fileSHA returns the blob sha of a file on the given ref, or an error if not found.
func (c *Client) fileSHA(ctx context.Context, path, ref string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", c.baseURL, c.owner, c.repo, path, ref)
	var result struct {
		SHA string `json:"sha"`
	}
	if err := c.do(ctx, http.MethodGet, url, nil, &result); err != nil {
		return "", err
	}
	return result.SHA, nil
}

// CreatePull opens a pull request and returns its number and URL.
// POST /repos/{owner}/{repo}/pulls
func (c *Client) CreatePull(ctx context.Context, title, body, head, base string, _ []string) (int, string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.baseURL, c.owner, c.repo)
	reqBody := map[string]interface{}{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}
	var result struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := c.do(ctx, http.MethodPost, url, reqBody, &result); err != nil {
		return 0, "", err
	}
	return result.Number, result.HTMLURL, nil
}

// ReadFile fetches file content from the repo and decodes it from base64.
// GET /repos/{owner}/{repo}/contents/{path}?ref={ref}
func (c *Client) ReadFile(ctx context.Context, path, ref string) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", c.baseURL, c.owner, c.repo, path, ref)
	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.do(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, err
	}
	// GitHub returns base64-encoded content with embedded newlines — strip them first.
	cleaned := strings.ReplaceAll(result.Content, "\n", "")
	data, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("decoding file content: %w", err)
	}
	return data, nil
}

// ListDir returns the paths of entries in a repository directory.
// GET /repos/{owner}/{repo}/contents/{path}?ref={ref}  (returns array for directories)
func (c *Client) ListDir(ctx context.Context, path, ref string) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", c.baseURL, c.owner, c.repo, path, ref)
	var items []struct {
		Path string `json:"path"`
		Type string `json:"type"`
	}
	if err := c.do(ctx, http.MethodGet, url, nil, &items); err != nil {
		return nil, err
	}
	paths := make([]string, len(items))
	for i, item := range items {
		paths[i] = item.Path
	}
	return paths, nil
}

// do is the shared HTTP helper: sets auth + content-type headers, decodes JSON response.
func (c *Client) do(ctx context.Context, method, url string, body, result interface{}) error {
	var reqBody *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitHub API %s %s returned %s", method, url, resp.Status)
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}
