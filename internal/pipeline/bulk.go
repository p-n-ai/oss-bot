package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/parser"
)

// TopicResult holds the result for a single chunk processed in a bulk import.
type TopicResult struct {
	ChunkIndex int
	Heading    string
	Output     string // Generated content
	Err        error  // Non-nil if this chunk failed
}

// BulkResult holds the combined results of a bulk import operation.
type BulkResult struct {
	Topics   []TopicResult
	Errors   []error
	Duration time.Duration
}

// BulkRequest configures a bulk import operation.
type BulkRequest struct {
	Chunks     []parser.Chunk  // Document chunks to process in parallel
	SyllabusID string          // Target syllabus identifier
	Mode       ExecutionMode   // ModePreview, ModeWriteFS, or ModeCreatePR
	Source     string          // "cli", "bot", "web" — for provenance
	Workers    int             // Concurrent workers (0 defaults to 3)
	Reporter   ProgressReporter // Real-time progress feedback
	Provider   ai.Provider      // AI provider to use for generation

	// Hooks for testing concurrency behaviour — called on each worker goroutine.
	OnWorkerStart func()
	OnWorkerDone  func()
}

// defaultWorkers is the default number of parallel workers.
const defaultWorkers = 3

// ExecuteBulk processes multiple document chunks in parallel using a worker pool.
// Per-chunk errors are collected in BulkResult.Errors; the top-level error is
// returned only for setup failures or context cancellation before any work starts.
func ExecuteBulk(ctx context.Context, req BulkRequest) (*BulkResult, error) {
	workers := req.Workers
	if workers <= 0 {
		workers = defaultWorkers
	}

	reporter := req.Reporter
	if reporter == nil {
		reporter = &NoopReporter{}
	}

	start := time.Now()
	total := len(req.Chunks)

	if total == 0 {
		return &BulkResult{Duration: time.Since(start)}, nil
	}

	// Check context before starting.
	if err := ctx.Err(); err != nil {
		return &BulkResult{Duration: time.Since(start)}, err
	}

	reporter.OnStart(total)

	// Fan work out to a bounded worker pool.
	type job struct {
		chunk parser.Chunk
	}
	type result struct {
		topicResult TopicResult
	}

	jobs := make(chan job, total)
	results := make(chan result, total)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if req.OnWorkerStart != nil {
					req.OnWorkerStart()
				}
				output, err := processChunk(ctx, req, j.chunk)
				if req.OnWorkerDone != nil {
					req.OnWorkerDone()
				}
				results <- result{
					topicResult: TopicResult{
						ChunkIndex: j.chunk.Index,
						Heading:    j.chunk.Heading,
						Output:     output,
						Err:        err,
					},
				}
			}
		}()
	}

	// Send jobs.
	go func() {
		for _, c := range req.Chunks {
			select {
			case jobs <- job{chunk: c}:
			case <-ctx.Done():
				break
			}
		}
		close(jobs)
	}()

	// Close results once all workers finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results.
	var topics []TopicResult
	var errs []error
	processed := 0

	for r := range results {
		tr := r.topicResult
		topics = append(topics, tr)
		processed++

		if tr.Err != nil {
			errs = append(errs, fmt.Errorf("chunk %d (%s): %w", tr.ChunkIndex, tr.Heading, tr.Err))
			reporter.OnError(tr.Heading, tr.Err)
		} else {
			reporter.OnProgress(processed, total, tr.Heading, "done")
		}
	}

	bulkResult := &BulkResult{
		Topics:   topics,
		Errors:   errs,
		Duration: time.Since(start),
	}

	reporter.OnComplete(bulkResult)
	return bulkResult, nil
}

// processChunk generates curriculum content for a single document chunk.
func processChunk(ctx context.Context, req BulkRequest, chunk parser.Chunk) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if req.Provider == nil {
		return "", fmt.Errorf("no AI provider configured")
	}

	prompt := fmt.Sprintf(
		`Extract curriculum topics from this document section for syllabus %q.

Heading: %s

Content:
%s

Output structured YAML listing any topics, learning objectives, and key concepts found.`,
		req.SyllabusID, chunk.Heading, chunk.Content,
	)

	resp, err := req.Provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are a curriculum analysis assistant. Extract structured learning content from source documents."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.3,
	})
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}
