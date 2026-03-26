package pipeline_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

// mockContentReader is a ContentReader that always returns the configured content.
type mockContentReader struct {
	content map[string][]byte
	err     error
}

func (m *mockContentReader) ReadFile(path, _ string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if data, ok := m.content[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("not found: %s", path)
}

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

func TestPipeline_MergeWithExisting(t *testing.T) {
	repoDir := setupPipelineTestRepoWithNotes(t)

	existingNotes := "## Overview\nExisting overview.\n\n## Teaching Sequence & Strategy\nExisting strategy.\n"
	topicRelPath := filepath.Join("curricula", "test", "topics", "algebra", "F1-01.teaching.md")

	reader := &mockContentReader{
		content: map[string][]byte{
			topicRelPath: []byte(existingNotes),
		},
	}

	// AI returns a new section not in existing
	mock := ai.NewMockProvider("## Overview\nNew.\n\n## High Alert Misconceptions\nNew section.\n")
	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir).WithContentReader(reader)

	result, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModePreview,
		Source:           "bot",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.MergeReport == nil {
		t.Fatal("MergeReport should be non-nil when existing content was found")
	}
	if result.MergeReport.Added < 1 {
		t.Errorf("expected at least 1 added section, got %d", result.MergeReport.Added)
	}
}

func TestPipeline_MergeNoExistingFile(t *testing.T) {
	repoDir := setupPipelineTestRepoWithNotes(t)

	// Reader returns error for all paths → no merge, no error
	reader := &mockContentReader{err: fmt.Errorf("not found")}
	mock := ai.NewMockProvider("## Overview\nFresh content.\n")
	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir).WithContentReader(reader)

	result, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModePreview,
		Source:           "bot",
	})
	if err != nil {
		t.Fatalf("Execute() should not error when existing file is missing: %v", err)
	}
	if result.MergeReport != nil {
		t.Error("MergeReport should be nil when no existing file was found")
	}
}

func setupPipelineTestRepoWithNotes(t *testing.T) string {
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
ai_teaching_notes: F1-01.teaching.md
`), 0o644)

	return dir
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
