package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestGenerateAssessments(t *testing.T) {
	mockYAML := `topic_id: F1-01
provenance: ai-generated

questions:
  - id: Q1
    text: "If x = 3, find 2x + 5"
    difficulty: easy
    learning_objective: 1.0.1
    tp_level: 3
    kbat: false
    answer:
      type: exact
      value: "11"
      working: "2(3) + 5 = 11"
    marks: 2
    rubric:
      - marks: 1
        criteria: "Correct substitution"
      - marks: 1
        criteria: "Correct answer"
    hints:
      - level: 1
        text: "Replace x with 3"
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

	result, err := generator.GenerateAssessments(context.Background(), mock, genCtx, 5, "medium")
	if err != nil {
		t.Fatalf("GenerateAssessments() error = %v", err)
	}

	if !strings.Contains(result.Content, "topic_id") {
		t.Error("Result should contain YAML with topic_id")
	}
}

func TestBuildAssessmentsPrompt(t *testing.T) {
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

	prompt := generator.BuildAssessmentsPrompt(genCtx, 5, "medium")
	if !strings.Contains(prompt, "5") {
		t.Error("Prompt should contain question count")
	}
	if !strings.Contains(prompt, "tp_level") {
		t.Error("Prompt should mention tp_level field")
	}
}

func TestBuildAssessmentsPrompt_WithSchemaRules(t *testing.T) {
	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SyllabusID: "test-syllabus",
			LearningObjectives: []generator.LearningObjective{
				{ID: "1.0.1", Text: "Test", Bloom: "apply"},
			},
		},
		SchemaRules: `{"type":"object","required":["topic_id","questions"]}`,
	}

	prompt := generator.BuildAssessmentsPrompt(genCtx, 3, "easy")
	if !strings.Contains(prompt, "JSON Schema") {
		t.Error("Prompt should contain JSON Schema section when SchemaRules is set")
	}
	if !strings.Contains(prompt, `"required":["topic_id","questions"]`) {
		t.Error("Prompt should contain the schema content")
	}
}

func TestBuildAssessmentsPrompt_WithoutSchemaRules(t *testing.T) {
	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SyllabusID: "test-syllabus",
			LearningObjectives: []generator.LearningObjective{
				{ID: "1.0.1", Text: "Test", Bloom: "apply"},
			},
		},
	}

	prompt := generator.BuildAssessmentsPrompt(genCtx, 3, "easy")
	if strings.Contains(prompt, "JSON Schema") {
		t.Error("Prompt should NOT contain JSON Schema section when SchemaRules is empty")
	}
}
