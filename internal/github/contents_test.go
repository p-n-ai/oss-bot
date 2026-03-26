package github_test

import (
	"errors"
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func TestMockContentsClient_ReadFile(t *testing.T) {
	client := &gh.MockContentsClient{
		Files: map[string][]byte{
			"topics/algebra/01.yaml": []byte("id: algebra-01\nname: Linear Equations"),
		},
	}

	t.Run("existing file", func(t *testing.T) {
		data, err := client.ReadFile("owner", "repo", "topics/algebra/01.yaml", "main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected file content, got empty")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := client.ReadFile("owner", "repo", "topics/algebra/99.yaml", "main")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

func TestMockContentsClient_ListDir(t *testing.T) {
	client := &gh.MockContentsClient{
		Dirs: map[string][]string{
			"topics/algebra": {"01.yaml", "02.yaml", "03.yaml"},
		},
	}

	t.Run("existing dir", func(t *testing.T) {
		entries, err := client.ListDir("owner", "repo", "topics/algebra", "main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 3 {
			t.Errorf("expected 3 entries, got %d", len(entries))
		}
	})

	t.Run("missing dir returns empty", func(t *testing.T) {
		entries, err := client.ListDir("owner", "repo", "topics/missing", "main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for missing dir, got %d", len(entries))
		}
	})
}

func TestMockContentsClient_Error(t *testing.T) {
	client := &gh.MockContentsClient{
		Err: errors.New("API error"),
	}

	t.Run("ReadFile returns error", func(t *testing.T) {
		_, err := client.ReadFile("owner", "repo", "any.yaml", "main")
		if err == nil {
			t.Error("expected error from mock, got nil")
		}
	})

	t.Run("ListDir returns error", func(t *testing.T) {
		_, err := client.ListDir("owner", "repo", "any/", "main")
		if err == nil {
			t.Error("expected error from mock, got nil")
		}
	})
}
