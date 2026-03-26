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
