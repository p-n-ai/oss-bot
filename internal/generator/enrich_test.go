package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnrichTopicYAML(t *testing.T) {
	topicYAML := `id: MT5-01
name: "Nombor Nisbah"
name_en: "Rational Numbers"
subject_grade_id: malaysia-kssm-matematik-tingkatan-5
subject_id: malaysia-kssm-matematik
syllabus_id: malaysia-kssm
country_id: malaysia
language: ms
difficulty: intermediate
tier: core
learning_objectives:
  - id: "1.1.1"
    text: "Mengenal pasti nombor nisbah."
    text_en: "Identify rational numbers."
    bloom: remember
prerequisites:
  required: []
bloom_levels:
  - remember
quality_level: 1
provenance: ai-generated
`

	enrichmentYAML := `teaching:
  sequence:
    - "Step 1: Review integers and fractions (10 min)"
    - "Step 2: Define rational numbers (15 min)"
    - "Step 3: Practice identifying rational vs irrational (15 min)"
  common_misconceptions:
    - misconception: "All decimals are irrational"
      remediation: "Show that 0.5 = 1/2 is rational"
    - misconception: "Negative numbers cannot be rational"
      remediation: "Demonstrate that -3 = -3/1"
engagement_hooks:
  - "Is pi a rational number? Why or why not?"
  - "Can you find a fraction equal to 0.333...?"
`

	// Write topic to a temp file
	dir := t.TempDir()
	topicFile := filepath.Join(dir, "MT5-01.yaml")
	if err := os.WriteFile(topicFile, []byte(topicYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := EnrichTopicYAML(topicFile, enrichmentYAML)
	if err != nil {
		t.Fatalf("EnrichTopicYAML failed: %v", err)
	}

	// Verify original fields preserved
	for _, want := range []string{"id: MT5-01", "name: \"Nombor Nisbah\"", "difficulty: intermediate", "quality_level: 1"} {
		if !strings.Contains(result, want) {
			t.Errorf("missing original field: %s", want)
		}
	}

	// Verify enrichment fields added
	for _, want := range []string{
		"teaching:",
		"sequence:",
		"Review integers and fractions",
		"common_misconceptions:",
		"All decimals are irrational",
		"remediation:",
		"engagement_hooks:",
		"Is pi a rational number",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("missing enrichment field: %s", want)
		}
	}
}

func TestEnrichTopicYAML_OverwritesExisting(t *testing.T) {
	topicYAML := `id: MT5-02
name: "Test Topic"
teaching:
  sequence:
    - "Old step"
engagement_hooks:
  - "Old hook"
`

	enrichmentYAML := `teaching:
  sequence:
    - "New step 1"
    - "New step 2"
  common_misconceptions:
    - misconception: "Wrong idea"
      remediation: "Fix it"
engagement_hooks:
  - "New hook"
`

	dir := t.TempDir()
	topicFile := filepath.Join(dir, "MT5-02.yaml")
	if err := os.WriteFile(topicFile, []byte(topicYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := EnrichTopicYAML(topicFile, enrichmentYAML)
	if err != nil {
		t.Fatalf("EnrichTopicYAML failed: %v", err)
	}

	if strings.Contains(result, "Old step") {
		t.Error("old teaching.sequence should be replaced")
	}
	if strings.Contains(result, "Old hook") {
		t.Error("old engagement_hooks should be replaced")
	}
	if !strings.Contains(result, "New step 1") {
		t.Error("new teaching.sequence not found")
	}
	if !strings.Contains(result, "New hook") {
		t.Error("new engagement_hooks not found")
	}
}

func TestBuildTopicEnrichmentPrompt(t *testing.T) {
	genCtx := &GenerationContext{
		Topic: Topic{
			ID:         "MT5-01",
			Name:       "Nombor Nisbah",
			SubjectID:  "malaysia-kssm-matematik",
			SyllabusID: "malaysia-kssm",
			Difficulty: "intermediate",
			LearningObjectives: []LearningObjective{
				{ID: "1.1.1", Bloom: "remember", Text: "Identify rational numbers"},
			},
		},
		TopicDir: t.TempDir(),
	}

	prompt := BuildTopicEnrichmentPrompt(genCtx)

	for _, want := range []string{
		"MT5-01",
		"Nombor Nisbah",
		"teaching:",
		"sequence:",
		"common_misconceptions:",
		"engagement_hooks:",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing: %s", want)
		}
	}
}
