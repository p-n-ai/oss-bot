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
	topicRelPath := filepath.Join("curricula", "test", "test-algebra", "test-algebra-1", "topics", "F1-01.teaching.md")

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

func TestPipeline_FilesMapPopulated(t *testing.T) {
	repoDir := setupPipelineTestRepoWithNotes(t) // topic has ai_teaching_notes set
	mock := ai.NewMockProvider("# Teaching Notes\n\nGenerated content.")

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
	if len(result.Files) == 0 {
		t.Error("Files map should be populated after generation when topic has ai_teaching_notes set")
	}
	for path, content := range result.Files {
		if path == "" {
			t.Error("file path must not be empty")
		}
		if content == "" {
			t.Error("file content must not be empty")
		}
	}
}

func TestPipeline_BloomValidationErrors_SurfacedInResult(t *testing.T) {
	repoDir := setupPipelineTestRepoInvalidBloom(t)
	mock := ai.NewMockProvider("# Teaching Notes\n\nContent.")

	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)
	result, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-BAD",
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModePreview,
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(result.ValidationErrors) == 0 {
		t.Error("expected ValidationErrors for topic with invalid Bloom level")
	}
}

func TestPipeline_BloomValidation_NoErrorsForValidTopic(t *testing.T) {
	repoDir := setupPipelineTestRepo(t) // topic has bloom: understand (valid)
	mock := ai.NewMockProvider("# Teaching Notes\n\nContent.")

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
	if len(result.ValidationErrors) != 0 {
		t.Errorf("expected no ValidationErrors for valid topic, got: %v", result.ValidationErrors)
	}
}

func setupPipelineTestRepoInvalidBloom(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	topicsDir := filepath.Join(dir, "curricula", "test", "test-algebra", "test-algebra-1", "topics")
	os.MkdirAll(topicsDir, 0o755)
	os.WriteFile(filepath.Join(topicsDir, "bad.yaml"), []byte(`
id: F1-BAD
name: "Bad Bloom Topic"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: 1.0.1
    text: "An objective"
    bloom: think_hard_about_it
prerequisites:
  required: []
quality_level: 1
provenance: human
`), 0o644)
	return dir
}

func setupPipelineTestRepoWithNotes(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	topicsDir := filepath.Join(dir, "curricula", "test", "test-algebra", "test-algebra-1", "topics")
	os.MkdirAll(topicsDir, 0o755)

	os.WriteFile(filepath.Join(topicsDir, "01-test.yaml"), []byte(`
id: F1-01
name: "Test Topic"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: 1.0.1
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

func TestStripThinkTags(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no think tags",
			in:   "id: MT5-03\nname: Test",
			want: "id: MT5-03\nname: Test",
		},
		{
			name: "full think block before content",
			in:   "<think>\nsome reasoning\n</think>\nid: MT5-03",
			want: "id: MT5-03",
		},
		{
			name: "orphaned closing tag only",
			in:   "</think>\nid: MT5-03",
			want: "id: MT5-03",
		},
		{
			name: "think block with content after",
			in:   "<think>deep thought about curriculum</think>\n\nid: MT5-03\nname: Test",
			want: "id: MT5-03\nname: Test",
		},
		{
			name: "opening tag without closing",
			in:   "<think>\nreasoning forever...",
			want: "",
		},
		{
			name: "multiple think blocks",
			in:   "<think>first</think>id: MT5-03<think>second</think>\nname: Test",
			want: "id: MT5-03\nname: Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pipeline.StripThinkTags(tt.in)
			if got != tt.want {
				t.Errorf("StripThinkTags()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestStripCodeFences_WithThinkTags(t *testing.T) {
	// Verify StripCodeFences also strips think tags (integration)
	in := "</think>\n```yaml\nid: MT5-03\n```"
	got := pipeline.StripCodeFences(in)
	if got != "id: MT5-03" {
		t.Errorf("StripCodeFences with think tags:\ngot:  %q\nwant: %q", got, "id: MT5-03")
	}
}

func TestSchemaTypeForContribution(t *testing.T) {
	tests := []struct {
		contribType string
		want        string
	}{
		{"assessments", "assessments"},
		{"examples", "examples"},
		{"topic_enrich", "topic"},
		{"teaching_notes", ""},
		{"unknown", ""},
	}
	for _, tt := range tests {
		got := pipeline.SchemaTypeForContribution(tt.contribType)
		if got != tt.want {
			t.Errorf("SchemaTypeForContribution(%q) = %q, want %q", tt.contribType, got, tt.want)
		}
	}
}

func TestPipeline_SchemaInjectedIntoPrompt(t *testing.T) {
	repoDir := setupPipelineTestRepoWithSchema(t)

	// Track what the AI receives
	var capturedPrompt string
	mock := &promptCapturingMock{
		response: validAssessmentsResponse,
		capturedPrompt: &capturedPrompt,
	}

	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)

	_, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "assessments",
		Mode:             pipeline.ModePreview,
		Options:          map[string]string{"count": "1", "difficulty": "easy"},
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if capturedPrompt == "" {
		t.Fatal("prompt was not captured")
	}
	if !containsStr(capturedPrompt, "JSON Schema") {
		t.Error("prompt should contain 'JSON Schema' section when schema exists")
	}
	if !containsStr(capturedPrompt, "topic_id") {
		t.Error("prompt should contain schema content")
	}
}

func TestPipeline_SchemaValidation_RetryOnFailure(t *testing.T) {
	repoDir := setupPipelineTestRepoWithSchema(t)

	// First call returns invalid YAML (missing required "marks" per schema), second returns valid
	callCount := 0
	mock := &sequentialMock{
		responses: []string{
			invalidAssessmentsResponse, // first: missing marks
			validAssessmentsResponse,   // retry: valid
		},
		callCount: &callCount,
	}

	p := pipeline.New(mock, &output.LocalWriter{}, "prompts/", repoDir)

	_, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        "F1-01",
		ContributionType: "assessments",
		Mode:             pipeline.ModePreview,
		Options:          map[string]string{"count": "1", "difficulty": "easy"},
		Source:           "cli",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if callCount < 2 {
		t.Errorf("expected at least 2 AI calls (initial + retry), got %d", callCount)
	}
}

// --- Schema test helpers ---

const validAssessmentsResponse = `topic_id: F1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 1+1?"
    difficulty: easy
    learning_objective: "1.0.1"
    answer:
      type: exact
      value: "2"
    marks: 1
`

const invalidAssessmentsResponse = `topic_id: F1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 1+1?"
    difficulty: easy
    learning_objective: "1.0.1"
    answer:
      type: exact
      value: "2"
`

// promptCapturingMock captures the user prompt from the AI call.
type promptCapturingMock struct {
	response       string
	capturedPrompt *string
}

func (m *promptCapturingMock) Complete(_ context.Context, req ai.CompletionRequest) (ai.CompletionResponse, error) {
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			*m.capturedPrompt = msg.Content
		}
	}
	return ai.CompletionResponse{Content: m.response, Model: "mock"}, nil
}

func (m *promptCapturingMock) StreamComplete(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *promptCapturingMock) Models() []ai.ModelInfo {
	return []ai.ModelInfo{{ID: "mock"}}
}

// sequentialMock returns different responses on consecutive calls.
type sequentialMock struct {
	responses []string
	callCount *int
}

func (m *sequentialMock) Complete(_ context.Context, _ ai.CompletionRequest) (ai.CompletionResponse, error) {
	idx := *m.callCount
	*m.callCount++
	if idx < len(m.responses) {
		return ai.CompletionResponse{Content: m.responses[idx], Model: "mock"}, nil
	}
	return ai.CompletionResponse{Content: m.responses[len(m.responses)-1], Model: "mock"}, nil
}

func (m *sequentialMock) StreamComplete(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *sequentialMock) Models() []ai.ModelInfo {
	return []ai.ModelInfo{{ID: "mock"}}
}

// setupPipelineTestRepoWithSchema creates a test repo with a subject-level schema
// that requires "marks" on assessments questions.
func setupPipelineTestRepoWithSchema(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Subject directory with subject.yaml
	subjectDir := filepath.Join(dir, "curricula", "test", "test-algebra")
	os.MkdirAll(subjectDir, 0o755)
	os.WriteFile(filepath.Join(subjectDir, "subject.yaml"), []byte("id: test-algebra\nname: Algebra\nsyllabus_id: test\ntopics: []\n"), 0o644)

	// Subject-level schema (assessments requires marks)
	schemasDir := filepath.Join(subjectDir, "schemas")
	os.MkdirAll(schemasDir, 0o755)
	assessmentsSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["topic_id", "questions"],
		"properties": {
			"topic_id": { "type": "string" },
			"provenance": { "type": "string" },
			"questions": {
				"type": "array", "minItems": 1,
				"items": {
					"type": "object",
					"required": ["id", "text", "difficulty", "learning_objective", "answer", "marks"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string" },
						"difficulty": { "type": "string", "enum": ["easy", "medium", "hard"] },
						"learning_objective": { "type": "string" },
						"answer": {
							"type": "object",
							"required": ["type", "value"],
							"properties": {
								"type": { "type": "string" },
								"value": { "type": "string" }
							},
							"additionalProperties": false
						},
						"marks": { "type": "integer", "minimum": 1 }
					},
					"additionalProperties": false
				}
			}
		},
		"additionalProperties": false
	}`
	os.WriteFile(filepath.Join(schemasDir, "assessments.schema.json"), []byte(assessmentsSchema), 0o644)

	// Topics
	topicsDir := filepath.Join(subjectDir, "test-algebra-1", "topics")
	os.MkdirAll(topicsDir, 0o755)
	os.WriteFile(filepath.Join(topicsDir, "01-test.yaml"), []byte(`
id: F1-01
name: "Test Topic"
subject_id: test-algebra
syllabus_id: test
difficulty: beginner
learning_objectives:
  - id: 1.0.1
    text: "Test objective"
    bloom: understand
prerequisites:
  required: []
quality_level: 1
provenance: human
`), 0o644)

	return dir
}

func setupPipelineTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	topicsDir := filepath.Join(dir, "curricula", "test", "test-algebra", "test-algebra-1", "topics")
	os.MkdirAll(topicsDir, 0o755)

	os.WriteFile(filepath.Join(topicsDir, "01-test.yaml"), []byte(`
id: F1-01
name: "Test Topic"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: 1.0.1
    text: "Test objective"
    bloom: understand
prerequisites:
  required: []
quality_level: 1
provenance: human
`), 0o644)

	return dir
}
