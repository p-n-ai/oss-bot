package pipeline_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

// TestEndToEnd_FullPipeline tests the complete flow:
// build context → generate → write to filesystem
func TestEndToEnd_FullPipeline(t *testing.T) {
	repoDir := setupPipelineTestRepo(t)
	outputDir := t.TempDir()

	// Step 1: Verify context builds correctly
	genCtx, err := generator.BuildContext(repoDir, "F1-01")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}
	if genCtx.Topic.Name == "" {
		t.Fatal("Topic name should not be empty")
	}

	// Step 2: Generate teaching notes via pipeline
	mock := ai.NewMockProvider("# Test Topic — Teaching Notes\n\n## Overview\nGenerated content.\n\n## Teaching Sequence & Strategy\n\n### 1. Introduction (15 min)\nStart here.")
	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)

	result, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModeWriteFS,
		OutputDir:        outputDir,
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Pipeline.Execute(teaching_notes) error = %v", err)
	}
	if result.StructuredOutput == "" {
		t.Error("StructuredOutput should not be empty")
	}

	// Step 3: Generate assessments via pipeline
	mock2 := ai.NewMockProvider("topic_id: F1-01\nprovenance: ai-generated\nquestions:\n  - id: Q1\n    text: \"Test question\"\n    difficulty: easy\n    learning_objective: LO1\n    tp_level: 2\n    kbat: false\n    answer:\n      type: exact\n      value: \"42\"\n      working: \"The answer is 42\"\n    marks: 1\n    rubric:\n      - marks: 1\n        criteria: \"Correct answer\"\n    hints:\n      - level: 1\n        text: \"Think about it\"")
	p2 := pipeline.New(mock2, &output.LocalWriter{}, "prompts/", repoDir)

	result2, err := p2.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "assessments",
		Mode:             pipeline.ModeWriteFS,
		OutputDir:        outputDir,
		Options:          map[string]string{"count": "5", "difficulty": "medium"},
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Pipeline.Execute(assessments) error = %v", err)
	}
	if result2.StructuredOutput == "" {
		t.Error("Assessment output should not be empty")
	}

	// Step 4: Generate examples via pipeline
	mock3 := ai.NewMockProvider("topic_id: F1-01\nprovenance: ai-generated\nworked_examples:\n  - id: WE-01\n    topic: \"Test\"\n    difficulty: easy\n    real_world_analogy: \"Like counting blocks\"\n    misconception_alert: \"Students confuse X with Y\"\n    scenario: \"Find the value\"\n    working: \"Step 1: Do this\"")
	p3 := pipeline.New(mock3, &output.LocalWriter{}, "prompts/", repoDir)

	result3, err := p3.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "examples",
		Mode:             pipeline.ModeWriteFS,
		OutputDir:        outputDir,
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Pipeline.Execute(examples) error = %v", err)
	}
	if result3.StructuredOutput == "" {
		t.Error("Examples output should not be empty")
	}

	// Step 5: Translate
	topic := genCtx.Topic
	mockTranslator := ai.NewMockProvider("name: \"Topik Ujian\"\nlearning_objectives:\n  - id: LO1\n    text: \"Terjemahan objektif\"")
	translationResult, err := generator.Translate(context.Background(), mockTranslator, &topic, "ms")
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	// Write translation output
	transPath := filepath.Join(outputDir, "F1-01.ms.yaml")
	if err := os.WriteFile(transPath, []byte(translationResult.Content), 0o644); err != nil {
		t.Fatalf("Writing translation: %v", err)
	}
	if _, err := os.Stat(transPath); os.IsNotExist(err) {
		t.Error("Translation file should exist")
	}
}
