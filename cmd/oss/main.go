package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		Use:   "teaching-notes <topic-path>",
		Short: "Generate teaching notes for a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
}

func generateAssessmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assessments <topic-path>",
		Short: "Generate assessment questions for a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
	cmd.Flags().IntP("count", "c", 5, "Number of questions to generate")
	cmd.Flags().StringP("difficulty", "d", "medium", "Difficulty level: easy, medium, hard")
	return cmd
}

func generateExamplesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "examples <topic-path>",
		Short: "Generate worked examples for a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
	cmd.Flags().IntP("count", "c", 3, "Number of examples to generate")
	return cmd
}

func qualityCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quality [path]",
		Short: "Generate quality report for curriculum content",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
}

func translateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate topic content to another language",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not yet implemented")
		},
	}
	cmd.Flags().String("topic", "", "Path to topic file")
	cmd.Flags().String("to", "", "Target language code (e.g., ms, zh, ta)")
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
