package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestGenerateTeachingNotes(t *testing.T) {
	mockResponse := `# Test Topic — Teaching Notes

## Overview
This topic covers test content.

## Prerequisites Check
- Basic arithmetic

## Teaching Sequence & Strategy

### 1. Introduction (15 min)
Start with examples.
- **Strategies:** Use visual aids
- **Check for Understanding (CFU):** Ask a question
- **The Trap:** Common mistake here

## High Alert Misconceptions

| Misconception | Why Students Think This | How to Fix |
|---------------|-------------------------|------------|
| Error | Reason | Fix |

## Engagement Hooks
- Real world example

## Bilingual Key Terms
| English | Bahasa Melayu |
|---------|---------------|
| Variable | Pemboleh ubah |
`

	mock := ai.NewMockProvider(mockResponse)

	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SubjectID:  "algebra",
			SyllabusID: "test-syllabus",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "LO1", Text: "Test objective", Bloom: "understand"},
			},
		},
	}

	result, err := generator.GenerateTeachingNotes(context.Background(), mock, genCtx)
	if err != nil {
		t.Fatalf("GenerateTeachingNotes() error = %v", err)
	}

	if !strings.Contains(result.Content, "Teaching Notes") {
		t.Error("Result should contain 'Teaching Notes'")
	}
	if result.Model == "" {
		t.Error("Result.Model should not be empty")
	}
}

func TestBuildTeachingNotesPrompt(t *testing.T) {
	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SubjectID:  "algebra",
			SyllabusID: "test-syllabus",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "LO1", Text: "Test", Bloom: "understand"},
			},
		},
	}

	prompt := generator.BuildTeachingNotesPrompt(genCtx)
	if prompt == "" {
		t.Error("BuildTeachingNotesPrompt() returned empty string")
	}
	if !strings.Contains(prompt, "F1-01") {
		t.Error("Prompt should contain topic ID")
	}
	if !strings.Contains(prompt, "test-syllabus") {
		t.Error("Prompt should contain syllabus ID")
	}
}
