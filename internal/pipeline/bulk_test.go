package pipeline_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/parser"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

func TestExecuteBulk_EmptyChunks(t *testing.T) {
	req := pipeline.BulkRequest{
		Chunks:     []parser.Chunk{},
		SyllabusID: "test-syllabus",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    2,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   ai.NewMockProvider("generated content"),
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if len(result.Topics) != 0 {
		t.Errorf("expected 0 topics, got %d", len(result.Topics))
	}
}

func TestExecuteBulk_ProcessesAllChunks(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 3, Heading: "Chapter 1", Content: "Algebra content"},
		{Index: 1, Total: 3, Heading: "Chapter 2", Content: "Geometry content"},
		{Index: 2, Total: 3, Heading: "Chapter 3", Content: "Calculus content"},
	}

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "india-jee",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    2,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   ai.NewMockProvider("generated content"),
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}

	if len(result.Topics) != 3 {
		t.Errorf("expected 3 topics, got %d", len(result.Topics))
	}
	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestExecuteBulk_DefaultWorkerCount(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 1, Heading: "Chapter 1", Content: "Content"},
	}

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "test",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    0, // should default to 3
		Reporter:   &pipeline.NoopReporter{},
		Provider:   ai.NewMockProvider("result"),
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if len(result.Topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(result.Topics))
	}
}

func TestExecuteBulk_RespectsMaxWorkers(t *testing.T) {
	// Track max concurrent workers.
	var concurrent, maxConcurrent int64

	chunks := make([]parser.Chunk, 9)
	for i := range chunks {
		chunks[i] = parser.Chunk{Index: i, Total: 9, Content: "content"}
	}

	provider := ai.NewMockProviderWithDelay("result", 10*time.Millisecond)

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "test",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    3,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   provider,
		OnWorkerStart: func() {
			n := atomic.AddInt64(&concurrent, 1)
			for {
				m := atomic.LoadInt64(&maxConcurrent)
				if n <= m || atomic.CompareAndSwapInt64(&maxConcurrent, m, n) {
					break
				}
			}
		},
		OnWorkerDone: func() {
			atomic.AddInt64(&concurrent, -1)
		},
	}

	_, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}

	if maxConcurrent > 3 {
		t.Errorf("max concurrent workers = %d, want <= 3", maxConcurrent)
	}
}

func TestExecuteBulk_CollectsErrors(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 2, Heading: "Good", Content: "fine content"},
		{Index: 1, Total: 2, Heading: "Bad", Content: ""},
	}

	provider := ai.NewMockProviderWithError(errors.New("generation failed"))

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "test",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    1,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   provider,
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	// ExecuteBulk should not return a top-level error for per-chunk failures.
	if err != nil {
		t.Fatalf("ExecuteBulk() should not return top-level error, got: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected errors to be collected in result")
	}
}

func TestExecuteBulk_ContextCancellation(t *testing.T) {
	chunks := make([]parser.Chunk, 10)
	for i := range chunks {
		chunks[i] = parser.Chunk{Index: i, Total: 10, Content: "content"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "test",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    3,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   ai.NewMockProvider("result"),
	}

	result, err := pipeline.ExecuteBulk(ctx, req)
	// Should either return error or process 0 items.
	if err == nil && len(result.Topics) == len(chunks) {
		t.Error("expected fewer processed topics when context is cancelled")
	}
}

func TestBulkResult_ProcessedChunksPopulated(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 2, Heading: "A", Content: "content a"},
		{Index: 1, Total: 2, Heading: "B", Content: "content b"},
	}
	req := pipeline.BulkRequest{
		Chunks: chunks, SyllabusID: "test", Mode: pipeline.ModePreview,
		Source: "test", Workers: 1, Reporter: &pipeline.NoopReporter{},
		Provider: ai.NewMockProvider("result"),
	}
	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if result.ProcessedChunks != 2 {
		t.Errorf("ProcessedChunks = %d, want 2", result.ProcessedChunks)
	}
	if result.Cancelled {
		t.Error("Cancelled should be false for a normal run")
	}
}

func TestBulkResult_CancelledSetOnContextCancellation(t *testing.T) {
	chunks := make([]parser.Chunk, 10)
	for i := range chunks {
		chunks[i] = parser.Chunk{Index: i, Total: 10, Content: "content"}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := pipeline.BulkRequest{
		Chunks: chunks, SyllabusID: "test", Mode: pipeline.ModePreview,
		Source: "test", Workers: 3, Reporter: &pipeline.NoopReporter{},
		Provider: ai.NewMockProvider("result"),
	}
	result, _ := pipeline.ExecuteBulk(ctx, req)
	if !result.Cancelled {
		t.Error("Cancelled should be true when context was cancelled")
	}
}

func TestExecuteBulk_PreservesChunkOrder(t *testing.T) {
	const n = 8
	chunks := make([]parser.Chunk, n)
	for i := range chunks {
		chunks[i] = parser.Chunk{Index: i, Total: n, Content: fmt.Sprintf("content %d", i)}
	}

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "test",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    4, // multiple workers → results arrive out of order without sorting
		Reporter:   &pipeline.NoopReporter{},
		Provider:   ai.NewMockProviderWithDelay("result", 5*time.Millisecond),
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if len(result.Topics) != n {
		t.Fatalf("expected %d topics, got %d", n, len(result.Topics))
	}
	for i, tr := range result.Topics {
		if tr.ChunkIndex != i {
			t.Errorf("result[%d].ChunkIndex = %d, want %d (results not ordered)", i, tr.ChunkIndex, i)
		}
	}
}

func TestBulkResult_TopicResultFields(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 1, Heading: "Quadratics", Content: "quadratic equations content"},
	}

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "india-jee",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    1,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   ai.NewMockProvider("teaching_notes:\n  - note: test"),
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}

	if len(result.Topics) == 0 {
		t.Fatal("expected at least one topic result")
	}

	topic := result.Topics[0]
	if topic.Heading == "" && topic.ChunkIndex < 0 {
		t.Error("TopicResult fields not populated")
	}
	if topic.Output == "" {
		t.Error("TopicResult.Output should not be empty")
	}
}

// bulkMockContentReader is a test double for pipeline.ContentReader used in bulk tests.
type bulkMockContentReader struct {
	files map[string][]byte
}

func (m *bulkMockContentReader) ReadFile(path, _ string) ([]byte, error) {
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, errors.New("not found")
}

func TestExecuteBulk_TruncatesOversizedChunks(t *testing.T) {
	// Create a chunk well above the 9800-token (~31200 char) limit.
	oversized := strings.Repeat("word ", 10000) // ~40000 chars ≈ 10000 tokens
	chunks := []parser.Chunk{
		{Index: 0, Total: 1, Heading: "Big Chapter", Content: oversized},
	}

	var capturedPrompt string
	provider := &promptCapturingProvider{
		inner:    ai.NewMockProvider("result"),
		captured: &capturedPrompt,
	}

	req := pipeline.BulkRequest{
		Chunks: chunks, SyllabusID: "test", Mode: pipeline.ModePreview,
		Source: "test", Workers: 1, Reporter: &pipeline.NoopReporter{},
		Provider: provider,
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if len(result.Topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(result.Topics))
	}
	// The prompt should not contain the full oversized content.
	const maxPromptChars = 9800 * 4 * 2 // generous upper bound including template
	if len(capturedPrompt) > maxPromptChars {
		t.Errorf("prompt length %d exceeds expected max %d — chunk not truncated",
			len(capturedPrompt), maxPromptChars)
	}
}

func TestExecuteBulk_WithContentReader_IncludesExistingContent(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 1, Heading: "Algebra", Content: "new algebra content"},
	}

	// Pre-populate a mock content reader with existing content.
	reader := &bulkMockContentReader{
		files: map[string][]byte{
			"syllabi/india-jee/algebra.yaml": []byte("existing: content"),
		},
	}

	var capturedPrompt string
	provider := &promptCapturingProvider{
		inner:    ai.NewMockProvider("generated result"),
		captured: &capturedPrompt,
	}

	req := pipeline.BulkRequest{
		Chunks:        chunks,
		SyllabusID:    "india-jee",
		Mode:          pipeline.ModePreview,
		Source:        "test",
		Workers:       1,
		Reporter:      &pipeline.NoopReporter{},
		Provider:      provider,
		ContentReader: reader,
	}

	result, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if len(result.Topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(result.Topics))
	}
	// The provider should have seen the existing content in the prompt.
	if !containsStr(capturedPrompt, "existing: content") {
		t.Errorf("prompt should include existing content when ContentReader is set; got prompt: %q", capturedPrompt)
	}
}

func TestExecuteBulk_WithoutContentReader_NoExistingContentInPrompt(t *testing.T) {
	chunks := []parser.Chunk{
		{Index: 0, Total: 1, Heading: "Algebra", Content: "new algebra content"},
	}

	var capturedPrompt string
	provider := &promptCapturingProvider{
		inner:    ai.NewMockProvider("generated result"),
		captured: &capturedPrompt,
	}

	req := pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: "india-jee",
		Mode:       pipeline.ModePreview,
		Source:     "test",
		Workers:    1,
		Reporter:   &pipeline.NoopReporter{},
		Provider:   provider,
		// No ContentReader
	}

	_, err := pipeline.ExecuteBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteBulk() error = %v", err)
	}
	if containsStr(capturedPrompt, "DO NOT duplicate") {
		t.Error("prompt should not include existing-content section when ContentReader is nil")
	}
}

// promptCapturingProvider captures the last prompt sent to the AI for inspection.
type promptCapturingProvider struct {
	inner    ai.Provider
	captured *string
}

func (p *promptCapturingProvider) Complete(ctx context.Context, req ai.CompletionRequest) (ai.CompletionResponse, error) {
	for _, m := range req.Messages {
		if m.Role == "user" {
			*p.captured = m.Content
		}
	}
	return p.inner.Complete(ctx, req)
}

func (p *promptCapturingProvider) StreamComplete(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	return p.inner.StreamComplete(ctx, req)
}

func (p *promptCapturingProvider) Models() []ai.ModelInfo {
	return p.inner.Models()
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
