// Package pipeline provides the unified orchestrator for all content generation.
// All three interfaces (CLI, Bot, Web Portal) call Pipeline.Execute().
package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/validator"
	"gopkg.in/yaml.v3"
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
	ContributionType string // "teaching_notes", "assessments", "examples", "topic_enrich"
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

	// Strip markdown code fences (```yaml ... ```) that AI models sometimes add.
	generated.Content = StripCodeFences(generated.Content)

	// Fix double-quoted YAML strings containing LaTeX backslash sequences
	// (e.g. \text, \sqrt, \frac) that would be corrupted by YAML escape processing.
	if req.ContributionType == "assessments" || req.ContributionType == "examples" {
		generated.Content = SanitizeYAMLQuoting(generated.Content)
		for k, v := range generated.Files {
			generated.Files[k] = SanitizeYAMLQuoting(v)
		}
	}

	// Schema validation + single retry for YAML content types.
	if schemaType := SchemaTypeForContribution(req.ContributionType); schemaType != "" {
		resolver := validator.NewSchemaResolver(filepath.Join(p.repoPath, "schema"))
		subjectDir := validator.FindSubjectDir(filepath.Join(genCtx.TopicDir, "x.yaml"))
		schemasDir := validator.SubjectSchemasDir(subjectDir)

		v := validator.NewWithResolver(resolver)
		vResult, vErr := v.ValidateContentResolved([]byte(generated.Content), schemaType, schemasDir)
		if vErr == nil && !vResult.Valid && len(genCtx.ValidationFeedback) == 0 {
			// First failure — retry with schema errors as feedback.
			genCtx.ValidationFeedback = vResult.Errors
			generated, err = p.generate(ctx, genCtx, req)
			if err != nil {
				return nil, fmt.Errorf("retry generation: %w", err)
			}
			generated.Content = StripCodeFences(generated.Content)
			if req.ContributionType == "assessments" || req.ContributionType == "examples" {
				generated.Content = SanitizeYAMLQuoting(generated.Content)
			}
		}
	}

	// For topic_enrich, merge the AI output into the existing topic YAML file.
	if req.ContributionType == "topic_enrich" {
		topicFile, err := generator.FindTopicFile(p.repoPath, req.TopicPath)
		if err != nil {
			return nil, fmt.Errorf("finding topic file for enrichment: %w", err)
		}
		enrichedYAML, err := generator.EnrichTopicYAML(topicFile, generated.Content)
		if err != nil {
			return nil, fmt.Errorf("enriching topic YAML: %w", err)
		}
		relPath, _ := filepath.Rel(p.repoPath, topicFile)
		generated.Content = enrichedYAML
		generated.Files = map[string]string{relPath: enrichedYAML}
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
		// Update the topic YAML to reference the generated companion file.
		if req.ContributionType != "topic_enrich" {
			if updateErr := updateTopicFileRef(genCtx, req.ContributionType); updateErr != nil {
				slog.Warn("failed to update topic YAML file reference", "type", req.ContributionType, "error", updateErr)
			}
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
	// Derive filename from topic ID when the YAML field is not set.
	if fileName == "" && genCtx.Topic.ID != "" {
		switch contribType {
		case "teaching_notes":
			fileName = genCtx.Topic.ID + ".teaching.md"
		case "assessments":
			fileName = genCtx.Topic.ID + ".assessments.yaml"
		case "examples":
			fileName = genCtx.Topic.ID + ".examples.yaml"
		}
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

// StripCodeFences removes markdown code fences (```yaml ... ``` or ```markdown ... ```)
// that AI models sometimes wrap around generated content.
func StripCodeFences(s string) string {
	s = StripThinkTags(s)
	s = strings.TrimSpace(s)
	// Check for opening fence: ```yaml, ```yml, ```markdown, or bare ```
	if strings.HasPrefix(s, "```") {
		// Remove the first line (the opening fence)
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		// Remove the closing fence
		if strings.HasSuffix(strings.TrimSpace(s), "```") {
			s = strings.TrimSpace(s)
			s = s[:len(s)-3]
			s = strings.TrimRight(s, "\n")
		}
	}
	return s
}

// StripThinkTags removes <think>...</think> blocks that reasoning models
// (e.g. DeepSeek R1) include in their responses as chain-of-thought output.
func StripThinkTags(s string) string {
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(s, "</think>")
		if end == -1 {
			// Opening tag without closing — strip from <think> to end
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("</think>"):]
	}
	// Handle orphaned </think> (model returned only the closing tag)
	s = strings.ReplaceAll(s, "</think>", "")
	return strings.TrimSpace(s)
}

// updateTopicFileRef updates the topic YAML file to set the file reference field
// (ai_teaching_notes, assessments_file, or examples_file) for the given contribution type.
// This ensures the topic YAML always points to the generated companion file.
func updateTopicFileRef(genCtx *generator.GenerationContext, contribType string) error {
	if genCtx.TopicDir == "" || genCtx.Topic.ID == "" {
		return nil
	}

	// Determine which YAML key and filename to set.
	var yamlKey, fileName string
	switch contribType {
	case "teaching_notes":
		fileName = genCtx.Topic.ID + ".teaching.md"
		yamlKey = "ai_teaching_notes"
		if genCtx.Topic.TeachingNotesFile == fileName {
			return nil // already correct
		}
	case "assessments":
		fileName = genCtx.Topic.ID + ".assessments.yaml"
		yamlKey = "assessments_file"
		if genCtx.Topic.AssessmentsFile == fileName {
			return nil
		}
	case "examples":
		fileName = genCtx.Topic.ID + ".examples.yaml"
		yamlKey = "examples_file"
		if genCtx.Topic.ExamplesFile == fileName {
			return nil
		}
	default:
		return nil
	}

	// Find the topic YAML file in TopicDir.
	topicFile := filepath.Join(genCtx.TopicDir, genCtx.Topic.ID+".yaml")
	data, err := os.ReadFile(topicFile)
	if err != nil {
		return fmt.Errorf("reading topic file %s: %w", topicFile, err)
	}

	var raw yaml.Node
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing topic YAML: %w", err)
	}
	if raw.Kind != yaml.DocumentNode || len(raw.Content) == 0 {
		return fmt.Errorf("unexpected YAML structure in %s", topicFile)
	}
	mapping := raw.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node in %s", topicFile)
	}

	// Set or update the key.
	found := false
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == yamlKey {
			mapping.Content[i+1].Value = fileName
			mapping.Content[i+1].Tag = "!!str"
			found = true
			break
		}
	}
	if !found {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: yamlKey, Tag: "!!str"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: fileName, Tag: "!!str"},
		)
	}

	out, err := yaml.Marshal(&raw)
	if err != nil {
		return fmt.Errorf("marshaling updated YAML: %w", err)
	}

	if err := os.WriteFile(topicFile, out, 0644); err != nil {
		return fmt.Errorf("writing updated topic file: %w", err)
	}

	return nil
}

// genLOsToValidatorLOs converts generator learning objectives to the validator package type.
func genLOsToValidatorLOs(los []generator.LearningObjective) []validator.LearningObjective {
	out := make([]validator.LearningObjective, len(los))
	for i, lo := range los {
		out[i] = validator.LearningObjective{ID: lo.ID, Text: lo.Text, Bloom: lo.Bloom}
	}
	return out
}

// SchemaTypeForContribution maps a contribution type to its schema type name.
// Returns "" for types that don't have a YAML schema (e.g. teaching_notes is markdown).
// SchemaTypeForContribution maps a contribution type to its schema type name.
func SchemaTypeForContribution(ct string) string {
	switch ct {
	case "assessments":
		return "assessments"
	case "examples":
		return "examples"
	case "topic_enrich":
		return "topic"
	default:
		return ""
	}
}

func (p *Pipeline) generate(ctx context.Context, genCtx *generator.GenerationContext, req Request) (*generator.GenerationResult, error) {
	// Resolve and inject schema into generation context for YAML content types.
	schemaType := SchemaTypeForContribution(req.ContributionType)
	if schemaType != "" && genCtx.SchemaRules == "" {
		resolver := validator.NewSchemaResolver(filepath.Join(p.repoPath, "schema"))
		subjectDir := validator.FindSubjectDir(filepath.Join(genCtx.TopicDir, "x.yaml"))
		schemasDir := validator.SubjectSchemasDir(subjectDir)
		if schemaPath, ok := resolver.ResolveSchemaPath(schemaType, schemasDir); ok {
			if data, err := os.ReadFile(schemaPath); err == nil {
				genCtx.SchemaRules = string(data)
			}
		}
	}

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
	case "topic_enrich":
		return generator.GenerateTopicEnrichment(ctx, p.aiProvider, genCtx)
	default:
		return nil, fmt.Errorf("unknown contribution type: %s", req.ContributionType)
	}
}
