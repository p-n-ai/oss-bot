// Package pipeline provides the unified orchestrator for all content generation.
// All three interfaces (CLI, Bot, Web Portal) call Pipeline.Execute().
package pipeline

import (
	"context"
	"fmt"
	"strconv"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
)

// ExecutionMode determines what happens after content is generated and validated.
type ExecutionMode int

const (
	// ModePreview generates and validates content, returns structured output.
	ModePreview ExecutionMode = iota

	// ModeWriteFS writes generated files to the local filesystem.
	ModeWriteFS

	// ModeCreatePR creates a GitHub PR with generated content.
	ModeCreatePR
)

// Request is the unified input for all content generation, regardless of interface.
type Request struct {
	TopicPath        string
	ContributionType string // "teaching_notes", "assessments", "examples"
	Content          string // Pre-extracted text (for import)
	Mode             ExecutionMode
	OutputDir        string            // For ModeWriteFS
	Options          map[string]string // count, difficulty, language, etc.
	Source           string            // "cli", "bot", "web" — for provenance
}

// Result is the unified output from the pipeline.
type Result struct {
	StructuredOutput string            // Generated YAML/Markdown
	Files            map[string]string // filepath -> content
	ValidationErrors []string
	PRUrl            string // Populated only in ModeCreatePR
	PRNumber         int
}

// Pipeline is the shared orchestrator for all content generation.
type Pipeline struct {
	aiProvider ai.Provider
	writer     output.Writer
	promptsDir string
	repoPath   string
}

// New creates a pipeline with the given dependencies.
func New(provider ai.Provider, w output.Writer, promptsDir, repoPath string) *Pipeline {
	return &Pipeline{
		aiProvider: provider,
		writer:     w,
		promptsDir: promptsDir,
		repoPath:   repoPath,
	}
}

// Execute runs the full content generation workflow:
//  1. Build context from topic
//  2. Generate content via AI
//  3. Execute based on mode (preview, write FS, create PR)
func (p *Pipeline) Execute(ctx context.Context, req Request) (*Result, error) {
	// 1. Build context
	genCtx, err := generator.BuildContext(p.repoPath, req.TopicPath)
	if err != nil {
		return nil, fmt.Errorf("building context: %w", err)
	}

	// 2. Generate based on contribution type
	generated, err := p.generate(ctx, genCtx, req)
	if err != nil {
		return nil, fmt.Errorf("generating content: %w", err)
	}

	result := &Result{
		StructuredOutput: generated.Content,
		Files:            generated.Files,
	}

	// 3. Execute based on mode
	switch req.Mode {
	case ModeWriteFS:
		if err := p.writer.WriteFiles(ctx, req.OutputDir, generated.Files); err != nil {
			return nil, fmt.Errorf("writing files: %w", err)
		}
	case ModeCreatePR:
		pr, err := p.writer.CreatePR(ctx, output.PRInput{
			Files:       generated.Files,
			TopicPath:   req.TopicPath,
			ContentType: req.ContributionType,
			Source:      req.Source,
		})
		if err != nil {
			return nil, fmt.Errorf("creating PR: %w", err)
		}
		result.PRUrl = pr.URL
		result.PRNumber = pr.Number
	case ModePreview:
		// No side effects — result already populated
	}

	return result, nil
}

func (p *Pipeline) generate(ctx context.Context, genCtx *generator.GenerationContext, req Request) (*generator.GenerationResult, error) {
	switch req.ContributionType {
	case "teaching_notes":
		return generator.GenerateTeachingNotes(ctx, p.aiProvider, genCtx)
	case "assessments":
		count := 5
		difficulty := "medium"
		if v, ok := req.Options["count"]; ok {
			if n, err := strconv.Atoi(v); err == nil {
				count = n
			}
		}
		if v, ok := req.Options["difficulty"]; ok {
			difficulty = v
		}
		return generator.GenerateAssessments(ctx, p.aiProvider, genCtx, count, difficulty)
	case "examples":
		return generator.GenerateExamples(ctx, p.aiProvider, genCtx)
	default:
		return nil, fmt.Errorf("unknown contribution type: %s", req.ContributionType)
	}
}
