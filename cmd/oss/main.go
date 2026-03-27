package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/parser"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
	"github.com/p-n-ai/oss-bot/internal/validator"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "oss",
		Short: "OSS Bot CLI — validate, generate, and manage curriculum content",
		Long: `OSS Bot CLI provides tools to validate curriculum YAML files,
generate AI-powered teaching content, import from PDFs, and translate topics.`,
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(qualityCmd())
	rootCmd.AddCommand(translateCmd())
	rootCmd.AddCommand(scaffoldCmd())
	rootCmd.AddCommand(importCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func validateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate YAML files against OSS schemas",
		Long:  `Validate all YAML files in a directory tree against the OSS JSON Schemas.`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runValidate,
	}
	cmd.Flags().StringP("file", "f", "", "Validate a single file")
	cmd.Flags().StringP("schema-dir", "s", "", "Path to schema directory (default: auto-detect from OSS repo)")
	return cmd
}

func generateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate AI-powered curriculum content",
	}
	cmd.AddCommand(generateTeachingNotesCmd())
	cmd.AddCommand(generateAssessmentsCmd())
	cmd.AddCommand(generateExamplesCmd())
	return cmd
}

func generateTeachingNotesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "teaching-notes <topic-id>",
		Short: "Generate teaching notes for a topic",
		Args:  cobra.ExactArgs(1),
		RunE:  runGenerate("teaching_notes"),
	}
}

func generateAssessmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assessments <topic-id>",
		Short: "Generate assessment questions for a topic",
		Args:  cobra.ExactArgs(1),
		RunE:  runGenerate("assessments"),
	}
	cmd.Flags().IntP("count", "c", 5, "Number of questions to generate")
	cmd.Flags().StringP("difficulty", "d", "medium", "Difficulty level: easy, medium, hard")
	return cmd
}

func generateExamplesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "examples <topic-id>",
		Short: "Generate worked examples for a topic",
		Args:  cobra.ExactArgs(1),
		RunE:  runGenerate("examples"),
	}
}

func qualityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quality [path]",
		Short: "Generate quality report for curriculum content",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runQuality,
	}
}

func translateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate topic content to another language",
		RunE:  runTranslate,
	}
	cmd.Flags().String("topic", "", "Topic ID to translate (required)")
	cmd.Flags().String("to", "", "Target language code: ms, zh, ta (required)")
	cmd.MarkFlagRequired("topic")
	cmd.MarkFlagRequired("to")
	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}

	singleFile, _ := cmd.Flags().GetString("file")
	schemaDir, _ := cmd.Flags().GetString("schema-dir")

	if schemaDir == "" {
		schemaDir = filepath.Join(repoPath, "schema")
	}

	v, err := validator.New(schemaDir)
	if err != nil {
		return fmt.Errorf("initializing validator: %w", err)
	}

	if singleFile != "" {
		schemaType := validator.DetectSchemaType(singleFile)
		if schemaType == "" {
			return fmt.Errorf("cannot detect schema type for %s", singleFile)
		}
		result, err := v.ValidateFile(singleFile, schemaType)
		if err != nil {
			return err
		}
		printResult(*result)
		if !result.Valid {
			os.Exit(1)
		}
		return nil
	}

	// Validate directory
	target := repoPath
	if len(args) > 0 {
		target = args[0]
	}

	results, err := v.ValidateDir(target)
	if err != nil {
		return fmt.Errorf("validating directory: %w", err)
	}

	hasErrors := false
	for _, r := range results {
		printResult(r)
		if !r.Valid {
			hasErrors = true
		}
	}

	if hasErrors {
		fmt.Fprintf(os.Stderr, "\n❌ Validation failed\n")
		os.Exit(1)
	}

	fmt.Printf("\n✅ All %d files valid\n", len(results))
	return nil
}

func runQuality(cmd *cobra.Command, args []string) error {
	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}
	target := repoPath
	if len(args) > 0 {
		target = args[0]
	}

	// Walk the target directory and assess quality of YAML topic files
	var topics []validator.TopicQuality
	levelCounts := make(map[int]int)

	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		schemaType := validator.DetectSchemaType(path)
		if schemaType != "topic" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		topicInfo := validator.TopicInfoFromYAML(data, path, target)
		actual := validator.AssessQuality(topicInfo)
		overclaimed := topicInfo.ClaimedLevel > actual

		tq := validator.TopicQuality{
			ID:           topicInfo.ID,
			Name:         topicInfo.Name,
			ActualLevel:  actual,
			ClaimedLevel: topicInfo.ClaimedLevel,
			Overclaimed:  overclaimed,
		}
		topics = append(topics, tq)
		levelCounts[actual]++
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	report := validator.QualityReport{
		Topics:      topics,
		LevelCounts: levelCounts,
	}
	fmt.Print(validator.FormatQualityReport(report))

	if len(topics) == 0 {
		fmt.Println("No topic files found.")
	}
	return nil
}

func runTranslate(cmd *cobra.Command, args []string) error {
	topicID, _ := cmd.Flags().GetString("topic")
	targetLang, _ := cmd.Flags().GetString("to")
	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}

	provider, err := createAIProvider()
	if err != nil {
		return err
	}

	genCtx, err := generator.BuildContext(repoPath, topicID)
	if err != nil {
		return fmt.Errorf("building context: %w", err)
	}

	result, err := generator.Translate(context.Background(), provider, &genCtx.Topic, targetLang)
	if err != nil {
		return err
	}

	fmt.Println(result.Content)
	return nil
}

// runGenerate returns a RunE function for any generate subcommand.
func runGenerate(contributionType string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		topicID := args[0]
		repoPath := os.Getenv("OSS_REPO_PATH")
		if repoPath == "" {
			repoPath = "."
		}
		promptsDir := os.Getenv("OSS_PROMPTS_DIR")
		if promptsDir == "" {
			promptsDir = "prompts/"
		}

		provider, err := createAIProvider()
		if err != nil {
			return err
		}

		p := pipeline.New(provider, &output.LocalWriter{}, promptsDir, repoPath)

		opts := make(map[string]string)
		if contributionType == "assessments" {
			count, _ := cmd.Flags().GetInt("count")
			difficulty, _ := cmd.Flags().GetString("difficulty")
			opts["count"] = strconv.Itoa(count)
			opts["difficulty"] = difficulty
		}

		result, err := p.Execute(context.Background(), pipeline.Request{
			TopicPath:        topicID,
			ContributionType: contributionType,
			Mode:             pipeline.ModePreview,
			OutputDir:        repoPath,
			Options:          opts,
			Source:           "cli",
		})
		if err != nil {
			return err
		}

		fmt.Println(result.StructuredOutput)

		if len(result.ValidationErrors) > 0 {
			fmt.Fprintf(os.Stderr, "\nValidation warnings:\n")
			for _, e := range result.ValidationErrors {
				fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
			}
		}
		return nil
	}
}

func scaffoldCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaffold",
		Short: "Scaffold new syllabus or subject directory structure",
	}
	cmd.AddCommand(scaffoldSyllabusCmd())
	cmd.AddCommand(scaffoldSubjectCmd())
	return cmd
}

func scaffoldSyllabusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "syllabus",
		Short: "Create a new syllabus directory from a curriculum document or URL",
		RunE:  runScaffoldSyllabus,
	}
	cmd.Flags().String("id", "", "Syllabus ID (required, e.g. india-jee)")
	cmd.Flags().String("country", "", "Country code (e.g. india)")
	cmd.Flags().String("from-file", "", "Path to curriculum document (PDF, DOCX, TXT)")
	cmd.Flags().String("from-url", "", "URL of curriculum specification page")
	cmd.Flags().String("from-text", "", "Curriculum description text")
	cmd.MarkFlagRequired("id")
	return cmd
}

func scaffoldSubjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subject",
		Short: "Create a new subject directory within an existing syllabus",
		RunE:  runScaffoldSubject,
	}
	cmd.Flags().String("syllabus", "", "Syllabus ID (required)")
	cmd.Flags().String("id", "", "Subject ID (required, e.g. mathematics)")
	cmd.Flags().String("country", "", "Country code (e.g. india)")
	cmd.Flags().String("from-file", "", "Path to subject document")
	cmd.Flags().String("from-url", "", "URL of subject specification page")
	cmd.Flags().String("from-text", "", "Subject description text")
	cmd.MarkFlagRequired("syllabus")
	cmd.MarkFlagRequired("id")
	return cmd
}

func runScaffoldSyllabus(cmd *cobra.Command, _ []string) error {
	syllabusID, _ := cmd.Flags().GetString("id")
	country, _ := cmd.Flags().GetString("country")
	fromFile, _ := cmd.Flags().GetString("from-file")
	fromText, _ := cmd.Flags().GetString("from-text")
	outputDir := os.Getenv("OSS_REPO_PATH")
	if outputDir == "" {
		outputDir = "."
	}

	sourceText, err := resolveSourceText(fromFile, fromText)
	if err != nil {
		return err
	}

	provider, _ := createAIProvider() // Optional; scaffolder works without AI

	s := generator.NewScaffolder(provider)
	result, err := s.ScaffoldSyllabus(context.Background(), generator.ScaffoldRequest{
		SyllabusID: syllabusID,
		Country:    country,
		SourceText: sourceText,
		OutputDir:  outputDir,
	})
	if err != nil {
		return fmt.Errorf("scaffolding syllabus: %w", err)
	}

	if err := s.WriteFiles(result, outputDir); err != nil {
		return fmt.Errorf("writing scaffold files: %w", err)
	}

	fmt.Println(result.Summary)
	for path := range result.Files {
		fmt.Printf("  created: %s\n", filepath.Join(outputDir, path))
	}
	return nil
}

func runScaffoldSubject(cmd *cobra.Command, _ []string) error {
	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectID, _ := cmd.Flags().GetString("id")
	country, _ := cmd.Flags().GetString("country")
	fromFile, _ := cmd.Flags().GetString("from-file")
	fromText, _ := cmd.Flags().GetString("from-text")
	outputDir := os.Getenv("OSS_REPO_PATH")
	if outputDir == "" {
		outputDir = "."
	}

	sourceText, err := resolveSourceText(fromFile, fromText)
	if err != nil {
		return err
	}

	provider, _ := createAIProvider() // Optional

	s := generator.NewScaffolder(provider)
	result, err := s.ScaffoldSubject(context.Background(), generator.ScaffoldRequest{
		SyllabusID: syllabusID,
		SubjectID:  subjectID,
		Country:    country,
		SourceText: sourceText,
		OutputDir:  outputDir,
	})
	if err != nil {
		return fmt.Errorf("scaffolding subject: %w", err)
	}

	if err := s.WriteFiles(result, outputDir); err != nil {
		return fmt.Errorf("writing scaffold files: %w", err)
	}

	fmt.Println(result.Summary)
	for path := range result.Files {
		fmt.Printf("  created: %s\n", filepath.Join(outputDir, path))
	}
	return nil
}

func importCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import curriculum content from a PDF into the OSS repo",
		Long: `Extract curriculum topics from a PDF document and generate structured
YAML files in the OSS repo. Uses parallel AI workers to process each
chapter/section concurrently.

Example:
  oss import --pdf DSKP-KSSM-Matematik-Tingkatan-4.pdf --syllabus malaysia-kssm --subject malaysia-kssm-matematik-tingkatan-4`,
		RunE: runImport,
	}
	cmd.Flags().String("pdf", "", "Path to PDF file (required)")
	cmd.Flags().String("syllabus", "", "Target syllabus ID (required, e.g. malaysia-kssm)")
	cmd.Flags().String("subject", "", "Target subject ID (e.g. malaysia-kssm-matematik-tingkatan-4)")
	cmd.Flags().Int("workers", 3, "Number of parallel AI workers (overrides OSS_WORKER_COUNT)")
	cmd.Flags().Int("chunk-size", 2000, "Max tokens per chunk (lower = more files, higher = less context loss)")
	cmd.Flags().Bool("pr", false, "Create a GitHub PR instead of writing to filesystem")
	cmd.Flags().Bool("force", false, "Overwrite existing topic files instead of AI-merging them")
	cmd.MarkFlagRequired("pdf")
	cmd.MarkFlagRequired("syllabus")
	return cmd
}

func runImport(cmd *cobra.Command, _ []string) error {
	pdfPath, _ := cmd.Flags().GetString("pdf")
	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectID, _ := cmd.Flags().GetString("subject")
	workers, _ := cmd.Flags().GetInt("workers")
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	createPR, _ := cmd.Flags().GetBool("pr")
	force, _ := cmd.Flags().GetBool("force")

	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}

	provider, err := createAIProvider()
	if err != nil {
		return err
	}

	// Wrap with a reasoning provider for bulk import and content merge.
	// Uses OSS_AI_REASONING_API_KEY + OSS_AI_REASONING_MODEL (default: deepseek/deepseek-r1).
	// Falls back to the base provider transparently when the key is not set.
	reasoningProvider := ai.NewReasoningProviderFromEnv(provider)
	if os.Getenv("OSS_AI_REASONING_API_KEY") != "" {
		model := os.Getenv("OSS_AI_REASONING_MODEL")
		if model == "" {
			model = "deepseek/deepseek-r1"
		}
		fmt.Printf("Using reasoning model: %s\n", model)
	}

	// 1. Extract text from PDF
	fmt.Printf("Extracting text from %s...\n", pdfPath)
	text, err := parser.ExtractPDFText(pdfPath)
	if err != nil {
		return fmt.Errorf("extracting PDF: %w", err)
	}
	fmt.Printf("Extracted %d characters\n", len(text))

	// 2. Try DSKP-specific extraction first; fall back to generic chunker.
	// DSKP (Dokumen Standard Kurikulum dan Pentaksiran) uses BIDANG PEMBELAJARAN
	// / TAJUK markers that are distinct from generic Markdown headings.
	var chunks []parser.Chunk
	if dskpTopics := extractDSKPTopics(text); len(dskpTopics) > 0 {
		fmt.Printf("Detected DSKP format: %d topics (BIDANG PEMBELAJARAN/TAJUK structure)\n", len(dskpTopics))
		chunks = dskpTopicsToChunks(dskpTopics)
	} else {
		chunks = parser.ChunkDocument(text, parser.ChunkOptions{
			MaxChunkSize: chunkSize,
			SplitOn:      []string{"# ", "## ", "### ", "Chapter ", "Bab ", "BAB ", "BAHAGIAN ", "Bahagian ", "TAJUK", "Tajuk"},
		})
	}
	fmt.Printf("Split into %d chunks\n", len(chunks))

	// 3. Resolve output directory — search for existing subject topics dir
	// created by scaffold, fall back to a flat output dir.
	topicsDir, err := findSubjectTopicsDir(repoPath, subjectID, syllabusID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		topicsDir = filepath.Join(repoPath, "import-output", syllabusID)
		fmt.Printf("Writing to fallback dir: %s\n", topicsDir)
	}
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		return fmt.Errorf("creating output dir %s: %w", topicsDir, err)
	}

	// 4. Run bulk import with progress reporting
	mode := pipeline.ModeWriteFS
	if createPR {
		mode = pipeline.ModeCreatePR
	}
	result, err := pipeline.ExecuteBulk(cmd.Context(), pipeline.BulkRequest{
		Chunks:     chunks,
		SyllabusID: syllabusID,
		SubjectID:  subjectID,
		Mode:       mode,
		Source:     "cli",
		Workers:    workers,
		Reporter:   pipeline.NewCLIReporter(),
		Provider:   reasoningProvider,
	})
	if err != nil {
		return fmt.Errorf("bulk import: %w", err)
	}

	// 5. Write each topic output to a YAML file.
	// Use the canonical OSS topic ID (e.g. MT4-01) when subjectID is known;
	// fall back to the heading slug otherwise.
	// If the file already exists, AI-merge the existing and new content into a
	// single coherent YAML before writing (no duplicate documents in the file).
	written := 0
	merged := 0
	for _, tr := range result.Topics {
		if tr.Err != nil || strings.TrimSpace(tr.Output) == "" {
			continue
		}
		var fileID string
		if subjectID != "" {
			fileID = topicFileID(subjectID, tr.Heading, tr.ChunkIndex)
		} else {
			slug := importSlug(tr.Heading)
			if slug == "" {
				slug = fmt.Sprintf("topic-%02d", tr.ChunkIndex+1)
			}
			fileID = slug
		}
		outPath := filepath.Join(topicsDir, fileID+".yaml")

		if existingData, readErr := os.ReadFile(outPath); readErr == nil {
			if force {
				// --force: overwrite existing file with new content directly.
				if err := os.WriteFile(outPath, []byte(tr.Output), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ writing %s: %v\n", outPath, err)
					continue
				}
				fmt.Printf("  replaced: %s\n", outPath)
				merged++
			} else {
				// Default: use AI to merge existing + new content.
				fmt.Printf("  merging: %s\n", outPath)
				mergedContent, mergeErr := mergeTopicYAML(cmd.Context(), reasoningProvider, string(existingData), tr.Output, tr.Heading)
				if mergeErr != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ AI merge failed for %s: %v — skipping\n", outPath, mergeErr)
					continue
				}
				if err := os.WriteFile(outPath, []byte(mergedContent), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ writing merged %s: %v\n", outPath, err)
					continue
				}
				fmt.Printf("  merged: %s\n", outPath)
				merged++
			}
		} else {
			// File does not exist — create fresh.
			if err := os.WriteFile(outPath, []byte(tr.Output), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ writing %s: %v\n", outPath, err)
				continue
			}
			fmt.Printf("  wrote: %s\n", outPath)
			written++
		}
	}

	// 6. Summary
	fmt.Printf("\nProcessed %d/%d chunks in %s — wrote %d new, merged %d existing file(s) in %s\n",
		result.ProcessedChunks, len(chunks), result.Duration.Round(time.Second),
		written, merged, topicsDir)

	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "%d chunks failed:\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
		}
	}
	return nil
}

// findSubjectTopicsDir searches under repoPath for an existing topics directory
// whose parent directory name matches subjectID or syllabusID. This locates
// directories created by "oss scaffold subject" without needing the country.
// Falls back to a path derived from syllabusID if no match is found.
func findSubjectTopicsDir(repoPath, subjectID, syllabusID string) (string, error) {
	if subjectID == "" && syllabusID == "" {
		return "", fmt.Errorf("no subject or syllabus ID provided")
	}

	searchID := subjectID
	if searchID == "" {
		searchID = syllabusID
	}

	var found string
	_ = filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if d.IsDir() && d.Name() == searchID {
			found = filepath.Join(path, "topics")
			return fs.SkipAll
		}
		return nil
	})

	if found != "" {
		return found, nil
	}
	return "", fmt.Errorf("directory %q not found under %s — run 'oss scaffold subject' first", searchID, repoPath)
}

// dskpTopic holds a single topic extracted from a DSKP (Dokumen Standard
// Kurikulum dan Pentaksiran) document, preserving the BIDANG PEMBELAJARAN
// (learning domain) that groups multiple topics.
type dskpTopic struct {
	LearningArea string // BIDANG PEMBELAJARAN label (e.g. "PERKAITAN DAN ALGEBRA")
	Number       string // Topic number (e.g. "1.0")
	Name         string // Topic name in Malay (e.g. "FUNGSI DAN PERSAMAAN KUADRATIK")
	Content      string // Raw text content of the topic section
}

// extractDSKPTopics scans PDF-extracted text for the DSKP section structure:
//
//	BIDANG PEMBELAJARAN
//	<area name>
//	TAJUK
//	<N.0 topic name>
//	<content until next TAJUK or BIDANG PEMBELAJARAN>
//
// Returns one entry per TAJUK found. Returns nil when the pattern is not
// detected (non-DSKP document).
func extractDSKPTopics(text string) []dskpTopic {
	lines := strings.Split(text, "\n")
	var topics []dskpTopic

	currentArea := ""
	var currentTopic *dskpTopic
	var contentLines []string

	saveCurrent := func() {
		if currentTopic != nil {
			currentTopic.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
			topics = append(topics, *currentTopic)
			currentTopic = nil
			contentLines = nil
		}
	}

	// nextNonEmpty advances past blank lines and returns the first non-blank
	// line and its index, or ("", len(lines)) if at EOF.
	nextNonEmpty := func(start int) (string, int) {
		for start < len(lines) {
			if s := strings.TrimSpace(lines[start]); s != "" {
				return s, start
			}
			start++
		}
		return "", start
	}

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		switch {
		case line == "BIDANG PEMBELAJARAN":
			saveCurrent()
			area, ni := nextNonEmpty(i + 1)
			if area != "" {
				currentArea = area
				i = ni + 1
			} else {
				i++
			}
		case line == "TAJUK":
			saveCurrent()
			topicLine, ni := nextNonEmpty(i + 1)
			if topicLine != "" {
				number, name := parseDSKPTopicLine(topicLine)
				currentTopic = &dskpTopic{
					LearningArea: currentArea,
					Number:       number,
					Name:         name,
				}
				contentLines = []string{topicLine}
				i = ni + 1
			} else {
				i++
			}
		default:
			if currentTopic != nil {
				contentLines = append(contentLines, lines[i])
			}
			i++
		}
	}
	saveCurrent()
	return topics
}

// parseDSKPTopicLine splits "1.0 TOPIC NAME" into number ("1.0") and name.
// Returns ("", line) for lines that don't match the N.0 pattern.
func parseDSKPTopicLine(line string) (number, name string) {
	idx := strings.Index(line, " ")
	if idx > 0 {
		candidate := line[:idx]
		if strings.Contains(candidate, ".") {
			return candidate, strings.TrimSpace(line[idx+1:])
		}
	}
	return "", line
}

// dskpTopicsToChunks converts DSKP topics into parser.Chunk values. Each chunk
// prepends a structured preamble so the AI knows the learning domain context.
func dskpTopicsToChunks(topics []dskpTopic) []parser.Chunk {
	chunks := make([]parser.Chunk, len(topics))
	total := len(topics)
	for i, t := range topics {
		heading := strings.TrimSpace(t.Number + " " + t.Name)
		content := fmt.Sprintf("BIDANG PEMBELAJARAN: %s\nTAJUK: %s\n\n%s",
			t.LearningArea, heading, t.Content)
		chunks[i] = parser.Chunk{
			Index:   i,
			Total:   total,
			Heading: heading,
			Content: content,
		}
	}
	return chunks
}

// topicFileID derives the canonical OSS topic file ID (e.g. "MT4-01") from
// the subject ID, chunk heading, and chunk index.
// Format: {PREFIX}{grade_num}-{NN} as defined in docs/id-conventions.md.
func topicFileID(subjectID, heading string, chunkIndex int) string {
	prefix := subjectPrefix(subjectID)
	grade := gradeNumber(subjectID)
	seq := topicSeqNum(heading, chunkIndex)
	return fmt.Sprintf("%s%s-%02d", prefix, grade, seq)
}

// subjectPrefix returns the 2-letter topic ID prefix for a subject ID.
// The prefix is always derived from the English subject name (language-neutral).
// See docs/id-conventions.md prefix table.
func subjectPrefix(subjectID string) string {
	prefixes := []struct{ pattern, prefix string }{
		{"matematik", "MT"}, {"matematika", "MT"}, {"mathematics", "MT"},
		{"sains", "SC"}, {"science", "SC"},
		{"fizik", "PH"}, {"fisika", "PH"}, {"physics", "PH"},
		{"kimia", "CH"}, {"chemistry", "CH"},
		{"biologi", "BI"}, {"biology", "BI"},
		{"sejarah", "HI"}, {"history", "HI"},
		{"geografi", "GE"}, {"geography", "GE"},
		{"bahasa-melayu", "BM"},
		{"english", "EN"},
		{"bahasa-arab", "AR"}, {"arabic", "AR"},
		{"indonesian", "ID"},
	}
	for _, p := range prefixes {
		if strings.Contains(subjectID, p.pattern) {
			return p.prefix
		}
	}
	// Fallback: first 2 uppercase letters of the last meaningful word
	gradeWords := map[string]bool{
		"tingkatan": true, "class": true, "year": true,
		"kelas": true, "chugaku": true, "koko": true,
	}
	parts := strings.Split(subjectID, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if _, err := strconv.Atoi(p); err != nil && !gradeWords[p] && len(p) >= 2 {
			return strings.ToUpper(p[:2])
		}
	}
	return "XX"
}

// gradeNumber extracts the numeric grade from a subject ID.
// Only values 1–20 qualify as grades; larger numbers are subject codes (e.g. 0580).
// Returns "" for exam-based syllabi with no grade (e.g. JEE, Cambridge IGCSE).
func gradeNumber(subjectID string) string {
	parts := strings.Split(subjectID, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		if n, err := strconv.Atoi(parts[i]); err == nil && n >= 1 && n <= 20 {
			return parts[i]
		}
	}
	return ""
}

// topicSeqNum extracts the integer sequence from a DSKP heading like
// "1.0 FUNGSI DAN PERSAMAAN" → 1. Falls back to chunkIndex+1.
func topicSeqNum(heading string, chunkIndex int) int {
	if fields := strings.Fields(heading); len(fields) > 0 {
		numStr := strings.SplitN(fields[0], ".", 2)[0]
		if n, err := strconv.Atoi(numStr); err == nil && n > 0 {
			return n
		}
	}
	return chunkIndex + 1
}

// importSlug converts a heading string into a lowercase hyphenated filename slug.
func importSlug(heading string) string {
	slug := strings.ToLower(heading)
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
	return strings.Trim(b.String(), "-")
}

// mergeTopicYAML uses AI to merge an existing OSS topic YAML file with newly
// imported content. The existing file is authoritative for identity fields
// (id, subject_id, syllabus_id, country_id, language); the AI supplements it
// with any new learning objectives or improved fields from the new content.
func mergeTopicYAML(ctx context.Context, provider ai.Provider, existing, newContent, heading string) (string, error) {
	prompt := fmt.Sprintf(`You are merging two versions of an OSS topic YAML file for topic "%s".

EXISTING FILE (authoritative — keep its structure and identity fields):
%s

NEW IMPORTED CONTENT (may contain additional objectives or corrections):
%s

Merge rules:
1. Keep the existing file's id, subject_id, syllabus_id, country_id, language, official_ref, name, mastery, ai_teaching_notes, provenance EXACTLY as-is
2. Update name_en only if the existing value is missing or clearly incorrect
3. Add any NEW learning_objectives from the new content not already present (deduplicate by id and by text similarity ≥ 85%%)
4. Set bloom_levels to the union of all bloom levels in the merged learning_objectives
5. Keep the higher of the two quality_level values
6. Keep prerequisites.required and prerequisites.recommended from the existing file (do not overwrite with empty lists)
7. Output ONLY valid YAML — no markdown fences, no explanatory text before or after`,
		heading, existing, newContent)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are a curriculum YAML merge assistant. Produce a single merged YAML file that preserves all existing content and adds new material without duplication."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.1,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// resolveSourceText reads text from a file path or returns the provided text directly.
func resolveSourceText(fromFile, fromText string) (string, error) {
	if fromFile != "" {
		data, err := os.ReadFile(fromFile)
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", fromFile, err)
		}
		return string(data), nil
	}
	return fromText, nil
}

// createAIProvider creates an AI provider from environment variables.
func createAIProvider() (ai.Provider, error) {
	providerName := os.Getenv("OSS_AI_PROVIDER")
	if providerName == "" {
		return nil, fmt.Errorf("OSS_AI_PROVIDER is required (set to: openai, anthropic, or ollama)")
	}
	apiKey := os.Getenv("OSS_AI_API_KEY")
	return ai.NewProvider(providerName, apiKey)
}

func printResult(r validator.ValidationResult) {
	if r.Valid {
		fmt.Printf("  ✅ %s (%s)\n", r.File, r.Type)
	} else {
		fmt.Printf("  ❌ %s (%s)\n", r.File, r.Type)
		for _, e := range r.Errors {
			fmt.Printf("     → %s\n", e)
		}
	}
}
