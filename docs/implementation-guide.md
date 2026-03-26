# Implementation Guide — OSS Bot

> **Companion to:** [development-timeline.md](development-timeline.md)
> **Architecture reference:** [technical-plan.md](technical-plan.md)
> **Duration:** Day 16 – Day 30 (Weeks 4-6)
> **Scope:** CLI validator, AI generation pipeline, GitHub Bot, Web Portal

This guide provides step-by-step executable instructions for every day of the oss-bot development timeline. Each day includes entry criteria, exact file paths, code templates, test specifications, validation commands, and exit checklists.

## How to Use This Guide

1. Work through days sequentially — each day builds on the previous
2. Check **entry criteria** before starting a day
3. **Write tests first** — every feature follows TDD (test → implement → verify)
4. Complete all tasks, run **validation commands**
5. Verify all **exit criteria** checkboxes before moving to the next day
6. Track cumulative progress in the dashboard at the bottom of each day

### Task Owner Legend

| Icon | Owner | Meaning |
|------|-------|---------|
| 🤖 | Developer / AI Agent | Can be executed autonomously |
| 🧑 | Education Lead | Requires human educator expertise |
| 🧑🤖 | Collaborative | AI drafts, educator reviews and edits |

### TDD Workflow (Mandatory)

Every feature follows this strict cycle:

```
1. Write tests first     → define expected behavior
2. Run tests (RED)       → confirm tests fail
3. Implement             → write minimum code to pass
4. Run package tests     → go test ./internal/<package>/...
5. Run FULL test suite   → go test ./...
6. Never skip step 5     → every feature must pass the full suite
```

---

## Prerequisites

### Required Tools

```bash
# Go 1.22+ (backend)
go version   # Expected: go1.22.x or higher

# Node.js 20 LTS (web portal)
node --version   # Expected: v20.x.x

# golangci-lint (Go linter)
golangci-lint --version   # Expected: ≥1.55

# Docker + Docker Compose (deployment)
docker --version && docker compose version

# Reasoning model API key (for bulk import, optional)
# Supported: Kimi K2.5, Qwen 3.5, OpenAI o3-mini
# Set via OSS_REASONING_API_KEY and OSS_REASONING_PROVIDER
```

### Verify Setup

```bash
# All should succeed without errors
go version && node --version && golangci-lint --version && docker --version
```

---

## WEEKS 1-3 — NO OSS-BOT WORK

OSS Bot is not needed during the content validation phase (Weeks 1-3). During this period, curriculum content is created manually in the [p-n-ai/oss](https://github.com/p-n-ai/oss) repository.

**Why wait:** Building tooling before the content schema stabilizes is premature. The content format must be validated with real curriculum data first.

---

## WEEK 4 — CLI TOOL + AI GENERATION PIPELINE

### Day 16 — Initialize Repo + CLI Scaffold

**Entry criteria:** Repository exists with documentation only (README.md, CLAUDE.md, AGENTS.md, docs/). No Go code exists yet.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 16.1 | `B-W4D16-1` | Initialize Go module + directory structure | 🤖 | `go.mod`, directories |
| 16.2 | `B-W4D16-2` | CLI scaffold with cobra: root + subcommands | 🤖 | `cmd/oss/main.go`, `cmd/bot/main.go` |
| 16.3 | `B-W4D16-3` | Schema validator core | 🤖 | `internal/validator/validator.go` |
| 16.4 | `B-W4D16-4` | `oss validate` command | 🤖 | CLI wiring |

#### 16.1 — Initialize Go Module + Directory Structure

```bash
# Initialize Go module
go mod init github.com/p-n-ai/oss-bot

# Create directory structure
mkdir -p cmd/oss cmd/bot
mkdir -p internal/{ai,generator,validator,parser,github,api,pipeline,output}
mkdir -p prompts
mkdir -p scripts
mkdir -p deploy/docker
mkdir -p .github/workflows
```

#### 16.2 — CLI Scaffold with Cobra

Install cobra dependency:

```bash
go get github.com/spf13/cobra@latest
```

**File:** `cmd/oss/main.go`

```go
package main

import (
	"fmt"
	"os"

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
			// Placeholder — implemented on Day 19
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
			// Placeholder — implemented on Day 19
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
			// Placeholder — implemented on Day 19
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
			// Placeholder — implemented on Day 17
			return fmt.Errorf("not yet implemented")
		},
	}
}

func translateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translate",
		Short: "Translate topic content to another language",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Placeholder — implemented on Day 20
			return fmt.Errorf("not yet implemented")
		},
	}
	cmd.Flags().String("topic", "", "Path to topic file")
	cmd.Flags().String("to", "", "Target language code (e.g., ms, zh, ta)")
	return cmd
}
```

The `runValidate` function is wired in task 16.4 after the validator is built.

**File:** `cmd/bot/main.go`

```go
package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Placeholder — webhook server implemented in Week 5
	fmt.Println("oss-bot server — not yet implemented")
	fmt.Println("See: go run ./cmd/oss for CLI commands")
	os.Exit(0)
}
```

#### 16.3 — Schema Validator Core (TDD)

**Step 1: Write tests first**

**File:** `internal/validator/validator_test.go`

```go
package validator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestNewValidator(t *testing.T) {
	schemaDir := setupTestSchemas(t)

	v, err := validator.New(schemaDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if v == nil {
		t.Fatal("New() returned nil validator")
	}
}

func TestValidateFile_ValidTopic(t *testing.T) {
	schemaDir := setupTestSchemas(t)
	v, err := validator.New(schemaDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	topicFile := createTempYAML(t, validTopicYAML)

	result, err := v.ValidateFile(topicFile, "topic")
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	if !result.Valid {
		t.Errorf("ValidateFile() expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFile_InvalidTopic(t *testing.T) {
	schemaDir := setupTestSchemas(t)
	v, err := validator.New(schemaDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	topicFile := createTempYAML(t, invalidTopicYAML)

	result, err := v.ValidateFile(topicFile, "topic")
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	if result.Valid {
		t.Error("ValidateFile() expected invalid, got valid")
	}
	if len(result.Errors) == 0 {
		t.Error("ValidateFile() expected errors, got none")
	}
}

func TestValidateDir(t *testing.T) {
	schemaDir := setupTestSchemas(t)
	v, err := validator.New(schemaDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dir := t.TempDir()

	// Create a topics subdirectory with valid YAML
	topicsDir := filepath.Join(dir, "topics", "algebra")
	if err := os.MkdirAll(topicsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(topicsDir, "01-test.yaml"), validTopicYAML)

	results, err := v.ValidateDir(dir)
	if err != nil {
		t.Fatalf("ValidateDir() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("ValidateDir() returned no results")
	}
}

func TestDetectSchemaType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"syllabus", "curricula/malaysia/kssm/syllabus.yaml", "syllabus"},
		{"subject", "curricula/malaysia/kssm/subjects/algebra.yaml", "subject"},
		{"topic", "curricula/malaysia/kssm/topics/algebra/01-test.yaml", "topic"},
		{"assessments", "curricula/malaysia/kssm/topics/algebra/01-test.assessments.yaml", "assessments"},
		{"examples", "curricula/malaysia/kssm/topics/algebra/01-test.examples.yaml", "examples"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.DetectSchemaType(tt.path)
			if got != tt.expected {
				t.Errorf("DetectSchemaType(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

// --- Test helpers ---

const validTopicYAML = `
id: F1-01
name: "Test Topic"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: LO1
    text: "Test objective"
    bloom: understand
quality_level: 1
provenance: human
`

const invalidTopicYAML = `
id: invalid id with spaces
name: ""
`

func setupTestSchemas(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Minimal topic schema for testing
	topicSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["id", "name", "subject_id", "syllabus_id", "difficulty", "learning_objectives", "quality_level", "provenance"],
		"properties": {
			"id": { "type": "string", "pattern": "^[A-Z][0-9]+-[0-9]{2}$" },
			"name": { "type": "string", "minLength": 1 },
			"subject_id": { "type": "string" },
			"syllabus_id": { "type": "string" },
			"difficulty": { "type": "string", "enum": ["beginner", "intermediate", "advanced"] },
			"learning_objectives": {
				"type": "array", "minItems": 1,
				"items": {
					"type": "object",
					"required": ["id", "text", "bloom"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string", "minLength": 1 },
						"bloom": { "type": "string", "enum": ["remember", "understand", "apply", "analyze", "evaluate", "create"] }
					},
					"additionalProperties": false
				}
			},
			"quality_level": { "type": "integer", "minimum": 0, "maximum": 5 },
			"provenance": { "type": "string", "enum": ["human", "ai-assisted", "ai-generated", "ai-observed"] }
		},
		"additionalProperties": true
	}`

	writeFile(t, filepath.Join(dir, "topic.schema.json"), topicSchema)
	return dir
}

func createTempYAML(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "test.yaml")
	writeFile(t, f, content)
	return f
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
```

**Step 2: Implement**

**File:** `internal/validator/validator.go`

```go
// Package validator provides JSON Schema validation for OSS curriculum YAML files.
package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// ValidationResult holds the result of validating a single file.
type ValidationResult struct {
	File   string
	Type   string
	Valid  bool
	Errors []string
}

// Validator validates YAML files against JSON Schemas.
type Validator struct {
	schemas map[string]*jsonschema.Schema
}

// New creates a Validator by loading all schemas from the given directory.
func New(schemaDir string) (*Validator, error) {
	v := &Validator{
		schemas: make(map[string]*jsonschema.Schema),
	}

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("reading schema directory: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".schema.json")
		path := filepath.Join(schemaDir, entry.Name())

		schema, err := compiler.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("compiling schema %s: %w", name, err)
		}

		v.schemas[name] = schema
	}

	return v, nil
}

// ValidateFile validates a single YAML file against the specified schema type.
func (v *Validator) ValidateFile(filePath, schemaType string) (*ValidationResult, error) {
	result := &ValidationResult{
		File: filePath,
		Type: schemaType,
	}

	schema, ok := v.schemas[schemaType]
	if !ok {
		return nil, fmt.Errorf("unknown schema type: %s (available: %v)", schemaType, v.SchemaTypes())
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse YAML to generic interface
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		result.Valid = false
		result.Errors = []string{fmt.Sprintf("YAML parse error: %v", err)}
		return result, nil
	}

	// Convert YAML types to JSON-compatible types
	jsonData := convertYAMLToJSON(yamlData)

	// Validate against schema
	if err := schema.Validate(jsonData); err != nil {
		result.Valid = false
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			result.Errors = flattenValidationErrors(ve)
		} else {
			result.Errors = []string{err.Error()}
		}
		return result, nil
	}

	result.Valid = true
	return result, nil
}

// ValidateDir validates all YAML files in a directory tree, auto-detecting schema types.
func (v *Validator) ValidateDir(dir string) ([]ValidationResult, error) {
	var results []ValidationResult

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		// Skip teaching notes markdown files
		if strings.HasSuffix(path, ".teaching.md") {
			return nil
		}

		schemaType := DetectSchemaType(path)
		if schemaType == "" {
			return nil // Skip files we can't classify
		}

		if _, ok := v.schemas[schemaType]; !ok {
			return nil // Skip if schema not loaded
		}

		result, err := v.ValidateFile(path, schemaType)
		if err != nil {
			results = append(results, ValidationResult{
				File:   path,
				Type:   schemaType,
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return nil
		}
		results = append(results, *result)
		return nil
	})

	return results, err
}

// SchemaTypes returns the list of loaded schema type names.
func (v *Validator) SchemaTypes() []string {
	types := make([]string, 0, len(v.schemas))
	for t := range v.schemas {
		types = append(types, t)
	}
	return types
}

// DetectSchemaType determines the schema type from a file path.
func DetectSchemaType(path string) string {
	base := filepath.Base(path)

	switch {
	case base == "syllabus.yaml" || base == "syllabus.yml":
		return "syllabus"
	case strings.HasSuffix(base, ".assessments.yaml"):
		return "assessments"
	case strings.HasSuffix(base, ".examples.yaml"):
		return "examples"
	case strings.Contains(path, "subjects/") || strings.Contains(path, "subjects\\"):
		return "subject"
	case strings.Contains(path, "topics/") || strings.Contains(path, "topics\\"):
		return "topic"
	case strings.Contains(path, "concepts/"):
		return "concept"
	case strings.Contains(path, "taxonomy/"):
		return "taxonomy"
	default:
		return ""
	}
}

// convertYAMLToJSON converts YAML-parsed data to JSON-compatible types.
// YAML parses integers as int, but JSON Schema expects float64.
// YAML parses maps as map[string]interface{}, which is already JSON-compatible.
func convertYAMLToJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = convertYAMLToJSON(v)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[fmt.Sprint(k)] = convertYAMLToJSON(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = convertYAMLToJSON(v)
		}
		return result
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	default:
		return val
	}
}

// flattenValidationErrors extracts human-readable error messages from validation errors.
func flattenValidationErrors(ve *jsonschema.ValidationError) []string {
	var errs []string

	if ve.Message != "" {
		location := ve.InstanceLocation
		if location == "" {
			location = "(root)"
		}
		errs = append(errs, fmt.Sprintf("%s: %s", location, ve.Message))
	}

	for _, cause := range ve.Causes {
		errs = append(errs, flattenValidationErrors(cause)...)
	}

	return errs
}

// MarshalYAMLToJSON converts a YAML file's content to a JSON-compatible map.
func MarshalYAMLToJSON(data []byte) (interface{}, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}
	return convertYAMLToJSON(yamlData), nil
}

// PrettyJSON converts an interface to formatted JSON string (for debugging).
func PrettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(b)
}
```

#### 16.4 — Wire `oss validate` Command

Add the `runValidate` function to `cmd/oss/main.go` (append before the closing of the file, or add as a method):

**Add to:** `cmd/oss/main.go`

```go
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
```

Add required imports to `cmd/oss/main.go`:

```go
import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/p-n-ai/oss-bot/internal/validator"
	"github.com/spf13/cobra"
)
```

#### 16.5 — Create Makefile

**File:** `Makefile`

```makefile
.PHONY: test lint build-cli build-bot docker setup

# Testing
test:
	go test ./...

test-v:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting
lint:
	golangci-lint run ./...

# Building
build-cli:
	CGO_ENABLED=0 go build -o bin/oss ./cmd/oss

build-bot:
	CGO_ENABLED=0 go build -o bin/oss-bot ./cmd/bot

build: build-cli build-bot

# Docker
docker:
	docker build -f deploy/docker/Dockerfile -t oss-bot .

# Setup
setup:
	cp -n .env.example .env 2>/dev/null || true
	go mod download
	@echo "Setup complete. Edit .env with your configuration."
```

#### 16.6 — Create `.env.example`

**File:** `.env.example`

```bash
# OSS Bot Configuration
# Copy to .env and fill in your values

# --- Required (CLI) ---
OSS_REPO_PATH=./oss                    # Path to local OSS clone

# --- Required (Bot) ---
OSS_GITHUB_APP_ID=
OSS_GITHUB_PRIVATE_KEY_PATH=
OSS_GITHUB_WEBHOOK_SECRET=

# --- Required (All) ---
OSS_AI_PROVIDER=ollama                  # openai | anthropic | ollama
OSS_AI_API_KEY=                         # Not needed for Ollama
OSS_REPO_OWNER=p-n-ai
OSS_REPO_NAME=oss

# --- Optional ---
OSS_AI_OLLAMA_URL=http://localhost:11434
OSS_AI_MODEL=                           # Override default model
OSS_WEB_PORT=3001
OSS_BOT_PORT=8090
OSS_LOG_LEVEL=info                      # debug | info | warn | error
OSS_PROMPTS_DIR=./prompts
```

#### 16.7 — Create CI Workflow

**File:** `.github/workflows/ci.yml`

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test ./...

      - name: Run linter
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Build CLI
        run: CGO_ENABLED=0 go build -o bin/oss ./cmd/oss

      - name: Build Bot
        run: CGO_ENABLED=0 go build -o bin/oss-bot ./cmd/bot
```

#### Day 16 Validation

```bash
# Install dependencies
go mod tidy

# Run tests (must pass)
go test ./...

# Build both binaries
go build ./cmd/oss
go build ./cmd/bot

# Verify CLI scaffold
go run ./cmd/oss --help
go run ./cmd/oss validate --help
go run ./cmd/oss generate --help
```

#### Day 16 Exit Criteria

- [ ] `go.mod` exists with Go 1.22+
- [ ] Directory structure created: `cmd/`, `internal/`, `prompts/`, `scripts/`, `deploy/`
- [ ] `cmd/oss/main.go` builds with root + subcommands (validate, generate, quality, translate)
- [ ] `cmd/bot/main.go` builds (placeholder)
- [ ] `internal/validator/validator.go` implemented with JSON Schema validation
- [ ] `internal/validator/validator_test.go` passes all tests
- [ ] `go test ./...` passes with zero failures
- [ ] `Makefile`, `.env.example`, `.github/workflows/ci.yml` created

**Progress:** CLI scaffold | 1 package (validator) | Tests passing | CI workflow

---

### Day 17 — Validation Tools

**Entry criteria:** Day 16 complete. `go test ./...` passes. Validator core works with JSON Schema.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 17.1 | `B-W4D17-1` | Bloom's taxonomy validator | 🤖 | `internal/validator/bloom.go` |
| 17.2 | `B-W4D17-2` | Prerequisite cycle detector | 🤖 | `internal/validator/prerequisites.go` |
| 17.3 | `B-W4D17-3` | Duplicate content detector | 🤖 | `internal/validator/duplicates.go` |
| 17.4 | `B-W4D17-4` | Quality level assessor | 🤖 | `internal/validator/quality.go` |
| 17.5 | `B-W4D17-5` | `oss quality` command wiring | 🤖 | Update `cmd/oss/main.go` |

#### 17.1 — Bloom's Taxonomy Validator (TDD)

Verifies that Bloom's levels in learning objectives match the verbs used in assessment questions.

**Step 1: Write tests**

**File:** `internal/validator/bloom_test.go`

```go
package validator_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestBloomLevel(t *testing.T) {
	tests := []struct {
		name     string
		verb     string
		expected string
	}{
		{"remember-list", "list", "remember"},
		{"remember-define", "define", "remember"},
		{"understand-explain", "explain", "understand"},
		{"understand-describe", "describe", "understand"},
		{"apply-solve", "solve", "apply"},
		{"apply-calculate", "calculate", "apply"},
		{"analyze-compare", "compare", "analyze"},
		{"evaluate-justify", "justify", "evaluate"},
		{"create-design", "design", "create"},
		{"unknown-verb", "xyzzy", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.BloomLevelForVerb(tt.verb)
			if got != tt.expected {
				t.Errorf("BloomLevelForVerb(%q) = %q, want %q", tt.verb, got, tt.expected)
			}
		})
	}
}

func TestValidateBloomConsistency(t *testing.T) {
	tests := []struct {
		name       string
		objectives []validator.LearningObjective
		questions  []validator.AssessmentQuestion
		wantErrors int
	}{
		{
			name: "consistent",
			objectives: []validator.LearningObjective{
				{ID: "LO1", Bloom: "apply"},
			},
			questions: []validator.AssessmentQuestion{
				{ID: "Q1", LearningObjective: "LO1", Text: "Solve the equation 2x + 3 = 7"},
			},
			wantErrors: 0,
		},
		{
			name: "question-exceeds-bloom",
			objectives: []validator.LearningObjective{
				{ID: "LO1", Bloom: "remember"},
			},
			questions: []validator.AssessmentQuestion{
				{ID: "Q1", LearningObjective: "LO1", Text: "Evaluate and compare the two approaches"},
			},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateBloomConsistency(tt.objectives, tt.questions)
			if len(errs) != tt.wantErrors {
				t.Errorf("ValidateBloomConsistency() returned %d errors, want %d: %v", len(errs), tt.wantErrors, errs)
			}
		})
	}
}
```

**Step 2: Implement**

**File:** `internal/validator/bloom.go`

```go
package validator

import "strings"

// LearningObjective represents a topic's learning objective for bloom validation.
type LearningObjective struct {
	ID    string
	Text  string
	Bloom string
}

// AssessmentQuestion represents an assessment question for bloom validation.
type AssessmentQuestion struct {
	ID                string
	Text              string
	Difficulty        string
	LearningObjective string
}

// bloomVerbs maps verbs to their Bloom's taxonomy level.
var bloomVerbs = map[string]string{
	// Remember
	"list": "remember", "define": "remember", "recall": "remember",
	"identify": "remember", "name": "remember", "state": "remember",
	"label": "remember", "recognise": "remember", "recognize": "remember",
	// Understand
	"explain": "understand", "describe": "understand", "summarise": "understand",
	"summarize": "understand", "interpret": "understand", "classify": "understand",
	"discuss": "understand", "distinguish": "understand", "paraphrase": "understand",
	// Apply
	"solve": "apply", "calculate": "apply", "use": "apply",
	"apply": "apply", "demonstrate": "apply", "compute": "apply",
	"determine": "apply", "construct": "apply", "show": "apply",
	// Analyze
	"compare": "analyze", "contrast": "analyze", "differentiate": "analyze",
	"analyse": "analyze", "analyze": "analyze", "examine": "analyze",
	"investigate": "analyze", "categorise": "analyze", "categorize": "analyze",
	// Evaluate
	"justify": "evaluate", "evaluate": "evaluate", "assess": "evaluate",
	"critique": "evaluate", "judge": "evaluate", "argue": "evaluate",
	"defend": "evaluate", "recommend": "evaluate",
	// Create
	"design": "create", "create": "create", "formulate": "create",
	"compose": "create", "develop": "create", "invent": "create",
	"plan": "create", "produce": "create", "propose": "create",
}

// bloomOrder defines the hierarchy of Bloom's levels (index = rank).
var bloomOrder = []string{"remember", "understand", "apply", "analyze", "evaluate", "create"}

// BloomLevelForVerb returns the Bloom's taxonomy level for a given verb.
// Returns empty string if the verb is not recognized.
func BloomLevelForVerb(verb string) string {
	return bloomVerbs[strings.ToLower(verb)]
}

// bloomRank returns the numeric rank of a Bloom's level (0=remember, 5=create).
// Returns -1 if the level is not recognized.
func bloomRank(level string) int {
	for i, l := range bloomOrder {
		if l == level {
			return i
		}
	}
	return -1
}

// ValidateBloomConsistency checks that assessment questions don't exceed
// the Bloom's level of their referenced learning objective.
func ValidateBloomConsistency(objectives []LearningObjective, questions []AssessmentQuestion) []string {
	objMap := make(map[string]string) // LO ID -> bloom level
	for _, o := range objectives {
		objMap[o.ID] = o.Bloom
	}

	var errors []string
	for _, q := range questions {
		loBloom, ok := objMap[q.LearningObjective]
		if !ok {
			errors = append(errors, "question "+q.ID+" references unknown learning objective "+q.LearningObjective)
			continue
		}

		// Extract first verb from question text
		questionBloom := detectBloomFromText(q.Text)
		if questionBloom == "" {
			continue // Can't detect, skip
		}

		loRank := bloomRank(loBloom)
		qRank := bloomRank(questionBloom)

		if qRank > loRank {
			errors = append(errors,
				"question "+q.ID+" uses "+questionBloom+"-level verb but learning objective "+
					q.LearningObjective+" is at "+loBloom+" level")
		}
	}

	return errors
}

// detectBloomFromText extracts the highest Bloom's level verb from text.
func detectBloomFromText(text string) string {
	words := strings.Fields(strings.ToLower(text))
	highestRank := -1
	highestLevel := ""

	for _, word := range words {
		// Strip punctuation
		word = strings.Trim(word, ".,;:!?()\"'")
		if level, ok := bloomVerbs[word]; ok {
			rank := bloomRank(level)
			if rank > highestRank {
				highestRank = rank
				highestLevel = level
			}
		}
	}

	return highestLevel
}
```

#### 17.2 — Prerequisite Cycle Detector (TDD)

**Step 1: Write tests**

**File:** `internal/validator/prerequisites_test.go`

```go
package validator_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestDetectCycles_NoCycles(t *testing.T) {
	graph := map[string][]string{
		"F1-01": {},
		"F1-02": {"F1-01"},
		"F1-03": {"F1-01"},
		"F2-01": {"F1-01"},
		"F2-02": {"F1-02", "F2-01"},
	}

	cycles := validator.DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("DetectCycles() found %d cycles, want 0: %v", len(cycles), cycles)
	}
}

func TestDetectCycles_WithCycle(t *testing.T) {
	graph := map[string][]string{
		"F1-01": {"F1-03"},
		"F1-02": {"F1-01"},
		"F1-03": {"F1-02"},
	}

	cycles := validator.DetectCycles(graph)
	if len(cycles) == 0 {
		t.Error("DetectCycles() found no cycles, expected at least one")
	}
}

func TestDetectCycles_SelfReference(t *testing.T) {
	graph := map[string][]string{
		"F1-01": {"F1-01"},
	}

	cycles := validator.DetectCycles(graph)
	if len(cycles) == 0 {
		t.Error("DetectCycles() found no cycles, expected self-reference cycle")
	}
}

func TestValidateMissingPrereqs(t *testing.T) {
	graph := map[string][]string{
		"F1-01": {},
		"F1-02": {"F1-01", "F1-99"},
	}

	missing := validator.FindMissingPrereqs(graph)
	if len(missing) == 0 {
		t.Error("FindMissingPrereqs() found no missing, expected F1-99")
	}
}
```

**Step 2: Implement**

**File:** `internal/validator/prerequisites.go`

```go
package validator

import "fmt"

// DetectCycles finds circular dependencies in a prerequisite graph using DFS.
// Returns a list of cycle descriptions (empty if no cycles).
func DetectCycles(graph map[string][]string) []string {
	var cycles []string

	const (
		white = 0 // unvisited
		gray  = 1 // in current path
		black = 2 // fully processed
	)

	colors := make(map[string]int)
	parent := make(map[string]string)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		colors[node] = gray

		for _, dep := range graph[node] {
			if colors[dep] == gray {
				// Found a cycle — reconstruct it
				cycle := reconstructCycle(parent, node, dep)
				cycles = append(cycles, fmt.Sprintf("cycle detected: %s", cycle))
				return true
			}
			if colors[dep] == white {
				parent[dep] = node
				if dfs(dep) {
					return false // Continue looking for more cycles
				}
			}
		}

		colors[node] = black
		return false
	}

	for node := range graph {
		if colors[node] == white {
			dfs(node)
		}
	}

	return cycles
}

// reconstructCycle builds a human-readable cycle string.
func reconstructCycle(parent map[string]string, from, to string) string {
	path := []string{to, from}
	current := from
	for current != to {
		p, ok := parent[current]
		if !ok {
			break
		}
		path = append(path, p)
		current = p
	}
	// Reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " → "
		}
		result += p
	}
	return result
}

// FindMissingPrereqs returns prerequisites that reference non-existent topic IDs.
func FindMissingPrereqs(graph map[string][]string) []string {
	var missing []string
	for topic, prereqs := range graph {
		for _, prereq := range prereqs {
			if _, exists := graph[prereq]; !exists {
				missing = append(missing, fmt.Sprintf("%s requires %s (not found)", topic, prereq))
			}
		}
	}
	return missing
}
```

#### 17.3 — Duplicate Content Detector (TDD)

**Step 1: Write tests**

**File:** `internal/validator/duplicates_test.go`

```go
package validator_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestTokenize(t *testing.T) {
	text := "Solve the equation 2x + 3 = 7"
	tokens := validator.Tokenize(text)
	if len(tokens) == 0 {
		t.Error("Tokenize() returned empty")
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		a, b      string
		threshold float64
		similar   bool
	}{
		{
			name:      "identical",
			a:         "Solve the equation 2x + 3 = 7",
			b:         "Solve the equation 2x + 3 = 7",
			threshold: 0.85,
			similar:   true,
		},
		{
			name:      "very-different",
			a:         "Solve the equation 2x + 3 = 7",
			b:         "Describe the process of photosynthesis in plants",
			threshold: 0.85,
			similar:   false,
		},
		{
			name:      "similar-but-different",
			a:         "Find the value of x in 3x + 5 = 20",
			b:         "Find the value of x in 4x - 2 = 14",
			threshold: 0.85,
			similar:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := validator.CosineSimilarity(tt.a, tt.b)
			got := sim >= tt.threshold
			if got != tt.similar {
				t.Errorf("CosineSimilarity(%q, %q) = %f, similar=%v want %v",
					tt.a, tt.b, sim, got, tt.similar)
			}
		})
	}
}

func TestFindDuplicates(t *testing.T) {
	questions := []string{
		"Solve 2x + 3 = 7",
		"Simplify 3a + 2b - a",
		"Solve 2x + 3 = 7",  // duplicate of first
	}

	dupes := validator.FindDuplicates(questions, 0.85)
	if len(dupes) == 0 {
		t.Error("FindDuplicates() found no duplicates, expected at least one pair")
	}
}
```

**Step 2: Implement**

**File:** `internal/validator/duplicates.go`

```go
package validator

import (
	"fmt"
	"math"
	"strings"
	"unicode"
)

// DuplicatePair represents two similar content items.
type DuplicatePair struct {
	IndexA     int
	IndexB     int
	Similarity float64
}

// Tokenize splits text into lowercase tokens, removing punctuation.
func Tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	return words
}

// CosineSimilarity computes cosine similarity between two texts using bag-of-words.
func CosineSimilarity(a, b string) float64 {
	tokensA := Tokenize(a)
	tokensB := Tokenize(b)

	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}

	// Build term frequency vectors
	freqA := termFrequency(tokensA)
	freqB := termFrequency(tokensB)

	// Compute dot product and magnitudes
	var dotProduct, magA, magB float64

	allTerms := make(map[string]bool)
	for t := range freqA {
		allTerms[t] = true
	}
	for t := range freqB {
		allTerms[t] = true
	}

	for term := range allTerms {
		a := freqA[term]
		b := freqB[term]
		dotProduct += a * b
		magA += a * a
		magB += b * b
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(magA) * math.Sqrt(magB))
}

// termFrequency builds a term frequency map from tokens.
func termFrequency(tokens []string) map[string]float64 {
	freq := make(map[string]float64)
	for _, t := range tokens {
		freq[t]++
	}
	return freq
}

// FindDuplicates finds pairs of texts that exceed the similarity threshold.
func FindDuplicates(texts []string, threshold float64) []DuplicatePair {
	var pairs []DuplicatePair

	for i := 0; i < len(texts); i++ {
		for j := i + 1; j < len(texts); j++ {
			sim := CosineSimilarity(texts[i], texts[j])
			if sim >= threshold {
				pairs = append(pairs, DuplicatePair{
					IndexA:     i,
					IndexB:     j,
					Similarity: sim,
				})
			}
		}
	}

	return pairs
}

// FormatDuplicateReport creates a human-readable report of duplicate pairs.
func FormatDuplicateReport(pairs []DuplicatePair, texts []string) []string {
	var report []string
	for _, p := range pairs {
		report = append(report, fmt.Sprintf(
			"%.0f%% similar: [%d] %q ↔ [%d] %q",
			p.Similarity*100,
			p.IndexA, truncate(texts[p.IndexA], 50),
			p.IndexB, truncate(texts[p.IndexB], 50),
		))
	}
	return report
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
```

#### 17.4 — Quality Level Assessor (TDD)

**Step 1: Write tests**

**File:** `internal/validator/quality_test.go`

```go
package validator_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

func TestAssessQuality(t *testing.T) {
	tests := []struct {
		name     string
		topic    validator.TopicInfo
		expected int
	}{
		{
			name: "level-0-minimal",
			topic: validator.TopicInfo{
				HasID:                true,
				HasName:              true,
				HasLearningObjectives: true,
			},
			expected: 0,
		},
		{
			name: "level-1-basic",
			topic: validator.TopicInfo{
				HasID:                true,
				HasName:              true,
				HasLearningObjectives: true,
				HasPrerequisites:     true,
				HasDifficulty:        true,
				HasBloomLevels:       true,
			},
			expected: 1,
		},
		{
			name: "level-2-structured",
			topic: validator.TopicInfo{
				HasID: true, HasName: true, HasLearningObjectives: true,
				HasPrerequisites: true, HasDifficulty: true, HasBloomLevels: true,
				HasTeachingSequence:   true,
				HasMisconceptions:     true,
				HasEngagementHooks:    true,
			},
			expected: 2,
		},
		{
			name: "level-3-teachable",
			topic: validator.TopicInfo{
				HasID: true, HasName: true, HasLearningObjectives: true,
				HasPrerequisites: true, HasDifficulty: true, HasBloomLevels: true,
				HasTeachingSequence: true, HasMisconceptions: true, HasEngagementHooks: true,
				HasTeachingNotes: true,
				HasExamples:      true,
				HasAssessments:   true,
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.AssessQuality(tt.topic)
			if got != tt.expected {
				t.Errorf("AssessQuality() = %d, want %d", got, tt.expected)
			}
		})
	}
}
```

**Step 2: Implement**

**File:** `internal/validator/quality.go`

```go
package validator

import "fmt"

// TopicInfo holds the presence of fields for quality assessment.
type TopicInfo struct {
	ID   string
	Name string

	// Level 0: basic identity
	HasID                bool
	HasName              bool
	HasLearningObjectives bool

	// Level 1: structured metadata
	HasPrerequisites bool
	HasDifficulty    bool
	HasBloomLevels   bool

	// Level 2: teaching content
	HasTeachingSequence bool
	HasMisconceptions   bool
	HasEngagementHooks  bool

	// Level 3: full content files
	HasTeachingNotes bool
	HasExamples      bool
	HasAssessments   bool

	// Level 4: translations + cross-curriculum
	HasTranslation    bool
	HasCrossCurriculum bool

	// Level 5: authority validated
	HasAuthorityValidation bool

	// Claimed level (from YAML)
	ClaimedLevel int
}

// QualityReport holds the quality assessment results.
type QualityReport struct {
	Topics      []TopicQuality
	LevelCounts map[int]int
}

// TopicQuality holds quality info for a single topic.
type TopicQuality struct {
	ID            string
	Name          string
	ActualLevel   int
	ClaimedLevel  int
	Overclaimed   bool
}

// AssessQuality determines the actual quality level of a topic based on present fields.
func AssessQuality(info TopicInfo) int {
	// Level 0: has id, name, learning_objectives
	if !info.HasID || !info.HasName || !info.HasLearningObjectives {
		return 0
	}

	// Level 1: + prerequisites, difficulty, bloom_levels
	if !info.HasPrerequisites || !info.HasDifficulty || !info.HasBloomLevels {
		return 0
	}

	// Level 2: + teaching.sequence, teaching.common_misconceptions, engagement_hooks
	if !info.HasTeachingSequence || !info.HasMisconceptions || !info.HasEngagementHooks {
		return 1
	}

	// Level 3: + teaching_notes file, examples file, assessments file
	if !info.HasTeachingNotes || !info.HasExamples || !info.HasAssessments {
		return 2
	}

	// Level 4: + translation, cross_curriculum
	if !info.HasTranslation || !info.HasCrossCurriculum {
		return 3
	}

	// Level 5: authority validated
	if !info.HasAuthorityValidation {
		return 4
	}

	return 5
}

// FormatQualityReport generates a human-readable quality report.
func FormatQualityReport(report QualityReport) string {
	result := "=== Quality Level Report ===\n"
	levelNames := map[int]string{
		0: "Stub", 1: "Basic", 2: "Structured",
		3: "Teachable", 4: "Complete", 5: "Gold",
	}

	for level := 5; level >= 0; level-- {
		count := report.LevelCounts[level]
		result += fmt.Sprintf("Level %d (%s): %d topics\n", level, levelNames[level], count)
	}

	// Flag overclaimed
	var overclaimed []TopicQuality
	for _, t := range report.Topics {
		if t.Overclaimed {
			overclaimed = append(overclaimed, t)
		}
	}
	if len(overclaimed) > 0 {
		result += "\n⚠️  Overclaimed quality levels:\n"
		for _, t := range overclaimed {
			result += fmt.Sprintf("  %s: claims Level %d, actual Level %d\n", t.ID, t.ClaimedLevel, t.ActualLevel)
		}
	}

	return result
}
```

#### 17.5 — Wire `oss quality` Command

Update the `qualityCmd()` in `cmd/oss/main.go` to call the quality assessor. This connects the validator quality package to the CLI by walking topic directories and printing a quality report.

#### Day 17 Validation

```bash
# Run all tests
go test ./...

# Run validator tests specifically
go test -v ./internal/validator/...

# Build CLI
go build ./cmd/oss
```

#### Day 17 Exit Criteria

- [ ] `bloom.go` + `bloom_test.go` — Bloom's verb detection and consistency checking
- [ ] `prerequisites.go` + `prerequisites_test.go` — cycle detection via DFS, missing prereq detection
- [ ] `duplicates.go` + `duplicates_test.go` — cosine similarity, duplicate pair detection
- [ ] `quality.go` + `quality_test.go` — quality level 0-5 assessment
- [ ] `oss quality` command wired and producing output
- [ ] `go test ./...` passes with zero failures

**Progress:** CLI scaffold | 1 package (validator: 5 files) | Tests passing | CI workflow

---

### Day 18 — AI Content Generation Pipeline

**Entry criteria:** Day 17 complete. All validator tests pass. `go test ./...` green.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 18.1 | `B-W4D18-1` | AI provider interface + mock provider | 🤖 | `internal/ai/provider.go`, `internal/ai/mock.go` |
| 18.2 | `B-W4D18-2` | OpenAI provider implementation | 🤖 | `internal/ai/openai.go` |
| 18.3 | `B-W4D18-3` | Anthropic provider implementation | 🤖 | `internal/ai/anthropic.go` |
| 18.4 | `B-W4D18-4` | Ollama provider implementation | 🤖 | `internal/ai/ollama.go` |
| 18.5 | `B-W4D18-5` | Context builder | 🤖 | `internal/generator/context.go` |
| 18.6 | `B-W4D18-6` | Create prompt templates | 🤖 | `prompts/teaching_notes.md`, `prompts/assessments.md` |
| 18.7 | `B-W4D18-7` | 🧑 Review and edit prompt templates | 🧑 | Edits to prompts |

#### 18.1 — AI Provider Interface (TDD)

**Step 1: Write tests**

**File:** `internal/ai/provider_test.go`

```go
package ai_test

import (
	"context"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

func TestMockProvider_Complete(t *testing.T) {
	mock := ai.NewMockProvider("test response")

	resp, err := mock.Complete(context.Background(), ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "user", Content: "Hello"},
		},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if resp.Content != "test response" {
		t.Errorf("Complete() content = %q, want %q", resp.Content, "test response")
	}
}

func TestMockProvider_Models(t *testing.T) {
	mock := ai.NewMockProvider("response")
	models := mock.Models()
	if len(models) == 0 {
		t.Error("Models() returned empty")
	}
}

func TestNewProvider_Unknown(t *testing.T) {
	_, err := ai.NewProvider("unknown", "")
	if err == nil {
		t.Error("NewProvider(unknown) should return error")
	}
}
```

**Step 2: Implement**

**File:** `internal/ai/provider.go`

```go
// Package ai provides a unified interface for AI content generation.
// This interface is shared with P&AI Bot for consistency.
package ai

import (
	"context"
	"fmt"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// CompletionRequest is the input to an AI completion.
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// CompletionResponse is the output from an AI completion.
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// StreamChunk represents a streaming response chunk.
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// ModelInfo describes an available model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MaxTokens   int    `json:"max_tokens"`
	Description string `json:"description"`
}

// Provider is the interface all AI providers must implement.
// This interface is shared with P&AI Bot.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	Models() []ModelInfo
}

// NewProvider creates a new AI provider based on the provider name.
func NewProvider(name, apiKey string) (Provider, error) {
	switch name {
	case "openai":
		return NewOpenAIProvider(apiKey)
	case "anthropic":
		return NewAnthropicProvider(apiKey)
	case "ollama":
		return NewOllamaProvider("")
	case "mock":
		return NewMockProvider(""), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s (supported: openai, anthropic, ollama)", name)
	}
}
```

**File:** `internal/ai/mock.go`

```go
package ai

import "context"

// MockProvider is a test double for AI providers.
type MockProvider struct {
	Response string
	Err      error
}

// NewMockProvider creates a mock provider that returns the given response.
func NewMockProvider(response string) *MockProvider {
	return &MockProvider{Response: response}
}

func (m *MockProvider) Complete(_ context.Context, _ CompletionRequest) (CompletionResponse, error) {
	if m.Err != nil {
		return CompletionResponse{}, m.Err
	}
	return CompletionResponse{
		Content:      m.Response,
		Model:        "mock",
		InputTokens:  10,
		OutputTokens: len(m.Response),
	}, nil
}

func (m *MockProvider) StreamComplete(_ context.Context, _ CompletionRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- StreamChunk{Content: m.Response, Done: true}
	}()
	return ch, nil
}

func (m *MockProvider) Models() []ModelInfo {
	return []ModelInfo{
		{ID: "mock", Name: "Mock Model", MaxTokens: 4096, Description: "Test mock"},
	}
}
```

#### 18.2–18.4 — Provider Implementations

Each provider (`openai.go`, `anthropic.go`, `ollama.go`) implements the `Provider` interface. They follow the same structure:

**File:** `internal/ai/openai.go`

```go
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAIProvider implements the Provider interface for OpenAI.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(apiKey string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required (set OSS_AI_API_KEY)")
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{},
	}, nil
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = "gpt-4o"
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	body := map[string]interface{}{
		"model":      model,
		"messages":   req.Messages,
		"max_tokens": maxTokens,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return CompletionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("OpenAI API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return CompletionResponse{}, fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return CompletionResponse{}, fmt.Errorf("parsing response: %w", err)
	}

	if len(result.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("OpenAI returned no choices")
	}

	return CompletionResponse{
		Content:      result.Choices[0].Message.Content,
		Model:        result.Model,
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
	}, nil
}

func (p *OpenAIProvider) StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	// Streaming implementation — similar to Complete but with SSE parsing
	// For now, fall back to non-streaming
	ch := make(chan StreamChunk, 1)
	go func() {
		defer close(ch)
		resp, err := p.Complete(ctx, req)
		if err != nil {
			ch <- StreamChunk{Error: err}
			return
		}
		ch <- StreamChunk{Content: resp.Content, Done: true}
	}()
	return ch, nil
}

func (p *OpenAIProvider) Models() []ModelInfo {
	return []ModelInfo{
		{ID: "gpt-4o", Name: "GPT-4o", MaxTokens: 128000, Description: "Most capable general model"},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", MaxTokens: 128000, Description: "Fast and affordable"},
	}
}
```

Create similar implementations for `anthropic.go` (using the Anthropic Messages API) and `ollama.go` (using the Ollama `/api/chat` endpoint). Both follow the same pattern — implement `Complete`, `StreamComplete`, and `Models`.

#### 18.5 — Context Builder (TDD)

**Step 1: Write tests**

**File:** `internal/generator/context_test.go`

```go
package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestBuildContext(t *testing.T) {
	repoDir := setupTestRepo(t)

	ctx, err := generator.BuildContext(repoDir, "F1-01")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}

	if ctx.Topic.ID != "F1-01" {
		t.Errorf("Topic.ID = %q, want %q", ctx.Topic.ID, "F1-01")
	}
	if ctx.Topic.Name == "" {
		t.Error("Topic.Name is empty")
	}
}

func TestBuildContext_WithPrerequisites(t *testing.T) {
	repoDir := setupTestRepo(t)

	ctx, err := generator.BuildContext(repoDir, "F1-02")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}

	if len(ctx.Prerequisites) == 0 {
		t.Error("Prerequisites should not be empty for F1-02")
	}
}

func TestBuildContext_NotFound(t *testing.T) {
	repoDir := setupTestRepo(t)

	_, err := generator.BuildContext(repoDir, "NONEXISTENT")
	if err == nil {
		t.Error("BuildContext() should error for non-existent topic")
	}
}

// setupTestRepo creates a minimal OSS repo structure for testing.
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	topicsDir := filepath.Join(dir, "curricula", "test", "topics", "algebra")
	os.MkdirAll(topicsDir, 0o755)

	// Topic F1-01 (no prerequisites)
	os.WriteFile(filepath.Join(topicsDir, "01-test.yaml"), []byte(`
id: F1-01
name: "Test Topic One"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: LO1
    text: "Test objective"
    bloom: understand
prerequisites:
  required: []
quality_level: 1
provenance: human
`), 0o644)

	// Topic F1-02 (requires F1-01)
	os.WriteFile(filepath.Join(topicsDir, "02-test.yaml"), []byte(`
id: F1-02
name: "Test Topic Two"
subject_id: algebra
syllabus_id: test-syllabus
difficulty: beginner
learning_objectives:
  - id: LO1
    text: "Test objective two"
    bloom: apply
prerequisites:
  required:
    - F1-01
quality_level: 1
provenance: human
`), 0o644)

	return dir
}
```

**Step 2: Implement**

**File:** `internal/generator/context.go`

```go
// Package generator implements the AI content generation pipeline.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Topic represents a parsed topic YAML file.
type Topic struct {
	ID                 string              `yaml:"id"`
	Name               string              `yaml:"name"`
	SubjectID          string              `yaml:"subject_id"`
	SyllabusID         string              `yaml:"syllabus_id"`
	Difficulty         string              `yaml:"difficulty"`
	LearningObjectives []LearningObjective `yaml:"learning_objectives"`
	Prerequisites      PrerequisiteList    `yaml:"prerequisites"`
	Teaching           *TeachingInfo       `yaml:"teaching,omitempty"`
	BloomLevels        []string            `yaml:"bloom_levels,omitempty"`
	QualityLevel       int                 `yaml:"quality_level"`
	Provenance         string              `yaml:"provenance"`
	TeachingNotesFile  string              `yaml:"ai_teaching_notes,omitempty"`
	ExamplesFile       string              `yaml:"examples_file,omitempty"`
	AssessmentsFile    string              `yaml:"assessments_file,omitempty"`
}

// LearningObjective represents a single learning objective.
type LearningObjective struct {
	ID    string `yaml:"id"`
	Text  string `yaml:"text"`
	Bloom string `yaml:"bloom"`
}

// PrerequisiteList holds required and recommended prerequisites.
type PrerequisiteList struct {
	Required    []string `yaml:"required"`
	Recommended []string `yaml:"recommended"`
}

// TeachingInfo holds teaching-related content.
type TeachingInfo struct {
	Sequence            []string          `yaml:"sequence,omitempty"`
	CommonMisconceptions []Misconception  `yaml:"common_misconceptions,omitempty"`
	EngagementHooks     []string          `yaml:"engagement_hooks,omitempty"`
}

// Misconception represents a common student misconception.
type Misconception struct {
	Misconception string `yaml:"misconception"`
	Remediation   string `yaml:"remediation"`
}

// GenerationContext holds all context needed for AI content generation.
type GenerationContext struct {
	Topic         Topic
	Prerequisites []Topic
	Siblings      []Topic
	ExistingNotes string
	SchemaRules   string
}

// BuildContext assembles the generation context for a given topic ID.
func BuildContext(repoDir, topicID string) (*GenerationContext, error) {
	// Find the topic file
	topicFile, err := findTopicFile(repoDir, topicID)
	if err != nil {
		return nil, fmt.Errorf("finding topic %s: %w", topicID, err)
	}

	topic, err := loadTopic(topicFile)
	if err != nil {
		return nil, fmt.Errorf("loading topic %s: %w", topicID, err)
	}

	ctx := &GenerationContext{
		Topic: *topic,
	}

	// Load prerequisites
	for _, prereqID := range topic.Prerequisites.Required {
		prereqFile, err := findTopicFile(repoDir, prereqID)
		if err != nil {
			continue // Prerequisite might not exist yet
		}
		prereq, err := loadTopic(prereqFile)
		if err != nil {
			continue
		}
		ctx.Prerequisites = append(ctx.Prerequisites, *prereq)
	}

	// Load siblings (other topics in the same directory)
	topicDir := filepath.Dir(topicFile)
	siblings, err := loadSiblingTopics(topicDir, topicID)
	if err == nil {
		ctx.Siblings = siblings
	}

	// Load existing teaching notes if they exist
	if topic.TeachingNotesFile != "" {
		notesPath := filepath.Join(topicDir, topic.TeachingNotesFile)
		if data, err := os.ReadFile(notesPath); err == nil {
			ctx.ExistingNotes = string(data)
		}
	}

	return ctx, nil
}

// findTopicFile searches the repo for a topic file with the given ID.
func findTopicFile(repoDir, topicID string) (string, error) {
	var found string

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		if strings.HasSuffix(path, ".assessments.yaml") || strings.HasSuffix(path, ".examples.yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var partial struct {
			ID string `yaml:"id"`
		}
		if err := yaml.Unmarshal(data, &partial); err != nil {
			return nil
		}

		if partial.ID == topicID {
			found = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("topic %s not found in %s", topicID, repoDir)
	}
	return found, nil
}

// loadTopic parses a topic YAML file.
func loadTopic(path string) (*Topic, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var topic Topic
	if err := yaml.Unmarshal(data, &topic); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &topic, nil
}

// loadSiblingTopics loads all topics in the same directory, excluding the given ID.
func loadSiblingTopics(dir, excludeID string) ([]Topic, error) {
	var siblings []Topic

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".assessments.yaml") ||
			strings.HasSuffix(entry.Name(), ".examples.yaml") {
			continue
		}

		topic, err := loadTopic(filepath.Join(dir, entry.Name()))
		if err != nil || topic.ID == excludeID {
			continue
		}
		siblings = append(siblings, *topic)
	}

	return siblings, nil
}
```

#### 18.6 — Create Prompt Templates

**File:** `prompts/teaching_notes.md`

````markdown
# Teaching Notes Generation Prompt

You are an expert educator creating teaching notes for a mathematics topic.

## Context

**Topic:** {{topic_name}} ({{topic_id}})
**Syllabus:** {{syllabus_id}}
**Difficulty:** {{difficulty}}
**Prerequisites:** {{prerequisites}}

### Learning Objectives
{{learning_objectives}}

### Existing Content (for style matching)
{{existing_notes}}

## Instructions

Generate comprehensive teaching notes following this exact structure:

```markdown
# {{topic_name}} — Teaching Notes

## Overview
[Brief description of what this topic covers and why it matters]

## Prerequisites Check
[What students should know before starting]

## Teaching Sequence

### 1. [Section Title] (XX min)
[Teaching instructions with concrete examples]

### 2. [Section Title] (XX min)
[Teaching instructions]

## Common Misconceptions

| Misconception | Why Students Think This | How to Address |
|---------------|------------------------|----------------|
| ... | ... | ... |

## Engagement Hooks
- [Real-world connection 1]
- [Real-world connection 2]

## Assessment Guidance
[Tips for assessing understanding]

## Bahasa Melayu Key Terms
| English | Bahasa Melayu |
|---------|---------------|
| ... | ... |
```

## Requirements
- Write for AI chat delivery (conversational, not textbook)
- Start with engagement hook, not definition
- Include scaffolding for when student is stuck
- Use mathematically correct notation
- Include correct BM (Bahasa Melayu) terminology
- Reference prerequisite knowledge where appropriate
- End each section with a forward look to what's next
````

**File:** `prompts/assessments.md`

````markdown
# Assessment Generation Prompt

You are an expert educator creating assessment questions for a mathematics topic.

## Context

**Topic:** {{topic_name}} ({{topic_id}})
**Syllabus:** {{syllabus_id}}
**Difficulty:** {{difficulty}}
**Count:** {{count}} questions
**Target difficulty:** {{target_difficulty}}

### Learning Objectives
{{learning_objectives}}

## Instructions

Generate {{count}} assessment questions as YAML. Each question must include:
- Worked solution (`answer.working`)
- Mark scheme (`rubric`)
- Progressive hints (at least 2)
- Common wrong answers with targeted feedback (`distractors`)

Output format (YAML):

```yaml
topic_id: {{topic_id}}
provenance: ai-generated

questions:
  - id: Q1
    text: "Question text. Supports $LaTeX$ notation."
    difficulty: easy
    learning_objective: LO1
    answer:
      type: exact
      value: "correct answer"
      working: |
        Step-by-step solution
    marks: 2
    rubric:
      - marks: 1
        criteria: "First mark criterion"
    hints:
      - level: 1
        text: "Gentle nudge"
      - level: 2
        text: "More explicit help"
    distractors:
      - value: "common wrong answer"
        feedback: "Targeted feedback"
```

## Requirements
- Distribute questions across available learning objectives
- Difficulty spread: mix of easy, medium, hard per {{target_difficulty}}
- Use KSSM exam format for Malaysian curriculum
- Include LaTeX for mathematical notation
- Each question must test a single concept clearly
- Distractors must reflect REAL student errors (not random wrong answers)
````

#### 18.7 — Education Lead Reviews Prompt Templates (🧑)

The Education Lead should review `prompts/teaching_notes.md` and `prompts/assessments.md` for:
- [ ] Correct pedagogical approach for KSSM
- [ ] Appropriate BM terminology references
- [ ] Accurate Bloom's taxonomy usage
- [ ] Realistic common misconceptions

#### Day 18 Validation

```bash
# Run all tests
go test ./...

# Run AI package tests specifically
go test -v ./internal/ai/...

# Run generator tests
go test -v ./internal/generator/...
```

#### Day 18 Exit Criteria

- [ ] `internal/ai/provider.go` — Provider interface defined
- [ ] `internal/ai/mock.go` — MockProvider for testing (all tests use this)
- [ ] `internal/ai/openai.go` — OpenAI implementation
- [ ] `internal/ai/anthropic.go` — Anthropic implementation
- [ ] `internal/ai/ollama.go` — Ollama implementation
- [ ] `internal/generator/context.go` — Context builder loads topic, prerequisites, siblings
- [ ] `prompts/teaching_notes.md` and `prompts/assessments.md` created
- [ ] Education Lead has reviewed prompt templates
- [ ] `go test ./...` passes with zero failures

**Progress:** CLI + validator + AI providers + context builder | 3 packages | Prompt templates

---

### Day 19 — Generation Commands

**Entry criteria:** Day 18 complete. AI providers and context builder work. Prompt templates reviewed.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 19.1 | `B-W4D19-1` | Teaching notes generator | 🤖 | `internal/generator/teaching_notes.go` |
| 19.2 | `B-W4D19-2` | Assessment generator | 🤖 | `internal/generator/assessments.go` |
| 19.3 | `B-W4D19-3` | Worked examples generator | 🤖 | `internal/generator/examples.go` |
| 19.4 | `B-W4D19-4` | Unified pipeline orchestrator + output writers | 🤖 | `internal/pipeline/pipeline.go`, `internal/output/writer.go`, `internal/output/github.go` |
| 19.5 | `B-W4D19-5` | Wire CLI commands via pipeline | 🤖 | Update `cmd/oss/main.go` |

#### 19.1 — Teaching Notes Generator (TDD)

**Step 1: Write tests**

**File:** `internal/generator/teaching_notes_test.go`

```go
package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestGenerateTeachingNotes(t *testing.T) {
	mockResponse := `# Test Topic — Teaching Notes

## Overview
This topic covers test content.

## Prerequisites Check
- Basic arithmetic

## Teaching Sequence

### 1. Introduction (15 min)
Start with examples.

## Common Misconceptions

| Misconception | Why Students Think This | How to Address |
|---------------|------------------------|----------------|
| Error | Reason | Fix |

## Engagement Hooks
- Real world example

## Bahasa Melayu Key Terms
| English | Bahasa Melayu |
|---------|---------------|
| Variable | Pemboleh ubah |
`

	mock := ai.NewMockProvider(mockResponse)

	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			SubjectID:  "algebra",
			SyllabusID: "test-syllabus",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "LO1", Text: "Test objective", Bloom: "understand"},
			},
		},
	}

	result, err := generator.GenerateTeachingNotes(context.Background(), mock, genCtx, "prompts/")
	if err != nil {
		t.Fatalf("GenerateTeachingNotes() error = %v", err)
	}

	if !strings.Contains(result.Content, "Teaching Notes") {
		t.Error("Result should contain 'Teaching Notes'")
	}
	if result.Model == "" {
		t.Error("Result.Model should not be empty")
	}
}

func TestBuildTeachingNotesPrompt(t *testing.T) {
	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "LO1", Text: "Test", Bloom: "understand"},
			},
		},
	}

	prompt := generator.BuildTeachingNotesPrompt(genCtx)
	if prompt == "" {
		t.Error("BuildTeachingNotesPrompt() returned empty string")
	}
	if !strings.Contains(prompt, "F1-01") {
		t.Error("Prompt should contain topic ID")
	}
}
```

**Step 2: Implement**

**File:** `internal/generator/teaching_notes.go`

```go
package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// GenerationResult holds the output of a content generation.
type GenerationResult struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int
}

// GenerateTeachingNotes generates teaching notes for a topic using AI.
func GenerateTeachingNotes(ctx context.Context, provider ai.Provider, genCtx *GenerationContext, promptsDir string) (*GenerationResult, error) {
	prompt := BuildTeachingNotesPrompt(genCtx)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are an expert mathematics educator creating teaching notes for the Malaysian KSSM curriculum."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// BuildTeachingNotesPrompt constructs the prompt for teaching notes generation.
func BuildTeachingNotesPrompt(genCtx *GenerationContext) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate comprehensive teaching notes for the topic: %s (%s)\n\n", genCtx.Topic.Name, genCtx.Topic.ID))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n", genCtx.Topic.SyllabusID))
	sb.WriteString(fmt.Sprintf("Difficulty: %s\n\n", genCtx.Topic.Difficulty))

	sb.WriteString("## Learning Objectives\n")
	for _, lo := range genCtx.Topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", lo.ID, lo.Bloom, lo.Text))
	}
	sb.WriteString("\n")

	if len(genCtx.Prerequisites) > 0 {
		sb.WriteString("## Prerequisites (students have already learned)\n")
		for _, p := range genCtx.Prerequisites {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", p.ID, p.Name))
		}
		sb.WriteString("\n")
	}

	if genCtx.ExistingNotes != "" {
		sb.WriteString("## Existing Notes (match this style)\n")
		sb.WriteString(genCtx.ExistingNotes)
		sb.WriteString("\n\n")
	}

	sb.WriteString(`## Output Format
Write in Markdown following this structure:
# [Topic Name] — Teaching Notes

## Overview
## Prerequisites Check
## Teaching Sequence (with time estimates)
## Common Misconceptions (table format)
## Engagement Hooks
## Assessment Guidance
## Bahasa Melayu Key Terms (table format)
`)

	return sb.String()
}
```

#### 19.2 — Assessment Generator (TDD)

**Step 1: Write tests**

**File:** `internal/generator/assessments_test.go`

```go
package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestGenerateAssessments(t *testing.T) {
	mockYAML := `topic_id: F1-01
provenance: ai-generated

questions:
  - id: Q1
    text: "If x = 3, find 2x + 5"
    difficulty: easy
    learning_objective: LO1
    answer:
      type: exact
      value: "11"
      working: "2(3) + 5 = 11"
    marks: 2
    rubric:
      - marks: 1
        criteria: "Correct substitution"
      - marks: 1
        criteria: "Correct answer"
    hints:
      - level: 1
        text: "Replace x with 3"
    distractors:
      - value: "235"
        feedback: "You wrote numbers side by side instead of multiplying"
`

	mock := ai.NewMockProvider(mockYAML)

	genCtx := &generator.GenerationContext{
		Topic: generator.Topic{
			ID:         "F1-01",
			Name:       "Test Topic",
			Difficulty: "beginner",
			LearningObjectives: []generator.LearningObjective{
				{ID: "LO1", Text: "Test", Bloom: "apply"},
			},
		},
	}

	result, err := generator.GenerateAssessments(context.Background(), mock, genCtx, 5, "medium")
	if err != nil {
		t.Fatalf("GenerateAssessments() error = %v", err)
	}

	if !strings.Contains(result.Content, "topic_id") {
		t.Error("Result should contain YAML with topic_id")
	}
}
```

**Step 2: Implement**

**File:** `internal/generator/assessments.go`

```go
package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// GenerateAssessments generates assessment questions for a topic using AI.
func GenerateAssessments(ctx context.Context, provider ai.Provider, genCtx *GenerationContext, count int, difficulty string) (*GenerationResult, error) {
	prompt := BuildAssessmentsPrompt(genCtx, count, difficulty)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are an expert mathematics educator creating assessment questions for the Malaysian KSSM curriculum. Output valid YAML only."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   4096,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// BuildAssessmentsPrompt constructs the prompt for assessment generation.
func BuildAssessmentsPrompt(genCtx *GenerationContext, count int, difficulty string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate %d assessment questions for: %s (%s)\n\n", count, genCtx.Topic.Name, genCtx.Topic.ID))
	sb.WriteString(fmt.Sprintf("Target difficulty: %s\n", difficulty))
	sb.WriteString(fmt.Sprintf("Syllabus: %s\n\n", genCtx.Topic.SyllabusID))

	sb.WriteString("## Learning Objectives\n")
	for _, lo := range genCtx.Topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", lo.ID, lo.Bloom, lo.Text))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf(`## Requirements
- Generate exactly %d questions
- Each question must include: worked solution, rubric, 2+ hints, distractors
- Distribute across learning objectives
- Use KSSM exam format
- Support LaTeX via $...$ notation
- Output as valid YAML matching the assessments schema

## Output Format
Output ONLY valid YAML (no markdown code fences):

topic_id: %s
provenance: ai-generated

questions:
  - id: Q1
    text: "..."
    difficulty: easy|medium|hard
    learning_objective: LO1
    answer:
      type: exact|range|multiple_choice|free_text
      value: "..."
      working: |
        Step by step solution
    marks: N
    rubric:
      - marks: 1
        criteria: "..."
    hints:
      - level: 1
        text: "..."
    distractors:
      - value: "..."
        feedback: "..."
`, count, genCtx.Topic.ID))

	return sb.String()
}
```

#### 19.3 — Worked Examples Generator (TDD)

Follow the same pattern as 19.1 and 19.2. Create:

**Files:**
- `internal/generator/examples_test.go` — tests with mock provider
- `internal/generator/examples.go` — `GenerateExamples()` and `BuildExamplesPrompt()`

The examples generator produces `.examples.yaml` files with step-by-step worked solutions.

#### 19.4 — Unified Pipeline Orchestrator

All three interfaces (CLI, Bot, Web Portal) execute the same content generation workflow. Instead of reimplementing this in each interface, a shared `Pipeline` orchestrator ensures one code path for all.

**File:** `internal/pipeline/pipeline.go`

```go
package pipeline

import (
	"context"
	"fmt"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
	"github.com/p-n-ai/oss-bot/internal/output"
	"github.com/p-n-ai/oss-bot/internal/validator"
)

// ExecutionMode determines what happens after content is generated and validated.
type ExecutionMode int

const (
	// ModePreview generates and validates content, returns structured output.
	// Used by: Web Portal preview, CLI dry-run.
	ModePreview ExecutionMode = iota

	// ModeWriteFS writes generated files to the local filesystem.
	// Used by: CLI tool.
	ModeWriteFS

	// ModeCreatePR creates a GitHub PR with generated content.
	// Used by: GitHub Bot, Web Portal submit, CLI --pr flag.
	ModeCreatePR
)

// Request is the unified input for all content generation, regardless of interface.
type Request struct {
	TopicPath        string
	ContributionType string // "teaching_notes", "assessments", "examples", "translation", "import"
	Content          string // Pre-extracted text (after input processing)
	Mode             ExecutionMode
	OutputDir        string            // For ModeWriteFS
	Options          map[string]string // count, difficulty, language, etc.
	Source           string            // "cli", "bot", "web" — for provenance
}

// Result is the unified output from the pipeline.
type Result struct {
	StructuredOutput string              // Generated YAML/Markdown
	Files            map[string]string   // filepath -> content
	ValidationErrors []string
	QualityLevel     int
	PRUrl            string // Populated only in ModeCreatePR
	PRNumber         int
}

// Pipeline is the shared orchestrator for all content generation.
type Pipeline struct {
	aiProvider ai.Provider
	validator  *validator.Validator
	writer     output.Writer
	promptsDir string
	repoPath   string
}

// New creates a pipeline with the given dependencies.
func New(provider ai.Provider, v *validator.Validator, w output.Writer, promptsDir, repoPath string) *Pipeline {
	return &Pipeline{
		aiProvider: provider,
		validator:  v,
		writer:     w,
		promptsDir: promptsDir,
		repoPath:   repoPath,
	}
}

// Execute runs the full content generation workflow:
//  1. Build context from topic
//  2. Generate content via AI
//  3. Validate output
//  4. Retry once if validation fails
//  5. Execute based on mode (preview, write FS, create PR)
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

	// 3. Validate
	validationErrors := p.validator.ValidateContent(generated.Files)

	// 4. Retry once if validation fails
	if len(validationErrors) > 0 {
		generated, err = p.generateWithFeedback(ctx, genCtx, req, validationErrors)
		if err != nil {
			return nil, fmt.Errorf("retry generation: %w", err)
		}
		validationErrors = p.validator.ValidateContent(generated.Files)
	}

	result := &Result{
		StructuredOutput: generated.Content,
		Files:            generated.Files,
		ValidationErrors: validationErrors,
		QualityLevel:     p.validator.AssessQuality(generated.Files),
	}

	// 5. Execute based on mode
	if len(validationErrors) == 0 {
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
				Quality:     result.QualityLevel,
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
	}

	return result, nil
}

func (p *Pipeline) generate(ctx context.Context, genCtx *generator.GenerationContext, req Request) (*generator.GenerationResult, error) {
	switch req.ContributionType {
	case "teaching_notes":
		return generator.GenerateTeachingNotes(ctx, p.aiProvider, genCtx, p.promptsDir)
	case "assessments":
		count := 5 // default
		difficulty := "medium"
		if v, ok := req.Options["count"]; ok {
			fmt.Sscanf(v, "%d", &count)
		}
		if v, ok := req.Options["difficulty"]; ok {
			difficulty = v
		}
		return generator.GenerateAssessments(ctx, p.aiProvider, genCtx, count, difficulty, p.promptsDir)
	case "examples":
		return generator.GenerateExamples(ctx, p.aiProvider, genCtx, p.promptsDir)
	case "translation":
		lang := req.Options["to"]
		return generator.Translate(ctx, p.aiProvider, genCtx, lang, p.promptsDir)
	case "import":
		return generator.ImportFromText(ctx, p.aiProvider, req.Content, req.Options, p.promptsDir)
	default:
		return nil, fmt.Errorf("unknown contribution type: %s", req.ContributionType)
	}
}

func (p *Pipeline) generateWithFeedback(ctx context.Context, genCtx *generator.GenerationContext, req Request, errors []string) (*generator.GenerationResult, error) {
	// Inject validation errors as feedback for the retry
	genCtx.ValidationFeedback = errors
	return p.generate(ctx, genCtx, req)
}
```

**File:** `internal/output/writer.go`

```go
package output

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Writer abstracts where generated content is written.
type Writer interface {
	// WriteFiles writes files to the local filesystem (CLI).
	WriteFiles(ctx context.Context, baseDir string, files map[string]string) error

	// CreatePR creates a GitHub PR with the given files (Bot, Web Portal).
	CreatePR(ctx context.Context, input PRInput) (*PROutput, error)
}

// PRInput holds the data needed to create a PR.
type PRInput struct {
	Files       map[string]string // filepath -> content
	TopicPath   string
	ContentType string
	Quality     int
	Source      string // "cli", "bot", "web"
}

// PROutput holds the result of creating a PR.
type PROutput struct {
	URL    string
	Number int
	Branch string
}

// LocalWriter writes files to the local filesystem. Used by CLI.
type LocalWriter struct{}

func (w *LocalWriter) WriteFiles(ctx context.Context, baseDir string, files map[string]string) error {
	for path, content := range files {
		fullPath := filepath.Join(baseDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

func (w *LocalWriter) CreatePR(ctx context.Context, input PRInput) (*PROutput, error) {
	return nil, fmt.Errorf("LocalWriter does not support PR creation — use GitHubWriter")
}
```

**File:** `internal/output/github.go`

```go
package output

import (
	"context"
	"fmt"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

// GitHubWriter creates PRs via the GitHub API. Used by Bot and Web Portal.
type GitHubWriter struct {
	app       *gh.App
	repoOwner string
	repoName  string
}

// NewGitHubWriter creates a writer that creates PRs via the GitHub API.
func NewGitHubWriter(app *gh.App, owner, repo string) *GitHubWriter {
	return &GitHubWriter{app: app, repoOwner: owner, repoName: repo}
}

func (w *GitHubWriter) WriteFiles(ctx context.Context, baseDir string, files map[string]string) error {
	return fmt.Errorf("GitHubWriter does not support local file writing — use LocalWriter")
}

func (w *GitHubWriter) CreatePR(ctx context.Context, input PRInput) (*PROutput, error) {
	branch := gh.GenerateBranchName("add", input.ContentType, input.TopicPath)

	var fileChanges []gh.FileChange
	for path, content := range input.Files {
		fileChanges = append(fileChanges, gh.FileChange{Path: path, Content: content})
	}

	labels := []string{
		"provenance:ai-generated",
		fmt.Sprintf("quality:level-%d", input.Quality),
		fmt.Sprintf("source:%s", input.Source),
	}

	result, err := w.app.CreatePR(ctx, gh.PRRequest{
		Owner:      w.repoOwner,
		Repo:       w.repoName,
		Title:      fmt.Sprintf("Add %s for %s", input.ContentType, input.TopicPath),
		Body:       fmt.Sprintf("Generated by oss-bot via %s.\n\nQuality level: %d", input.Source, input.Quality),
		BranchName: branch,
		BaseBranch: "main",
		Files:      fileChanges,
		Labels:     labels,
	})
	if err != nil {
		return nil, err
	}

	return &PROutput{URL: result.URL, Number: result.Number, Branch: result.Branch}, nil
}
```

This means all three interfaces use the **same code path**:

| Interface | How it calls the pipeline |
|-----------|--------------------------|
| **CLI** | `pipeline.Execute(ctx, Request{Mode: ModeWriteFS, OutputDir: repoPath})` |
| **Bot** | `pipeline.Execute(ctx, Request{Mode: ModeCreatePR, Source: "bot"})` |
| **Web (preview)** | `pipeline.Execute(ctx, Request{Mode: ModePreview})` |
| **Web (submit)** | `pipeline.Execute(ctx, Request{Mode: ModeCreatePR, Source: "web"})` |
| **CLI --pr** | `pipeline.Execute(ctx, Request{Mode: ModeCreatePR, Source: "cli"})` |

#### 19.5 — Wire CLI Commands

Update `cmd/oss/main.go` to connect the `generate teaching-notes` and `generate assessments` commands to the **pipeline** (not directly to generators):

1. Parse `OSS_AI_PROVIDER` and `OSS_AI_API_KEY` environment variables
2. Create the AI provider using `ai.NewProvider()`
3. Create the pipeline: `pipeline.New(provider, validator, &output.LocalWriter{}, promptsDir, repoPath)`
4. Call `pipeline.Execute()` with the appropriate `Request`
5. Print success/failure with file path

```go
// Example: CLI wiring for generate teaching-notes
func runTeachingNotes(cmd *cobra.Command, args []string) error {
	topicPath := args[0]
	repoPath := os.Getenv("OSS_REPO_PATH")

	p := pipeline.New(provider, v, &output.LocalWriter{}, "prompts/", repoPath)

	result, err := p.Execute(context.Background(), pipeline.Request{
		TopicPath:        topicPath,
		ContributionType: "teaching_notes",
		Mode:             pipeline.ModeWriteFS,
		OutputDir:        repoPath,
		Source:           "cli",
	})
	if err != nil {
		return err
	}

	if len(result.ValidationErrors) > 0 {
		for _, e := range result.ValidationErrors {
			fmt.Fprintf(os.Stderr, "  ⚠ %s\n", e)
		}
	}

	for path := range result.Files {
		fmt.Printf("  ✅ Written: %s\n", path)
	}
	return nil
}
```

**Create also:** `prompts/examples.md` — template for worked examples generation.

#### Day 19 Validation

```bash
# Run all tests
go test ./...

# Run generator tests specifically
go test -v ./internal/generator/...

# Verify CLI commands exist
go run ./cmd/oss generate --help
go run ./cmd/oss generate teaching-notes --help
go run ./cmd/oss generate assessments --help
```

#### Day 19 Exit Criteria

- [ ] `internal/generator/teaching_notes.go` + tests — generates teaching notes via AI
- [ ] `internal/generator/assessments.go` + tests — generates assessment YAML via AI
- [ ] `internal/generator/examples.go` + tests — generates worked examples via AI
- [ ] `internal/pipeline/pipeline.go` + tests — unified pipeline (Preview, WriteFS, CreatePR modes)
- [ ] `internal/output/writer.go` — `Writer` interface + `LocalWriter` (filesystem) + `GitHubWriter` (PRs)
- [ ] CLI commands wired through `pipeline.Execute()` (not directly to generators)
- [ ] `oss generate teaching-notes <topic>` command wired and working
- [ ] `oss generate assessments <topic> --count 5 --difficulty medium` command wired
- [ ] All generators use mock provider in tests (no real API calls)
- [ ] `go test ./...` passes with zero failures

**Progress:** CLI fully functional (validate + generate) | 5 packages (+ pipeline, output) | 3 prompt templates

---

### Day 20 — Translation + End-to-End Testing

**Entry criteria:** Day 19 complete. All generate commands work with mock provider.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 20.1 | `B-W4D20-1` | Create translation prompt template | 🤖 | `prompts/translation.md` |
| 20.2 | `B-W4D20-2` | Translation generator + `oss translate` command | 🤖 | `internal/generator/translator.go` |
| 20.3 | `B-W4D20-3` | End-to-end pipeline test | 🤖🧑 | Integration test |
| 20.4 | `B-W4D20-4` | 🧑 Education Lead evaluates AI output quality | 🧑 | Decision only |

#### 20.1 — Translation Prompt Template

**File:** `prompts/translation.md`

````markdown
# Translation Prompt

You are a professional translator specializing in mathematics education content.

## Context

**Source language:** English
**Target language:** {{target_language}}
**Topic:** {{topic_name}} ({{topic_id}})

## Source Content
{{source_content}}

## Instructions

Translate the content following these rules:

1. **Preserve YAML structure exactly** — only translate human-readable text values
2. **Do not translate:** `id`, `type`, `bloom`, `difficulty`, `provenance` field values
3. **Do translate:** `name`, `text`, learning objective text, misconception text, hints, feedback
4. **Use correct mathematical terminology** in the target language
5. **For Bahasa Melayu specifically:**
   - Variable → Pemboleh ubah
   - Coefficient → Pekali
   - Constant → Pemalar
   - Equation → Persamaan
   - Expression → Ungkapan
   - Inequality → Ketaksamaan

## Output Format

Output ONLY the translated YAML (no code fences, no commentary):

```yaml
name: "Translated name"

learning_objectives:
  - id: LO1
    text: "Translated text"
```
````

#### 20.2 — Translation Generator (TDD)

**Step 1: Write tests**

**File:** `internal/generator/translator_test.go`

```go
package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestTranslate(t *testing.T) {
	mockTranslation := `name: "Pemboleh ubah & Ungkapan Algebra"

learning_objectives:
  - id: LO1
    text: "Menggunakan huruf untuk mewakili kuantiti yang tidak diketahui"
`

	mock := ai.NewMockProvider(mockTranslation)

	topic := generator.Topic{
		ID:         "F1-01",
		Name:       "Variables & Algebraic Expressions",
		Difficulty: "beginner",
		LearningObjectives: []generator.LearningObjective{
			{ID: "LO1", Text: "Use letters to represent unknown quantities", Bloom: "remember"},
		},
	}

	result, err := generator.Translate(context.Background(), mock, &topic, "ms")
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}

	if !strings.Contains(result.Content, "Pemboleh ubah") {
		t.Error("Translation should contain BM terminology")
	}
}
```

**Step 2: Implement**

**File:** `internal/generator/translator.go`

```go
package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// LanguageNames maps language codes to full names.
var LanguageNames = map[string]string{
	"ms": "Bahasa Melayu",
	"zh": "Chinese (Simplified)",
	"ta": "Tamil",
	"en": "English",
}

// Translate translates a topic's content to the target language.
func Translate(ctx context.Context, provider ai.Provider, topic *Topic, targetLang string) (*GenerationResult, error) {
	langName, ok := LanguageNames[targetLang]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s (supported: %v)", targetLang, supportedLanguages())
	}

	prompt := buildTranslationPrompt(topic, langName)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are a professional translator specializing in mathematics education. Translate accurately while preserving YAML structure. Output ONLY valid YAML."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   2048,
		Temperature: 0.3, // Lower temperature for translation accuracy
	})
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	return &GenerationResult{
		Content:      resp.Content,
		Model:        resp.Model,
		InputTokens:  resp.InputTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

func buildTranslationPrompt(topic *Topic, targetLang string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Translate the following topic content to %s.\n\n", targetLang))
	sb.WriteString(fmt.Sprintf("Topic: %s (%s)\n\n", topic.Name, topic.ID))

	sb.WriteString("## Content to translate\n\n")
	sb.WriteString(fmt.Sprintf("name: %q\n\n", topic.Name))
	sb.WriteString("learning_objectives:\n")
	for _, lo := range topic.LearningObjectives {
		sb.WriteString(fmt.Sprintf("  - id: %s\n    text: %q\n", lo.ID, lo.Text))
	}

	sb.WriteString("\n## Rules\n")
	sb.WriteString("- Only translate human-readable text (name, text fields)\n")
	sb.WriteString("- Do NOT translate: id, bloom, difficulty, provenance values\n")
	sb.WriteString("- Use mathematically correct terminology\n")
	sb.WriteString("- Output ONLY the translated YAML\n")

	return sb.String()
}

func supportedLanguages() []string {
	langs := make([]string, 0, len(LanguageNames))
	for code := range LanguageNames {
		langs = append(langs, code)
	}
	return langs
}
```

#### 20.3 — End-to-End Pipeline Test

**File:** `internal/generator/pipeline_test.go`

```go
package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/generator"
)

func TestFullPipeline(t *testing.T) {
	// Setup: create a test repo with a topic
	repoDir := setupTestRepo(t)

	// Step 1: Build context
	genCtx, err := generator.BuildContext(repoDir, "F1-01")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}

	// Step 2: Generate teaching notes
	mock := ai.NewMockProvider("# F1-01 — Teaching Notes\n\n## Overview\nTest content.")
	notes, err := generator.GenerateTeachingNotes(context.Background(), mock, genCtx, "")
	if err != nil {
		t.Fatalf("GenerateTeachingNotes() error = %v", err)
	}
	if notes.Content == "" {
		t.Error("Teaching notes should not be empty")
	}

	// Step 3: Write output
	outPath := filepath.Join(repoDir, "curricula", "test", "topics", "algebra", "01-test.teaching.md")
	if err := os.WriteFile(outPath, []byte(notes.Content), 0o644); err != nil {
		t.Fatalf("Writing teaching notes: %v", err)
	}

	// Step 4: Verify file exists
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("Teaching notes file should exist after writing")
	}
}
```

#### 20.4 — Education Lead Evaluates AI Output Quality (🧑)

Using a real AI provider (not mock), generate content for 1 Form 2 topic and evaluate:

```bash
# Generate teaching notes with a real provider
OSS_AI_PROVIDER=anthropic OSS_AI_API_KEY=<key> OSS_REPO_PATH=../oss \
  go run ./cmd/oss generate teaching-notes F2-01

# Generate assessments
OSS_AI_PROVIDER=anthropic OSS_AI_API_KEY=<key> OSS_REPO_PATH=../oss \
  go run ./cmd/oss generate assessments F2-01 --count 5 --difficulty medium
```

**Education Lead review checklist:**
- [ ] Teaching notes are pedagogically sound for KSSM
- [ ] Assessment questions are at appropriate difficulty
- [ ] BM terminology is correct
- [ ] Content quality is acceptable with light editing
- [ ] What needs to improve in prompt templates?

#### Day 20 Validation

```bash
# Run ALL tests (mandatory)
go test ./...

# Build and verify all CLI commands
go build ./cmd/oss
./bin/oss --help
./bin/oss validate --help
./bin/oss generate --help
./bin/oss translate --help
./bin/oss quality --help
```

#### Day 20 Exit Criteria

- [ ] `prompts/translation.md` created
- [ ] `internal/generator/translator.go` + tests — translates topics
- [ ] `oss translate --topic <path> --to ms` command works
- [ ] End-to-end pipeline test passes (context → generate → write)
- [ ] Education Lead has evaluated AI-generated content quality
- [ ] Prompt template improvements applied based on feedback
- [ ] `go test ./...` passes with zero failures

**Week 4 Output:** Working CLI with `validate`, `generate` (teaching-notes, assessments, examples), `quality`, `translate` commands. AI provider interface (OpenAI, Anthropic, Ollama). Prompt templates for KSSM content.

**Progress:** CLI fully functional | 3 packages (validator, ai, generator) | 4 prompt templates | All tests green

---

## WEEK 5 — GITHUB BOT + DOCUMENT IMPORT

### Day 21 — GitHub App Setup

**Entry criteria:** Week 4 complete. CLI works with validate, generate, quality, translate. All tests green.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 21.1 | `B-W5D21-1` | GitHub App authentication (JWT + installation tokens) | 🤖 | `internal/github/app.go` |
| 21.2 | `B-W5D21-2` | Webhook handler with HMAC verification | 🤖 | `internal/github/webhook.go` |
| 21.3 | `B-W5D21-3` | Command parser for @oss-bot commands | 🤖 | `internal/parser/command.go` |
| 21.4 | `B-W5D21-4` | 🧑 Register GitHub App on GitHub | 🧑 | GitHub App created |

#### 21.1 — GitHub App Authentication (TDD)

**Step 1: Write tests**

**File:** `internal/github/app_test.go`

```go
package github_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func TestNewApp(t *testing.T) {
	keyPath := generateTestKey(t)

	app, err := gh.NewApp(12345, keyPath)
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}
}

func TestNewApp_MissingKey(t *testing.T) {
	_, err := gh.NewApp(12345, "/nonexistent/key.pem")
	if err == nil {
		t.Error("NewApp() should fail with missing key file")
	}
}

func TestGenerateJWT(t *testing.T) {
	keyPath := generateTestKey(t)
	app, err := gh.NewApp(12345, keyPath)
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	token, err := app.GenerateJWT()
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}
}

// generateTestKey creates a temporary RSA private key for testing.
func generateTestKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}

	path := filepath.Join(t.TempDir(), "test-key.pem")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := pem.Encode(f, pemBlock); err != nil {
		t.Fatal(err)
	}

	return path
}
```

**Step 2: Implement**

**File:** `internal/github/app.go`

```go
// Package github provides GitHub App authentication and API integration.
package github

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// App represents a GitHub App for authentication.
type App struct {
	AppID      int64
	PrivateKey *rsa.PrivateKey
}

// NewApp creates a GitHub App instance from an app ID and private key file.
func NewApp(appID int64, keyPath string) (*App, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %s", keyPath)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	return &App{
		AppID:      appID,
		PrivateKey: key,
	}, nil
}

// GenerateJWT creates a signed JWT for GitHub App authentication.
// The JWT is valid for 10 minutes (GitHub's maximum).
func (a *App) GenerateJWT() (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    fmt.Sprintf("%d", a.AppID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("signing JWT: %w", err)
	}

	return signed, nil
}
```

#### 21.2 — Webhook Handler (TDD)

**Step 1: Write tests**

**File:** `internal/github/webhook_test.go`

```go
package github_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func TestVerifySignature_Valid(t *testing.T) {
	secret := "test-secret"
	body := `{"action":"created","comment":{"body":"@oss-bot add teaching notes for F1-01"}}`
	signature := computeHMAC(body, secret)

	err := gh.VerifySignature([]byte(body), "sha256="+signature, secret)
	if err != nil {
		t.Errorf("VerifySignature() error = %v", err)
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	err := gh.VerifySignature([]byte("body"), "sha256=invalid", "secret")
	if err == nil {
		t.Error("VerifySignature() should fail with invalid signature")
	}
}

func TestWebhookHandler_IssueComment(t *testing.T) {
	handler := gh.NewWebhookHandler("test-secret", func(cmd gh.BotCommand) error {
		if cmd.Action != "add" {
			t.Errorf("Action = %q, want %q", cmd.Action, "add")
		}
		return nil
	})

	body := `{
		"action": "created",
		"comment": {
			"body": "@oss-bot add teaching notes for F1-01",
			"user": {"login": "testuser"}
		},
		"issue": {
			"number": 42
		},
		"repository": {
			"full_name": "p-n-ai/oss"
		}
	}`

	secret := "test-secret"
	signature := "sha256=" + computeHMAC(body, secret)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", signature)
	req.Header.Set("X-GitHub-Event", "issue_comment")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		body, _ := io.ReadAll(rr.Body)
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, string(body))
	}
}

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	handler := gh.NewWebhookHandler("test-secret", nil)

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader("{}"))
	req.Header.Set("X-Hub-Signature-256", "sha256=invalid")
	req.Header.Set("X-GitHub-Event", "issue_comment")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func computeHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
```

**Step 2: Implement**

**File:** `internal/github/webhook.go`

```go
package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// BotCommand represents a parsed @oss-bot command from an issue comment.
type BotCommand struct {
	Action     string   // "add", "translate", "scaffold", "import", "quality", "enrich"
	ContentType string  // "teaching notes", "assessments", "examples"
	TopicPath  string   // Path or ID of the target topic
	Options    map[string]string // Additional options (count, difficulty, language, url, etc.)
	User       string   // GitHub username who issued the command
	IssueNum   int      // Issue number where the command was posted
	RepoFullName string // "owner/repo"
	CommentBody  string // Full comment body (for enrich command)
	Attachments []Attachment // File attachments (for import command)
}

// Attachment represents a file attached to a GitHub comment.
type Attachment struct {
	URL      string // Download URL of the attachment
	FileName string // Original filename (e.g., "syllabus.pdf")
	MimeType string // Detected MIME type
}

// CommandHandler processes a parsed bot command.
type CommandHandler func(cmd BotCommand) error

// WebhookHandler handles incoming GitHub webhook events.
type WebhookHandler struct {
	secret  string
	handler CommandHandler
}

// NewWebhookHandler creates a new webhook handler with HMAC verification.
func NewWebhookHandler(secret string, handler CommandHandler) *WebhookHandler {
	return &WebhookHandler{secret: secret, handler: handler}
}

// ServeHTTP implements the http.Handler interface.
func (wh *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Verify HMAC signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if err := VerifySignature(body, signature, wh.secret); err != nil {
		slog.Warn("webhook signature verification failed", "error", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")

	switch eventType {
	case "issue_comment":
		wh.handleIssueComment(w, body)
	case "ping":
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	default:
		slog.Debug("ignoring event type", "type", eventType)
		w.WriteHeader(http.StatusOK)
	}
}

func (wh *WebhookHandler) handleIssueComment(w http.ResponseWriter, body []byte) {
	var event struct {
		Action  string `json:"action"`
		Comment struct {
			Body string `json:"body"`
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"comment"`
		Issue struct {
			Number int `json:"number"`
		} `json:"issue"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if event.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check for @oss-bot mention
	if !strings.Contains(event.Comment.Body, "@oss-bot") {
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd, err := ParseCommand(event.Comment.Body)
	if err != nil {
		slog.Warn("failed to parse command", "error", err, "body", event.Comment.Body)
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd.User = event.Comment.User.Login
	cmd.IssueNum = event.Issue.Number
	cmd.RepoFullName = event.Repository.FullName
	cmd.CommentBody = event.Comment.Body

	if wh.handler != nil {
		if err := wh.handler(*cmd); err != nil {
			slog.Error("command handler failed", "error", err, "command", cmd.Action)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// VerifySignature validates the HMAC-SHA256 signature of a webhook payload.
func VerifySignature(body []byte, signature, secret string) error {
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature format (expected sha256=...)")
	}

	expectedMAC := strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	actualMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(actualMAC), []byte(expectedMAC)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
```

#### 21.3 — Command Parser (TDD)

**Step 1: Write tests**

**File:** `internal/parser/command_test.go`

```go
package parser_test

import (
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAction string
		wantType   string
		wantTopic  string
		wantErr    bool
	}{
		{
			name:       "add-teaching-notes",
			input:      "@oss-bot add teaching notes for F2-01",
			wantAction: "add",
			wantType:   "teaching notes",
			wantTopic:  "F2-01",
		},
		{
			name:       "add-assessments-with-count",
			input:      "@oss-bot add 5 assessments for F1-01 difficulty:medium",
			wantAction: "add",
			wantType:   "assessments",
			wantTopic:  "F1-01",
		},
		{
			name:       "translate",
			input:      "@oss-bot translate F1-01 to ms",
			wantAction: "translate",
			wantTopic:  "F1-01",
		},
		{
			name:       "quality",
			input:      "@oss-bot quality malaysia-kssm-matematik-tingkatan1",
			wantAction: "quality",
			wantTopic:  "malaysia-kssm-matematik-tingkatan1",
		},
		{
			name:       "scaffold",
			input:      "@oss-bot scaffold syllabus india/cbse/mathematics-class10",
			wantAction: "scaffold",
			wantTopic:  "india/cbse/mathematics-class10",
		},
		{
			name:       "import-url",
			input:      "@oss-bot import https://example.org/curriculum-spec",
			wantAction: "import",
			wantTopic:  "",
		},
		{
			name:       "import-attachment",
			input:      "@oss-bot import",
			wantAction: "import",
			wantTopic:  "",
		},
		{
			name:       "import-attachment-with-vision",
			input:      "@oss-bot import vision:true",
			wantAction: "import",
			wantTopic:  "",
		},
		{
			name:    "no-bot-mention",
			input:   "Just a regular comment",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := gh.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if cmd.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", cmd.Action, tt.wantAction)
			}
			if tt.wantType != "" && cmd.ContentType != tt.wantType {
				t.Errorf("ContentType = %q, want %q", cmd.ContentType, tt.wantType)
			}
			if cmd.TopicPath != tt.wantTopic {
				t.Errorf("TopicPath = %q, want %q", cmd.TopicPath, tt.wantTopic)
			}
		})
	}
}
```

**Step 2: Implement**

The `ParseCommand` function goes in `internal/github/webhook.go` (or a separate `internal/parser/command.go` if preferred — both are valid, but keeping it in the github package avoids circular imports):

```go
// ParseCommand extracts a BotCommand from a comment body containing @oss-bot.
func ParseCommand(body string) (*BotCommand, error) {
	if !strings.Contains(body, "@oss-bot") {
		return nil, fmt.Errorf("no @oss-bot mention found")
	}

	// Extract the command portion after @oss-bot
	idx := strings.Index(body, "@oss-bot")
	rest := strings.TrimSpace(body[idx+len("@oss-bot"):])

	if rest == "" {
		return nil, fmt.Errorf("no command after @oss-bot")
	}

	cmd := &BotCommand{
		Options: make(map[string]string),
	}

	// Parse key:value options and remove them from rest
	parts := strings.Fields(rest)
	var cleanParts []string
	for _, p := range parts {
		if strings.Contains(p, ":") && !strings.HasPrefix(p, "/") {
			kv := strings.SplitN(p, ":", 2)
			if len(kv) == 2 {
				cmd.Options[kv[0]] = kv[1]
				continue
			}
		}
		cleanParts = append(cleanParts, p)
	}
	rest = strings.Join(cleanParts, " ")

	// Match command patterns
	switch {
	case strings.HasPrefix(rest, "add teaching notes"):
		cmd.Action = "add"
		cmd.ContentType = "teaching notes"
		cmd.TopicPath = extractTopicAfter(rest, "for")

	case strings.Contains(rest, "assessments"):
		cmd.Action = "add"
		cmd.ContentType = "assessments"
		cmd.TopicPath = extractTopicAfter(rest, "for")
		// Extract count if present (e.g., "add 5 assessments")
		for _, p := range cleanParts {
			if _, err := fmt.Sscanf(p, "%d", new(int)); err == nil {
				cmd.Options["count"] = p
			}
		}

	case strings.HasPrefix(rest, "translate"):
		cmd.Action = "translate"
		remaining := strings.TrimPrefix(rest, "translate ")
		toParts := strings.SplitN(remaining, " to ", 2)
		if len(toParts) == 2 {
			cmd.TopicPath = strings.TrimSpace(toParts[0])
			cmd.Options["to"] = strings.TrimSpace(toParts[1])
		} else {
			cmd.TopicPath = strings.TrimSpace(remaining)
		}

	case strings.HasPrefix(rest, "scaffold"):
		cmd.Action = "scaffold"
		cmd.TopicPath = strings.TrimSpace(strings.TrimPrefix(rest, "scaffold syllabus "))
		cmd.TopicPath = strings.TrimSpace(strings.TrimPrefix(cmd.TopicPath, "scaffold "))

	case strings.HasPrefix(rest, "quality"):
		cmd.Action = "quality"
		cmd.TopicPath = strings.TrimSpace(strings.TrimPrefix(rest, "quality "))

	case strings.HasPrefix(rest, "import"):
		cmd.Action = "import"
		remaining := strings.TrimSpace(strings.TrimPrefix(rest, "import"))
		// If remaining looks like a URL, store it as an option
		if strings.HasPrefix(remaining, "http://") || strings.HasPrefix(remaining, "https://") {
			cmd.Options["url"] = remaining
		} else {
			cmd.TopicPath = remaining
		}

	case strings.HasPrefix(rest, "enrich"):
		cmd.Action = "enrich"
		cmd.TopicPath = extractTopicAfter(rest, "enrich")

	default:
		return nil, fmt.Errorf("unrecognized command: %s", rest)
	}

	return cmd, nil
}

func extractTopicAfter(text, keyword string) string {
	idx := strings.Index(text, keyword)
	if idx < 0 {
		// Try to find the last word as topic
		parts := strings.Fields(text)
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return ""
	}
	return strings.TrimSpace(text[idx+len(keyword):])
}
```

#### 21.4 — Register GitHub App (🧑 Human)

Manual step. Create a GitHub App at `github.com/organizations/p-n-ai/settings/apps/new`:

- **App name:** `oss-bot`
- **Webhook URL:** `https://smee.io/<channel>` (dev) or production URL
- **Webhook secret:** generate with `openssl rand -hex 32`
- **Permissions:** Issues (R/W), Pull Requests (R/W), Contents (R/W)
- **Subscribe to events:** `issue_comment`, `pull_request`
- **Install on:** `p-n-ai/oss` repository
- **Download:** Private key `.pem` file

Set environment variables:
```bash
OSS_GITHUB_APP_ID=<from GitHub>
OSS_GITHUB_PRIVATE_KEY_PATH=./oss-bot.pem
OSS_GITHUB_WEBHOOK_SECRET=<the generated secret>
```

#### Day 21 Validation

```bash
# Run all tests
go test ./...

# Run github package tests
go test -v ./internal/github/...

# Run parser tests (if separate)
go test -v ./internal/parser/...
```

#### Day 21 Exit Criteria

- [ ] `internal/github/app.go` + tests — JWT generation from private key
- [ ] `internal/github/webhook.go` + tests — HMAC verification, event routing
- [ ] Command parser handles: add, translate, scaffold, quality, import (URL and attachment), enrich
- [ ] 🧑 GitHub App registered and installed on p-n-ai/oss
- [ ] `go test ./...` passes with zero failures

**Progress:** CLI + GitHub App auth + webhook handler + command parser | 4 packages

---

### Day 22 — Bot → PR Pipeline

**Entry criteria:** Day 21 complete. GitHub App registered. Webhook handler and parser work.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 22.1 | `B-W5D22-1` | PR creation (branch, commit, open PR) | 🤖 | `internal/github/pr.go` |
| 22.2 | `B-W5D22-2` | GitHub Contents API (read files from repo) | 🤖 | `internal/github/contents.go` |
| 22.3 | `B-W5D22-3` | Bot command flow: parse command → call shared pipeline (ModeCreatePR) → react with PR link | 🤖 | Wiring in `cmd/bot/main.go` |
| 22.4 | `B-W5D22-4` | Bot responds to issue with PR link | 🤖 | Part of PR flow |
| 22.5 | `B-W5D22-5` | Content merge strategies (assessments, examples, teaching notes) | 🤖 | `internal/generator/merge.go` |
| 22.6 | `B-W5D22-6` | Pipeline merge stage integration + MergeReport | 🤖 | Update `internal/pipeline/pipeline.go` |

> **Architecture note:** The bot reuses the same `pipeline.Execute()` from Day 19. It creates a `GitHubWriter` (instead of `LocalWriter`) and calls `pipeline.Execute(ctx, Request{Mode: ModeCreatePR, Source: "bot"})`. No generation logic is reimplemented — only the webhook-to-command parsing and GitHub reaction posting are bot-specific.

#### 22.1 — PR Creation (TDD)

**File:** `internal/github/pr_test.go`

```go
package github_test

import (
	"testing"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func TestNewPRRequest(t *testing.T) {
	pr := gh.PRRequest{
		Owner:       "p-n-ai",
		Repo:        "oss",
		Title:       "Add teaching notes for F1-01",
		Body:        "Generated by oss-bot",
		BranchName:  "oss-bot/teaching-notes-F1-01-20260228",
		BaseBranch:  "main",
		Files: []gh.FileChange{
			{Path: "topics/algebra/01.teaching.md", Content: "# Notes"},
		},
		Labels:      []string{"provenance:ai-generated"},
	}

	if pr.Owner == "" || pr.Repo == "" {
		t.Error("PR request should have owner and repo")
	}
	if len(pr.Files) == 0 {
		t.Error("PR request should have files")
	}
	if pr.BranchName == "" {
		t.Error("PR request should have branch name")
	}
}

func TestGenerateBranchName(t *testing.T) {
	name := gh.GenerateBranchName("add", "teaching-notes", "F1-01")
	if name == "" {
		t.Error("GenerateBranchName() returned empty")
	}
	if len(name) > 255 {
		t.Error("Branch name too long")
	}
}
```

**File:** `internal/github/pr.go`

```go
package github

import (
	"fmt"
	"strings"
	"time"
)

// PRRequest holds all data needed to create a pull request.
type PRRequest struct {
	Owner      string
	Repo       string
	Title      string
	Body       string
	BranchName string
	BaseBranch string
	Files      []FileChange
	Labels     []string
	Reviewers  []string
}

// FileChange represents a file to be created or modified in the PR.
type FileChange struct {
	Path    string
	Content string
}

// PRResult holds the result of creating a pull request.
type PRResult struct {
	Number int
	URL    string
	Branch string
}

// GenerateBranchName creates a consistent branch name for bot-created PRs.
func GenerateBranchName(action, contentType, topicID string) string {
	timestamp := time.Now().Format("20060102-150405")
	safeTopic := strings.ReplaceAll(topicID, "/", "-")
	safeType := strings.ReplaceAll(contentType, " ", "-")
	return fmt.Sprintf("oss-bot/%s-%s-%s-%s", action, safeType, safeTopic, timestamp)
}

// BuildPRBody creates the PR description with provenance metadata.
func BuildPRBody(cmd BotCommand, model string, generatedAt time.Time) string {
	var sb strings.Builder

	sb.WriteString("## Auto-generated by OSS Bot\n\n")
	sb.WriteString(fmt.Sprintf("**Requested by:** @%s\n", cmd.User))
	sb.WriteString(fmt.Sprintf("**Command:** `@oss-bot %s`\n", cmd.Action))
	sb.WriteString(fmt.Sprintf("**Topic:** `%s`\n", cmd.TopicPath))
	sb.WriteString(fmt.Sprintf("**Model:** %s\n", model))
	sb.WriteString(fmt.Sprintf("**Generated at:** %s\n\n", generatedAt.Format(time.RFC3339)))

	sb.WriteString("### Provenance\n")
	sb.WriteString("```yaml\n")
	sb.WriteString("provenance: ai-generated\n")
	sb.WriteString(fmt.Sprintf("model: %s\n", model))
	sb.WriteString(fmt.Sprintf("generator: oss-bot\n"))
	sb.WriteString(fmt.Sprintf("generated_at: %q\n", generatedAt.Format(time.RFC3339)))
	sb.WriteString("```\n\n")

	sb.WriteString("### Review Checklist\n")
	sb.WriteString("- [ ] Content is pedagogically sound\n")
	sb.WriteString("- [ ] Mathematical notation is correct\n")
	sb.WriteString("- [ ] BM terminology is accurate\n")
	sb.WriteString("- [ ] Passes schema validation\n")

	return sb.String()
}
```

#### 22.2 — GitHub Contents API (TDD)

**File:** `internal/github/contents.go`

```go
package github

// ContentsClient reads and writes files via the GitHub Contents API.
// In production, this uses go-github/v62. In tests, it uses a mock.
type ContentsClient interface {
	ReadFile(owner, repo, path, ref string) ([]byte, error)
	ListDir(owner, repo, path, ref string) ([]string, error)
}
```

Create mock implementation for tests and real implementation using `google/go-github/v62`.

#### 22.3 — Wire Bot Command Flow

Update `cmd/bot/main.go` to create the full server:

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	gh "github.com/p-n-ai/oss-bot/internal/github"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	port := os.Getenv("OSS_BOT_PORT")
	if port == "" {
		port = "8090"
	}

	webhookSecret := os.Getenv("OSS_GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		slog.Error("OSS_GITHUB_WEBHOOK_SECRET is required")
		os.Exit(1)
	}

	handler := gh.NewWebhookHandler(webhookSecret, handleCommand)

	mux := http.NewServeMux()
	mux.Handle("POST /webhook", handler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	slog.Info("oss-bot server starting", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handleCommand(cmd gh.BotCommand) error {
	slog.Info("received command",
		"action", cmd.Action,
		"topic", cmd.TopicPath,
		"user", cmd.User,
		"issue", cmd.IssueNum,
	)

	// Route to appropriate handler
	switch cmd.Action {
	case "add":
		return handleAdd(cmd)
	case "translate":
		return handleTranslate(cmd)
	case "quality":
		return handleQuality(cmd)
	case "scaffold":
		return handleScaffold(cmd)
	default:
		return fmt.Errorf("unhandled action: %s", cmd.Action)
	}
}

func handleAdd(cmd gh.BotCommand) error {
	// 1. Read topic from GitHub Contents API
	// 2. Build generation context
	// 3. Call AI provider
	// 4. Validate output
	// 5. Create branch + commit + PR
	// 6. Comment on issue with PR link
	slog.Info("handling add command", "type", cmd.ContentType, "topic", cmd.TopicPath)
	return nil // Placeholder — wired with real logic
}

func handleTranslate(cmd gh.BotCommand) error {
	slog.Info("handling translate command", "topic", cmd.TopicPath, "to", cmd.Options["to"])
	return nil
}

func handleQuality(cmd gh.BotCommand) error {
	slog.Info("handling quality command", "topic", cmd.TopicPath)
	return nil
}

func handleScaffold(cmd gh.BotCommand) error {
	slog.Info("handling scaffold command", "topic", cmd.TopicPath)
	return nil
}
```

#### 22.5 — Content Merge Strategies (TDD)

**File:** `internal/generator/merge.go`

When generating content for topics that already have existing data, the pipeline must merge new content with existing content rather than overwriting.

Three merge strategies:

| Content Type | Strategy | Behavior |
|-------------|----------|----------|
| Assessments | Append + dedup | Add new questions, skip duplicates by matching question text (fuzzy) |
| Examples | Append + dedup + re-sort | Add new examples, skip duplicates, re-sort by difficulty/complexity |
| Teaching Notes | Additive | Merge sections additively — never remove existing notes, append new sections |

```go
// MergeReport tracks what changed during a merge for PR descriptions.
type MergeReport struct {
    Added   int      // Number of new items added
    Skipped int      // Number of duplicates skipped
    Sections []string // Which sections were modified
}

// MergeAssessments appends new assessments to existing, deduplicating by question text.
func MergeAssessments(existing, generated []Assessment) ([]Assessment, MergeReport) { ... }

// MergeExamples appends new examples, deduplicates, and re-sorts by difficulty.
func MergeExamples(existing, generated []Example) ([]Example, MergeReport) { ... }

// MergeTeachingNotes additively merges teaching note sections.
func MergeTeachingNotes(existing, generated TeachingNotes) (TeachingNotes, MergeReport) { ... }
```

The `MergeReport` struct is used by the pipeline to build informative PR descriptions showing exactly what was added vs. skipped.

#### 22.6 — Pipeline Merge Stage Integration

Update `internal/pipeline/pipeline.go` to call the appropriate merge function when existing content is detected. The pipeline reads current content via the `ContentsClient` (GitHub API or filesystem), then merges before writing output.

#### Day 22 Validation

```bash
# Run all tests
go test ./...

# Build bot server
go build ./cmd/bot

# Test server starts (ctrl+C to stop)
OSS_GITHUB_WEBHOOK_SECRET=test go run ./cmd/bot &
curl http://localhost:8090/health
kill %1
```

#### Day 22 Exit Criteria

- [ ] `internal/github/pr.go` + tests — PR request builder, branch naming, PR body with provenance
- [ ] `internal/github/contents.go` + tests — read files from GitHub API
- [ ] `cmd/bot/main.go` — working HTTP server with webhook route and health check
- [ ] Bot command flow: webhook → parse → route to handler
- [ ] `internal/generator/merge.go` + tests — MergeAssessments (append + dedup), MergeExamples (append + dedup + re-sort), MergeTeachingNotes (additive)
- [ ] Pipeline merge stage integration — merges with existing content when present
- [ ] MergeReport struct populates PR descriptions with added/skipped counts
- [ ] `go test ./...` passes with zero failures

**Progress:** CLI + Bot server + GitHub integration + content merge | 5 packages

---

### Day 23 — Content Import (URL, Upload, Text) + More Commands

**Entry criteria:** Day 22 complete. Bot server starts and handles webhooks.

#### Architecture Decision: Three Input Methods

Users can contribute content via three input methods, all supported across every interface (Web Portal, GitHub Bot, CLI):

1. **URL** — paste a link to a curriculum page or hosted document. The system fetches the page, extracts text, and structures it into YAML.
2. **Text** — type or paste content directly (structured or freeform natural language, any language).
3. **Upload** — attach a file. Supported formats: PDF (`.pdf`), Word (`.docx`), PowerPoint (`.pptx`), plain text (`.txt`), and images (`.png`, `.jpg`, `.jpeg` — extracted via OCR for printed text or AI Vision for handwriting/diagrams).

#### Architecture Decision: Hybrid Content Extraction

The project uses a **hybrid approach** for content extraction:
- **CLI:** Uses `ledongthuc/pdf` (Go-native) for PDFs, Tesseract for image OCR, and AI Vision (via the existing AI provider interface) for handwriting/diagrams. Lightweight, no Docker needed.
- **Server (Bot + Web Portal):** Uses Apache Tika as a Docker sidecar via `google/go-tika`. Handles PDF, DOCX, PPTX, TXT, images (OCR via Tika's Tesseract, AI Vision via AI provider), and 1000+ other formats.
- **URL fetching:** Shared across CLI and server. Uses Go `net/http` with optional headless rendering for JavaScript-heavy pages.

All extractors share the `ContentExtractor` interface, allowing the import pipeline to work identically regardless of source.

**Why hybrid:**
- CLI stays lightweight and self-contained (single Go binary, no Docker)
- Server gets broad format support where it matters (teachers submit DOCX, PPTX, images — not just PDF)
- Apache Tika is battle-tested (15+ years, Apache Foundation), has a native Go client (`google/go-tika`), and runs as a simple Docker sidecar
- URL fetching enables importing from government curriculum pages, publisher sites, and online specifications

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 23.1 | `B-W5D23-1` | Content import prompt template | 🤖 | `prompts/document_import.md` |
| 23.2 | `B-W5D23-2` | ContentExtractor interface | 🤖 | `internal/parser/document.go` |
| 23.3 | `B-W5D23-3` | Go-native PDF extraction (CLI) | 🤖 | `internal/parser/pdf.go` |
| 23.4 | `B-W5D23-4` | Apache Tika client (Server) | 🤖 | `internal/parser/tika.go` |
| 23.5 | `B-W5D23-5` | URL fetcher | 🤖 | `internal/parser/url.go` |
| 23.6 | `B-W5D23-6` | Image extraction (OCR + AI Vision) | 🤖 | `internal/parser/image.go` |
| 23.7 | `B-W5D23-7` | Scaffolder (any source → syllabus structure) | 🤖 | `internal/generator/scaffolder.go` |
| 23.8 | `B-W5D23-8` | `@oss-bot quality` command implementation | 🤖 | Update bot handlers |
| 23.9 | `B-W5D23-9` | `oss scaffold syllabus` and `oss scaffold subject` CLI commands | 🤖 | Update `cmd/oss/main.go` |
| 23.10 | `B-W5D23-10` | Bulk import prompt template | 🤖 | `prompts/bulk_import.md` |

#### 23.1 — Content Import Prompt Template

**File:** `prompts/document_import.md`

> **Curriculum-agnostic:** All prompt templates must use `{{syllabus_id}}` template variables and not hardcode any specific curriculum (e.g., KSSM, BM). The syllabus ID is passed at runtime.

```markdown
# Content Import Prompt

You are an expert at extracting curriculum structure from educational sources.

## Context

**Syllabus:** {{syllabus_id}}
**Source content (pre-extracted text):**
{{document_text}}

**Source type:** {{source_type}}  <!-- "url", "pdf", "docx", "pptx", "txt", "image_ocr", "image_vision", "text" -->
**Image extraction method (if image):** {{image_method}}  <!-- "ocr", "vision", or empty -->
**Source URL (if applicable):** {{source_url}}
**Source format:** {{source_format}}
**Target board:** {{board}}
**Target level:** {{level}}

## Instructions

Extract the curriculum structure from the source text and output as YAML:

1. Identify subjects/strands
2. Identify individual topics within each subject
3. For each topic, determine:
   - A unique ID (format: XX-NN)
   - Name in source language
   - Learning objectives (with Bloom's levels inferred from verbs)
   - Difficulty (beginner/intermediate/advanced)
   - Prerequisites (which topics should come before)

## Output Format (YAML)

subjects:
  - id: subject-id
    name: "Subject Name"
    topics:
      - id: XX-01
        name: "Topic Name"
        difficulty: beginner
        learning_objectives:
          - id: LO1
            text: "..."
            bloom: understand
        prerequisites: []
```

#### 23.2 — ContentExtractor Interface (TDD)

**File:** `internal/parser/document.go`

```go
// Package parser handles input parsing (content extraction, natural language, commands).
package parser

import "context"

// ContentExtractor extracts text from various sources for the AI import pipeline.
// Implementations: PDFParser (CLI), TikaParser (server), URLFetcher, ImageExtractor.
type ContentExtractor interface {
	// Extract converts a source to plain text for AI processing.
	// input is the raw file bytes (for files/images) or nil (for URL fetcher).
	// mimeType hints at the format (e.g., "application/pdf", "image/png").
	Extract(ctx context.Context, input []byte, mimeType string) (string, error)

	// SupportedTypes returns the MIME types this extractor handles.
	SupportedTypes() []string
}

// ImageExtractionMode controls how images are processed.
type ImageExtractionMode int

const (
	// ImageModeAuto tries OCR first; if OCR returns low-confidence or sparse
	// text, falls back to AI Vision.
	ImageModeAuto ImageExtractionMode = iota
	// ImageModeOCR forces Tesseract/Tika OCR only (fast, no API cost).
	ImageModeOCR
	// ImageModeVision forces AI Vision via the AI provider (GPT-4o/Claude).
	// Best for handwriting, diagrams, flowcharts, whiteboard photos, complex layouts.
	ImageModeVision
)

// URLFetcher fetches and extracts text from web pages.
type URLFetcher interface {
	// Fetch retrieves a web page and returns its text content.
	// Handles static HTML and optionally renders JavaScript-heavy pages.
	Fetch(ctx context.Context, url string) (string, error)
}

// InputSource represents the three ways users can provide content.
type InputSource struct {
	Type      string              // "url", "text", "file"
	URL       string              // For URL input
	Text      string              // For text (copy-paste) input
	FileData  []byte              // For file upload input
	FileName  string              // Original filename (used to detect MIME type)
	MimeType  string              // MIME type of uploaded file
	ImageMode ImageExtractionMode // For images: Auto, OCR, or Vision
}
```

#### 23.3 — Go-Native PDF Extraction for CLI (TDD)

**File:** `internal/parser/pdf_test.go`

```go
package parser_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestPDFParser(t *testing.T) {
	p := parser.NewPDFParser()

	t.Run("supported-types", func(t *testing.T) {
		types := p.SupportedTypes()
		if len(types) != 1 || types[0] != "application/pdf" {
			t.Errorf("SupportedTypes() = %v, want [application/pdf]", types)
		}
	})

	t.Run("non-pdf-content", func(t *testing.T) {
		_, err := p.Extract(context.Background(), []byte("not a pdf"), "text/plain")
		if err == nil {
			t.Error("Extract() should error for non-PDF content")
		}
	})

	t.Run("empty-input", func(t *testing.T) {
		_, err := p.Extract(context.Background(), nil, "application/pdf")
		if err == nil {
			t.Error("Extract() should error for empty input")
		}
	})
}

// Legacy function tests for backwards compatibility
func TestExtractPDFText(t *testing.T) {
	t.Run("non-existent-file", func(t *testing.T) {
		_, err := parser.ExtractPDFText("/nonexistent/file.pdf")
		if err == nil {
			t.Error("ExtractPDFText() should error for non-existent file")
		}
	})

	t.Run("non-pdf-file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "test.txt")
		os.WriteFile(f, []byte("not a pdf"), 0o644)
		_, err := parser.ExtractPDFText(f)
		if err == nil {
			t.Error("ExtractPDFText() should error for non-PDF file")
		}
	})
}
```

**File:** `internal/parser/pdf.go`

```go
package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PDFParser implements DocumentParser using Go-native PDF extraction.
// Used by the CLI for standalone operation without external dependencies.
type PDFParser struct{}

// NewPDFParser creates a new Go-native PDF parser.
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Extract(ctx context.Context, input []byte, mimeType string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("empty input")
	}
	if mimeType != "" && mimeType != "application/pdf" {
		return "", fmt.Errorf("unsupported MIME type for PDFParser: %s (only application/pdf supported)", mimeType)
	}

	// PDF extraction using ledongthuc/pdf
	// TODO: Implement with ledongthuc/pdf library
	return "", fmt.Errorf("PDF extraction not yet implemented — install ledongthuc/pdf")
}

func (p *PDFParser) SupportedTypes() []string {
	return []string{"application/pdf"}
}

// ExtractPDFText is a convenience function for CLI file-based extraction.
func ExtractPDFText(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".pdf" {
		return "", fmt.Errorf("not a PDF file: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}

	p := NewPDFParser()
	return p.Extract(context.Background(), data, "application/pdf")
}
```

#### 23.4 — Apache Tika Client for Server (TDD)

**File:** `internal/parser/tika_test.go`

```go
package parser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestTikaParser(t *testing.T) {
	// Mock Tika server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Extracted text from document"))
	}))
	defer server.Close()

	p := parser.NewTikaParser(server.URL)

	t.Run("supported-types", func(t *testing.T) {
		types := p.SupportedTypes()
		if len(types) < 5 {
			t.Errorf("SupportedTypes() should return many types, got %d", len(types))
		}
	})

	t.Run("extract-document", func(t *testing.T) {
		text, err := p.Extract(context.Background(), []byte("fake doc content"), "application/pdf")
		if err != nil {
			t.Fatalf("Extract() error = %v", err)
		}
		if text == "" {
			t.Error("Extract() returned empty text")
		}
	})

	t.Run("empty-input", func(t *testing.T) {
		_, err := p.Extract(context.Background(), nil, "application/pdf")
		if err == nil {
			t.Error("Extract() should error for empty input")
		}
	})
}

func TestTikaParserUnreachable(t *testing.T) {
	p := parser.NewTikaParser("http://localhost:1") // unreachable

	_, err := p.Extract(context.Background(), []byte("content"), "application/pdf")
	if err == nil {
		t.Error("Extract() should error when Tika is unreachable")
	}
}
```

**File:** `internal/parser/tika.go`

```go
package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"bytes"
)

// TikaParser implements DocumentParser using Apache Tika server.
// Used by the Bot and Web Portal for multi-format document extraction.
// Requires a running Tika instance (Docker sidecar).
type TikaParser struct {
	tikaURL string
	client  *http.Client
}

// NewTikaParser creates a parser that connects to an Apache Tika server.
// tikaURL is typically "http://tika:9998" in Docker or "http://localhost:9998" locally.
func NewTikaParser(tikaURL string) *TikaParser {
	return &TikaParser{
		tikaURL: tikaURL,
		client:  &http.Client{},
	}
}

func (p *TikaParser) Extract(ctx context.Context, input []byte, mimeType string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("empty input")
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", p.tikaURL+"/tika", bytes.NewReader(input))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "text/plain")
	if mimeType != "" {
		req.Header.Set("Content-Type", mimeType)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Tika server at %s: %w", p.tikaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Tika returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading Tika response: %w", err)
	}

	return string(body), nil
}

func (p *TikaParser) SupportedTypes() []string {
	return []string{
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",   // DOCX
		"application/vnd.openxmlformats-officedocument.presentationml.presentation", // PPTX
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",         // XLSX
		"text/plain",       // TXT
		"text/html",
		"image/png",        // Images (Tika uses Tesseract OCR; AI Vision handled by ImageExtractor)
		"image/jpeg",       // JPG/JPEG (same as above)
		"application/rtf",
		"application/epub+zip",
	}
}
```

#### 23.5 — URL Fetcher (TDD)

**File:** `internal/parser/url_test.go`

```go
package parser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestURLFetcher(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><h1>Mathematics Syllabus</h1><p>Topic 1: Algebra</p></body></html>`))
	}))
	defer server.Close()

	f := parser.NewURLFetcher()

	t.Run("fetch-html-page", func(t *testing.T) {
		text, err := f.Fetch(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("Fetch() error = %v", err)
		}
		if text == "" {
			t.Error("Fetch() returned empty text")
		}
	})

	t.Run("invalid-url", func(t *testing.T) {
		_, err := f.Fetch(context.Background(), "http://localhost:1/nonexistent")
		if err == nil {
			t.Error("Fetch() should error for unreachable URL")
		}
	})

	t.Run("empty-url", func(t *testing.T) {
		_, err := f.Fetch(context.Background(), "")
		if err == nil {
			t.Error("Fetch() should error for empty URL")
		}
	})
}
```

**File:** `internal/parser/url.go`

```go
package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// HTTPURLFetcher implements URLFetcher using Go net/http.
// Fetches web pages and extracts visible text content.
type HTTPURLFetcher struct {
	client *http.Client
}

// NewURLFetcher creates a new URL fetcher with sensible defaults.
func NewURLFetcher() *HTTPURLFetcher {
	return &HTTPURLFetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *HTTPURLFetcher) Fetch(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "oss-bot/1.0 (curriculum importer)")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching URL %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("URL returned status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")

	// If the response is HTML, extract text from the DOM
	if strings.Contains(contentType, "text/html") {
		return extractTextFromHTML(resp.Body)
	}

	// For other text types, read directly
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}
	return string(body), nil
}

// extractTextFromHTML walks the HTML DOM and extracts visible text.
func extractTextFromHTML(r io.Reader) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", fmt.Errorf("parsing HTML: %w", err)
	}

	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		// Skip script, style, and nav elements
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "nav", "footer", "header":
				return
			}
		}
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				sb.WriteString(text)
				sb.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return strings.TrimSpace(sb.String()), nil
}
```

#### 23.6 — Image Extraction: OCR + AI Vision (TDD)

The image extractor supports two extraction methods:

| Method | Best For | Implementation | Cost |
|--------|----------|----------------|------|
| **OCR** (Tesseract/Tika) | Clean printed text — scanned docs, typed content in photos | `os/exec` → `tesseract` (CLI) or Tika's built-in Tesseract (server) | Free |
| **AI Vision** (GPT-4o/Claude) | Handwritten notes, diagrams, flowcharts, whiteboard photos, textbook page layouts, tables | Sends base64 image to multimodal AI via the existing `ai.Provider` interface | API cost per image |

**Auto-detection logic:** Try OCR first. If OCR returns fewer than 20 characters or Tesseract reports confidence below 60%, fall back to AI Vision. Users can override with `--vision` (CLI), `vision:true` (bot), or the "Use AI Vision" toggle (web portal).

**File:** `internal/parser/image_test.go`

```go
package parser_test

import (
	"context"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestImageExtractor(t *testing.T) {
	// OCR-only extractor (no AI provider)
	p := parser.NewImageExtractor(nil, parser.ImageModeOCR)

	t.Run("supported-types", func(t *testing.T) {
		types := p.SupportedTypes()
		if len(types) < 2 {
			t.Errorf("SupportedTypes() should include png and jpeg, got %v", types)
		}
	})

	t.Run("empty-input", func(t *testing.T) {
		_, err := p.Extract(context.Background(), nil, "image/png")
		if err == nil {
			t.Error("Extract() should error for empty input")
		}
	})

	t.Run("non-image-type", func(t *testing.T) {
		_, err := p.Extract(context.Background(), []byte("not an image"), "application/pdf")
		if err == nil {
			t.Error("Extract() should error for non-image MIME type")
		}
	})
}

func TestImageExtractorVisionMode(t *testing.T) {
	// Mock AI provider that returns a fixed response
	mockProvider := &mockAIProvider{
		response: "Topic: Algebra\nLearning Objective: Solve linear equations",
	}
	p := parser.NewImageExtractor(mockProvider, parser.ImageModeVision)

	t.Run("vision-extracts-content", func(t *testing.T) {
		// Minimal valid PNG header
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		text, err := p.Extract(context.Background(), pngHeader, "image/png")
		if err != nil {
			t.Fatalf("Extract() error = %v", err)
		}
		if text == "" {
			t.Error("Extract() returned empty text from vision")
		}
	})

	t.Run("vision-requires-provider", func(t *testing.T) {
		noProvider := parser.NewImageExtractor(nil, parser.ImageModeVision)
		_, err := noProvider.Extract(context.Background(), []byte{0x89}, "image/png")
		if err == nil {
			t.Error("Extract() should error when AI provider is nil in vision mode")
		}
	})
}
```

**File:** `internal/parser/image.go`

```go
package parser

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// ImageExtractor implements ContentExtractor for image files.
// Supports two extraction methods:
//   - OCR (Tesseract/Tika): fast, free, best for clean printed text
//   - AI Vision (GPT-4o/Claude): handles handwriting, diagrams, flowcharts,
//     whiteboard photos, complex layouts, and tables in images
type ImageExtractor struct {
	aiProvider ai.Provider
	mode       ImageExtractionMode
}

// NewImageExtractor creates a new image extractor.
// aiProvider is required for Vision mode (can be nil for OCR-only).
// mode controls extraction: ImageModeAuto, ImageModeOCR, or ImageModeVision.
func NewImageExtractor(provider ai.Provider, mode ImageExtractionMode) *ImageExtractor {
	return &ImageExtractor{
		aiProvider: provider,
		mode:       mode,
	}
}

func (p *ImageExtractor) Extract(ctx context.Context, input []byte, mimeType string) (string, error) {
	if len(input) == 0 {
		return "", fmt.Errorf("empty input")
	}
	if !p.isImageType(mimeType) {
		return "", fmt.Errorf("unsupported MIME type for ImageExtractor: %s", mimeType)
	}

	switch p.mode {
	case ImageModeOCR:
		return p.extractOCR(ctx, input)
	case ImageModeVision:
		return p.extractVision(ctx, input, mimeType)
	case ImageModeAuto:
		// Try OCR first; fall back to Vision if result is sparse
		text, err := p.extractOCR(ctx, input)
		if err == nil && len(strings.TrimSpace(text)) >= 20 {
			return text, nil
		}
		// OCR returned sparse/empty text — try AI Vision
		if p.aiProvider != nil {
			return p.extractVision(ctx, input, mimeType)
		}
		// No AI provider available — return whatever OCR got
		if err != nil {
			return "", fmt.Errorf("OCR failed and no AI provider for vision fallback: %w", err)
		}
		return text, nil
	default:
		return "", fmt.Errorf("unknown image extraction mode: %d", p.mode)
	}
}

// extractOCR uses Tesseract CLI for text extraction.
func (p *ImageExtractor) extractOCR(ctx context.Context, input []byte) (string, error) {
	// TODO: Implement with os/exec call to tesseract binary
	// Write input to temp file, run: tesseract <input> stdout
	// Parse stdout for extracted text
	return "", fmt.Errorf("OCR extraction not yet implemented — install tesseract")
}

// extractVision sends the image to a multimodal AI model.
func (p *ImageExtractor) extractVision(ctx context.Context, input []byte, mimeType string) (string, error) {
	if p.aiProvider == nil {
		return "", fmt.Errorf("AI provider required for vision extraction")
	}

	// Encode image as base64 for the AI provider
	b64 := base64.StdEncoding.EncodeToString(input)

	resp, err := p.aiProvider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{
				Role: "user",
				Content: []ai.ContentBlock{
					{
						Type: "image",
						Source: &ai.ImageSource{
							Type:      "base64",
							MediaType: mimeType,
							Data:      b64,
						},
					},
					{
						Type: "text",
						Text: "Extract all educational content from this image. " +
							"Identify: subject areas, topic names, learning objectives, " +
							"assessment questions, teaching notes, diagrams descriptions, " +
							"and any curriculum structure. " +
							"If there is handwritten text, transcribe it accurately. " +
							"If there are diagrams or flowcharts, describe their structure and content. " +
							"Output as plain text, preserving the logical structure.",
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("AI vision extraction failed: %w", err)
	}

	return resp.Text, nil
}

func (p *ImageExtractor) SupportedTypes() []string {
	return []string{"image/png", "image/jpeg"}
}

func (p *ImageExtractor) isImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}
```

#### 23.7 — Scaffolder (TDD)

**File:** `internal/generator/scaffolder_test.go` and `internal/generator/scaffolder.go`

The scaffolder takes extracted content (from any source — URL, file, or text) and generates a complete syllabus directory structure with Level 0-1 topic stubs.

> **Important:** The scaffolder must handle creating entirely new curricula from scratch (not just importing into existing ones). This means it must be able to generate the top-level `syllabus.yaml`, subject directories, and all topic stubs from a single source document or URL, without requiring any pre-existing OSS repo content.

#### 23.9 — Scaffold CLI Commands

Add `oss scaffold syllabus` and `oss scaffold subject` subcommands:

```bash
# Create an entirely new syllabus from a curriculum specification document
oss scaffold syllabus --from-url https://example.gov/curriculum-spec.pdf --id my-syllabus
oss scaffold syllabus --from-file curriculum.pdf --id my-syllabus

# Create a new subject within an existing syllabus
oss scaffold subject --syllabus my-syllabus --from-file math-spec.docx --id mathematics
```

Both commands delegate to the scaffolder (`internal/generator/scaffolder.go`) and output the generated directory tree to the filesystem. The `--id` flag sets the `syllabus_id` used in template variables.

#### 23.10 — Bulk Import Prompt Template

**File:** `prompts/bulk_import.md`

The bulk import prompt is used when importing large documents that contain multiple topics. It guides the AI to extract and structure many topics at once.

> **Curriculum-agnostic:** This template (like `document_import.md` and `contribution_parser.md`) must use `{{syllabus_id}}` template variables and not hardcode any specific curriculum (e.g., KSSM, BM). The syllabus ID is passed at runtime.

```markdown
# Bulk Import Prompt

You are extracting curriculum structure from a large educational document.

## Context

**Syllabus:** {{syllabus_id}}
**Source content (chunk {{chunk_index}} of {{total_chunks}}):**
{{document_chunk}}

**Source type:** {{source_type}}
**Previously extracted topics (for continuity):**
{{previous_topics}}

## Instructions

Extract all curriculum topics from this chunk. For each topic:
1. Assign a unique ID following the pattern used in {{syllabus_id}}
2. Extract the topic name in the source language
3. Identify learning objectives with Bloom's taxonomy levels
4. Determine difficulty (beginner/intermediate/advanced)
5. Identify prerequisites (referencing topic IDs from this or previous chunks)
6. Extract any teaching notes, examples, or assessment items present

Maintain consistency with topics extracted from previous chunks.

## Output Format (YAML)

topics:
  - id: XX-NN
    name: "Topic Name"
    difficulty: beginner
    learning_objectives:
      - id: LO1
        text: "..."
        bloom: understand
    prerequisites: []
    teaching_notes: "..." # if present in source
    examples: []          # if present in source
    assessments: []       # if present in source
```

#### Day 23 Validation

```bash
go test ./...
```

#### Day 23 Exit Criteria

- [ ] `prompts/document_import.md` created (supports URL, file, and text sources; uses `{{syllabus_id}}` not hardcoded curricula)
- [ ] `internal/parser/document.go` — `ContentExtractor` interface + `InputSource` struct
- [ ] `internal/parser/pdf.go` + tests — Go-native PDF extraction (CLI)
- [ ] `internal/parser/tika.go` + tests — Apache Tika multi-format extraction including images (server)
- [ ] `internal/parser/url.go` + tests — URL fetcher with HTML text extraction
- [ ] `internal/parser/image.go` + tests — Dual image extraction: OCR for printed text, AI Vision for handwriting/diagrams (with auto-detection)
- [ ] `internal/generator/scaffolder.go` + tests — generates syllabus structure from any source, including entirely new curricula from scratch
- [ ] `oss scaffold syllabus` and `oss scaffold subject` CLI commands wired and functional
- [ ] `prompts/bulk_import.md` created (uses `{{syllabus_id}}` template variable, curriculum-agnostic)
- [ ] `@oss-bot quality` responds with quality report as issue comment
- [ ] `go test ./...` passes

**Progress:** CLI + Bot + content import (URL, upload with OCR + AI Vision, text) + scaffolder + scaffold commands | 5 packages | 6 prompt templates

---

### Day 24 — Contribution Parser + Feedback API

**Entry criteria:** Day 23 complete. Document import (hybrid) and scaffolder work.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 24.1 | `B-W5D24-1` | Contribution parser prompt template | 🤖 | `prompts/contribution_parser.md` |
| 24.2 | `B-W5D24-2` | Natural language → structured data parser | 🤖 | `internal/parser/contribution.go` |
| 24.3 | `B-W5D24-3` | `POST /api/feedback` endpoint | 🤖 | `internal/api/feedback.go` |
| 24.4 | `B-W5D24-4` | Feedback → PR pipeline | 🤖 | Wiring |
| 24.5 | `B-W5D24-5` | Large document chunker | 🤖 | `internal/parser/chunker.go` |
| 24.6 | `B-W5D24-6` | Bulk import parallel worker pool | 🤖 | `internal/pipeline/bulk.go` |
| 24.7 | `B-W5D24-7` | Progress reporter interface | 🤖 | `internal/pipeline/progress.go` |
| 24.8 | `B-W5D24-8` | Reasoning model provider | 🤖 | `internal/ai/reasoning.go` |
| 24.9 | `B-W5D24-9` | Extended Bloom verb sets for cross-subject support | 🤖 | Update `internal/validator/bloom.go` |

#### 24.1 — Contribution Parser Prompt

**File:** `prompts/contribution_parser.md`

```markdown
# Contribution Parser Prompt

You are helping parse a teacher's natural language contribution into structured curriculum data.

## Teacher's Input
{{contribution_text}}

## Target Topic
{{topic_name}} ({{topic_id}})

## Instructions

Parse the teacher's input and identify what type of contribution it is:
- misconception: a common student error the teacher has observed
- teaching_note: a teaching strategy or approach
- assessment: a question the teacher uses
- example: a worked example
- correction: fixing existing content

Output as structured YAML matching the appropriate schema.

Preserve the teacher's voice and specific observations where possible.
```

#### 24.2 — Contribution Parser (TDD)

**File:** `internal/parser/contribution_test.go`

```go
package parser_test

import (
	"context"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestParseContribution(t *testing.T) {
	mockResponse := `type: misconception
topic_id: F1-01
content:
  misconception: "Students always try to add 3x + 2y as 5xy"
  remediation: "Use algebra tiles to show different shapes cannot combine"
`

	mock := ai.NewMockProvider(mockResponse)
	input := "My students always confuse the negative sign when expanding brackets like 3(x-2)"

	result, err := parser.ParseContribution(context.Background(), mock, input, "F1-01")
	if err != nil {
		t.Fatalf("ParseContribution() error = %v", err)
	}
	if !strings.Contains(result, "misconception") && !strings.Contains(result, "type") {
		t.Error("Result should contain structured contribution data")
	}
}
```

**File:** `internal/parser/contribution.go`

```go
package parser

import (
	"context"
	"fmt"

	"github.com/p-n-ai/oss-bot/internal/ai"
)

// ParseContribution takes natural language teacher input and structures it into YAML.
func ParseContribution(ctx context.Context, provider ai.Provider, input, topicID string) (string, error) {
	prompt := fmt.Sprintf(`Parse this teacher's contribution for topic %s:

"%s"

Identify the contribution type (misconception, teaching_note, assessment, example, correction)
and output structured YAML. Preserve the teacher's specific observations.`, topicID, input)

	resp, err := provider.Complete(ctx, ai.CompletionRequest{
		Messages: []ai.Message{
			{Role: "system", Content: "You are helping structure a teacher's curriculum contribution into valid YAML."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1024,
		Temperature: 0.3,
	})
	if err != nil {
		return "", fmt.Errorf("parsing contribution: %w", err)
	}

	return resp.Content, nil
}
```

#### 24.3 — Feedback API Endpoint (TDD)

**File:** `internal/api/feedback_test.go`

```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/api"
)

func TestFeedbackHandler_ValidRequest(t *testing.T) {
	handler := api.NewFeedbackHandler(nil) // nil pipeline for now

	body := `{
		"type": "misconception_observed",
		"topic_path": "topics/algebra/05-quadratic-equations",
		"data": {
			"misconception": "Students confuse sign errors",
			"frequency": 0.73,
			"sample_size": 142
		}
	}`

	req := httptest.NewRequest("POST", "/api/feedback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusAccepted)
	}
}

func TestFeedbackHandler_InvalidJSON(t *testing.T) {
	handler := api.NewFeedbackHandler(nil)

	req := httptest.NewRequest("POST", "/api/feedback", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
```

**File:** `internal/api/feedback.go`

```go
// Package api provides HTTP handlers for the web portal backend.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// FeedbackRequest represents incoming feedback from P&AI Bot.
type FeedbackRequest struct {
	Type      string                 `json:"type"`
	TopicPath string                 `json:"topic_path"`
	Data      map[string]interface{} `json:"data"`
}

// FeedbackHandler handles POST /api/feedback.
type FeedbackHandler struct {
	// Pipeline to process feedback into PRs (injected dependency)
	pipeline interface{}
}

// NewFeedbackHandler creates a new feedback handler.
func NewFeedbackHandler(pipeline interface{}) *FeedbackHandler {
	return &FeedbackHandler{pipeline: pipeline}
}

func (h *FeedbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Type == "" || req.TopicPath == "" {
		http.Error(w, "type and topic_path are required", http.StatusBadRequest)
		return
	}

	slog.Info("received feedback",
		"type", req.Type,
		"topic", req.TopicPath,
	)

	// Process asynchronously — create PR with provenance:ai-observed label
	// (Full implementation connects to generation pipeline)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
		"message": "Feedback queued for processing",
	})
}
```

#### 24.5 — Large Document Chunker (TDD)

**File:** `internal/parser/chunker.go`

The chunker splits large documents into semantically meaningful chunks for parallel bulk import. It splits by chapter/heading boundaries rather than arbitrary byte offsets, preserving the logical structure of the source document.

```go
// ChunkOptions controls how documents are split.
type ChunkOptions struct {
    MaxChunkSize  int      // Max tokens per chunk (default: 8000)
    OverlapTokens int      // Overlap between chunks for context continuity (default: 200)
    SplitOn       []string // Heading patterns to split on (e.g., "# ", "## ", "Chapter ")
}

// Chunk represents a portion of a large document.
type Chunk struct {
    Index   int    // 0-based chunk index
    Total   int    // Total number of chunks
    Content string // The chunk text
    Heading string // The heading that starts this chunk (if any)
}

// ChunkDocument splits a large document into chunks at heading boundaries.
func ChunkDocument(text string, opts ChunkOptions) []Chunk { ... }
```

#### 24.6 — Bulk Import Parallel Worker Pool (TDD)

**File:** `internal/pipeline/bulk.go`

The bulk import pipeline processes multiple topics from a large document in parallel using a worker pool. Default concurrency is 3 agents (configurable via `OSS_BULK_WORKERS`).

```go
// BulkRequest configures a bulk import operation.
type BulkRequest struct {
    Chunks      []parser.Chunk  // Document chunks to process
    SyllabusID  string          // Target syllabus
    Mode        ExecutionMode   // ModePreview, ModeWriteFS, or ModeCreatePR
    Source      string          // "cli", "bot", "web"
    Workers     int             // Concurrent workers (default: 3)
    Reporter    ProgressReporter // Progress callback
}

// BulkResult holds the combined results of a bulk import.
type BulkResult struct {
    Topics    []TopicResult   // Per-topic results
    Errors    []error         // Any errors encountered
    Duration  time.Duration   // Total processing time
}

// ExecuteBulk processes multiple chunks in parallel.
func ExecuteBulk(ctx context.Context, req BulkRequest) (*BulkResult, error) { ... }
```

#### 24.7 — Progress Reporter Interface (TDD)

**File:** `internal/pipeline/progress.go`

The progress reporter provides real-time feedback across all three interfaces:

| Interface | Implementation | Behavior |
|-----------|---------------|----------|
| CLI | Terminal progress bar | Shows `[3/12] Processing topic: Algebra...` |
| Bot | Edit GitHub comment | Updates the bot's reply comment in-place with progress |
| Web Portal | Server-Sent Events (SSE) | Streams progress updates to the browser |

```go
// ProgressReporter receives progress updates during bulk operations.
type ProgressReporter interface {
    // OnStart is called when processing begins.
    OnStart(totalItems int)
    // OnProgress is called when an item completes.
    OnProgress(current, total int, itemName string, status string)
    // OnComplete is called when all processing finishes.
    OnComplete(result *BulkResult)
    // OnError is called when an item fails.
    OnError(itemName string, err error)
}
```

#### 24.8 — Reasoning Model Provider (TDD)

**File:** `internal/ai/reasoning.go`

For complex bulk import tasks (e.g., extracting curriculum structure from ambiguous source documents), a reasoning model provides better results. Instead of implementing separate providers for each reasoning model, we use **OpenRouter** as a unified API gateway. OpenRouter provides a single OpenAI-compatible endpoint (`https://openrouter.ai/api/v1`) that routes to 100+ models, so `reasoning.go` only needs one provider implementation.

Supported reasoning models (all via OpenRouter):
- **DeepSeek R1** (`deepseek/deepseek-r1`) — DeepSeek reasoning model (default)
- **Kimi K2.5** (`moonshotai/kimi-k2.5`) — Moonshot AI reasoning model
- **Qwen 3.5** (`qwen/qwen3.5`) — Alibaba reasoning model
- **OpenAI o3-mini** (`openai/o3-mini`) — OpenAI reasoning model

```go
// ReasoningProvider uses OpenRouter as a unified API gateway for reasoning models.
// OpenRouter provides an OpenAI-compatible API, so this reuses the OpenAI provider
// with a custom base URL (https://openrouter.ai/api/v1) and model name.
// Used for complex extraction tasks where step-by-step reasoning improves accuracy.
type ReasoningProvider struct {
    provider Provider  // OpenAI-compatible provider pointed at OpenRouter
    model    string    // e.g., "deepseek/deepseek-r1", "openai/o3-mini"
}

// NewReasoningProvider creates a reasoning-capable provider via OpenRouter.
// Falls back to the standard provider if no reasoning model is configured.
func NewReasoningProvider(base Provider, model string) *ReasoningProvider { ... }
```

Configured via `OSS_AI_REASONING_PROVIDER=openrouter`, `OSS_AI_REASONING_API_KEY` (OpenRouter API key), and `OSS_AI_REASONING_MODEL=deepseek/deepseek-r1` environment variables. When not configured, the pipeline falls back to the standard AI provider.

#### 24.9 — Extended Bloom Verb Sets (TDD)

Update `internal/validator/bloom.go` to include Bloom's taxonomy verbs beyond mathematics, supporting cross-subject curricula:

| Domain | Example Verbs Added |
|--------|-------------------|
| Science | hypothesize, experiment, observe, classify, measure, predict |
| Humanities | interpret, argue, critique, compare, contextualize, empathize |
| General | design, create, evaluate, justify, synthesize, reflect |

The existing Bloom validator must continue to work for mathematics while accepting these extended verbs when validating topics from other subjects. Subject detection is based on the topic's `syllabus_id`.

#### Day 24 Validation

```bash
go test ./...
```

#### Day 24 Exit Criteria

- [ ] `prompts/contribution_parser.md` created (uses `{{syllabus_id}}` template variable, curriculum-agnostic)
- [ ] `internal/parser/contribution.go` + tests — natural language → structured YAML
- [ ] `internal/api/feedback.go` + tests — POST /api/feedback endpoint
- [ ] Feedback handler accepts valid requests and returns 202
- [ ] `internal/parser/chunker.go` + tests — splits large documents at heading boundaries
- [ ] `internal/pipeline/bulk.go` + tests — parallel worker pool (default 3 agents) for multi-topic extraction
- [ ] `internal/pipeline/progress.go` + tests — ProgressReporter interface with CLI (terminal bar), Bot (edit comment), Web (SSE) implementations
- [ ] `internal/ai/reasoning.go` + tests — reasoning model provider via OpenRouter (DeepSeek R1, Kimi K2.5, Qwen 3.5, o3-mini) with fallback
- [ ] `internal/validator/bloom.go` updated with extended verb sets for science, humanities, and general subjects
- [ ] `go test ./...` passes

**Progress:** CLI + Bot + API endpoints + bulk import infrastructure | 7 packages | 7 prompt templates

---

### Day 25 — Docker + End-to-End Testing

**Entry criteria:** Day 24 complete. All packages have tests. Bot and API endpoints work.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 25.1 | `B-W5D25-1` | Dockerfile (multi-stage build) | 🤖 | `deploy/docker/Dockerfile` |
| 25.2 | `B-W5D25-2` | docker-compose.yml (bot + Tika sidecar + optional Ollama) | 🤖 | `docker-compose.yml` |
| 25.3 | `B-W5D25-3` | Webhook test script | 🤖 | `scripts/test-webhook.sh` |
| 25.4 | `B-W5D25-4` | End-to-end test: issue comment → PR | 🤖🧑 | Integration test |
| 25.5 | `B-W5D25-5` | 🧑 Education Lead reviews AI-generated PRs | 🧑 | Decision only |
| 25.6 | `B-W5D25-6` | End-to-end test: bulk import (scaffold + multi-topic extraction) | 🤖 | Integration test |

#### 25.1 — Dockerfile

**File:** `deploy/docker/Dockerfile`

```dockerfile
# Stage 1: Build Go binaries
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /oss-bot ./cmd/bot
RUN CGO_ENABLED=0 go build -o /oss-cli ./cmd/oss

# Stage 2: Final image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /oss-bot /usr/local/bin/oss-bot
COPY --from=go-builder /oss-cli /usr/local/bin/oss
COPY prompts/ /prompts/
ENV OSS_PROMPTS_DIR=/prompts
EXPOSE 8090
ENTRYPOINT ["oss-bot"]
```

#### 25.2 — docker-compose.yml

**File:** `docker-compose.yml`

```yaml
services:
  bot:
    build:
      context: .
      dockerfile: deploy/docker/Dockerfile
    ports:
      - "8090:8090"
    env_file:
      - .env
    environment:
      - OSS_TIKA_URL=http://tika:9998
    depends_on:
      tika:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8090/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  # Apache Tika sidecar for multi-format document extraction (PDF, DOCX, PPTX, etc.)
  tika:
    image: apache/tika:latest
    ports:
      - "9998:9998"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9998/tika"]
      interval: 30s
      timeout: 5s
      retries: 3

  # Optional: local Ollama for free AI generation
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    profiles:
      - ollama

volumes:
  ollama-data:
```

#### 25.3 — Webhook Test Script

**File:** `scripts/test-webhook.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

# Test the webhook handler with a simulated GitHub issue_comment event
BOT_URL="${1:-http://localhost:8090}"
SECRET="${OSS_GITHUB_WEBHOOK_SECRET:-test-secret}"

BODY='{"action":"created","comment":{"body":"@oss-bot quality F1-01","user":{"login":"testuser"}},"issue":{"number":1},"repository":{"full_name":"p-n-ai/oss"}}'

# Compute HMAC
SIGNATURE=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print $2}')

echo "Sending test webhook to $BOT_URL/webhook"
curl -X POST "$BOT_URL/webhook" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: issue_comment" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -d "$BODY" \
  -w "\nHTTP Status: %{http_code}\n"
```

```bash
chmod +x scripts/test-webhook.sh
```

#### 25.4 — End-to-End Test (🤖🧑)

```bash
# Start bot server locally
OSS_GITHUB_WEBHOOK_SECRET=test-secret go run ./cmd/bot &

# Send test webhook
./scripts/test-webhook.sh

# Verify logs show command processed
# Stop server
kill %1
```

#### 25.5 — Education Lead Reviews (🧑)

Review 3 AI-generated PRs (or mock outputs):
- [ ] Would you approve the teaching notes?
- [ ] Would you approve the assessment questions?
- [ ] What needs improvement in the prompt templates?

#### 25.6 — Bulk Import End-to-End Test

Test the full bulk import pipeline:

```bash
# Test scaffold + bulk import from a local file
oss scaffold syllabus --from-file test-curriculum.pdf --id test-syllabus --dry-run

# Verify: directory structure created, topics extracted, progress reported
# Verify: chunker splits document correctly
# Verify: worker pool processes chunks in parallel (check logs for concurrent execution)
# Verify: merge strategies work when importing into existing content
```

#### Day 25 Validation

```bash
# Run all tests
go test ./...

# Build Docker image
docker build -f deploy/docker/Dockerfile -t oss-bot .

# Test Docker container
docker run --rm -e OSS_GITHUB_WEBHOOK_SECRET=test oss-bot &
sleep 2
curl http://localhost:8090/health
docker stop $(docker ps -q --filter ancestor=oss-bot)
```

#### Day 25 Exit Criteria

- [ ] `deploy/docker/Dockerfile` — multi-stage build produces working image
- [ ] `docker-compose.yml` — bot + Tika sidecar + optional Ollama
- [ ] `scripts/test-webhook.sh` — sends valid test webhook
- [ ] End-to-end test: webhook → parse → handler called
- [ ] End-to-end test: bulk import (scaffold → chunk → parallel extract → merge → output)
- [ ] Docker image builds and starts successfully
- [ ] Tika sidecar is healthy and accessible from bot container
- [ ] 🧑 Education Lead has reviewed AI-generated content quality
- [ ] `go test ./...` passes

**Week 5 Output:** Working GitHub bot that receives webhooks and routes commands. CLI with all commands (including scaffold). Bulk import pipeline with chunking, parallel workers, and progress reporting. Content merge strategies. Feedback API. Docker deployment with Tika sidecar. Reasoning model support. All tests green.

**Progress:** CLI + Bot + API + Docker + Bulk Import | 7 packages | 7 prompt templates | Docker image

---

## WEEK 6 — WEB PORTAL + LAUNCH

### Day 26 — Web Portal Scaffold

**Entry criteria:** Week 5 complete. Bot server works with webhooks. Docker image builds. All tests green.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 26.1 | `B-W6D26-1` | Scaffold Next.js web portal | 🤖 | `web/` directory |
| 26.2 | `B-W6D26-2` | Contribution form component | 🤖 | `web/src/app/contribute/page.tsx` |
| 26.3 | `B-W6D26-3` | `POST /api/preview` endpoint | 🤖 | `internal/api/preview.go` |

#### 26.1 — Scaffold Next.js Web Portal

```bash
cd web
npx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir --no-import-alias
npm install @tanstack/react-query react-hook-form zod @hookform/resolvers
npx shadcn@latest init
npx shadcn@latest add button card input label select textarea tabs badge
```

**File:** `web/src/lib/api.ts`

```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090';

export interface PreviewRequest {
  syllabusId: string;
  topicId: string;
  contributionType: 'teaching_note' | 'assessment' | 'example' | 'misconception' | 'correction' | 'translation';
  inputMethod: 'text' | 'url' | 'file';
  content: string;       // For text input: the pasted/typed content
  url?: string;          // For URL input: the page to fetch
  useVision?: boolean;   // For image uploads: use AI Vision (handwriting, diagrams)
  language?: string;
  // File uploads use multipart/form-data via a separate uploadFile() call
}

export interface PreviewResponse {
  structured: string;       // Structured YAML/Markdown output
  validationErrors: string[];
  qualityLevel: number;
}

export interface SubmitResponse {
  prUrl: string;
  prNumber: number;
}

export async function preview(req: PreviewRequest): Promise<PreviewResponse> {
  const res = await fetch(`${API_BASE}/api/preview`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error(`Preview failed: ${res.statusText}`);
  return res.json();
}

export async function submit(req: PreviewRequest): Promise<SubmitResponse> {
  const res = await fetch(`${API_BASE}/api/submit`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error(`Submit failed: ${res.statusText}`);
  return res.json();
}

export interface Curriculum {
  id: string;
  name: string;
  subjects: { id: string; name: string; topics: { id: string; name: string; qualityLevel: number }[] }[];
}

export async function uploadFile(
  file: File,
  syllabusId: string,
  topicId: string,
  contributionType: PreviewRequest['contributionType'],
  useVision = false,
): Promise<PreviewResponse> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('syllabusId', syllabusId);
  formData.append('topicId', topicId);
  formData.append('contributionType', contributionType);
  if (useVision) formData.append('useVision', 'true');

  const res = await fetch(`${API_BASE}/api/preview`, {
    method: 'POST',
    body: formData,
  });
  if (!res.ok) throw new Error(`Upload failed: ${res.statusText}`);
  return res.json();
}

export async function getCurricula(): Promise<Curriculum[]> {
  const res = await fetch(`${API_BASE}/api/curricula`);
  if (!res.ok) throw new Error(`Failed to fetch curricula: ${res.statusText}`);
  return res.json();
}
```

#### 26.2 — Contribution Form

**File:** `web/src/app/contribute/page.tsx`

```tsx
'use client';

import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { getCurricula, preview, submit, uploadFile, PreviewRequest } from '@/lib/api';

export default function ContributePage() {
  const [step, setStep] = useState<'select' | 'write' | 'preview' | 'done'>('select');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [useVision, setUseVision] = useState(false);

  const isImageFile = (name: string) =>
    /\.(png|jpg|jpeg)$/i.test(name);

  const [form, setForm] = useState<PreviewRequest>({
    syllabusId: '',
    topicId: '',
    contributionType: 'teaching_note',
    inputMethod: 'text',
    content: '',
  });

  const curricula = useQuery({ queryKey: ['curricula'], queryFn: getCurricula });

  const previewMutation = useMutation({
    mutationFn: preview,
    onSuccess: () => setStep('preview'),
  });

  const uploadFileMutation = useMutation({
    mutationFn: (file: File) => uploadFile(file, form.syllabusId, form.topicId, form.contributionType, useVision),
    onSuccess: () => setStep('preview'),
  });

  const submitMutation = useMutation({
    mutationFn: submit,
    onSuccess: () => setStep('done'),
  });

  return (
    <main className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Contribute to Open School Syllabus</h1>

      {step === 'select' && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">1. Select Topic</h2>
          {/* Syllabus and topic selection dropdowns */}
          {/* Contribution type selector */}
          <button
            onClick={() => setStep('write')}
            disabled={!form.syllabusId || !form.topicId}
            className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}

      {step === 'write' && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">2. Provide Your Content</h2>

          {/* Input method tabs */}
          <div className="flex gap-2 border-b">
            {(['text', 'url', 'file'] as const).map((method) => (
              <button
                key={method}
                onClick={() => setForm({ ...form, inputMethod: method })}
                className={`px-4 py-2 ${form.inputMethod === method ? 'border-b-2 border-blue-600 font-semibold' : 'text-gray-500'}`}
              >
                {method === 'text' ? 'Type / Paste' : method === 'url' ? 'Paste URL' : 'Upload File'}
              </button>
            ))}
          </div>

          {/* Text input */}
          {form.inputMethod === 'text' && (
            <>
              <p className="text-gray-600">
                Write in any language. Our AI will structure it into the correct format.
              </p>
              <textarea
                className="w-full h-48 p-3 border rounded"
                placeholder="Share your teaching experience, a common misconception, an example problem..."
                value={form.content}
                onChange={(e) => setForm({ ...form, content: e.target.value })}
              />
            </>
          )}

          {/* URL input */}
          {form.inputMethod === 'url' && (
            <>
              <p className="text-gray-600">
                Paste a link to a curriculum page, syllabus specification, or educational resource.
              </p>
              <input
                type="url"
                className="w-full p-3 border rounded"
                placeholder="https://example.org/curriculum-specification"
                value={form.url || ''}
                onChange={(e) => setForm({ ...form, url: e.target.value })}
              />
            </>
          )}

          {/* File upload input */}
          {form.inputMethod === 'file' && (
            <>
              <p className="text-gray-600">
                Upload a PDF, Word document, PowerPoint, text file, or image (PNG, JPG).
              </p>
              <input
                type="file"
                accept=".pdf,.docx,.pptx,.txt,.png,.jpg,.jpeg"
                className="w-full p-3 border rounded"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) setSelectedFile(file);
                }}
              />
              {selectedFile && (
                <p className="text-sm text-gray-500">Selected: {selectedFile.name}</p>
              )}
              {selectedFile && isImageFile(selectedFile.name) && (
                <label className="flex items-center gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={useVision}
                    onChange={(e) => setUseVision(e.target.checked)}
                  />
                  <span>Use AI Vision</span>
                  <span className="text-gray-400">(for handwritten notes, diagrams, whiteboard photos)</span>
                </label>
              )}
            </>
          )}

          <button
            onClick={() => {
              if (form.inputMethod === 'file' && selectedFile) {
                uploadFileMutation.mutate(selectedFile);
              } else {
                previewMutation.mutate(form);
              }
            }}
            disabled={
              (form.inputMethod === 'text' && !form.content) ||
              (form.inputMethod === 'url' && !form.url) ||
              (form.inputMethod === 'file' && !selectedFile) ||
              previewMutation.isPending || uploadFileMutation.isPending
            }
            className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
          >
            {(previewMutation.isPending || uploadFileMutation.isPending) ? 'Processing...' : 'Preview'}
          </button>
        </div>
      )}

      {step === 'preview' && previewMutation.data && (
        <div className="space-y-4">
          <h2 className="text-lg font-semibold">3. Review & Submit</h2>
          <pre className="bg-gray-50 p-4 rounded text-sm overflow-auto">
            {previewMutation.data.structured}
          </pre>
          {previewMutation.data.validationErrors.length > 0 && (
            <div className="text-red-600">
              {previewMutation.data.validationErrors.map((e, i) => (
                <p key={i}>⚠ {e}</p>
              ))}
            </div>
          )}
          <div className="flex gap-2">
            <button onClick={() => setStep('write')} className="px-4 py-2 border rounded">
              Edit
            </button>
            <button
              onClick={() => submitMutation.mutate(form)}
              disabled={submitMutation.isPending || previewMutation.data.validationErrors.length > 0}
              className="px-4 py-2 bg-green-600 text-white rounded disabled:opacity-50"
            >
              {submitMutation.isPending ? 'Submitting...' : 'Submit as PR'}
            </button>
          </div>
        </div>
      )}

      {step === 'done' && submitMutation.data && (
        <div className="space-y-4 text-center">
          <h2 className="text-lg font-semibold text-green-600">Contribution Submitted!</h2>
          <p>Your contribution has been submitted as a pull request.</p>
          <a
            href={submitMutation.data.prUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 underline"
          >
            View PR #{submitMutation.data.prNumber}
          </a>
        </div>
      )}
    </main>
  );
}
```

#### 26.3 — Preview API Endpoint (TDD)

**File:** `internal/api/preview_test.go`

```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/api"
)

func TestPreviewHandler_Valid(t *testing.T) {
	handler := api.NewPreviewHandler(nil, nil) // nil deps for now

	body := `{
		"syllabusId": "malaysia-kssm-matematik-tingkatan1",
		"topicId": "F1-01",
		"contributionType": "misconception",
		"content": "Students confuse 3x + 2y = 5xy"
	}`

	req := httptest.NewRequest("POST", "/api/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}
```

**File:** `internal/api/preview.go`

```go
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

// PreviewRequest is the web portal's preview request.
// Supports three input methods: text (content field), URL (url field), or file (multipart upload).
type PreviewRequest struct {
	SyllabusID       string `json:"syllabusId"`
	TopicID          string `json:"topicId"`
	ContributionType string `json:"contributionType"`
	InputMethod      string `json:"inputMethod"`          // "text", "url", or "file"
	Content          string `json:"content,omitempty"`     // For text input
	URL              string `json:"url,omitempty"`         // For URL input
	UseVision        bool   `json:"useVision,omitempty"`   // For image uploads: use AI Vision instead of OCR
	Language         string `json:"language,omitempty"`
	// File uploads are handled via multipart/form-data (not JSON)
}

// PreviewResponse returns the structured output and validation status.
type PreviewResponse struct {
	Structured       string   `json:"structured"`
	ValidationErrors []string `json:"validationErrors"`
	QualityLevel     int      `json:"qualityLevel"`
}

// PreviewHandler handles POST /api/preview.
// Uses the shared pipeline in ModePreview — no PR created, just structured output returned.
type PreviewHandler struct {
	pipe *pipeline.Pipeline
}

// NewPreviewHandler creates a new preview handler backed by the shared pipeline.
func NewPreviewHandler(pipe *pipeline.Pipeline) *PreviewHandler {
	return &PreviewHandler{pipe: pipe}
}

func (h *PreviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TopicID == "" {
		http.Error(w, "topicId is required", http.StatusBadRequest)
		return
	}

	// Validate that at least one input method has content
	hasInput := req.Content != "" || req.URL != ""
	if !hasInput {
		// Check for multipart file upload
		http.Error(w, "content, url, or file upload is required", http.StatusBadRequest)
		return
	}

	slog.Info("preview requested",
		"topic", req.TopicID,
		"type", req.ContributionType,
		"inputMethod", req.InputMethod,
	)

	// Extract content based on input method
	content, err := h.extractContent(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the shared pipeline in Preview mode — same code path as CLI and Bot
	result, err := h.pipe.Execute(r.Context(), pipeline.Request{
		TopicPath:        req.TopicID,
		ContributionType: req.ContributionType,
		Content:          content,
		Mode:             pipeline.ModePreview,
		Source:           "web",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := PreviewResponse{
		Structured:       result.StructuredOutput,
		ValidationErrors: result.ValidationErrors,
		QualityLevel:     result.QualityLevel,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
```

#### Day 26 Validation

```bash
# Go tests
go test ./...

# Web portal
cd web && npm install && npm run build
```

#### Day 26 Exit Criteria

- [ ] `web/` scaffolded with Next.js 15 + TypeScript + Tailwind + shadcn/ui
- [ ] Contribution form: select topic → write content → preview → submit
- [ ] `web/src/lib/api.ts` — API client for Go backend
- [ ] `internal/api/preview.go` + tests — POST /api/preview endpoint
- [ ] Web portal builds without errors
- [ ] `go test ./...` passes

**Progress:** CLI + Bot + API + Web Portal | 6 packages | Web portal scaffold

---

### Day 27 — Submit + Preview Flow

**Entry criteria:** Day 26 complete. Web portal scaffold exists. Preview endpoint works.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 27.1 | `B-W6D27-1` | Preview component with syntax highlighting | 🤖 | `web/src/components/yaml-preview.tsx` |
| 27.2 | `B-W6D27-2` | `POST /api/submit` endpoint | 🤖 | `internal/api/submit.go` |
| 27.3 | `B-W6D27-3` | Real-time schema validation in preview | 🤖 | Update preview handler |

#### 27.2 — Submit Endpoint (TDD)

**File:** `internal/api/submit_test.go`

```go
package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/api"
)

func TestSubmitHandler_Valid(t *testing.T) {
	handler := api.NewSubmitHandler(nil) // nil GitHub client for now

	body := `{
		"syllabusId": "malaysia-kssm-matematik-tingkatan1",
		"topicId": "F1-01",
		"contributionType": "misconception",
		"content": "Students confuse terms"
	}`

	req := httptest.NewRequest("POST", "/api/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should return 201 Created or 202 Accepted
	if rr.Code != http.StatusAccepted && rr.Code != http.StatusCreated {
		t.Errorf("Status = %d, want 201 or 202", rr.Code)
	}
}
```

**File:** `internal/api/submit.go`

```go
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// SubmitResponse holds the result of creating a PR from a contribution.
type SubmitResponse struct {
	PRUrl    string `json:"prUrl"`
	PRNumber int    `json:"prNumber"`
	Status   string `json:"status"`
}

// SubmitHandler handles POST /api/submit.
// Uses the shared pipeline in ModeCreatePR — same code path as GitHub Bot.
type SubmitHandler struct {
	pipe *pipeline.Pipeline
}

// NewSubmitHandler creates a new submit handler backed by the shared pipeline.
func NewSubmitHandler(pipe *pipeline.Pipeline) *SubmitHandler {
	return &SubmitHandler{pipe: pipe}
}

func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.TopicID == "" {
		http.Error(w, "topicId is required", http.StatusBadRequest)
		return
	}

	slog.Info("contribution submitted",
		"topic", req.TopicID,
		"type", req.ContributionType,
	)

	// Extract content based on input method (same as preview)
	content, err := extractContent(r, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the shared pipeline in CreatePR mode — same pipeline as Bot and CLI --pr
	result, err := h.pipe.Execute(r.Context(), pipeline.Request{
		TopicPath:        req.TopicID,
		ContributionType: req.ContributionType,
		Content:          content,
		Mode:             pipeline.ModeCreatePR,
		Source:           "web",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(result.ValidationErrors) > 0 {
		http.Error(w, "validation failed: "+result.ValidationErrors[0], http.StatusUnprocessableEntity)
		return
	}

	resp := SubmitResponse{
		PRUrl:    result.PRUrl,
		PRNumber: result.PRNumber,
		Status:   "submitted",
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
```

#### Day 27 Validation

```bash
go test ./...
cd web && npm run build
```

#### Day 27 Exit Criteria

- [ ] YAML preview component renders structured output
- [ ] `internal/api/submit.go` + tests — POST /api/submit creates PR
- [ ] Schema validation integrated into preview flow
- [ ] `go test ./...` and `npm run build` both pass

**Progress:** CLI + Bot + API (preview + submit + feedback) + Web Portal

---

### Day 28 — Curricula Browser

**Entry criteria:** Day 27 complete. Submit flow works end-to-end.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 28.1 | `B-W6D28-1` | `GET /api/curricula` endpoint | 🤖 | `internal/api/curricula.go` |
| 28.2 | `B-W6D28-2` | Browse page: tree view with quality badges | 🤖 | `web/src/app/page.tsx` |
| 28.3 | `B-W6D28-3` | Topic detail page with "Improve this" buttons | 🤖 | `web/src/app/topic/[id]/page.tsx` |

#### 28.1 — Curricula Endpoint (TDD)

**File:** `internal/api/curricula_test.go`

```go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/api"
)

func TestCurriculaHandler(t *testing.T) {
	handler := api.NewCurriculaHandler("")

	req := httptest.NewRequest("GET", "/api/curricula", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rr.Code, http.StatusOK)
	}

	var result []interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Response should be valid JSON array: %v", err)
	}
}
```

**File:** `internal/api/curricula.go`

```go
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// CurriculaHandler handles GET /api/curricula.
type CurriculaHandler struct {
	repoPath string // Path to local OSS clone or GitHub API
}

// NewCurriculaHandler creates a new curricula handler.
func NewCurriculaHandler(repoPath string) *CurriculaHandler {
	return &CurriculaHandler{repoPath: repoPath}
}

func (h *CurriculaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("curricula list requested")

	// In full implementation: walk the OSS repo or query GitHub API
	// to build the curricula tree (syllabi → subjects → topics)

	// Return empty array for now (populated when connected to OSS repo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}
```

#### 28.2 — API Router

Wire all API endpoints together:

**File:** `internal/api/router.go`

```go
package api

import (
	"net/http"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/validator"
)

// Router creates the HTTP mux for all API endpoints.
func Router(provider ai.Provider, v *validator.Validator, repoPath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /api/preview", NewPreviewHandler(provider, v))
	mux.Handle("POST /api/submit", NewSubmitHandler(nil))
	mux.Handle("POST /api/feedback", NewFeedbackHandler(nil))
	mux.Handle("GET /api/curricula", NewCurriculaHandler(repoPath))

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	return mux
}
```

#### Day 28 Validation

```bash
go test ./...
cd web && npm run build
```

#### Day 28 Exit Criteria

- [ ] `internal/api/curricula.go` + tests — GET /api/curricula returns JSON
- [ ] `internal/api/router.go` — all API routes wired in one place
- [ ] Browse page shows curricula tree with quality badges
- [ ] Topic detail page shows content with "Improve this" links
- [ ] `go test ./...` and `npm run build` pass

**Progress:** Full API surface (preview, submit, feedback, curricula) + Web Portal pages

---

### Day 29 — Deploy + Documentation

**Entry criteria:** Day 28 complete. All API endpoints work. Web portal builds.

#### Tasks

| # | Task ID | Task | Owner | Files Created |
|---|---------|------|-------|---------------|
| 29.1 | `B-W6D29-1` | Update Dockerfile for web portal | 🤖 | Update `deploy/docker/Dockerfile` |
| 29.2 | `B-W6D29-2` | Update docker-compose for full stack | 🤖 | Update `docker-compose.yml` |
| 29.3 | `B-W6D29-3` | 🧑 Test web portal with teachers | 🧑 | User testing |

#### 29.1 — Full Dockerfile (Go + Web)

Update `deploy/docker/Dockerfile` to include the web portal build:

```dockerfile
# Stage 1: Build Go binaries
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /oss-bot ./cmd/bot
RUN CGO_ENABLED=0 go build -o /oss-cli ./cmd/oss

# Stage 2: Build web portal
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# Stage 3: Final image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates nodejs npm
COPY --from=go-builder /oss-bot /usr/local/bin/oss-bot
COPY --from=go-builder /oss-cli /usr/local/bin/oss
COPY --from=web-builder /web/.next /web/.next
COPY --from=web-builder /web/public /web/public
COPY --from=web-builder /web/node_modules /web/node_modules
COPY --from=web-builder /web/package.json /web/package.json
COPY prompts/ /prompts/
ENV OSS_PROMPTS_DIR=/prompts
EXPOSE 8090 3001
COPY deploy/docker/start.sh /start.sh
RUN chmod +x /start.sh
ENTRYPOINT ["/start.sh"]
```

**File:** `deploy/docker/start.sh`

```bash
#!/bin/sh
# Start both Go bot server and Next.js web portal
cd /web && npx next start -p 3001 &
oss-bot
```

#### 29.2 — Full docker-compose

Update `docker-compose.yml`:

```yaml
services:
  bot:
    build:
      context: .
      dockerfile: deploy/docker/Dockerfile
    ports:
      - "8090:8090"
      - "3001:3001"
    env_file:
      - .env
    environment:
      - OSS_TIKA_URL=http://tika:9998
    depends_on:
      tika:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8090/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  tika:
    image: apache/tika:latest
    ports:
      - "9998:9998"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9998/tika"]
      interval: 30s
      timeout: 5s
      retries: 3

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    profiles:
      - ollama

volumes:
  ollama-data:
```

#### 29.3 — User Testing (🧑)

Test with 2 teachers who have no Git experience:

- [ ] Can they navigate to contribute.p-n-ai.org?
- [ ] Can they select a topic to contribute to?
- [ ] Can they write a contribution in natural language?
- [ ] Does the preview show understandable structured output?
- [ ] Can they submit without confusion?
- [ ] Is the confirmation with PR link clear?

#### Day 29 Validation

```bash
# Full test suite
go test ./...

# Build complete Docker image
docker build -f deploy/docker/Dockerfile -t oss-bot .

# Test full stack
docker compose up -d
curl http://localhost:8090/health
curl http://localhost:3001
docker compose down
```

#### Day 29 Exit Criteria

- [ ] Full Dockerfile builds (Go + Web in single image)
- [ ] docker-compose starts both bot and web portal
- [ ] Health check passes
- [ ] 🧑 2 teachers tested web portal successfully

**Progress:** Full stack deployed | CLI + Bot + API + Web Portal | Docker

---

### Day 30 — Launch + Report

**Entry criteria:** Day 29 complete. Full stack deployed and tested.

#### Tasks

| # | Task ID | Task | Owner |
|---|---------|------|-------|
| 30.1 | `B-W6D30-1` | 🧑 Announce web portal in launch materials | 🧑 |
| 30.2 | `B-W6D30-2` | 🧑 Write oss-bot section of 6-week report | 🧑 |

#### 30.1 — Launch (🧑 Human)

- [ ] Deploy to production VPS
- [ ] Configure GitHub App webhook URL to production
- [ ] Verify `@oss-bot` responds to commands in p-n-ai/oss
- [ ] Verify web portal accessible at contribute.p-n-ai.org
- [ ] Announce in launch materials

#### 30.2 — Report (🧑 Human)

Write the oss-bot section of the 6-week report covering:

- AI generation quality assessment
- Number of bot-created PRs during testing
- Web portal usability results from teacher testing
- Performance metrics vs targets
- Known issues and roadmap

#### Day 30 Exit Criteria

- [ ] Bot responding to `@oss-bot` commands in production
- [ ] Web portal live at contribute.p-n-ai.org
- [ ] CLI distributed as pre-built binary
- [ ] Report section written

**Week 6 Output:** Web portal live. GitHub bot responding. CLI distributed. Full stack deployed.

---

## Appendix A — Complete Package Reference

| Package | Created On | Key Files | Purpose |
|---------|-----------|-----------|---------|
| `internal/validator` | Day 16-17 | `validator.go`, `bloom.go`, `prerequisites.go`, `duplicates.go`, `quality.go` | JSON Schema validation, content quality checks |
| `internal/ai` | Day 18 | `provider.go`, `mock.go`, `openai.go`, `anthropic.go`, `ollama.go` | AI provider interface (shared with P&AI Bot) |
| `internal/generator` | Day 18-20 | `context.go`, `teaching_notes.go`, `assessments.go`, `examples.go`, `translator.go`, `scaffolder.go` | Content generation (individual generators) |
| `internal/pipeline` | Day 19 | `pipeline.go` | **Shared orchestrator** — all three interfaces call `pipeline.Execute()` with different modes (Preview, WriteFS, CreatePR) |
| `internal/output` | Day 19 | `writer.go`, `github.go` | Output writers: `LocalWriter` (CLI filesystem), `GitHubWriter` (Bot/Web PR creation) |
| `internal/parser` | Day 23-24 | `document.go`, `pdf.go`, `tika.go`, `url.go`, `image.go`, `contribution.go` | ContentExtractor interface, Go-native PDF (CLI), Tika multi-format (server), URL fetcher, image extraction (OCR + AI Vision), natural language parsing |
| `internal/github` | Day 21-22 | `app.go`, `webhook.go`, `pr.go`, `contents.go` | GitHub App auth, webhooks, PR creation |
| `internal/api` | Day 24-28 | `router.go`, `preview.go`, `submit.go`, `feedback.go`, `curricula.go` | Web portal HTTP API (thin layer — delegates to shared pipeline) |

---

## Appendix B — Prompt Template Reference

| Template | Created On | Purpose |
|----------|-----------|---------|
| `prompts/teaching_notes.md` | Day 18 | Generate `.teaching.md` files |
| `prompts/assessments.md` | Day 18 | Generate `.assessments.yaml` files |
| `prompts/examples.md` | Day 19 | Generate `.examples.yaml` files |
| `prompts/translation.md` | Day 20 | Translate topic files to other languages |
| `prompts/document_import.md` | Day 23 | Extract curriculum structure from any source (URL, PDF, DOCX, PPTX, TXT, images) |
| `prompts/contribution_parser.md` | Day 24 | Parse natural language into structured data |

---

## Appendix C — Environment Variables Quick Reference

| Variable | Week | Required For | Default |
|----------|------|-------------|---------|
| `OSS_REPO_PATH` | 4 | CLI | `./oss` |
| `OSS_AI_PROVIDER` | 4 | CLI, Bot | — |
| `OSS_AI_API_KEY` | 4 | CLI, Bot (not Ollama) | — |
| `OSS_AI_MODEL` | 4 | CLI, Bot | Provider default |
| `OSS_GITHUB_APP_ID` | 5 | Bot | — |
| `OSS_GITHUB_PRIVATE_KEY_PATH` | 5 | Bot | — |
| `OSS_GITHUB_WEBHOOK_SECRET` | 5 | Bot | — |
| `OSS_REPO_OWNER` | 5 | Bot | `p-n-ai` |
| `OSS_REPO_NAME` | 5 | Bot | `oss` |
| `OSS_BOT_PORT` | 5 | Bot | `8090` |
| `OSS_WEB_PORT` | 6 | Web Portal | `3001` |
| `OSS_PROMPTS_DIR` | 4 | All | `./prompts` |
| `OSS_TIKA_URL` | 5 | Bot, Web Portal | `http://tika:9998` |
| `OSS_LOG_LEVEL` | 4 | All | `info` |

---

## Appendix D — Performance Targets

| Operation | Target | Validation Command |
|-----------|--------|--------------------|
| `oss validate` (full repo) | <2s | `time go run ./cmd/oss validate ../oss` |
| Teaching notes generation | <15s | `time go run ./cmd/oss generate teaching-notes F1-01` |
| Assessment generation (5 questions) | <10s | `time go run ./cmd/oss generate assessments F1-01 -c 5` |
| PDF import, CLI (50-page syllabus) | <60s | `time go run ./cmd/oss import --pdf test.pdf` |
| URL import (fetch + extract) | <30s | `time go run ./cmd/oss import --url https://example.org/spec` |
| Image extraction (OCR) | <5s | `time go run ./cmd/oss import --file photo.jpg` |
| Image extraction (AI Vision) | <15s | `time go run ./cmd/oss import --file photo.jpg --vision` |
| Document import, server (50-page, any format) | <90s | Measure via API with DOCX/PPTX/image input |
| Bot webhook → PR created | <30s | Measure from webhook receipt to PR URL comment |
| Web portal preview | <5s | Measure from submit to preview render |
| CLI startup | <100ms | `time go run ./cmd/oss --help` |

---

## Appendix E — Progress Tracking Dashboard

| Day | Packages | Tests | CLI Commands | API Endpoints | Prompt Templates | Docker |
|-----|----------|-------|-------------|---------------|-----------------|--------|
| 16 | 1 (validator) | ✅ | validate | — | — | — |
| 17 | 1 (validator: 5 files) | ✅ | validate, quality | — | — | — |
| 18 | 3 (validator, ai, generator) | ✅ | validate, quality | — | 2 | — |
| 19 | 5 (+pipeline, output) | ✅ | validate, generate (3), quality | — | 3 | — |
| 20 | 3 | ✅ | validate, generate (3), quality, translate | — | 4 | — |
| 21 | 5 (+github, parser) | ✅ | All CLI | — | 4 | — |
| 22 | 5 | ✅ | All CLI | — | 4 | — |
| 23 | 5 | ✅ | All CLI + import (URL, file, text) | — | 5 | — |
| 24 | 6 (+api) | ✅ | All CLI | feedback | 6 | — |
| 25 | 6 | ✅ | All CLI | feedback | 6 | ✅ |
| 26 | 6 | ✅ | All CLI | feedback, preview, curricula | 6 | ✅ |
| 27 | 6 | ✅ | All CLI | feedback, preview, submit, curricula | 6 | ✅ |
| 28 | 6 | ✅ | All CLI | All 4 endpoints | 6 | ✅ |
| 29 | 6 | ✅ | All CLI | All 4 endpoints | 6 | ✅ (full) |
| 30 | 6 | ✅ | All CLI | All 4 endpoints | 6 | ✅ (deployed) |

---

## Appendix F — Task Count Summary

| Week | 🤖 Claude Code | 🧑 Human | Total |
|------|----------------|----------|-------|
| 4 (Days 16-20) | 16 | 2 | 18 |
| 5 (Days 21-25) | 16 | 2 | 18 |
| 6 (Days 26-30) | 10 | 2 | 12 |
| **Total** | **42** | **6** | **48** |
