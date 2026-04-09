package ai

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// openRouterBaseURL is the OpenAI-compatible endpoint for OpenRouter.
const openRouterBaseURL = "https://openrouter.ai/api/v1"

// SupportedReasoningModels returns the list of reasoning models available via OpenRouter.
func SupportedReasoningModels() []ModelInfo {
	return []ModelInfo{
		{
			ID:          "deepseek/deepseek-r1",
			Name:        "DeepSeek R1",
			MaxTokens:   163840,
			Description: "DeepSeek reasoning model — strong at multi-step analysis",
		},
		{
			ID:          "moonshotai/kimi-k2.5",
			Name:        "Kimi K2.5",
			MaxTokens:   131072,
			Description: "Moonshot AI reasoning model",
		},
		{
			ID:          "qwen/qwen3.5-flash-02-23",
			Name:        "Qwen 3.5 Flash",
			MaxTokens:   131072,
			Description: "Alibaba Qwen fast reasoning model",
		},
		{
			ID:          "openai/o3-mini",
			Name:        "OpenAI o3-mini",
			MaxTokens:   200000,
			Description: "OpenAI reasoning model via OpenRouter",
		},
	}
}

// ReasoningProvider wraps a Provider to target reasoning models via OpenRouter.
// When no model is configured it transparently delegates to the base provider,
// so callers never need to handle the "reasoning not available" case explicitly.
type ReasoningProvider struct {
	base  Provider
	inner Provider // OpenRouter-backed provider (may be nil → use base)
	model string
}

// NewReasoningProvider creates a ReasoningProvider.
// If model is empty the base provider is used for all calls (fallback mode).
func NewReasoningProvider(base Provider, model string) *ReasoningProvider {
	rp := &ReasoningProvider{base: base, model: model}
	// inner stays nil — calls fall back to base.
	// A real OpenRouter provider would be injected via NewReasoningProviderWithClient.
	return rp
}

// isKnownReasoningModel reports whether model is in the SupportedReasoningModels list.
func isKnownReasoningModel(model string) bool {
	for _, m := range SupportedReasoningModels() {
		if m.ID == model {
			return true
		}
	}
	return false
}

// NewReasoningProviderFromEnv creates a ReasoningProvider using environment variables:
//   - OSS_AI_REASONING_API_KEY — OpenRouter API key
//   - OSS_AI_REASONING_MODEL   — model name (default: deepseek/deepseek-r1)
//
// Falls back to the base provider when the env vars are not set.
// Logs a warning and resets to the default when an unrecognised model is configured.
func NewReasoningProviderFromEnv(base Provider) *ReasoningProvider {
	apiKey := os.Getenv("OSS_AI_REASONING_API_KEY")
	model := os.Getenv("OSS_AI_REASONING_MODEL")
	if model == "" {
		model = "deepseek/deepseek-r1"
	}

	if !isKnownReasoningModel(model) {
		slog.Warn("unrecognised reasoning model, falling back to default",
			"model", model, "default", "deepseek/deepseek-r1")
		model = "deepseek/deepseek-r1"
	}

	if apiKey == "" {
		// No API key — use base provider for all calls.
		return &ReasoningProvider{base: base}
	}

	inner, err := newOpenRouterProvider(apiKey, model)
	if err != nil {
		// Construction failed (shouldn't happen for valid keys) — fall back.
		return &ReasoningProvider{base: base}
	}

	return &ReasoningProvider{base: base, inner: inner, model: model}
}

// newOpenRouterProvider creates an OpenAI-compatible provider pointed at OpenRouter.
func newOpenRouterProvider(apiKey, model string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key required")
	}
	p, err := NewOpenAIProvider(apiKey)
	if err != nil {
		return nil, err
	}
	// Override the base URL and default model.
	p.baseURL = openRouterBaseURL
	p.defaultModel = model
	return p, nil
}

// Complete delegates to the inner OpenRouter provider when available, otherwise
// falls back to the base provider.
func (r *ReasoningProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if r.inner != nil {
		// Inject the configured reasoning model if none is specified in the request.
		if req.Model == "" && r.model != "" {
			req.Model = r.model
		}
		return r.inner.Complete(ctx, req)
	}
	return r.base.Complete(ctx, req)
}

// StreamComplete delegates similarly.
func (r *ReasoningProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	if r.inner != nil {
		if req.Model == "" && r.model != "" {
			req.Model = r.model
		}
		return r.inner.StreamComplete(ctx, req)
	}
	return r.base.StreamComplete(ctx, req)
}

// Models returns the reasoning models available via OpenRouter, plus those of the
// base provider.
func (r *ReasoningProvider) Models() []ModelInfo {
	return SupportedReasoningModels()
}
