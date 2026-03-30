package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

// TranslateFile translates the content of a companion file (teaching notes,
// assessments, examples) to the target language.
func TranslateFile(ctx context.Context, provider ai.Provider, topicID, filePath, targetLang string) (*GenerationResult, error) {
	langName, ok := LanguageNames[targetLang]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s (supported: %v)", targetLang, supportedLanguages())
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	fileName := filepath.Base(filePath)
	fileType := "YAML"
	if strings.HasSuffix(fileName, ".md") {
		fileType = "Markdown"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Translate the following %s education content file to %s.\n\n", fileType, langName))
	sb.WriteString(fmt.Sprintf("Topic: %s\n", topicID))
	sb.WriteString(fmt.Sprintf("File: %s\n\n", fileName))
	sb.WriteString("## Content to translate\n\n")
	sb.WriteString(string(content))
	sb.WriteString("\n\n## Rules\n")
	sb.WriteString("- Only translate human-readable text (descriptions, explanations, questions, answers)\n")
	sb.WriteString("- Do NOT translate: id, bloom, difficulty, provenance, metadata fields, YAML keys\n")
	sb.WriteString("- Preserve LaTeX notation ($...$) unchanged\n")
	sb.WriteString("- Preserve Markdown formatting (headers, lists, code blocks) unchanged\n")
	sb.WriteString("- Use mathematically correct terminology in the target language\n")
	sb.WriteString("- Maintain the same structure and indentation\n")
	sb.WriteString(fmt.Sprintf("- Output ONLY the translated %s content\n", fileType))

	systemPrompt := "You are a professional translator specializing in education content. Translate accurately while preserving the file structure. Output ONLY the translated content."

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: sb.String()},
		},
		MaxTokens:   4096,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("translation of %s failed: %w", fileName, err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// WriteTranslationFile writes translated content to the translations/{lang}/ directory
// following the id-conventions.md spec. The topicsDir is the directory containing
// the topic files, and fileName is the output filename (e.g. "MT3-09.yaml",
// "MT3-09.teaching.md", "MT3-09.assessments.yaml").
func WriteTranslationFile(topicsDir, langCode, fileName, translationContent string) error {
	// Create translations/{lang}/ directory
	translationDir := filepath.Join(topicsDir, "translations", langCode)
	if err := os.MkdirAll(translationDir, 0o755); err != nil {
		return fmt.Errorf("creating translation directory: %w", err)
	}

	outPath := filepath.Join(translationDir, fileName)
	if err := os.WriteFile(outPath, []byte(translationContent), 0o644); err != nil {
		return fmt.Errorf("writing translation file: %w", err)
	}

	return nil
}
