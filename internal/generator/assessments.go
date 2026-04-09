package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// GenerateAssessments generates assessment questions for a topic using AI.
func GenerateAssessments(ctx context.Context, provider ai.Provider, genCtx *GenerationContext, count int, difficulty string) (*GenerationResult, error) {
	prompt := BuildAssessmentsPrompt(genCtx, count, difficulty)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: fmt.Sprintf("You are an expert educator creating assessment questions for the %s curriculum. Output valid YAML only.", genCtx.Topic.SyllabusID)},
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

// BuildAssessmentsPrompt constructs the prompt for assessment generation.
func BuildAssessmentsPrompt(genCtx *GenerationContext, count int, difficulty string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate %d assessment questions for: %s (%s)\n\n", count, genCtx.Topic.Name, genCtx.Topic.ID))
	sb.WriteString(fmt.Sprintf("Target difficulty: %s\n", difficulty))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", genCtx.Topic.SubjectID))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n\n", genCtx.Topic.SyllabusID))

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

	sb.WriteString(fmt.Sprintf(`## Requirements
- Generate exactly %d questions
- Each question must include: worked solution, rubric with partial marks, progressive hints
- Distribute across learning objectives
- Follow the exam format and conventions of the %s curriculum
- Support LaTeX via $...$ notation. IMPORTANT: use single quotes (not double quotes) for any YAML string containing backslashes or LaTeX, e.g. text: 'Solve $\sqrt{x}$' — double-quoted strings process escape sequences (\t → tab, \n → newline) which corrupts LaTeX
- Include tp_level (performance/mastery level from the syllabus scale)
- Set kbat: true for questions at analyze/evaluate/create Bloom's levels
- Use a mix of answer types: exact, multiple_choice, free_text
- Hints should address specific misconceptions (prefix with "MISCONCEPTION ALERT:" where relevant)
- Group questions by learning objective with YAML comments

## Output Format
Output ONLY valid YAML (no markdown code fences):

topic_id: %s
provenance: ai-generated

questions:
  # Group by learning objective
  - id: Q1
    text: "..."
    difficulty: easy|medium|hard
    learning_objective: 1.0.1
    tp_level: 2
    kbat: false
    answer:
      type: exact|multiple_choice|free_text
      value: "..."
      working: |
        Step by step solution
    marks: N
    rubric:
      - marks: 1
        criteria: "..."
    hints:
      - level: 1
        text: "..."
`, count, genCtx.Topic.SyllabusID, genCtx.Topic.ID))

	return sb.String()
}
