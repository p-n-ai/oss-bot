package pipeline

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnforceSchemaRequiredFields_AddsMissing(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["id", "name", "provenance", "quality_level"],
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"provenance": {"type": "string", "enum": ["human", "ai-generated", "ai-assisted"]},
			"quality_level": {"type": "integer"}
		}
	}`

	// Input YAML missing "provenance" and "quality_level".
	input := "id: MT4-01\nname: \"Algebra\"\n"

	result := EnforceSchemaRequiredFields(input, schema)

	// Parse result to verify.
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result is not valid YAML: %v\n%s", err, result)
	}

	if _, ok := data["provenance"]; !ok {
		t.Error("missing required field 'provenance' should have been added")
	}
	if _, ok := data["quality_level"]; !ok {
		t.Error("missing required field 'quality_level' should have been added")
	}
	// provenance should default to first enum value.
	if data["provenance"] != "human" {
		t.Errorf("provenance should default to first enum 'human', got %v", data["provenance"])
	}
}

func TestEnforceSchemaRequiredFields_NoChangesWhenComplete(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["id", "name"],
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"}
		}
	}`

	input := "id: MT4-01\nname: \"Algebra\"\n"
	result := EnforceSchemaRequiredFields(input, schema)

	// Should be unchanged (or at least semantically identical).
	var orig, res map[string]interface{}
	yaml.Unmarshal([]byte(input), &orig)
	yaml.Unmarshal([]byte(result), &res)

	if len(res) != len(orig) {
		t.Errorf("no fields should be added when all required are present; orig=%d, result=%d", len(orig), len(res))
	}
}

func TestEnforceSchemaRequiredFields_NestedObject(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["mastery"],
		"properties": {
			"mastery": {
				"type": "object",
				"required": ["minimum_score", "assessment_count"],
				"properties": {
					"minimum_score": {"type": "number"},
					"assessment_count": {"type": "integer"}
				}
			}
		}
	}`

	// mastery exists but is missing assessment_count.
	input := "mastery:\n  minimum_score: 0.75\n"
	result := EnforceSchemaRequiredFields(input, schema)

	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("invalid YAML: %v\n%s", err, result)
	}

	mastery, ok := data["mastery"].(map[string]interface{})
	if !ok {
		t.Fatal("mastery should be a map")
	}
	if _, ok := mastery["assessment_count"]; !ok {
		t.Error("nested required field 'assessment_count' should have been added")
	}
}

func TestEnforceSchemaRequiredFields_ArrayItems(t *testing.T) {
	schema := `{
		"type": "object",
		"required": ["questions"],
		"properties": {
			"questions": {
				"type": "array",
				"items": {
					"type": "object",
					"required": ["id", "text", "difficulty"],
					"properties": {
						"id": {"type": "string"},
						"text": {"type": "string"},
						"difficulty": {"type": "string", "enum": ["easy", "medium", "hard"]}
					}
				}
			}
		}
	}`

	// Array item missing "difficulty".
	input := "questions:\n  - id: Q1\n    text: \"What is 2+2?\"\n"
	result := EnforceSchemaRequiredFields(input, schema)

	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("invalid YAML: %v\n%s", err, result)
	}

	questions, ok := data["questions"].([]interface{})
	if !ok || len(questions) == 0 {
		t.Fatal("questions should be a non-empty array")
	}
	q1, ok := questions[0].(map[string]interface{})
	if !ok {
		t.Fatal("question should be a map")
	}
	if _, ok := q1["difficulty"]; !ok {
		t.Error("missing required array item field 'difficulty' should have been added")
	}
}

func TestEnforceSchemaRequiredFields_InvalidInputs(t *testing.T) {
	// Invalid schema JSON.
	result := EnforceSchemaRequiredFields("id: test\n", "not json")
	if result != "id: test\n" {
		t.Error("should return input unchanged for invalid schema")
	}

	// Invalid YAML.
	result = EnforceSchemaRequiredFields("{{invalid yaml", `{"type":"object","required":["id"],"properties":{"id":{"type":"string"}}}`)
	if result != "{{invalid yaml" {
		t.Error("should return input unchanged for invalid YAML")
	}
}

func TestEnforceStringQuoting_QuotesUnquotedStrings(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"quality_level": {"type": "integer"},
			"provenance": {"type": "string"}
		}
	}`

	// "id" and "provenance" are unquoted string values.
	input := "id: MT4-01\nname: \"Algebra\"\nquality_level: 1\nprovenance: ai-generated\n"
	result := EnforceStringQuoting(input, schema)

	// id and provenance should now be double-quoted.
	if !strings.Contains(result, `"MT4-01"`) {
		t.Errorf("id should be double-quoted in result:\n%s", result)
	}
	if !strings.Contains(result, `"ai-generated"`) {
		t.Errorf("provenance should be double-quoted in result:\n%s", result)
	}
	// name was already quoted, should stay quoted.
	if !strings.Contains(result, `"Algebra"`) {
		t.Errorf("name should remain double-quoted in result:\n%s", result)
	}

	// Parse to verify still valid YAML.
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result is not valid YAML: %v\n%s", err, result)
	}
	if data["id"] != "MT4-01" {
		t.Errorf("id value should be preserved, got %v", data["id"])
	}
}

func TestEnforceStringQuoting_PreservesAlreadyQuoted(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`

	input := "name: \"Already Quoted\"\n"
	result := EnforceStringQuoting(input, schema)

	if !strings.Contains(result, `"Already Quoted"`) {
		t.Errorf("already quoted value should remain quoted:\n%s", result)
	}
}

func TestEnforceStringQuoting_PreservesSingleQuoted(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"text": {"type": "string"}
		}
	}`

	input := "text: 'single quoted value'\n"
	result := EnforceStringQuoting(input, schema)

	// Single-quoted should not be changed to double-quoted.
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result is not valid YAML: %v\n%s", err, result)
	}
	if data["text"] != "single quoted value" {
		t.Errorf("value should be preserved, got %v", data["text"])
	}
}

func TestEnforceStringQuoting_NestedAndArray(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"questions": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"id": {"type": "string"},
						"text": {"type": "string"},
						"marks": {"type": "integer"}
					}
				}
			}
		}
	}`

	input := "questions:\n  - id: Q1\n    text: What is 2+2\n    marks: 3\n"
	result := EnforceStringQuoting(input, schema)

	if !strings.Contains(result, `"Q1"`) {
		t.Errorf("array item string field 'id' should be double-quoted:\n%s", result)
	}
	if !strings.Contains(result, `"What is 2+2"`) {
		t.Errorf("array item string field 'text' should be double-quoted:\n%s", result)
	}

	// Verify valid YAML.
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result is not valid YAML: %v\n%s", err, result)
	}
}

func TestEnforceStringQuoting_DoesNotQuoteNonStrings(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"quality_level": {"type": "integer"},
			"active": {"type": "boolean"}
		}
	}`

	input := "quality_level: 1\nactive: true\n"
	result := EnforceStringQuoting(input, schema)

	// Non-string fields should not be quoted.
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("result is not valid YAML: %v\n%s", err, result)
	}
	if data["quality_level"] != 1 {
		t.Errorf("integer should remain unquoted, got %v (%T)", data["quality_level"], data["quality_level"])
	}
}

func TestEnforceStringQuoting_InvalidInputs(t *testing.T) {
	result := EnforceStringQuoting("id: test\n", "not json")
	if result != "id: test\n" {
		t.Error("should return input unchanged for invalid schema")
	}

	result = EnforceStringQuoting("{{invalid", `{"type":"object","properties":{"id":{"type":"string"}}}`)
	if result != "{{invalid" {
		t.Error("should return input unchanged for invalid YAML")
	}
}
