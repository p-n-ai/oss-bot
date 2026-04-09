package validator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/validator"
)

// setupResolvedTestDirs creates a temp directory with:
//
//	root/
//	  schema/                                  <- global schemas (topic + assessments)
//	  curricula/country/syllabus/subject/
//	    subject.yaml
//	    schemas/                                <- subject override (stricter assessments)
//	      assessments.schema.json
//	    grade/
//	      subject-grade.yaml
//	      topics/
//	        MT1-01.yaml
//	        MT1-01.assessments.yaml
func setupResolvedTestDirs(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()

	// Global schemas
	globalDir := filepath.Join(root, "schema")
	os.MkdirAll(globalDir, 0o755)

	// Global topic schema (same as used in existing tests)
	globalTopicSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["id", "name", "subject_id", "syllabus_id", "difficulty", "learning_objectives", "quality_level", "provenance"],
		"properties": {
			"id": { "type": "string", "pattern": "^[A-Z]{2,3}[0-9]*-[0-9]{2}$" },
			"name": { "type": "string", "minLength": 1 },
			"name_en": { "type": "string" },
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
					"required": ["id", "bloom"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string" },
						"text_en": { "type": "string" },
						"bloom": { "type": "string", "enum": ["remember", "understand", "apply", "analyze", "evaluate", "create"] }
					},
					"additionalProperties": false
				}
			},
			"prerequisites": { "type": "object" },
			"teaching": { "type": "object" },
			"bloom_levels": { "type": "array" },
			"mastery": { "type": "object" },
			"tier": { "type": "string" },
			"ai_teaching_notes": { "type": "string" },
			"examples_file": { "type": "string" },
			"assessments_file": { "type": "string" },
			"quality_level": { "type": "integer", "minimum": 0, "maximum": 5 },
			"provenance": { "type": "string", "enum": ["human", "ai-assisted", "ai-generated", "ai-observed"] },
			"generated_at": { "type": "string" },
			"engagement_hooks": { "type": "array" }
		},
		"additionalProperties": false
	}`
	os.WriteFile(filepath.Join(globalDir, "topic.schema.json"), []byte(globalTopicSchema), 0o644)

	// Global assessments schema — permissive (marks NOT required)
	globalAssessmentsSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["topic_id", "questions"],
		"properties": {
			"topic_id": { "type": "string" },
			"provenance": { "type": "string" },
			"questions": {
				"type": "array", "minItems": 1,
				"items": {
					"type": "object",
					"required": ["id", "text", "difficulty", "learning_objective", "answer"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string" },
						"difficulty": { "type": "string", "enum": ["easy", "medium", "hard"] },
						"learning_objective": { "type": "string" },
						"answer": {
							"type": "object",
							"required": ["type", "value"],
							"properties": {
								"type": { "type": "string" },
								"value": { "type": "string" }
							},
							"additionalProperties": false
						},
						"marks": { "type": "integer" }
					},
					"additionalProperties": false
				}
			}
		},
		"additionalProperties": false
	}`
	os.WriteFile(filepath.Join(globalDir, "assessments.schema.json"), []byte(globalAssessmentsSchema), 0o644)

	// Subject directory
	subjectDir := filepath.Join(root, "curricula", "country", "syllabus", "subject")
	os.MkdirAll(subjectDir, 0o755)
	os.WriteFile(filepath.Join(subjectDir, "subject.yaml"), []byte("id: subject\nname: Test\nsyllabus_id: syllabus\ntopics: []\n"), 0o644)

	// Subject-level schema override — stricter assessments (marks IS required)
	subjectSchemas := filepath.Join(subjectDir, "schemas")
	os.MkdirAll(subjectSchemas, 0o755)
	stricterAssessmentsSchema := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type": "object",
		"required": ["topic_id", "questions"],
		"properties": {
			"topic_id": { "type": "string" },
			"provenance": { "type": "string" },
			"questions": {
				"type": "array", "minItems": 1,
				"items": {
					"type": "object",
					"required": ["id", "text", "difficulty", "learning_objective", "answer", "marks"],
					"properties": {
						"id": { "type": "string" },
						"text": { "type": "string" },
						"difficulty": { "type": "string", "enum": ["easy", "medium", "hard"] },
						"learning_objective": { "type": "string" },
						"answer": {
							"type": "object",
							"required": ["type", "value"],
							"properties": {
								"type": { "type": "string" },
								"value": { "type": "string" }
							},
							"additionalProperties": false
						},
						"marks": { "type": "integer", "minimum": 1 }
					},
					"additionalProperties": false
				}
			}
		},
		"additionalProperties": false
	}`
	os.WriteFile(filepath.Join(subjectSchemas, "assessments.schema.json"), []byte(stricterAssessmentsSchema), 0o644)

	// Topics directory with files
	topicsDir := filepath.Join(subjectDir, "grade", "topics")
	os.MkdirAll(topicsDir, 0o755)
	os.WriteFile(filepath.Join(subjectDir, "grade", "subject-grade.yaml"), []byte("id: grade\n"), 0o644)

	topicContent := `id: MT1-01
name: "Test Topic"
subject_id: subject
syllabus_id: syllabus
difficulty: beginner
learning_objectives:
  - id: "1.1.1"
    bloom: understand
quality_level: 1
provenance: human
`
	os.WriteFile(filepath.Join(topicsDir, "MT1-01.yaml"), []byte(topicContent), 0o644)

	// Assessment WITHOUT marks (valid under global, invalid under subject schema)
	assessmentNoMarks := `topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
    learning_objective: "1.1.1"
    answer:
      type: exact
      value: "4"
`
	os.WriteFile(filepath.Join(topicsDir, "MT1-01.assessments.yaml"), []byte(assessmentNoMarks), 0o644)

	return root
}

func TestNewWithResolver(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))

	v := validator.NewWithResolver(resolver)
	if v == nil {
		t.Fatal("NewWithResolver() returned nil")
	}
}

func TestValidateFileResolved_UsesSubjectSchema(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))
	v := validator.NewWithResolver(resolver)

	assessmentFile := filepath.Join(root, "curricula", "country", "syllabus", "subject", "grade", "topics", "MT1-01.assessments.yaml")

	// This file has no "marks" field. Subject schema requires marks, global does not.
	result, err := v.ValidateFileResolved(assessmentFile, "assessments")
	if err != nil {
		t.Fatalf("ValidateFileResolved() error = %v", err)
	}
	if result.Valid {
		t.Error("expected invalid (subject schema requires marks), got valid")
	}
}

func TestValidateFileResolved_FallsBackToGlobal(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))
	v := validator.NewWithResolver(resolver)

	// Topic schema only exists globally (no subject override)
	topicFile := filepath.Join(root, "curricula", "country", "syllabus", "subject", "grade", "topics", "MT1-01.yaml")

	result, err := v.ValidateFileResolved(topicFile, "topic")
	if err != nil {
		t.Fatalf("ValidateFileResolved() error = %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid (global schema), got errors: %v", result.Errors)
	}
}

func TestValidateContentResolved_Valid(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))
	v := validator.NewWithResolver(resolver)

	validContent := []byte(`topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
    learning_objective: "1.1.1"
    answer:
      type: exact
      value: "4"
    marks: 2
`)
	subjectSchemasDir := filepath.Join(root, "curricula", "country", "syllabus", "subject", "schemas")

	result, err := v.ValidateContentResolved(validContent, "assessments", subjectSchemasDir)
	if err != nil {
		t.Fatalf("ValidateContentResolved() error = %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateContentResolved_Invalid(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))
	v := validator.NewWithResolver(resolver)

	// Missing marks — subject schema requires it
	invalidContent := []byte(`topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
    learning_objective: "1.1.1"
    answer:
      type: exact
      value: "4"
`)
	subjectSchemasDir := filepath.Join(root, "curricula", "country", "syllabus", "subject", "schemas")

	result, err := v.ValidateContentResolved(invalidContent, "assessments", subjectSchemasDir)
	if err != nil {
		t.Fatalf("ValidateContentResolved() error = %v", err)
	}
	if result.Valid {
		t.Error("expected invalid (marks missing), got valid")
	}
	if len(result.Errors) == 0 {
		t.Error("expected errors, got none")
	}
}

func TestValidateContentResolved_GlobalFallback(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))
	v := validator.NewWithResolver(resolver)

	// Without marks, but using empty subjectSchemasDir (global fallback, which doesn't require marks)
	content := []byte(`topic_id: MT1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
    learning_objective: "1.1.1"
    answer:
      type: exact
      value: "4"
`)

	result, err := v.ValidateContentResolved(content, "assessments", "")
	if err != nil {
		t.Fatalf("ValidateContentResolved() error = %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid (global schema, marks not required), got errors: %v", result.Errors)
	}
}

func TestValidateDirResolved(t *testing.T) {
	root := setupResolvedTestDirs(t)
	resolver := validator.NewSchemaResolver(filepath.Join(root, "schema"))
	v := validator.NewWithResolver(resolver)

	targetDir := filepath.Join(root, "curricula")

	results, err := v.ValidateDirResolved(targetDir)
	if err != nil {
		t.Fatalf("ValidateDirResolved() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("ValidateDirResolved() returned no results")
	}

	// The topic file should pass (global schema), the assessments file should fail (subject schema)
	var topicResult, assessmentsResult *validator.ValidationResult
	for i := range results {
		if results[i].Type == "topic" {
			topicResult = &results[i]
		}
		if results[i].Type == "assessments" {
			assessmentsResult = &results[i]
		}
	}

	if topicResult == nil {
		t.Fatal("no topic validation result found")
	}
	if !topicResult.Valid {
		t.Errorf("topic should be valid, got errors: %v", topicResult.Errors)
	}

	if assessmentsResult == nil {
		t.Fatal("no assessments validation result found")
	}
	if assessmentsResult.Valid {
		t.Error("assessments should be invalid (subject schema requires marks)")
	}
}
