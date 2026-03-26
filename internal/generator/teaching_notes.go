package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// GenerationResult holds the output of a content generation.
type GenerationResult struct {
	Content      string
	Files        map[string]string
	Model        string
	InputTokens  int
	OutputTokens int
}

// GenerateTeachingNotes generates teaching notes for a topic using AI.
func GenerateTeachingNotes(ctx context.Context, provider ai.Provider, genCtx *GenerationContext) (*GenerationResult, error) {
	prompt := BuildTeachingNotesPrompt(genCtx)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: fmt.Sprintf("You are an expert educator creating teaching notes for the %s curriculum. Follow the output structure exactly.", genCtx.Topic.SyllabusID)},
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

// BuildTeachingNotesPrompt constructs the prompt for teaching notes generation.
func BuildTeachingNotesPrompt(genCtx *GenerationContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate comprehensive teaching notes for the topic: %s (%s)\n\n", genCtx.Topic.Name, genCtx.Topic.ID))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", genCtx.Topic.SubjectID))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n", genCtx.Topic.SyllabusID))
	sb.WriteString(fmt.Sprintf("Difficulty: %s\n", genCtx.Topic.Difficulty))
	if genCtx.Topic.Tier != "" {
		sb.WriteString(fmt.Sprintf("Tier: %s\n", genCtx.Topic.Tier))
	}
	sb.WriteString("\n")

	sb.WriteString("## Learning Objectives\n")
	for _, lo := range genCtx.Topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", lo.ID, lo.Bloom, lo.Text))
	}
	sb.WriteString("\n")

	if len(genCtx.Prerequisites) > 0 {
		sb.WriteString("## Prerequisites (students have already learned)\n")
		for _, p := range genCtx.Prerequisites {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", p.ID, p.Name))
		}
		sb.WriteString("\n")
	}

	if len(genCtx.ValidationFeedback) > 0 {
		sb.WriteString("## Previous Attempt Feedback (fix these issues)\n")
		for _, e := range genCtx.ValidationFeedback {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("\n")
	}

	if genCtx.ExistingNotes != "" {
		sb.WriteString("## Existing Notes (match this style)\n")
		sb.WriteString(genCtx.ExistingNotes)
		sb.WriteString("\n\n")
	}

	sb.WriteString(`## Output Format
Write in Markdown following this exact structure:

# [Topic Name] — Teaching Notes

## Overview
[Brief description]

> [!IMPORTANT]
> **Chatbot Delivery Rules:**
> - Bite-sized pacing, max 2 short paragraphs per message
> - Casual, encouraging tone

## Curriculum Standards & Taxonomy
[Map learning objectives to official curriculum standard codes and performance levels]

## Prerequisites Check
[What students should know before starting]

## Teaching Sequence & Strategy

### 1. [Section Title] (XX min)
[Teaching instructions]
- **Strategies:** [Specific pedagogy, manipulatives, visual aids]
- **Check for Understanding (CFU):** [Question to ask, then wait]
- **The Trap:** [Common mistake in this section]

## High Alert Misconceptions
| Misconception | Why Students Think This | How to Fix |
|---------------|-------------------------|------------|

## Engagement Hooks
- [Real-world connections using locally relevant contexts]

## Assessment Guidance
[Tips for assessing understanding]

## Bilingual Key Terms
| English | Local Term |
|---------|------------|
`)

	return sb.String()
}
