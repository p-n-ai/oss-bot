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

	// Create a topics subdirectory with valid YAML (new structure: subject/subject_grade/topics/)
	topicsDir := filepath.Join(dir, "test-subject", "test-subject-1", "topics")
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

func TestValidateFile_ValidAssessments(t *testing.T) {
	schemaDir := setupTestSchemas(t)
	v, err := validator.New(schemaDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tests := []struct {
		name string
		yaml string
	}{
		{"basic assessment", validAssessmentsYAML},
		{"options as array", validAssessmentsOptionsArrayYAML},
		{"options as object", validAssessmentsOptionsObjectYAML},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := createTempAssessmentsYAML(t, tt.yaml)
			result, err := v.ValidateFile(f, "assessments")
			if err != nil {
				t.Fatalf("ValidateFile() error = %v", err)
			}
			if !result.Valid {
				t.Errorf("ValidateFile() expected valid, got errors: %v", result.Errors)
			}
		})
	}
}

func TestValidateFile_InvalidAssessmentsOptions(t *testing.T) {
	schemaDir := setupTestSchemas(t)
	v, err := validator.New(schemaDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	f := createTempAssessmentsYAML(t, invalidAssessmentsOptionsYAML)
	result, err := v.ValidateFile(f, "assessments")
	if err != nil {
		t.Fatalf("ValidateFile() error = %v", err)
	}
	if result.Valid {
		t.Error("ValidateFile() expected invalid for non-array/non-object options, got valid")
	}
	if len(result.Errors) == 0 {
		t.Error("ValidateFile() expected errors, got none")
	}
}

func TestDetectSchemaType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"syllabus", "curricula/malaysia/malaysia-kssm/syllabus.yaml", "syllabus"},
		{"subject", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/subject.yaml", "subject"},
		{"subject_grade", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-3/subject-grade.yaml", "subject_grade"},
		{"topic", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-3/topics/MT3-01.yaml", "topic"},
		{"assessments", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-3/topics/MT3-01.assessments.yaml", "assessments"},
		{"examples", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-3/topics/MT3-01.examples.yaml", "examples"},
		{"teaching_md_not_topic", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-3/topics/MT3-01.teaching.md", ""},
		{"teaching_yaml_not_topic", "curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-3/topics/MT3-01.teaching.yaml", ""},
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
id: MT1-01
name: "Test Topic"
name_en: "Test Topic"
subject_id: test-matematik
subject_grade_id: test-matematik-tingkatan-1
syllabus_id: test-syllabus
country_id: test
language: ms
difficulty: beginner
learning_objectives:
  - id: "1.1.1"
    text: "Objektif ujian"
    text_en: "Test objective"
    bloom: understand
quality_level: 1
provenance: human
`

const invalidTopicYAML = `
id: invalid id with spaces
name: ""
`

const validAssessmentsYAML = `
topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
    learning_objective: "1.1.1"
    tp_level: 2
    kbat: false
    answer:
      type: exact
      value: "4"
      working: "2+2 = 4"
    marks: 1
    rubric:
      - marks: 1
        criteria: "Correct answer"
    hints:
      - level: 1
        text: "Count on your fingers"
`

const validAssessmentsOptionsArrayYAML = `
topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "Which is the largest prime under 10?"
    difficulty: medium
    learning_objective: "1.1.1"
    answer:
      type: multiple_choice
      value: "7"
      options:
        - "3"
        - "5"
        - "7"
        - "9"
`

const validAssessmentsOptionsObjectYAML = `
topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "Which is the largest prime under 10?"
    difficulty: medium
    learning_objective: "1.1.1"
    answer:
      type: multiple_choice
      value: "C"
      options:
        A: "3"
        B: "5"
        C: "7"
        D: "9"
`

const invalidAssessmentsOptionsYAML = `
topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "Which is the largest prime under 10?"
    difficulty: medium
    learning_objective: "1.1.1"
    answer:
      type: multiple_choice
      value: "C"
      options: 42
`

func setupTestSchemas(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Minimal topic schema for testing — mirrors pnai-oss/schema/topic.schema.json
	topicSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["id", "name", "subject_id", "syllabus_id", "difficulty", "learning_objectives", "quality_level", "provenance"],
		"properties": {
			"id": { "type": "string", "pattern": "^[A-Z]{2,3}[0-9]*-[0-9]{2}$" },
			"official_ref": {
				"oneOf": [
					{ "type": "string" },
					{ "type": "array", "items": { "type": "string" } }
				]
			},
			"name": { "type": "string", "minLength": 1 },
			"name_en": { "type": "string", "minLength": 1 },
			"subject_grade_id": { "type": "string" },
			"subject_id": { "type": "string" },
			"syllabus_id": { "type": "string" },
			"country_id": { "type": "string" },
			"language": { "type": "string" },
			"difficulty": { "type": "string", "enum": ["beginner", "intermediate", "advanced"] },
			"learning_objectives": {
				"type": "array", "minItems": 1,
				"items": {
					"type": "object",
					"required": ["id", "text", "bloom"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string", "minLength": 1 },
						"text_en": { "type": "string", "minLength": 1 },
						"bloom": { "type": "string", "enum": ["remember", "understand", "apply", "analyze", "evaluate", "create"] }
					},
					"additionalProperties": false
				}
			},
			"prerequisites": {
				"type": "object",
				"properties": {
					"required": { "type": "array", "items": { "type": "string" } },
					"recommended": { "type": "array", "items": { "type": "string" } }
				},
				"additionalProperties": false
			},
			"teaching": {
				"type": "object",
				"properties": {
					"sequence": { "type": "array", "items": { "type": "string" } },
					"common_misconceptions": {
						"type": "array",
						"items": {
							"type": "object",
							"required": ["misconception", "remediation"],
							"properties": {
								"misconception": { "type": "string" },
								"remediation": { "type": "string" }
							},
							"additionalProperties": false
						}
					},
					"engagement_hooks": { "type": "array", "items": { "type": "string" } }
				},
				"additionalProperties": false
			},
			"bloom_levels": {
				"type": "array",
				"items": { "type": "string", "enum": ["remember", "understand", "apply", "analyze", "evaluate", "create"] }
			},
			"mastery": {
				"type": "object",
				"properties": {
					"minimum_score": { "type": "number", "minimum": 0, "maximum": 1 },
					"assessment_count": { "type": "integer", "minimum": 1 },
					"spaced_repetition": {
						"type": "object",
						"properties": {
							"initial_interval_days": { "type": "integer", "minimum": 1 },
							"multiplier": { "type": "number", "minimum": 1 }
						},
						"additionalProperties": false
					}
				},
				"additionalProperties": false
			},
			"tier": { "type": "string", "enum": ["core", "extended"] },
			"ai_teaching_notes": { "type": "string" },
			"examples_file": { "type": "string" },
			"assessments_file": { "type": "string" },
			"quality_level": { "type": "integer", "minimum": 0, "maximum": 5 },
			"provenance": { "type": "string", "enum": ["human", "ai-assisted", "ai-generated", "ai-observed"] },
			"generated_at": { "type": "string" },
			"engagement_hooks": { "type": "array", "items": { "type": "string" } },
			"translations": { "type": "object" }
		},
		"additionalProperties": false
	}`

	writeFile(t, filepath.Join(dir, "topic.schema.json"), topicSchema)

	// Minimal assessments schema for testing — mirrors pnai-oss/schema/assessments.schema.json
	assessmentsSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["topic_id", "questions"],
		"properties": {
			"topic_id": { "type": "string" },
			"provenance": { "type": "string", "enum": ["human", "ai-assisted", "ai-generated", "ai-observed"] },
			"questions": {
				"type": "array",
				"minItems": 1,
				"items": {
					"type": "object",
					"required": ["id", "text", "difficulty", "learning_objective", "answer"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string", "minLength": 1 },
						"difficulty": { "type": "string", "enum": ["easy", "medium", "hard"] },
						"learning_objective": { "type": "string" },
						"tp_level": { "type": "integer", "minimum": 1 },
						"kbat": { "type": "boolean" },
						"answer": {
							"type": "object",
							"required": ["type", "value"],
							"properties": {
								"type": { "type": "string", "enum": ["exact", "multiple_choice", "free_text"] },
								"value": { "type": "string" },
								"working": { "type": "string" },
								"options": {
									"description": "Answer options for multiple_choice questions",
									"oneOf": [
										{
											"type": "array",
											"items": { "type": "string" }
										},
										{
											"type": "object",
											"additionalProperties": { "type": "string" }
										}
									]
								}
							},
							"additionalProperties": false
						},
						"marks": { "type": "integer", "minimum": 1 },
						"rubric": {
							"type": "array",
							"items": {
								"type": "object",
								"required": ["marks", "criteria"],
								"properties": {
									"marks": { "type": "integer" },
									"criteria": { "type": "string" }
								},
								"additionalProperties": false
							}
						},
						"hints": {
							"type": "array",
							"items": {
								"type": "object",
								"required": ["level", "text"],
								"properties": {
									"level": { "type": "integer" },
									"text": { "type": "string" }
								},
								"additionalProperties": false
							}
						}
					},
					"additionalProperties": false
				}
			}
		},
		"additionalProperties": false
	}`

	writeFile(t, filepath.Join(dir, "assessments.schema.json"), assessmentsSchema)
	return dir
}

func createTempYAML(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "test.yaml")
	writeFile(t, f, content)
	return f
}

func createTempAssessmentsYAML(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "MT1-01.assessments.yaml")
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
