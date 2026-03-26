package ai_test

import (
	"context"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

func TestNewReasoningProvider_UsesBaseWhenNotConfigured(t *testing.T) {
	base := ai.NewMockProvider("base response")
	rp := ai.NewReasoningProvider(base, "")

	// Without a model, should fall back to base provider.
	resp, err := rp.Complete(context.Background(), ai.CompletionRequest{
		Messages: []ai.Message{{Role: "user", Content: "test"}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "base response" {
		t.Errorf("expected base provider response, got %q", resp.Content)
	}
}

func TestNewReasoningProvider_ImplementsProvider(t *testing.T) {
	base := ai.NewMockProvider("ok")
	var _ ai.Provider = ai.NewReasoningProvider(base, "deepseek/deepseek-r1")
}

func TestNewReasoningProvider_Models(t *testing.T) {
	base := ai.NewMockProvider("ok")
	rp := ai.NewReasoningProvider(base, "deepseek/deepseek-r1")

	models := rp.Models()
	if len(models) == 0 {
		t.Error("Models() should return at least one model")
	}

	// Should expose the reasoning models available via OpenRouter.
	found := false
	for _, m := range models {
		if m.ID == "deepseek/deepseek-r1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Models() should include deepseek/deepseek-r1")
	}
}

func TestNewReasoningProvider_StreamComplete(t *testing.T) {
	base := ai.NewMockProvider("stream response")
	rp := ai.NewReasoningProvider(base, "")

	ch, err := rp.StreamComplete(context.Background(), ai.CompletionRequest{
		Messages: []ai.Message{{Role: "user", Content: "test"}},
	})
	if err != nil {
		t.Fatalf("StreamComplete() error = %v", err)
	}

	var got string
	for chunk := range ch {
		if chunk.Error != nil {
			t.Fatalf("stream error: %v", chunk.Error)
		}
		got += chunk.Content
	}
	if got == "" {
		t.Error("expected stream response, got empty string")
	}
}

func TestNewReasoningProviderFromEnv_FallsBackWithoutConfig(t *testing.T) {
	base := ai.NewMockProvider("fallback")

	// Without env vars set, should use base provider.
	rp := ai.NewReasoningProviderFromEnv(base)

	resp, err := rp.Complete(context.Background(), ai.CompletionRequest{
		Messages: []ai.Message{{Role: "user", Content: "test"}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "fallback" {
		t.Errorf("expected fallback response, got %q", resp.Content)
	}
}

func TestNewReasoningProviderFromEnv_UnknownModelFallsBackToDefault(t *testing.T) {
	t.Setenv("OSS_AI_REASONING_MODEL", "not-a-real-model/v999")
	// No API key set, so it falls back to base regardless — but model validation
	// must not panic and must reset the model to the default.
	base := ai.NewMockProvider("fallback")
	rp := ai.NewReasoningProviderFromEnv(base)

	// Without an API key the base provider is used; just verify it doesn't panic
	// and still works.
	resp, err := rp.Complete(context.Background(), ai.CompletionRequest{
		Messages: []ai.Message{{Role: "user", Content: "test"}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "fallback" {
		t.Errorf("expected fallback response, got %q", resp.Content)
	}
}

func TestSupportedReasoningModels(t *testing.T) {
	models := ai.SupportedReasoningModels()
	expected := []string{
		"deepseek/deepseek-r1",
		"moonshotai/kimi-k2.5",
		"qwen/qwen3.5",
		"openai/o3-mini",
	}

	for _, want := range expected {
		found := false
		for _, m := range models {
			if m.ID == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedReasoningModels() missing %q", want)
		}
	}
}
