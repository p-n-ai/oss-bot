package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestGenerateExamples(t *testing.T) {
	mockYAML := `topic_id: F1-01
provenance: ai-generated
description: "Worked examples for Test Topic"
worked_examples:
  - id: WE-01
    topic: "Test Topic"
    difficulty: easy
    real_world_analogy: "Think of a bag of marbles."
    misconception_alert: "Students confuse variables with constants."
    scenario: "Form an expression for the total cost."
    working: |
      Step 1: Identify unknowns.
      Step 2: Build expression.
`

	mock := ai.NewMockProvider(mockYAML)

	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SyllabusID: "test-syllabus",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "1.0.1", Text: "Test", Bloom: "apply"},
			},
		},
	}

	result, err := generator.GenerateExamples(context.Background(), mock, genCtx)
	if err != nil {
		t.Fatalf("GenerateExamples() error = %v", err)
	}

	if !strings.Contains(result.Content, "worked_examples") {
		t.Error("Result should contain YAML with worked_examples")
	}
}

func TestBuildExamplesPrompt(t *testing.T) {
	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SyllabusID: "test-syllabus",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "1.0.1", Text: "Test", Bloom: "apply"},
			},
		},
	}

	prompt := generator.BuildExamplesPrompt(genCtx)
	if !strings.Contains(prompt, "worked_examples") {
		t.Error("Prompt should mention worked_examples format")
	}
	if !strings.Contains(prompt, "real_world_analogy") {
		t.Error("Prompt should mention real_world_analogy field")
	}
}
