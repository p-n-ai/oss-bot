package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient creates a Client pointed at the given test server URL.
func newTestClient(token, owner, repo, baseURL string) *Client {
	return &Client{
		token:      token,
		owner:      owner,
		repo:       repo,
		httpClient: http.DefaultClient,
		baseURL:    baseURL,
	}
}

func TestClient_GetRef(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("GetRef: expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/git/ref/") {
			t.Errorf("GetRef: unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("GetRef: missing/wrong auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"object": map[string]string{"sha": "abc123def456"},
		})
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	sha, err := client.GetRef(context.Background(), "heads/main")
	if err != nil {
		t.Fatalf("GetRef() error = %v", err)
	}
	if sha != "abc123def456" {
		t.Errorf("GetRef() sha = %q, want %q", sha, "abc123def456")
	}
}

func TestClient_CreateRef(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("CreateRef: expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/git/refs") {
			t.Errorf("CreateRef: unexpected path %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	err := client.CreateRef(context.Background(), "refs/heads/oss-bot/add-notes", "abc123")
	if err != nil {
		t.Fatalf("CreateRef() error = %v", err)
	}
	if gotBody["ref"] != "refs/heads/oss-bot/add-notes" {
		t.Errorf("CreateRef() body ref = %q, want refs/heads/oss-bot/add-notes", gotBody["ref"])
	}
	if gotBody["sha"] != "abc123" {
		t.Errorf("CreateRef() body sha = %q, want abc123", gotBody["sha"])
	}
}

func TestClient_PutContents_NewFile(t *testing.T) {
	// New file: GET returns 404, PUT creates it — no sha required.
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/contents/") {
			t.Errorf("PutContents: unexpected path %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodGet:
			http.Error(w, "not found", http.StatusNotFound)
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("PutContents: unexpected method %s", r.Method)
		}
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	err := client.PutContents(context.Background(), "topics/01/topic.yaml", "Add teaching notes", "content here", "oss-bot/branch")
	if err != nil {
		t.Fatalf("PutContents() error = %v", err)
	}
	if gotBody["message"] != "Add teaching notes" {
		t.Errorf("PutContents() message = %q", gotBody["message"])
	}
	decoded, err := base64.StdEncoding.DecodeString(gotBody["content"].(string))
	if err != nil {
		t.Fatalf("PutContents() content not valid base64: %v", err)
	}
	if string(decoded) != "content here" {
		t.Errorf("PutContents() decoded content = %q, want %q", decoded, "content here")
	}
	if gotBody["branch"] != "oss-bot/branch" {
		t.Errorf("PutContents() branch = %q", gotBody["branch"])
	}
	// No sha field when file is new
	if _, hasSHA := gotBody["sha"]; hasSHA {
		t.Error("PutContents() should not include sha for new file")
	}
}

func TestClient_PutContents_ExistingFile(t *testing.T) {
	// Existing file: GET returns sha, PUT must include it.
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]string{"sha": "existing-sha-abc"})
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	err := client.PutContents(context.Background(), "topics/01/existing.md", "Update notes", "new content", "oss-bot/branch")
	if err != nil {
		t.Fatalf("PutContents() error = %v", err)
	}
	if gotBody["sha"] != "existing-sha-abc" {
		t.Errorf("PutContents() sha = %q, want existing-sha-abc", gotBody["sha"])
	}
}

func TestClient_CreatePull(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("CreatePull: expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/pulls") {
			t.Errorf("CreatePull: unexpected path %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"number":   42,
			"html_url": "https://github.com/owner/repo/pull/42",
		})
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	num, url, err := client.CreatePull(context.Background(), "Add teaching notes", "PR body", "oss-bot/branch", "main", nil)
	if err != nil {
		t.Fatalf("CreatePull() error = %v", err)
	}
	if num != 42 {
		t.Errorf("CreatePull() number = %d, want 42", num)
	}
	if url != "https://github.com/owner/repo/pull/42" {
		t.Errorf("CreatePull() url = %q", url)
	}
	if gotBody["title"] != "Add teaching notes" {
		t.Errorf("CreatePull() title = %v", gotBody["title"])
	}
	if gotBody["head"] != "oss-bot/branch" {
		t.Errorf("CreatePull() head = %v", gotBody["head"])
	}
	if gotBody["base"] != "main" {
		t.Errorf("CreatePull() base = %v", gotBody["base"])
	}
}

func TestClient_ReadFile(t *testing.T) {
	fileContent := "id: algebra-01\nname: Linear Equations\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(fileContent))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("ReadFile: expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/contents/") {
			t.Errorf("ReadFile: unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("ref") != "main" {
			t.Errorf("ReadFile: expected ref=main, got %q", r.URL.Query().Get("ref"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"content":  encoded,
			"encoding": "base64",
		})
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	data, err := client.ReadFile(context.Background(), "topics/algebra/01.yaml", "main")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != fileContent {
		t.Errorf("ReadFile() content = %q, want %q", data, fileContent)
	}
}

func TestClient_ListDir(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("ListDir: expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]string{
			{"name": "01.yaml", "path": "topics/algebra/01.yaml", "type": "file"},
			{"name": "02.yaml", "path": "topics/algebra/02.yaml", "type": "file"},
		})
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)
	entries, err := client.ListDir(context.Background(), "topics/algebra", "main")
	if err != nil {
		t.Fatalf("ListDir() error = %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("ListDir() count = %d, want 2", len(entries))
	}
	if entries[0] != "topics/algebra/01.yaml" {
		t.Errorf("ListDir() entries[0] = %q, want topics/algebra/01.yaml", entries[0])
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient("test-token", "owner", "repo", srv.URL)

	t.Run("GetRef returns error on 404", func(t *testing.T) {
		_, err := client.GetRef(context.Background(), "heads/nonexistent")
		if err == nil {
			t.Error("expected error for 404 response, got nil")
		}
	})

	t.Run("ReadFile returns error on 404", func(t *testing.T) {
		_, err := client.ReadFile(context.Background(), "missing/file.yaml", "main")
		if err == nil {
			t.Error("expected error for 404 response, got nil")
		}
	})
}
