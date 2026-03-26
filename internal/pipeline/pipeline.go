// Package pipeline provides the unified orchestrator for all content generation.
// All three interfaces (CLI, Bot, Web Portal) call Pipeline.Execute().
package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/validator"
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

// ContentReader reads existing content from the target repository.
// Used by the pipeline merge stage to fetch current file contents before merging.
// Implementations: in tests use a mock; in production use GitHubContentsReader.
type ContentReader interface {
	ReadFile(path, ref string) ([]byte, error)
}

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
	PRUrl            string                 // Populated only in ModeCreatePR
	PRNumber         int                    // Populated only in ModeCreatePR
	MergeReport      *generator.MergeReport // Non-nil if existing content was merged
}

// Pipeline is the shared orchestrator for all content generation.
type Pipeline struct {
	aiProvider    ai.Provider
	writer        output.Writer
	promptsDir    string
	repoPath      string
	contentReader ContentReader // Optional; nil means skip merge stage.
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

// WithContentReader attaches an optional content reader for the merge stage.
// Returns the pipeline for chaining.
func (p *Pipeline) WithContentReader(cr ContentReader) *Pipeline {
	p.contentReader = cr
	return p
}

// Execute runs the full content generation workflow:
//  1. Build context from topic
//  2. Generate content via AI
//  3. Merge with existing content (if ContentReader set)
//  4. Execute based on mode (preview, write FS, create PR)
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

	// Populate Files map if the generator did not set it.
	if len(generated.Files) == 0 && generated.Content != "" {
		generated.Files = buildFilesMap(genCtx, req.ContributionType, generated.Content, p.repoPath)
	}

	// Validate Bloom levels declared on the topic's learning objectives.
	bloomErrors := validator.ValidateBloomLevels(genLOsToValidatorLOs(genCtx.Topic.LearningObjectives))

	result := &Result{
		StructuredOutput: generated.Content,
		Files:            generated.Files,
		ValidationErrors: bloomErrors,
	}

	// 3. Merge with existing content when a reader is available
	if p.contentReader != nil {
		mergedContent, report, mergeErr := p.mergeWithExisting(req, genCtx, generated.Content)
		if mergeErr != nil {
			slog.Warn("merge stage failed, using generated content as-is", "error", mergeErr)
		} else {
			generated.Content = mergedContent
			result.StructuredOutput = mergedContent
			if report != nil {
				result.MergeReport = report
			}
		}
	}

	// 4. Execute based on mode
	switch req.Mode {
	case ModeWriteFS:
		if err := p.writer.WriteFiles(ctx, req.OutputDir, generated.Files); err != nil {
			return nil, fmt.Errorf("writing files: %w", err)
		}
	case ModeCreatePR:
		mergeDetails := ""
		if result.MergeReport != nil {
			mergeDetails = result.MergeReport.String()
		}
		pr, err := p.writer.CreatePR(ctx, output.PRInput{
			Files:        generated.Files,
			TopicPath:    req.TopicPath,
			ContentType:  req.ContributionType,
			Source:       req.Source,
			MergeDetails: mergeDetails,
		})
		if err != nil {
			return nil, fmt.Errorf("creating PR: %w", err)
		}
		result.PRUrl = pr.URL
		result.PRNumber = pr.Number
	case ModePreview:
		// No side effects — result already populated.
	}

	return result, nil
}

// mergeWithExisting reads existing content from the repo and merges it with
// the newly generated content. Returns the merged content and a MergeReport.
// If the existing file does not exist, generated is returned unchanged.
func (p *Pipeline) mergeWithExisting(
	req Request,
	genCtx *generator.GenerationContext,
	generated string,
) (string, *generator.MergeReport, error) {
	// Determine the filename for existing content based on contribution type.
	var fileName string
	switch req.ContributionType {
	case "teaching_notes":
		fileName = genCtx.Topic.TeachingNotesFile
	case "assessments":
		fileName = genCtx.Topic.AssessmentsFile
	case "examples":
		fileName = genCtx.Topic.ExamplesFile
	}
	if fileName == "" {
		return generated, nil, nil // No existing file configured.
	}

	// Construct the repo-relative path to the existing file.
	path := fileName
	if p.repoPath != "" && genCtx.TopicDir != "" {
		if relDir, err := filepath.Rel(p.repoPath, genCtx.TopicDir); err == nil {
			path = filepath.Join(relDir, fileName)
		}
	}

	// Read the existing file; a "not found" error just means no merge needed.
	existingData, err := p.contentReader.ReadFile(path, "main")
	if err != nil {
		return generated, nil, nil
	}

	// Merge based on contribution type.
	switch req.ContributionType {
	case "teaching_notes":
		merged, report := generator.MergeTeachingNotes(string(existingData), generated)
		return merged, &report, nil
	case "assessments":
		merged, report, err := generator.MergeAssessmentsYAML(string(existingData), generated)
		if err != nil {
			return generated, nil, err
		}
		return merged, &report, nil
	case "examples":
		merged, report, err := generator.MergeExamplesYAML(string(existingData), generated)
		if err != nil {
			return generated, nil, err
		}
		return merged, &report, nil
	}

	return generated, nil, nil
}

// buildFilesMap constructs the repo-relative file path → content map for
// generated content. Uses the topic's file reference fields and TopicDir.
// Returns nil if the topic has no file reference configured for contribType.
func buildFilesMap(genCtx *generator.GenerationContext, contribType, content, repoPath string) map[string]string {
	var fileName string
	switch contribType {
	case "teaching_notes":
		fileName = genCtx.Topic.TeachingNotesFile
	case "assessments":
		fileName = genCtx.Topic.AssessmentsFile
	case "examples":
		fileName = genCtx.Topic.ExamplesFile
	}
	if fileName == "" {
		return nil
	}

	filePath := fileName
	if repoPath != "" && genCtx.TopicDir != "" {
		if relDir, err := filepath.Rel(repoPath, genCtx.TopicDir); err == nil {
			filePath = filepath.Join(relDir, fileName)
		}
	}

	return map[string]string{filePath: content}
}

// genLOsToValidatorLOs converts generator learning objectives to the validator package type.
func genLOsToValidatorLOs(los []generator.LearningObjective) []validator.LearningObjective {
	out := make([]validator.LearningObjective, len(los))
	for i, lo := range los {
		out[i] = validator.LearningObjective{ID: lo.ID, Text: lo.Text, Bloom: lo.Bloom}
	}
	return out
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
