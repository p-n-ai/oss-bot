package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// ContributionInput holds a teacher's natural language input and the context needed
// to convert it into structured curriculum content.
type ContributionInput struct {
	TeacherInput string // Free-form text from the teacher
	TopicPath    string // Repo-relative topic path or topic ID
	ContentType  string // "misconception", "teaching_note", or "example"
	SyllabusID   string // Optional syllabus identifier for context
}

// ParseContribution converts a teacher's natural language input into a structured
// YAML entry using the contribution_parser prompt template and the given AI provider.
// promptsDir is the directory containing contribution_parser.md.
func ParseContribution(ctx context.Context, provider ai.Provider, input ContributionInput, promptsDir string) (string, error) {
	if input.TeacherInput == "" {
		return "", fmt.Errorf("teacher input must not be empty")
	}
	if input.TopicPath == "" {
		return "", fmt.Errorf("topic path must not be empty")
	}
	if input.ContentType == "" {
		return "", fmt.Errorf("content type must not be empty")
	}
	if provider == nil {
		return "", fmt.Errorf("AI provider must not be nil")
	}

	prompt, err := buildContributionPrompt(input, promptsDir)
	if err != nil {
		return "", fmt.Errorf("building prompt: %w", err)
	}

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are an expert curriculum designer converting teacher observations into structured YAML."},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 1024,
	})
	if err != nil {
		return "", fmt.Errorf("AI completion failed: %w", err)
	}

	return strings.TrimSpace(resp.Content), nil
}

// buildContributionPrompt loads the prompt template and substitutes input values.
func buildContributionPrompt(input ContributionInput, promptsDir string) (string, error) {
	templatePath := filepath.Join(promptsDir, "contribution_parser.md")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		// Fall back to an inline minimal prompt if the template file is missing.
		return buildFallbackPrompt(input), nil
	}

	prompt := string(data)
	prompt = strings.ReplaceAll(prompt, "{{topic}}", input.TopicPath)
	prompt = strings.ReplaceAll(prompt, "{{content_type}}", input.ContentType)
	prompt = strings.ReplaceAll(prompt, "{{syllabus_id}}", input.SyllabusID)
	prompt = strings.ReplaceAll(prompt, "{{teacher_input}}", input.TeacherInput)

	return prompt, nil
}

// buildFallbackPrompt returns a minimal prompt when the template file is unavailable.
func buildFallbackPrompt(input ContributionInput) string {
	return fmt.Sprintf(`Convert the following teacher observation into a structured YAML entry of type %q for topic %q.
Preserve the teacher's voice. Output only the YAML block.

Teacher's input:
%s`, input.ContentType, input.TopicPath, input.TeacherInput)
}
