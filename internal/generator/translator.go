package generator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"gopkg.in/yaml.v3"
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

// WriteTranslationToTopic reads a topic YAML file and adds or updates a
// translation entry under the "translations" mapping for the given language code.
func WriteTranslationToTopic(topicFilePath, langCode, translationContent string) error {
	data, err := os.ReadFile(topicFilePath)
	if err != nil {
		return fmt.Errorf("reading topic file: %w", err)
	}

	var raw yaml.Node
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing topic YAML: %w", err)
	}

	if raw.Kind != yaml.DocumentNode || len(raw.Content) == 0 {
		return fmt.Errorf("unexpected YAML structure")
	}
	mapping := raw.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %d", mapping.Kind)
	}

	// Find or create the "translations" mapping
	var translationsNode *yaml.Node
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == "translations" {
			translationsNode = mapping.Content[i+1]
			break
		}
	}

	if translationsNode == nil || translationsNode.Kind != yaml.MappingNode {
		// Create a new translations mapping node
		translationsNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		setMappingKey(mapping, "translations", translationsNode)
	}

	// Build a scalar node for the translation content
	contentNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: translationContent,
		Style: yaml.LiteralStyle, // Use | block scalar for readability
		Tag:   "!!str",
	}

	// Set or replace the language entry within translations
	setMappingKey(translationsNode, langCode, contentNode)

	out, err := yaml.Marshal(&raw)
	if err != nil {
		return fmt.Errorf("marshaling updated YAML: %w", err)
	}

	if err := os.WriteFile(topicFilePath, out, 0o644); err != nil {
		return fmt.Errorf("writing topic file: %w", err)
	}

	return nil
}
