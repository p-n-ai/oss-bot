package pipeline_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

func TestPipeline_Preview(t *testing.T) {
	repoDir := setupPipelineTestRepo(t)
	mock := ai.NewMockProvider("# Teaching Notes\n\nGenerated content here.")

	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)

	result, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModePreview,
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.StructuredOutput == "" {
		t.Error("StructuredOutput should not be empty")
	}
}

func TestPipeline_WriteFS(t *testing.T) {
	repoDir := setupPipelineTestRepo(t)
	mock := ai.NewMockProvider("# Teaching Notes\n\nGenerated content here.")

	outputDir := t.TempDir()
	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)

	_, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModeWriteFS,
		OutputDir:        outputDir,
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestPipeline_UnknownType(t *testing.T) {
	repoDir := setupPipelineTestRepo(t)
	mock := ai.NewMockProvider("content")

	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)

	_, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "unknown_type",
		Mode:             pipeline.ModePreview,
		Source:           "cli",
	})
	if err == nil {
		t.Error("Execute() should error for unknown contribution type")
	}
}

func setupPipelineTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	topicsDir := filepath.Join(dir, "curricula", "test", "topics", "algebra")
	os.MkdirAll(topicsDir, 0o755)

	os.WriteFile(filepath.Join(topicsDir, "01-test.yaml"), []byte(`
id: F1-01
name: "Test Topic"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: LO1
    text: "Test objective"
    bloom: understand
prerequisites:
  required: []
quality_level: 1
provenance: human
`), 0o644)

	return dir
}
