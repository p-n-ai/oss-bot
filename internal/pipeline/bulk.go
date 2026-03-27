package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
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
	SubjectGradeID string       // Target subject grade (e.g. malaysia-kssm-matematik-tingkatan-4); optional
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

// chunkSubjectGradeID returns the effective subject grade ID for a bulk request:
// SubjectGradeID if set, otherwise SyllabusID as a fallback.
func chunkSubjectGradeID(req BulkRequest) string {
	if req.SubjectGradeID != "" {
		return req.SubjectGradeID
	}
	return req.SyllabusID
}

// chunkTopicFileID derives the canonical OSS topic file ID (e.g. "MT4-01")
// from a subject ID, chunk heading, and chunk index.
func chunkTopicFileID(subjectID, heading string, index int) string {
	if subjectID == "" {
		return ""
	}
	prefix := chunkSubjectPrefix(subjectID)
	grade := chunkGradeNumber(subjectID)
	seq := chunkTopicSeqNum(heading, index)
	return fmt.Sprintf("%s%s-%02d", prefix, grade, seq)
}

func chunkSubjectPrefix(subjectID string) string {
	prefixes := []struct{ pattern, prefix string }{
		{"matematik", "MT"}, {"matematika", "MT"}, {"mathematics", "MT"},
		{"sains", "SC"}, {"science", "SC"},
		{"fizik", "PH"}, {"fisika", "PH"}, {"physics", "PH"},
		{"kimia", "CH"}, {"chemistry", "CH"},
		{"biologi", "BI"}, {"biology", "BI"},
		{"sejarah", "HI"}, {"history", "HI"},
		{"geografi", "GE"}, {"geography", "GE"},
		{"bahasa-melayu", "BM"}, {"english", "EN"},
		{"bahasa-arab", "AR"}, {"arabic", "AR"},
	}
	for _, p := range prefixes {
		if strings.Contains(subjectID, p.pattern) {
			return p.prefix
		}
	}
	gradeWords := map[string]bool{"tingkatan": true, "class": true, "year": true, "kelas": true}
	parts := strings.Split(subjectID, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if _, err := strconv.Atoi(p); err != nil && !gradeWords[p] && len(p) >= 2 {
			return strings.ToUpper(p[:2])
		}
	}
	return "XX"
}

func chunkGradeNumber(subjectID string) string {
	parts := strings.Split(subjectID, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		if n, err := strconv.Atoi(parts[i]); err == nil && n >= 1 && n <= 20 {
			return parts[i]
		}
	}
	return ""
}

func chunkTopicSeqNum(heading string, index int) int {
	if fields := strings.Fields(heading); len(fields) > 0 {
		numStr := strings.SplitN(fields[0], ".", 2)[0]
		if n, err := strconv.Atoi(numStr); err == nil && n > 0 {
			return n
		}
	}
	return index + 1
}

func chunkCountryFromSubject(id string) string {
	if idx := strings.Index(id, "-"); idx > 0 {
		return id[:idx]
	}
	return id
}

func chunkLanguageForCountry(countryID string) string {
	langs := map[string]string{
		"malaysia":  "ms",
		"indonesia": "id",
		"japan":     "ja",
		"uae":       "ar",
	}
	if l, ok := langs[countryID]; ok {
		return l
	}
	return "en"
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

	// Compute per-topic OSS metadata. Pre-filling known fields in the prompt
	// prevents the AI from hallucinating subject_id, country_id, etc.
	subjectID := chunkSubjectGradeID(req)
	topicID := chunkTopicFileID(subjectID, chunk.Heading, chunk.Index)
	if topicID == "" {
		slug := strings.ToLower(strings.Join(strings.Fields(chunk.Heading), "-"))
		if slug == "" {
			slug = fmt.Sprintf("topic-%02d", chunk.Index+1)
		}
		topicID = slug
	}
	countryID := chunkCountryFromSubject(subjectID)
	language := chunkLanguageForCountry(countryID)

	prompt := fmt.Sprintf(
		`You are extracting curriculum content from an official curriculum document and generating an Open School Syllabus (OSS) topic YAML file.

Topic: %s
Document content:
%s%s

Generate a SINGLE YAML document. Pre-filled fields MUST be kept exactly as shown. Fill in all FILL_ placeholders from the document content.

id: %s
official_ref: "FILL_OFFICIAL_REF"   # chapter/section code as printed in document, e.g. "1.0" or "Bab 1"
name: "FILL_NAME"                   # topic name in document language (Malay for KSSM)
name_en: "FILL_NAME_EN"             # English translation of name — always translate
subject_id: %s
syllabus_id: %s
country_id: %s
language: %s
difficulty: FILL_DIFFICULTY         # beginner | intermediate | advanced
tier: core                          # core | extension

learning_objectives:
  - id: FILL_SP_CODE                # STANDARD PEMBELAJARAN code, e.g. 1.1.1
    text: "FILL_TEXT"               # objective text in the document language (%s)
    text_en: "FILL_TEXT_EN"         # English translation of the objective
    bloom: FILL_BLOOM               # remember | understand | apply | analyze | evaluate | create
  # repeat for ALL objectives in the document

prerequisites:
  required: []
  recommended: []

bloom_levels:
  - FILL_BLOOM_LEVELS               # list all distinct bloom levels used above

mastery:
  minimum_score: 0.75
  assessment_count: 3
  spaced_repetition:
    initial_interval_days: 3
    multiplier: 2.5

ai_teaching_notes: "%s.teaching.md"
quality_level: 1
provenance: ai-assisted

RULES:
- Output ONLY valid YAML — no markdown fences, no explanatory text before or after
- Extract ALL learning objectives from the STANDARD PEMBELAJARAN section
- name MUST be in the document language (%s), name_en MUST be the English translation
- learning_objectives text MUST be in the document language (%s), text_en MUST be the English translation
- bloom levels: remember | understand | apply | analyze | evaluate | create`,
		chunk.Heading, chunk.Content, existingSection,
		topicID, subjectID, req.SyllabusID, countryID, language,
		language,
		topicID,
		language, language,
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
