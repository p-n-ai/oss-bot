package pipeline

import (
	"strings"
	"testing"
)

func TestExtractSchemaDescriptions_BasicFields(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["topic_id", "questions"],
		"properties": {
			"topic_id": {
				"type": "string",
				"description": "The unique identifier for this topic, e.g. MT4-01"
			},
			"provenance": {
				"type": "string",
				"description": "Origin of the content",
				"enum": ["human", "ai-generated", "ai-assisted"]
			},
			"questions": {
				"type": "array",
				"description": "List of assessment questions for the topic",
				"items": {
					"type": "object",
					"required": ["id", "text", "difficulty"],
					"properties": {
						"id": {
							"type": "string",
							"description": "Unique question identifier like Q1, Q2"
						},
						"text": {
							"type": "string",
							"description": "The question text presented to students"
						},
						"difficulty": {
							"type": "string",
							"description": "Difficulty level of the question",
							"enum": ["easy", "medium", "hard"]
						}
					}
				}
			}
		}
	}`

	result := ExtractSchemaDescriptions(schema)

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// Should contain the field guide header.
	if !strings.Contains(result, "Field Descriptions") {
		t.Error("missing field guide header")
	}

	// Should contain field descriptions.
	checks := []string{
		"topic_id (string, required): The unique identifier",
		"questions (array, required): List of assessment questions",
		"provenance (string): Origin of the content",
		"Values: human | ai-generated | ai-assisted",
		"id (string, required): Unique question identifier",
		"text (string, required): The question text",
		"difficulty (string, required): Difficulty level",
		"Values: easy | medium | hard",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("missing expected content: %q\nGot:\n%s", check, result)
		}
	}
}

func TestExtractSchemaDescriptions_NestedObject(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["teaching"],
		"properties": {
			"teaching": {
				"type": "object",
				"description": "Teaching methodology and sequence",
				"required": ["sequence"],
				"properties": {
					"sequence": {
						"type": "array",
						"description": "Ordered teaching steps with durations"
					},
					"common_misconceptions": {
						"type": "array",
						"description": "Known student misconceptions with remediations"
					}
				}
			}
		}
	}`

	result := ExtractSchemaDescriptions(schema)

	if !strings.Contains(result, "teaching (object, required): Teaching methodology") {
		t.Errorf("missing teaching field description\nGot:\n%s", result)
	}
	if !strings.Contains(result, "  - sequence (array, required): Ordered teaching steps") {
		t.Errorf("missing nested sequence field\nGot:\n%s", result)
	}
	if !strings.Contains(result, "  - common_misconceptions (array): Known student misconceptions") {
		t.Errorf("missing nested misconceptions field\nGot:\n%s", result)
	}
}

func TestExtractSchemaDescriptions_InvalidJSON(t *testing.T) {
	result := ExtractSchemaDescriptions("not json")
	if result != "" {
		t.Errorf("expected empty result for invalid JSON, got: %s", result)
	}
}

func TestExtractSchemaDescriptions_NoProperties(t *testing.T) {
	result := ExtractSchemaDescriptions(`{"type": "object"}`)
	if result != "" {
		t.Errorf("expected empty result for schema without properties, got: %s", result)
	}
}

func TestExtractSchemaDescriptions_NoDescriptions(t *testing.T) {
	// Fields without descriptions at top level should still appear.
	schema := `{
		"type": "object",
		"required": ["id"],
		"properties": {
			"id": {"type": "string"}
		}
	}`
	result := ExtractSchemaDescriptions(schema)
	if !strings.Contains(result, "id (string, required)") {
		t.Errorf("required top-level field without description should still appear\nGot:\n%s", result)
	}
}
