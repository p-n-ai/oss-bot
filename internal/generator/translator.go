package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// LanguageNames maps language codes to full names.
var LanguageNames = map[string]string{
	"ms": "Bahasa Melayu",
	"zh": "Chinese (Simplified)",
	"ta": "Tamil",
	"en": "English",
}

// Translate translates a topic's content to the target language.
func Translate(ctx context.Context, provider ai.Provider, topic *Topic, targetLang string) (*GenerationResult, error) {
	langName, ok := LanguageNames[targetLang]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s (supported: %v)", targetLang, supportedLanguages())
	}

	prompt := BuildTranslationPrompt(topic, langName)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are a professional translator specializing in education content. Translate accurately while preserving YAML structure. Output ONLY valid YAML."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.3, // Lower temperature for translation accuracy
	})
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// BuildTranslationPrompt constructs the prompt for translation.
func BuildTranslationPrompt(topic *Topic, targetLang string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Translate the following topic content to %s.\n\n", targetLang))
	sb.WriteString(fmt.Sprintf("Topic: %s (%s)\n", topic.Name, topic.ID))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n\n", topic.SyllabusID))

	sb.WriteString("## Content to translate\n\n")
	sb.WriteString(fmt.Sprintf("name: %q\n\n", topic.Name))
	sb.WriteString("learning_objectives:\n")
	for _, lo := range topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("  - id: %s\n    text: %q\n    bloom: %s\n", lo.ID, lo.Text, lo.Bloom))
	}

	sb.WriteString("\n## Rules\n")
	sb.WriteString("- Only translate human-readable text (name, text, description fields)\n")
	sb.WriteString("- Do NOT translate: id, bloom, difficulty, provenance, tp_level, kbat values\n")
	sb.WriteString("- Preserve LaTeX notation ($...$) unchanged\n")
	sb.WriteString("- Use mathematically correct terminology in the target language\n")
	sb.WriteString("- Maintain the same YAML indentation and structure\n")
	sb.WriteString("- Output ONLY the translated YAML\n")

	return sb.String()
}

func supportedLanguages() []string {
	langs := make([]string, 0, len(LanguageNames))
	for code := range LanguageNames {
		langs = append(langs, code)
	}
	return langs
}
