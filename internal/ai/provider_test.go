package ai_test

import (
	"context"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

func TestMockProvider_Complete(t *testing.T) {
	mock := ai.NewMockProvider("test response")

	resp, err := mock.Complete(context.Background(), ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "user", Content: "Hello"},
		},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "test response" {
		t.Errorf("Complete() content = %q, want %q", resp.Content, "test response")
	}
}

func TestMockProvider_Models(t *testing.T) {
	mock := ai.NewMockProvider("response")
	models := mock.Models()
	if len(models) == 0 {
		t.Error("Models() returned empty")
	}
}

func TestNewProvider_Unknown(t *testing.T) {
	_, err := ai.NewProvider("unknown", "")
	if err == nil {
		t.Error("NewProvider(unknown) should return error")
	}
}
