package pipeline_test

import (
	"context"
	"errors"
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
