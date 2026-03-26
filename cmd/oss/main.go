package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
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
