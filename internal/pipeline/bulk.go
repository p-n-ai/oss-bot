package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
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
	Topics          []TopicResult
	Errors          []error
	Duration        time.Duration
	ProcessedChunks int  // Number of chunks that completed (including those with errors).
	Cancelled       bool // True when the operation stopped early due to context cancellation.
}

// BulkRequest configures a bulk import operation.
type BulkRequest struct {
	Chunks     []parser.Chunk   // Document chunks to process in parallel
	SyllabusID string           // Target syllabus identifier
	Mode       ExecutionMode    // ModePreview, ModeWriteFS, or ModeCreatePR
	Source     string           // "cli", "bot", "web" — for provenance
	Workers    int              // Concurrent workers (0 defaults to 3)
	Reporter   ProgressReporter // Real-time progress feedback
	Provider   ai.Provider      // AI provider to use for generation

	// ContentReader optionally reads existing repo content before generation.
	// When set, processChunk fetches any existing content at the chunk's derived
	// topic path and includes it in the prompt so the AI generates supplementary
	// material rather than duplicating what's already there.
	// Nil means treat every chunk as new content (initial import).
	ContentReader ContentReader

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
		return &BulkResult{Duration: time.Since(start), Cancelled: true}, err
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

	// Send jobs. Return (not break) on cancellation to stop enqueuing.
	go func() {
		defer close(jobs)
		for _, c := range req.Chunks {
			select {
			case jobs <- job{chunk: c}:
			case <-ctx.Done():
				return
			}
		}
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
			reporter.OnProgress(processed, total, tr.Heading, StatusDone)
		}
	}

	// Restore chunk order — parallel workers complete in arbitrary order.
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].ChunkIndex < topics[j].ChunkIndex
	})

	bulkResult := &BulkResult{
		Topics:          topics,
		Errors:          errs,
		Duration:        time.Since(start),
		ProcessedChunks: processed,
		Cancelled:       ctx.Err() != nil,
	}

	reporter.OnComplete(bulkResult)
	return bulkResult, nil
}

// syllabusTopicPath derives a candidate repo-relative file path from a syllabus ID
// and a chunk heading. Used when ContentReader is set to check for existing content.
// Example: ("india-jee", "Chapter 3: Algebra") → "syllabi/india-jee/chapter-3-algebra.yaml"
func syllabusTopicPath(syllabusID, heading string) string {
	slug := strings.ToLower(heading)
	// Replace non-alphanumeric runs with a single hyphen.
	var b strings.Builder
	prevHyphen := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	slug = strings.Trim(b.String(), "-")
	return "syllabi/" + syllabusID + "/" + slug + ".yaml"
}

// processChunk generates curriculum content for a single document chunk.
func processChunk(ctx context.Context, req BulkRequest, chunk parser.Chunk) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if req.Provider == nil {
		return "", fmt.Errorf("no AI provider configured")
	}

	// Guard: truncate chunk content if it would push the total prompt over a safe
	// token limit. The prompt template + system message adds ~150 tokens overhead;
	// leave the rest for the content. 1 token ≈ 4 chars.
	const maxContentTokens = 9800 // leaves ~200 tokens headroom within an 10K chunk
	if contentTokens := len(chunk.Content) / 4; contentTokens > maxContentTokens {
		slog.Warn("chunk content exceeds safe token limit, truncating",
			"chunk_index", chunk.Index,
			"tokens_estimated", contentTokens,
			"limit", maxContentTokens,
		)
		chunk.Content = chunk.Content[:maxContentTokens*4]
	}

	// Optionally include existing content so the AI avoids duplication.
	existingSection := ""
	if req.ContentReader != nil && chunk.Heading != "" {
		// Derive a candidate repo-relative path from the syllabus and heading.
		candidatePath := syllabusTopicPath(req.SyllabusID, chunk.Heading)
		if data, err := req.ContentReader.ReadFile(candidatePath, "main"); err == nil && len(data) > 0 {
			existingSection = fmt.Sprintf("\n\nExisting content at this path (DO NOT duplicate):\n%s", string(data))
		}
	}

	prompt := fmt.Sprintf(
		`Extract curriculum topics from this document section for syllabus %q.

Heading: %s

Content:
%s%s

Output structured YAML listing any topics, learning objectives, and key concepts found.
If existing content is provided above, generate only new or supplementary material.`,
		req.SyllabusID, chunk.Heading, chunk.Content, existingSection,
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
