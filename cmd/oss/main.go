package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/parser"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
	"github.com/p-n-ai/oss-bot/internal/validator"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
		Long: `Validate all YAML files in a directory tree against the OSS JSON Schemas.

You can specify a directory path directly, or use --syllabus and --subject-grade
flags to resolve the path automatically.

Examples:
  oss validate
  oss validate /path/to/topics
  oss validate --file topic.yaml
  oss validate --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5
  oss validate --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5 --topic-id MT5-01`,
		Args: cobra.MaximumNArgs(1),
		RunE: runValidate,
	}
	cmd.Flags().StringP("file", "f", "", "Validate a single file")
	cmd.Flags().StringP("schema-dir", "s", "", "Path to schema directory (default: auto-detect from OSS repo)")
	cmd.Flags().String("syllabus", "", "Syllabus ID (e.g. malaysia-kssm)")
	cmd.Flags().String("subject-grade", "", "Subject grade ID (e.g. malaysia-kssm-matematik-tingkatan-5)")
	cmd.Flags().String("topic-id", "", "Validate only the specified topic (e.g. MT2-12)")
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
	cmd.AddCommand(generateAllCmd())
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

func generateAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Generate teaching-notes, assessments, and examples for all topics in a subject-grade",
		Long: `Discover all topic YAML files under a subject-grade directory and generate
teaching-notes, assessments, and examples for each topic using parallel workers.

Example:
  oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4
  oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4 --workers 5
  oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4 --dry-run`,
		RunE: runGenerateAll,
	}
	cmd.Flags().String("syllabus", "", "Syllabus ID (required, e.g. malaysia-kssm)")
	cmd.Flags().String("subject-grade", "", "Subject grade ID (required, e.g. malaysia-kssm-matematik-tingkatan-4)")
	cmd.Flags().String("topic-id", "", "Generate only for the specified topic (e.g. MT4-01)")
	cmd.Flags().Int("workers", 3, "Number of parallel workers")
	cmd.Flags().Bool("dry-run", false, "List discovered topics without generating")
	cmd.MarkFlagRequired("syllabus")
	cmd.MarkFlagRequired("subject-grade")
	return cmd
}

func runGenerateAll(cmd *cobra.Command, _ []string) error {
	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectGradeID, _ := cmd.Flags().GetString("subject-grade")
	filterTopicID, _ := cmd.Flags().GetString("topic-id")
	workers, _ := cmd.Flags().GetInt("workers")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}
	promptsDir := os.Getenv("OSS_PROMPTS_DIR")
	if promptsDir == "" {
		promptsDir = "prompts/"
	}

	// Discover topics directory
	subjectID := subjectBaseID(subjectGradeID)
	topicsDir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectID, syllabusID)
	if err != nil {
		return fmt.Errorf("finding topics directory: %w", err)
	}

	// Discover topic IDs
	topicIDs, err := discoverTopicIDs(topicsDir)
	if err != nil {
		return fmt.Errorf("discovering topics: %w", err)
	}

	// Filter to a single topic when --topic-id is set
	if filterTopicID != "" {
		found := false
		for _, id := range topicIDs {
			if id == filterTopicID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("topic %q not found in %s", filterTopicID, topicsDir)
		}
		topicIDs = []string{filterTopicID}
	}

	if len(topicIDs) == 0 {
		fmt.Println("No topic files found.")
		return nil
	}

	fmt.Printf("Found %d topics in %s\n", len(topicIDs), topicsDir)
	for _, id := range topicIDs {
		fmt.Printf("  %s\n", id)
	}

	if dryRun {
		return nil
	}

	// Create AI provider
	provider, err := createAIProvider()
	if err != nil {
		return err
	}

	contentTypes := []string{"teaching_notes", "assessments", "examples", "topic_enrich"}
	totalJobs := len(topicIDs) * len(contentTypes)
	completed := 0
	var genErrors []string

	// Worker pool
	type job struct {
		topicID          string
		contributionType string
	}
	jobs := make(chan job, totalJobs)
	results := make(chan error, totalJobs)

	for w := 0; w < workers; w++ {
		go func() {
			for j := range jobs {
				p := pipeline.New(provider, &output.LocalWriter{}, promptsDir, repoPath)
				_, err := p.Execute(context.Background(), pipeline.Request{
					TopicPath:        j.topicID,
					ContributionType: j.contributionType,
					Mode:             pipeline.ModeWriteFS,
					OutputDir:        repoPath,
					Source:           "cli",
				})
				results <- err
			}
		}()
	}

	// Enqueue jobs
	for _, topicID := range topicIDs {
		for _, ct := range contentTypes {
			jobs <- job{topicID: topicID, contributionType: ct}
		}
	}
	close(jobs)

	// Collect results
	for i := 0; i < totalJobs; i++ {
		err := <-results
		completed++
		if err != nil {
			genErrors = append(genErrors, err.Error())
		}
		fmt.Printf("\r  Progress: %d/%d", completed, totalJobs)
	}
	fmt.Println()

	fmt.Printf("Completed: %d/%d successful\n", totalJobs-len(genErrors), totalJobs)
	if len(genErrors) > 0 {
		fmt.Fprintf(os.Stderr, "%d errors:\n", len(genErrors))
		for _, e := range genErrors {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
		}
	}
	return nil
}

// discoverTopicIDs reads a topics directory and returns sorted topic IDs
// from YAML files that contain an `id` field. Excludes supplementary files
// like .assessments.yaml and .examples.yaml.
func discoverTopicIDs(topicsDir string) ([]string, error) {
	entries, err := os.ReadDir(topicsDir)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".yaml") {
			continue
		}
		// Skip supplementary files
		if strings.Contains(name, ".assessments.") || strings.Contains(name, ".examples.") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(topicsDir, name))
		if err != nil {
			continue
		}

		// Extract id field via simple line scan
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "id:") {
				id := strings.TrimSpace(strings.TrimPrefix(line, "id:"))
				if id != "" {
					ids = append(ids, id)
				}
				break
			}
		}
	}

	sort.Strings(ids)
	return ids, nil
}

func qualityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quality [path]",
		Short: "Generate quality report for curriculum content",
		Long: `Analyze quality levels of topic YAML files and generate a report.

You can specify a directory path directly, or use --syllabus and --subject-grade
flags to resolve the path automatically.

Examples:
  oss quality /path/to/topics
  oss quality --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5`,
		Args: cobra.MaximumNArgs(1),
		RunE: runQuality,
	}
	cmd.Flags().String("syllabus", "", "Syllabus ID (e.g. malaysia-kssm)")
	cmd.Flags().String("subject-grade", "", "Subject grade ID (e.g. malaysia-kssm-matematik-tingkatan-5)")
	cmd.Flags().String("topic-id", "", "Show quality for only the specified topic (e.g. MT5-01)")
	return cmd
}

func translateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate topic content to another language",
		Long: `Translate a single topic or all topics in a subject-grade to another language.

Examples:
  oss translate --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5 --topic-id MT5-01 --to en
  oss translate --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5 --to en
  oss translate --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5 --to en --workers 5`,
		RunE: runTranslate,
	}
	cmd.Flags().String("topic-id", "", "Topic ID to translate (e.g. MT5-01) — requires --syllabus and --subject-grade")
	cmd.Flags().String("to", "", "Target language code: ms, zh, ta, en (required)")
	cmd.Flags().String("syllabus", "", "Syllabus ID (e.g. malaysia-kssm)")
	cmd.Flags().String("subject-grade", "", "Subject grade ID (e.g. malaysia-kssm-matematik-tingkatan-5)")
	cmd.Flags().Int("workers", 3, "Number of parallel workers (batch mode only)")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("syllabus")
	cmd.MarkFlagRequired("subject-grade")
	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}

	singleFile, _ := cmd.Flags().GetString("file")
	schemaDir, _ := cmd.Flags().GetString("schema-dir")
	topicID, _ := cmd.Flags().GetString("topic-id")

	globalSchemaDir := schemaDir
	if globalSchemaDir == "" {
		globalSchemaDir = filepath.Join(repoPath, "schema")
	}

	// Use resolver-based validation: per-subject schema overrides + global fallback.
	resolver := validator.NewSchemaResolver(globalSchemaDir)
	v := validator.NewWithResolver(resolver)

	// Single topic by ID — validates all YAML files for that topic (e.g. MT1-03.yaml,
	// MT1-03.assessments.yaml, MT1-03.examples.yaml).
	// Requires --syllabus and --subject-grade to scope the search.
	if topicID != "" {
		syllabusID, _ := cmd.Flags().GetString("syllabus")
		subjectGradeID, _ := cmd.Flags().GetString("subject-grade")
		if syllabusID == "" || subjectGradeID == "" {
			return fmt.Errorf("--topic-id requires both --syllabus and --subject-grade flags")
		}
		topicsDir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectBaseID(subjectGradeID), syllabusID)
		if err != nil {
			return fmt.Errorf("finding topics directory: %w", err)
		}

		// Find all files matching the topic ID prefix (e.g. MT1-03.*)
		entries, err := os.ReadDir(topicsDir)
		if err != nil {
			return fmt.Errorf("reading topics directory: %w", err)
		}
		prefix := topicID + "."
		var topicFiles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), ".yaml") {
				topicFiles = append(topicFiles, filepath.Join(topicsDir, e.Name()))
			}
		}
		if len(topicFiles) == 0 {
			return fmt.Errorf("no files found for topic %s in %s", topicID, topicsDir)
		}

		hasErrors := false
		validated := 0
		for _, f := range topicFiles {
			schemaType := validator.DetectSchemaType(f)
			if schemaType == "" {
				continue
			}
			result, err := v.ValidateFileResolved(f, schemaType)
			if err != nil {
				return err
			}
			printResult(*result)
			validated++
			if !result.Valid {
				hasErrors = true
			}
		}

		if hasErrors {
			fmt.Fprintf(os.Stderr, "\n❌ Validation failed\n")
			os.Exit(1)
		}
		fmt.Printf("\n✅ All %d files valid for topic %s\n", validated, topicID)
		return nil
	}

	if singleFile != "" {
		schemaType := validator.DetectSchemaType(singleFile)
		if schemaType == "" {
			return fmt.Errorf("cannot detect schema type for %s", singleFile)
		}
		result, err := v.ValidateFileResolved(singleFile, schemaType)
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
	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectGradeID, _ := cmd.Flags().GetString("subject-grade")

	var target string
	switch {
	case len(args) > 0:
		target = args[0]
	case subjectGradeID != "" || syllabusID != "":
		dir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectBaseID(subjectGradeID), syllabusID)
		if err != nil {
			return err
		}
		target = dir
	default:
		target = repoPath
	}

	results, err := v.ValidateDirResolved(target)
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

	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectGradeID, _ := cmd.Flags().GetString("subject-grade")
	filterTopicID, _ := cmd.Flags().GetString("topic-id")

	// Single topic by ID — requires --syllabus and --subject-grade to scope the search
	if filterTopicID != "" {
		if syllabusID == "" || subjectGradeID == "" {
			return fmt.Errorf("--topic-id requires both --syllabus and --subject-grade flags")
		}
		topicsDir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectBaseID(subjectGradeID), syllabusID)
		if err != nil {
			return fmt.Errorf("finding topics directory: %w", err)
		}
		topicFile := filepath.Join(topicsDir, filterTopicID+".yaml")
		if _, statErr := os.Stat(topicFile); statErr != nil {
			return fmt.Errorf("topic file not found: %s", topicFile)
		}
		data, err := os.ReadFile(topicFile)
		if err != nil {
			return fmt.Errorf("reading %s: %w", topicFile, err)
		}
		topicInfo := validator.TopicInfoFromYAML(data, topicFile, filepath.Dir(topicFile))
		actual := validator.AssessQuality(topicInfo)
		overclaimed := topicInfo.ClaimedLevel > actual

		report := validator.QualityReport{
			Topics: []validator.TopicQuality{{
				ID:           topicInfo.ID,
				Name:         topicInfo.Name,
				ActualLevel:  actual,
				ClaimedLevel: topicInfo.ClaimedLevel,
				Overclaimed:  overclaimed,
			}},
			LevelCounts: map[int]int{actual: 1},
		}
		fmt.Print(validator.FormatQualityReport(report))
		return nil
	}

	var target string
	switch {
	case len(args) > 0:
		target = args[0]
	case subjectGradeID != "" || syllabusID != "":
		dir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectBaseID(subjectGradeID), syllabusID)
		if err != nil {
			return err
		}
		target = dir
	default:
		target = repoPath
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
	topicID, _ := cmd.Flags().GetString("topic-id")
	targetLang, _ := cmd.Flags().GetString("to")
	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectGradeID, _ := cmd.Flags().GetString("subject-grade")
	workers, _ := cmd.Flags().GetInt("workers")

	repoPath := os.Getenv("OSS_REPO_PATH")
	if repoPath == "" {
		repoPath = "."
	}

	provider, err := createAIProvider()
	if err != nil {
		return err
	}

	// Discover topics directory (--syllabus and --subject-grade are always required)
	subjectID := subjectBaseID(subjectGradeID)
	topicsDir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectID, syllabusID)
	if err != nil {
		return fmt.Errorf("finding topics directory: %w", err)
	}

	// Single topic mode
	if topicID != "" {
		return translateSingleTopic(provider, repoPath, topicID, targetLang)
	}

	// Batch mode
	topicIDs, err := discoverTopicIDs(topicsDir)
	if err != nil {
		return fmt.Errorf("discovering topics: %w", err)
	}
	if len(topicIDs) == 0 {
		fmt.Println("No topic files found.")
		return nil
	}

	fmt.Printf("Translating %d topics to %s\n", len(topicIDs), targetLang)

	// Worker pool
	jobs := make(chan string, len(topicIDs))
	results := make(chan error, len(topicIDs))

	for w := 0; w < workers; w++ {
		go func() {
			for id := range jobs {
				results <- translateSingleTopic(provider, repoPath, id, targetLang)
			}
		}()
	}

	for _, id := range topicIDs {
		jobs <- id
	}
	close(jobs)

	completed := 0
	var translateErrors []string
	for range topicIDs {
		err := <-results
		completed++
		if err != nil {
			translateErrors = append(translateErrors, err.Error())
		}
		fmt.Printf("\r  Progress: %d/%d", completed, len(topicIDs))
	}
	fmt.Println()

	fmt.Printf("Completed: %d/%d successful\n", len(topicIDs)-len(translateErrors), len(topicIDs))
	if len(translateErrors) > 0 {
		fmt.Fprintf(os.Stderr, "%d errors:\n", len(translateErrors))
		for _, e := range translateErrors {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
		}
	}
	return nil
}

// translateSingleTopic translates a topic YAML and its companion files
// (teaching notes, assessments, examples) to the target language, writing
// each translated file into translations/{lang}/ per id-conventions.md.
func translateSingleTopic(provider ai.Provider, repoPath, topicID, targetLang string) error {
	genCtx, err := generator.BuildContext(repoPath, topicID)
	if err != nil {
		return fmt.Errorf("building context for %s: %w", topicID, err)
	}

	topicFile, err := generator.FindTopicFile(repoPath, topicID)
	if err != nil {
		return fmt.Errorf("finding topic file for %s: %w", topicID, err)
	}

	topicsDir := filepath.Dir(topicFile)
	base := strings.TrimSuffix(filepath.Base(topicFile), filepath.Ext(topicFile))

	// 1. Translate topic YAML
	result, err := generator.Translate(context.Background(), provider, &genCtx.Topic, targetLang)
	if err != nil {
		return fmt.Errorf("translating %s: %w", topicID, err)
	}

	content := pipeline.StripCodeFences(result.Content)
	if err := generator.WriteTranslationFile(topicsDir, targetLang, base+".yaml", content); err != nil {
		return fmt.Errorf("writing translation for %s: %w", topicID, err)
	}
	fmt.Printf("  ✓ %s.yaml translated to %s\n", topicID, targetLang)

	// 2. Translate companion files if they exist
	companions := []string{
		base + ".teaching.md",
		base + ".assessments.yaml",
		base + ".examples.yaml",
	}

	for _, fileName := range companions {
		filePath := filepath.Join(topicsDir, fileName)
		if _, err := os.Stat(filePath); err != nil {
			continue // file doesn't exist, skip
		}

		res, err := generator.TranslateFile(context.Background(), provider, topicID, filePath, targetLang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ %s: %v\n", fileName, err)
			continue
		}

		translated := pipeline.StripCodeFences(res.Content)
		if err := generator.WriteTranslationFile(topicsDir, targetLang, fileName, translated); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ writing %s: %v\n", fileName, err)
			continue
		}
		fmt.Printf("  ✓ %s translated to %s\n", fileName, targetLang)
	}

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
		Short: "Create a new subject + subject_grade directory within an existing syllabus",
		Long: `Scaffold a subject directory with the new three-level structure:

  {subject_id}/subject.yaml
  {subject_id}/{subject_grade_id}/subject-grade.yaml
  {subject_id}/{subject_grade_id}/topics/

Examples:
  oss scaffold subject --syllabus malaysia-kssm --id malaysia-kssm-matematik --grade-id malaysia-kssm-matematik-tingkatan-3 --country malaysia
  oss scaffold subject --syllabus india-jee --id india-jee-mathematics --grade-id india-jee-mathematics-class-11 --country india`,
		RunE: runScaffoldSubject,
	}
	cmd.Flags().String("syllabus", "", "Syllabus ID (required)")
	cmd.Flags().String("id", "", "Subject ID — grade-less (required, e.g. malaysia-kssm-matematik)")
	cmd.Flags().String("grade-id", "", "Subject grade ID — with grade (e.g. malaysia-kssm-matematik-tingkatan-3)")
	cmd.Flags().String("country", "", "Country code (e.g. malaysia)")
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
	subjectGradeID, _ := cmd.Flags().GetString("grade-id")
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
		SyllabusID:      syllabusID,
		SubjectID:       subjectID,
		SubjectGradeID:  subjectGradeID,
		Country:         country,
		SourceText:      sourceText,
		OutputDir:       outputDir,
		GlobalSchemaDir: filepath.Join(outputDir, "schema"),
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
YAML files in the OSS repo.

Default mode (whole-PDF): sends the entire PDF content to a reasoning model
with a large context window. If a scaffold exists, topic names and IDs are
included as reference so the AI can locate each topic in the document.

Chunk mode (--chunk): splits the PDF by the given keyword and processes each
chunk in parallel with separate AI calls.

Examples:
  # Whole-PDF mode (default, more robust)
  oss import --pdf Tingkatan-1.pdf --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-1

  # Whole-PDF with topic name hints
  oss import --pdf Tingkatan-1.pdf --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-1 --from-text "1. Fungsi\n2. Algebra"

  # Chunk mode (legacy, for DSKP-style documents)
  oss import --pdf DSKP-KSSM-Matematik-Tingkatan-4.pdf --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4 --chunk TAJUK`,
		RunE: runImport,
	}
	cmd.Flags().String("pdf", "", "Path to PDF file (required)")
	cmd.Flags().String("syllabus", "", "Target syllabus ID (required, e.g. malaysia-kssm)")
	cmd.Flags().String("subject-grade", "", "Target subject grade ID (e.g. malaysia-kssm-matematik-tingkatan-4)")
	cmd.Flags().String("chunk", "", "Enable chunk mode: split PDF by this keyword (e.g. TAJUK) and process chunks in parallel")
	cmd.Flags().Int("workers", 3, "Number of parallel AI workers (chunk mode only)")
	cmd.Flags().Int("chunk-size", 2000, "Max tokens per chunk (chunk mode only)")
	cmd.Flags().Bool("pr", false, "Create a GitHub PR instead of writing to filesystem")
	cmd.Flags().Bool("force", false, "Overwrite existing topic files instead of AI-merging them")
	cmd.Flags().String("topic-id", "", "Import only the specified topic (e.g. MT4-01)")
	cmd.Flags().String("from-text", "", "Topic name reference text (one topic per line, used as hints for AI)")
	cmd.Flags().String("from-file", "", "Path to file containing topic name references (alternative to --from-text)")
	cmd.MarkFlagRequired("pdf")
	cmd.MarkFlagRequired("syllabus")
	return cmd
}

func runImport(cmd *cobra.Command, _ []string) error {
	pdfPath, _ := cmd.Flags().GetString("pdf")
	syllabusID, _ := cmd.Flags().GetString("syllabus")
	subjectGradeID, _ := cmd.Flags().GetString("subject-grade")
	filterTopicID, _ := cmd.Flags().GetString("topic-id")
	workers, _ := cmd.Flags().GetInt("workers")
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	createPR, _ := cmd.Flags().GetBool("pr")
	force, _ := cmd.Flags().GetBool("force")
	chunkKeyword, _ := cmd.Flags().GetString("chunk")
	fromText, _ := cmd.Flags().GetString("from-text")
	fromFile, _ := cmd.Flags().GetString("from-file")

	if filterTopicID != "" && subjectGradeID == "" {
		return fmt.Errorf("--topic-id requires --subject-grade flag")
	}

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
	fmt.Printf("Extracted %d characters (~%d tokens)\n", len(text), len(text)/4)

	// 2. Resolve output directory — search for existing subject topics dir
	// created by scaffold, fall back to a flat output dir.
	topicsDir, err := findSubjectTopicsDir(repoPath, subjectGradeID, subjectBaseID(subjectGradeID), syllabusID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		topicsDir = filepath.Join(repoPath, "import-output", syllabusID)
		fmt.Printf("Writing to fallback dir: %s\n", topicsDir)
	}
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		return fmt.Errorf("creating output dir %s: %w", topicsDir, err)
	}

	mode := pipeline.ModeWriteFS
	if createPR {
		mode = pipeline.ModeCreatePR
	}

	// Branch: chunk mode (--chunk) vs whole-PDF mode (default)
	if chunkKeyword != "" {
		return runImportChunkMode(cmd, text, importChunkOpts{
			syllabusID:     syllabusID,
			subjectGradeID: subjectGradeID,
			filterTopicID:  filterTopicID,
			chunkKeyword:   chunkKeyword,
			chunkSize:      chunkSize,
			workers:        workers,
			mode:           mode,
			force:          force,
			topicsDir:      topicsDir,
			repoPath:       repoPath,
			provider:       reasoningProvider,
			baseProvider:   provider,
		})
	}

	return runImportWholeMode(cmd, text, importWholeOpts{
		syllabusID:     syllabusID,
		subjectGradeID: subjectGradeID,
		filterTopicID:  filterTopicID,
		fromText:       fromText,
		fromFile:       fromFile,
		mode:           mode,
		force:          force,
		topicsDir:      topicsDir,
		repoPath:       repoPath,
		provider:       reasoningProvider,
		baseProvider:   provider,
	})
}

// importChunkOpts holds options for chunk-based import mode.
type importChunkOpts struct {
	syllabusID     string
	subjectGradeID string
	filterTopicID  string
	chunkKeyword   string
	chunkSize      int
	workers        int
	mode           pipeline.ExecutionMode
	force          bool
	topicsDir      string
	repoPath       string
	provider       ai.Provider
	baseProvider   ai.Provider // faster model for merge (non-reasoning)
}

// runImportChunkMode runs the legacy chunk-based import, splitting the PDF by a keyword
// (e.g. TAJUK) and processing each chunk in parallel.
func runImportChunkMode(cmd *cobra.Command, text string, opts importChunkOpts) error {
	fmt.Printf("Chunk mode: splitting by %q\n", opts.chunkKeyword)

	// Try DSKP-specific extraction first; fall back to generic chunker.
	var chunks []parser.Chunk
	if dskpTopics := extractDSKPTopics(text); len(dskpTopics) > 0 {
		fmt.Printf("Detected DSKP format: %d topics (chapter structure)\n", len(dskpTopics))
		chunks = dskpTopicsToChunks(dskpTopics)
	} else {
		chunks = parser.ChunkDocument(text, parser.ChunkOptions{
			MaxChunkSize: opts.chunkSize,
			SplitOn:      []string{"# ", "## ", "### ", "Chapter ", "Bab ", "BAB ", "BAHAGIAN ", "Bahagian ", opts.chunkKeyword},
		})
	}
	fmt.Printf("Split into %d chunks\n", len(chunks))

	// Filter chunks early when --topic-id is set.
	if opts.filterTopicID != "" && opts.subjectGradeID != "" {
		var filtered []parser.Chunk
		for _, c := range chunks {
			if topicFileID(opts.subjectGradeID, c.Heading, c.Index) == opts.filterTopicID {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no chunk matches --topic-id %s (available: %s)", opts.filterTopicID, availableTopicIDs(opts.subjectGradeID, chunks))
		}
		fmt.Printf("Filtered to %d chunk(s) matching topic %s\n", len(filtered), opts.filterTopicID)
		chunks = filtered
	}

	// Resolve the topic schema for this subject (subject-level override or global fallback).
	var topicSchemaContent string
	globalSchemaDir := filepath.Join(opts.repoPath, "schema")
	schemaResolver := validator.NewSchemaResolver(globalSchemaDir)
	subjectDir := validator.FindSubjectDir(filepath.Join(opts.topicsDir, "x.yaml"))
	schemaDir := validator.SubjectSchemaDir(subjectDir)
	if schemaPath, ok := schemaResolver.ResolveSchemaPath("topic", schemaDir); ok {
		if data, err := os.ReadFile(schemaPath); err == nil {
			topicSchemaContent = string(data)
			fmt.Printf("Using topic schema: %s\n", schemaPath)
		}
	}

	result, err := pipeline.ExecuteBulk(cmd.Context(), pipeline.BulkRequest{
		Chunks:         chunks,
		SyllabusID:     opts.syllabusID,
		SubjectGradeID: opts.subjectGradeID,
		Mode:           opts.mode,
		Source:         "cli",
		Workers:        opts.workers,
		Reporter:       pipeline.NewCLIReporter(),
		Provider:       opts.provider,
		TopicSchema:    topicSchemaContent,
	})
	if err != nil {
		return fmt.Errorf("bulk import: %w", err)
	}

	return writeImportResults(cmd.Context(), result.Topics, opts.baseProvider, importWriteOpts{
		subjectGradeID: opts.subjectGradeID,
		filterTopicID:  opts.filterTopicID,
		force:          opts.force,
		topicsDir:      opts.topicsDir,
		processedCount: result.ProcessedChunks,
		totalCount:     len(chunks),
		duration:       result.Duration,
		errors:         result.Errors,
	})
}

// importWholeOpts holds options for whole-PDF import mode.
type importWholeOpts struct {
	syllabusID     string
	subjectGradeID string
	filterTopicID  string
	fromText       string
	fromFile       string
	mode           pipeline.ExecutionMode
	force          bool
	topicsDir      string
	repoPath       string
	provider       ai.Provider
	baseProvider   ai.Provider // faster model for merge (non-reasoning)
}

// runImportWholeMode sends the entire PDF content to the AI as context,
// using the reasoning model's large context window. This is more robust than
// chunk mode because the AI sees the full document and can cross-reference
// topics, prerequisites, and learning areas.
func runImportWholeMode(cmd *cobra.Command, text string, opts importWholeOpts) error {
	fmt.Println("Whole-PDF mode: sending entire document to AI")

	// Resolve topic name reference text (--from-text or --from-file).
	topicRefText, err := resolveSourceText(opts.fromFile, opts.fromText)
	if err != nil {
		return fmt.Errorf("reading topic reference: %w", err)
	}

	// Load scaffold topic references if a scaffold exists.
	scaffoldTopics := loadScaffoldTopics(opts.topicsDir)
	if len(scaffoldTopics) > 0 {
		fmt.Printf("Found %d scaffold topic stubs as reference\n", len(scaffoldTopics))
	}

	// Estimate tokens and warn if the PDF is very large.
	estimatedTokens := len(text) / 4
	fmt.Printf("Estimated input: ~%d tokens\n", estimatedTokens)

	// Determine max output tokens from the reasoning model.
	maxOutputTokens := 16384 // reasonable default
	if rp, ok := opts.provider.(*ai.ReasoningProvider); ok {
		models := rp.Models()
		for _, m := range models {
			if m.MaxTokens > 0 {
				// Use a fraction of the model's context for output;
				// the rest is consumed by the input PDF.
				candidate := m.MaxTokens - estimatedTokens
				if candidate > maxOutputTokens {
					maxOutputTokens = candidate
				}
				break
			}
		}
	}
	// Cap output tokens at a practical ceiling.
	if maxOutputTokens > 65536 {
		maxOutputTokens = 65536
	}
	if maxOutputTokens < 4096 {
		maxOutputTokens = 4096
	}

	// Build metadata for the prompt.
	subjectID := opts.subjectGradeID
	if subjectID == "" {
		subjectID = opts.syllabusID
	}
	countryID := countryFromSubject(subjectID)
	language := languageForCountry(countryID)
	prefix := subjectPrefix(subjectID)
	grade := gradeNumber(subjectID)

	// Build topic reference section for the prompt.
	var topicRefSection string
	if len(scaffoldTopics) > 0 {
		var sb strings.Builder
		sb.WriteString("KNOWN TOPICS (from scaffold — use these IDs and match to content in the document):\n")
		for _, t := range scaffoldTopics {
			sb.WriteString(fmt.Sprintf("  - %s: %q", t.ID, t.Name))
			if t.NameEn != "" && t.NameEn != t.Name {
				sb.WriteString(fmt.Sprintf(" (%s)", t.NameEn))
			}
			sb.WriteString("\n")
		}
		topicRefSection = sb.String()
	}
	if topicRefText != "" {
		topicRefSection += "\nTOPIC NAME REFERENCE (use these as hints to identify topics in the document):\n" + topicRefText + "\n"
	}

	// Resolve the topic schema (subject-level override or global fallback)
	// so the AI knows which fields are required in the output.
	var topicSchemaContent string
	globalSchemaDir := filepath.Join(opts.repoPath, "schema")
	resolver := validator.NewSchemaResolver(globalSchemaDir)
	// Find subject dir from topicsDir (walk up to find subject.yaml).
	subjectDir := validator.FindSubjectDir(filepath.Join(opts.topicsDir, "x.yaml"))
	schemaDir := validator.SubjectSchemaDir(subjectDir)
	if schemaPath, ok := resolver.ResolveSchemaPath("topic", schemaDir); ok {
		if data, err := os.ReadFile(schemaPath); err == nil {
			topicSchemaContent = string(data)
			fmt.Printf("Using topic schema: %s\n", schemaPath)
		}
	}

	// Build the prompt.
	prompt := buildWholePDFPrompt(text, wholePDFPromptOpts{
		syllabusID:      opts.syllabusID,
		subjectGradeID:  opts.subjectGradeID,
		countryID:       countryID,
		language:        language,
		prefix:          prefix,
		grade:           grade,
		topicRefSection: topicRefSection,
		scaffoldTopics:  scaffoldTopics,
		filterTopicID:   opts.filterTopicID,
		topicSchema:     topicSchemaContent,
	})

	fmt.Printf("Calling AI with ~%d input tokens, max %d output tokens...\n", estimatedTokens, maxOutputTokens)

	start := time.Now()
	resp, err := opts.provider.Complete(cmd.Context(), ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are a curriculum analysis assistant. Extract structured learning content from source documents and generate OSS-format YAML."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   maxOutputTokens,
		Temperature: 0.3,
	})
	if err != nil {
		return fmt.Errorf("AI completion: %w", err)
	}
	duration := time.Since(start)
	fmt.Printf("AI responded in %s (%d input tokens, %d output tokens)\n",
		duration.Round(time.Second), resp.InputTokens, resp.OutputTokens)

	// Parse the multi-topic YAML output into individual topic results.
	topics := parseWholePDFOutput(resp.Content, opts.subjectGradeID, prefix, grade, topicSchemaContent)
	if len(topics) == 0 {
		return fmt.Errorf("AI returned no parseable topics — try chunk mode with --chunk")
	}
	fmt.Printf("Extracted %d topics from AI response\n", len(topics))

	// Post-process: fix topic IDs that the AI may have mangled (e.g. TP6-SC1-02 → SC1-02).
	// When scaffold topics exist, we have a known-good set of IDs. If the AI generated
	// an ID that contains a known scaffold ID as a substring, correct it.
	if len(scaffoldTopics) > 0 {
		knownIDs := make(map[string]bool, len(scaffoldTopics))
		for _, t := range scaffoldTopics {
			knownIDs[t.ID] = true
		}
		for i, tr := range topics {
			yamlID, _ := extractTopicIDAndName(tr.Output)
			if yamlID == "" || knownIDs[yamlID] {
				continue // already correct or unparseable
			}
			// Check if a known ID is embedded in the mangled ID (e.g. "TP6-SC1-02" contains "SC1-02").
			for kid := range knownIDs {
				if strings.Contains(yamlID, kid) {
					fmt.Printf("  fixing topic ID: %s → %s\n", yamlID, kid)
					topics[i].Output = strings.Replace(tr.Output, "id: "+yamlID, "id: "+kid, 1)
					break
				}
			}
		}
	}

	// Filter to the single requested topic when --topic-id is set.
	if opts.filterTopicID != "" {
		var filtered []pipeline.TopicResult
		for _, tr := range topics {
			yamlID, _ := extractTopicIDAndName(tr.Output)
			if yamlID == opts.filterTopicID {
				filtered = append(filtered, tr)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("AI response did not contain topic %s", opts.filterTopicID)
		}
		topics = filtered
		fmt.Printf("Filtered to %d topic(s) matching --topic-id %s\n", len(topics), opts.filterTopicID)
	}

	return writeImportResults(cmd.Context(), topics, opts.baseProvider, importWriteOpts{
		subjectGradeID: opts.subjectGradeID,
		filterTopicID:  opts.filterTopicID,
		force:          opts.force,
		topicsDir:      opts.topicsDir,
		processedCount: len(topics),
		totalCount:     len(topics),
		duration:       duration,
		errors:         nil,
	})
}

// importWriteOpts holds options shared by both import modes for writing results.
type importWriteOpts struct {
	subjectGradeID string
	filterTopicID  string
	force          bool
	topicsDir      string
	processedCount int
	totalCount     int
	duration       time.Duration
	errors         []error
}

// mergeJob describes one AI merge to run in parallel.
type mergeJob struct {
	outPath      string
	existingData string
	newContent   string
	heading      string
}

// mergeResult holds the outcome of a parallel merge.
type mergeResult struct {
	outPath  string
	content  string
	duration time.Duration
	err      error
}

// writeImportResults writes topic results to YAML files, merging with existing content
// when appropriate. This is shared by both chunk and whole-PDF import modes.
// AI merges use the base provider (faster than reasoning) and run in parallel.
func writeImportResults(ctx context.Context, topics []pipeline.TopicResult, provider ai.Provider, opts importWriteOpts) error {
	written := 0
	merged := 0
	var mergeJobs []mergeJob

	// First pass: write new files, replace forced files, collect merge jobs.
	for _, tr := range topics {
		if tr.Err != nil || strings.TrimSpace(tr.Output) == "" {
			continue
		}

		// Validate YAML is parseable before writing — skip broken AI output.
		var yamlCheck interface{}
		if err := yaml.Unmarshal([]byte(tr.Output), &yamlCheck); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ skipping topic %q: invalid YAML: %v\n", tr.Heading, err)
			continue
		}

		// Prefer the id embedded in the YAML output; fall back to derivation.
		yamlID, _ := extractTopicIDAndName(tr.Output)
		var fileID string
		if yamlID != "" {
			fileID = yamlID
		} else if opts.subjectGradeID != "" {
			fileID = topicFileID(opts.subjectGradeID, tr.Heading, tr.ChunkIndex)
		} else {
			slug := importSlug(tr.Heading)
			if slug == "" {
				slug = fmt.Sprintf("topic-%02d", tr.ChunkIndex+1)
			}
			fileID = slug
		}

		// Skip topics that don't match --topic-id filter.
		if opts.filterTopicID != "" && fileID != opts.filterTopicID {
			continue
		}

		outPath := filepath.Join(opts.topicsDir, fileID+".yaml")

		if existingData, readErr := os.ReadFile(outPath); readErr == nil {
			if opts.force {
				if err := os.WriteFile(outPath, []byte(tr.Output), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "  ⚠ writing %s: %v\n", outPath, err)
					continue
				}
				fmt.Printf("  replaced: %s\n", outPath)
				merged++
			} else {
				// Queue for parallel AI merge.
				mergeJobs = append(mergeJobs, mergeJob{
					outPath:      outPath,
					existingData: string(existingData),
					newContent:   tr.Output,
					heading:      tr.Heading,
				})
			}
		} else {
			if err := os.WriteFile(outPath, []byte(tr.Output), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ writing %s: %v\n", outPath, err)
				continue
			}
			fmt.Printf("  wrote: %s\n", outPath)
			written++
		}
	}

	// Second pass: run AI merges in parallel.
	if len(mergeJobs) > 0 {
		fmt.Printf("Merging %d existing files in parallel...\n", len(mergeJobs))
		results := make([]mergeResult, len(mergeJobs))
		var wg sync.WaitGroup
		for i, job := range mergeJobs {
			wg.Add(1)
			go func(idx int, j mergeJob) {
				defer wg.Done()
				start := time.Now()
				content, err := mergeTopicYAML(ctx, provider, j.existingData, j.newContent, j.heading)
				results[idx] = mergeResult{
					outPath:  j.outPath,
					content:  content,
					duration: time.Since(start),
					err:      err,
				}
			}(i, job)
		}
		wg.Wait()

		// Write results and report.
		for _, r := range results {
			if r.err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ AI merge failed for %s: %v — skipping\n", r.outPath, r.err)
				continue
			}
			if err := os.WriteFile(r.outPath, []byte(r.content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ writing merged %s: %v\n", r.outPath, err)
				continue
			}
			fmt.Printf("  merged in %s: %s\n", r.duration.Round(10*time.Millisecond), r.outPath)
			merged++
		}
	}

	fmt.Printf("\nProcessed %d/%d topics in %s — wrote %d new, merged %d existing file(s) in %s\n",
		opts.processedCount, opts.totalCount, opts.duration.Round(time.Second),
		written, merged, opts.topicsDir)

	if len(opts.errors) > 0 {
		fmt.Fprintf(os.Stderr, "%d topics failed:\n", len(opts.errors))
		for _, e := range opts.errors {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
		}
	}
	return nil
}

// findSubjectTopicsDir searches under repoPath for an existing topics directory
// following the new directory structure: {subject_id}/{subject_grade_id}/topics/.
// It searches for a directory named subjectGradeID (or subjectID, or syllabusID
// as fallbacks) and returns the topics/ path within it.
func findSubjectTopicsDir(repoPath, subjectGradeID, subjectID, syllabusID string) (string, error) {
	if subjectGradeID == "" && subjectID == "" && syllabusID == "" {
		return "", fmt.Errorf("no subject grade, subject, or syllabus ID provided")
	}

	// Search in order: subjectGradeID > subjectID > syllabusID
	searchID := subjectGradeID
	if searchID == "" {
		searchID = subjectID
	}
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

// subjectBaseID strips the grade portion from a subject_grade_id to get the
// grade-less subject_id. e.g. "malaysia-kssm-matematik-tingkatan-3" → "malaysia-kssm-matematik".
// If no grade portion is detected, returns the input unchanged.
func subjectBaseID(subjectGradeID string) string {
	if subjectGradeID == "" {
		return ""
	}
	parts := strings.Split(subjectGradeID, "-")
	for i := len(parts) - 1; i >= 1; i-- {
		p := parts[i]
		allDigits := len(p) > 0 && len(p) <= 2
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			// The word before the number is the grade label; strip both.
			return strings.Join(parts[:i-1], "-")
		}
	}
	return subjectGradeID
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
			// Reassemble the topic line after TAJUK.  PDF extraction
			// often fragments "13.0  KEBARANGKALIAN MUDAH" into separate
			// lines like "1", "3", ".", "0", "", "KEBARANGKALIAN MUDAH".
			// We collect short numeric/dot fragments and join them, then
			// append the first real text line as the topic name.
			topicLine, ni := reassembleDSKPTopicLine(lines, i+1)
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

// reassembleDSKPTopicLine collects fragments after a TAJUK marker.
// PDF extraction often splits "13.0  KEBARANGKALIAN MUDAH" across lines:
//
//	"1"
//	"3"
//	"."
//	"0"
//	""
//	"KEBARANGKALIAN MUDAH"
//
// This function joins short digit/dot fragments into a number token, then
// appends the first non-fragment line as the topic name.  It returns the
// reassembled topic line and the index of the first line NOT consumed.
func reassembleDSKPTopicLine(lines []string, start int) (string, int) {
	var numParts []string
	idx := start
	// Collect short numeric/dot fragments (single chars or small tokens).
	for idx < len(lines) {
		s := strings.TrimSpace(lines[idx])
		if s == "" {
			idx++
			continue
		}
		// A fragment is a short token (≤3 chars) made of digits and dots.
		if len(s) <= 3 && isNumericFragment(s) {
			numParts = append(numParts, s)
			idx++
			continue
		}
		break
	}

	// If we collected fragments, join them and look for the name line.
	if len(numParts) > 0 {
		number := strings.Join(numParts, "")
		// Find the next non-empty line for the topic name.
		for idx < len(lines) {
			s := strings.TrimSpace(lines[idx])
			if s == "" {
				idx++
				continue
			}
			return number + " " + s, idx
		}
		return number, idx
	}

	// No fragments found — fall back to returning the first non-empty line.
	for idx < len(lines) {
		s := strings.TrimSpace(lines[idx])
		if s != "" {
			return s, idx
		}
		idx++
	}
	return "", idx
}

// isNumericFragment returns true if s consists only of digits and dots.
func isNumericFragment(s string) bool {
	for _, r := range s {
		if r != '.' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
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

// availableTopicIDs returns a comma-separated list of topic IDs for the given chunks.
func availableTopicIDs(subjectID string, chunks []parser.Chunk) string {
	ids := make([]string, len(chunks))
	for i, c := range chunks {
		ids[i] = topicFileID(subjectID, c.Heading, c.Index)
	}
	return strings.Join(ids, ", ")
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

// countryFromSubject extracts the country portion from a subject ID.
// e.g. "malaysia-kssm-matematik-tingkatan-4" → "malaysia"
func countryFromSubject(id string) string {
	if idx := strings.Index(id, "-"); idx > 0 {
		return id[:idx]
	}
	return id
}

// languageForCountry returns the BCP 47 language code for a country's MOE language.
func languageForCountry(countryID string) string {
	langs := map[string]string{
		"malaysia":  "ms",
		"indonesia": "id",
		"japan":     "ja",
		"uae":       "ar",
		"thailand":  "th",
		"vietnam":   "vi",
		"china":     "zh-hans",
		"korea":     "ko",
	}
	if l, ok := langs[countryID]; ok {
		return l
	}
	return "en"
}

// scaffoldTopic holds a topic reference loaded from scaffold stubs.
type scaffoldTopic struct {
	ID     string
	Name   string
	NameEn string
}

// loadScaffoldTopics reads existing scaffold topic stubs from the topics directory
// and returns their IDs and names for use as AI reference.
func loadScaffoldTopics(topicsDir string) []scaffoldTopic {
	entries, err := os.ReadDir(topicsDir)
	if err != nil {
		return nil
	}

	var topics []scaffoldTopic
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		// Skip companion files (teaching, assessments, examples).
		if strings.Contains(e.Name(), ".teaching.") ||
			strings.Contains(e.Name(), ".assessments.") ||
			strings.Contains(e.Name(), ".examples.") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(topicsDir, e.Name()))
		if err != nil {
			continue
		}

		// Extract id, name, and name_en from the YAML stub.
		t := scaffoldTopic{}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "id:") {
				t.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			} else if strings.HasPrefix(line, "name:") && !strings.HasPrefix(line, "name_en:") {
				t.Name = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "name:")), "\"")
			} else if strings.HasPrefix(line, "name_en:") {
				t.NameEn = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "name_en:")), "\"")
			}
		}
		if t.ID != "" {
			topics = append(topics, t)
		}
	}

	// Sort by ID for deterministic ordering.
	sort.Slice(topics, func(i, j int) bool { return topics[i].ID < topics[j].ID })
	return topics
}

// wholePDFPromptOpts holds parameters for building the whole-PDF import prompt.
type wholePDFPromptOpts struct {
	syllabusID      string
	subjectGradeID  string
	countryID       string
	language        string
	prefix          string
	grade           string
	topicRefSection string
	scaffoldTopics  []scaffoldTopic
	filterTopicID   string
	topicSchema     string // resolved topic JSON Schema content (if available)
}

// buildWholePDFPrompt constructs the prompt for whole-PDF import mode.
// The entire PDF text is included as context so the AI can cross-reference
// topics, prerequisites, and learning areas across the full document.
func buildWholePDFPrompt(pdfText string, opts wholePDFPromptOpts) string {
	subjectID := opts.subjectGradeID
	if subjectID == "" {
		subjectID = opts.syllabusID
	}

	// Build topic ID instructions.
	topicIDInstruction := fmt.Sprintf(
		"Topic IDs MUST follow the format: %s%s-NN (e.g. %s%s-01, %s%s-02, ...)",
		opts.prefix, opts.grade, opts.prefix, opts.grade, opts.prefix, opts.grade,
	)
	if len(opts.scaffoldTopics) > 0 {
		var idList []string
		for _, t := range opts.scaffoldTopics {
			idList = append(idList, t.ID)
		}
		topicIDInstruction = fmt.Sprintf(
			"CRITICAL — Topic IDs: You MUST use EXACTLY one of these IDs for each topic: %s. "+
				"Match each document topic to the closest scaffold topic by name. "+
				"DO NOT invent new IDs, DO NOT combine theme/chapter numbers with these IDs, DO NOT add prefixes.",
			strings.Join(idList, ", "),
		)
	}

	filterInstruction := ""
	if opts.filterTopicID != "" {
		filterInstruction = fmt.Sprintf("\n\nIMPORTANT: Only extract and generate YAML for topic %s. Ignore all other topics.", opts.filterTopicID)
	}

	// Build the schema section: if a resolved topic schema is available,
	// include it so the AI generates all required fields (e.g. content_standards).
	schemaSection := ""
	if opts.topicSchema != "" {
		schemaSection = fmt.Sprintf(`
JSON SCHEMA (your output MUST conform to this schema — include ALL required fields):
`+"```json\n%s\n```"+`
`, opts.topicSchema)
		if fieldGuide := pipeline.ExtractSchemaDescriptions(opts.topicSchema); fieldGuide != "" {
			schemaSection += "\n" + fieldGuide
		}
	}

	prompt := fmt.Sprintf(`You are extracting ALL curriculum topics from an educational document and generating OSS-format YAML files.

FULL DOCUMENT CONTENT:
%s

%s
METADATA:
- syllabus_id: %s
- subject_grade_id: %s
- country_id: %s
- language: %s
- %s%s
%s
INSTRUCTIONS:
For EACH topic found in the document, generate a complete YAML document. Separate each topic with a line containing only "---".

Each topic YAML must include ALL fields marked as "required" in the JSON Schema above (if provided). At minimum, each topic must contain:
- id, name, name_en, subject_id, syllabus_id, country_id, language, difficulty
- learning_objectives (array with id, text, text_en, bloom)
- content_standards (if required by schema — extract from the document's content/learning standards)
- quality_level, provenance

Use these fixed values:
- subject_id: %s
- syllabus_id: %s
- country_id: %s
- language: %s
- provenance: ai-assisted
- quality_level: 1

RULES:
- Output ONLY valid YAML — no markdown fences, no explanatory text before or after
- Separate each topic with a line containing ONLY "---"
- Extract ALL topics from the document, not just the first few
- Extract ALL learning objectives from each topic section
- Topic IDs: if KNOWN TOPICS are listed above, use EXACTLY those IDs — never invent new IDs or modify them
- name MUST be in the document language (%s), name_en MUST be English
- learning_objectives text MUST be in the document language, text_en MUST be English
- bloom levels: remember | understand | apply | analyze | evaluate | create
- Infer bloom levels from verbs: list/recall/define → remember, explain/describe → understand, solve/calculate → apply, differentiate/examine → analyze, assess/justify → evaluate, design/develop → create
- Prerequisites should reference topic IDs from this same document where applicable`,
		pdfText,
		opts.topicRefSection,
		opts.syllabusID, subjectID, opts.countryID, opts.language,
		topicIDInstruction, filterInstruction,
		schemaSection,
		subjectID, opts.syllabusID, opts.countryID, opts.language,
		opts.language,
	)

	return prompt
}

// parseWholePDFOutput splits the AI's multi-topic YAML response into individual
// TopicResult entries. Topics are separated by "---" lines.
func parseWholePDFOutput(output string, subjectGradeID, prefix, grade, topicSchema string) []pipeline.TopicResult {
	output = pipeline.StripCodeFences(output)

	// Split by YAML document separator.
	docs := strings.Split(output, "\n---")
	var results []pipeline.TopicResult

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" || doc == "---" {
			continue
		}
		// Remove leading "---" if present (first document).
		doc = strings.TrimPrefix(doc, "---")
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// Sanitize AI-generated YAML issues.
		doc = pipeline.FixYAMLColonSpacing(doc)
		doc = pipeline.RemoveDuplicateKeys(doc)
		doc = pipeline.SanitizeYAMLQuoting(doc)

		// Extract topic ID and name from the YAML for heading/metadata.
		topicID, topicName := extractTopicIDAndName(doc)

		// Build heading from topic ID or name.
		heading := topicName
		if heading == "" {
			heading = topicID
		}
		if heading == "" {
			heading = fmt.Sprintf("topic-%02d", i+1)
		}

		// If the YAML has a topic ID, use it directly as the heading
		// so topicFileID() doesn't need to re-derive it.
		if topicID != "" {
			heading = topicID + " " + topicName
		}

		// Ensure required fields.
		if topicID != "" {
			doc = pipeline.EnsureTopicFields(doc, topicID)
		}

		// Enforce schema constraints: add missing required fields, quote strings.
		if topicSchema != "" {
			doc = pipeline.EnforceSchemaRequiredFields(doc, topicSchema)
			doc = pipeline.EnforceStringQuoting(doc, topicSchema)
		}

		results = append(results, pipeline.TopicResult{
			ChunkIndex: i,
			Heading:    heading,
			Output:     doc,
		})
	}

	return results
}

// extractTopicIDAndName extracts the id and name fields from a YAML string
// using simple line scanning (avoids full YAML parse for robustness).
func extractTopicIDAndName(yamlContent string) (id, name string) {
	for _, line := range strings.Split(yamlContent, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "id:") && !strings.Contains(line, "_") {
			id = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			id = strings.Trim(id, "\"")
		} else if strings.HasPrefix(line, "name:") && !strings.HasPrefix(line, "name_en:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			name = strings.Trim(name, "\"")
		}
		if id != "" && name != "" {
			break
		}
	}
	return id, name
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
	return pipeline.StripCodeFences(resp.Content), nil
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
		if r.SchemaPath != "" {
			fmt.Printf("     schema: %s\n", r.SchemaPath)
		}
		for _, e := range r.Errors {
			fmt.Printf("     → %s\n", e)
		}
	}
}
