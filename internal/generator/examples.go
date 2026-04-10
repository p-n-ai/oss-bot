package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// GenerateExamples generates worked examples for a topic using AI.
func GenerateExamples(ctx context.Context, provider ai.Provider, genCtx *GenerationContext) (*GenerationResult, error) {
	prompt := BuildExamplesPrompt(genCtx)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: fmt.Sprintf("You are an expert educator creating worked examples for the %s curriculum. Output valid YAML only.", genCtx.Topic.SyllabusID)},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// BuildExamplesPrompt constructs the prompt for worked examples generation.
func BuildExamplesPrompt(genCtx *GenerationContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate worked examples for: %s (%s)\n\n", genCtx.Topic.Name, genCtx.Topic.ID))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", genCtx.Topic.SubjectID))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n", genCtx.Topic.SyllabusID))
	sb.WriteString(fmt.Sprintf("Difficulty: %s\n\n", genCtx.Topic.Difficulty))

	sb.WriteString("## Learning Objectives\n")
	for _, lo := range genCtx.Topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", lo.ID, lo.Bloom, lo.Text))
	}
	sb.WriteString("\n")

	if len(genCtx.ValidationFeedback) > 0 {
		sb.WriteString("## Previous Attempt Feedback (fix these issues)\n")
		for _, e := range genCtx.ValidationFeedback {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("\n")
	}

	if genCtx.SchemaRules != "" {
		sb.WriteString("## JSON Schema (your output MUST conform to this schema)\n")
		sb.WriteString("```json\n")
		sb.WriteString(genCtx.SchemaRules)
		sb.WriteString("\n```\n\n")
	}

	if genCtx.SchemaFieldGuide != "" {
		sb.WriteString(genCtx.SchemaFieldGuide)
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`## Requirements
- Generate 3 worked examples covering different difficulty levels (easy, medium, hard)
- Each example must include a real_world_analogy, misconception_alert, scenario, and step-by-step working
- Working must be broken into clearly numbered steps
- Use real-world scenarios relevant to the %s curriculum context
- Each example should target different learning objectives where possible
- IMPORTANT: use single quotes (not double quotes) for any YAML string containing backslashes or LaTeX, e.g. scenario: 'Find $\sqrt{x}$' — double-quoted strings process escape sequences (\t → tab, \n → newline) which corrupts LaTeX

## Output Format
Output ONLY valid YAML (no markdown code fences):

topic_id: %s
provenance: ai-generated
description: "Worked examples for %s"
worked_examples:
  - id: WE-01
    topic: "Section or subtopic name"
    difficulty: easy
    real_world_analogy: "A relatable analogy to explain the concept"
    misconception_alert: "Common mistake students make and why"
    scenario: "The problem statement in context"
    working: |
      Step 1: [First step with explanation]
      Step 2: [Next step]
      Step 3: [Final step with answer]
`, genCtx.Topic.SyllabusID, genCtx.Topic.ID, genCtx.Topic.Name))

	return sb.String()
}
