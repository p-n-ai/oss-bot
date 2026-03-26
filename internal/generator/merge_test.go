package generator_test

import (
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/generator"
)

// --- MergeAssessments ---

func TestMergeAssessments_AppendAndDedup(t *testing.T) {
	existing := []generator.Assessment{
		{ID: "Q1", Text: "What is 2+2?", Difficulty: "easy"},
		{ID: "Q2", Text: "Solve: 3x = 12", Difficulty: "medium"},
	}
	generated := []generator.Assessment{
		{ID: "Q1", Text: "What is 2+2?", Difficulty: "easy"}, // near-duplicate
		{ID: "Q3", Text: "Find the roots of x^2 - 5x + 6 = 0", Difficulty: "hard"},
	}

	merged, report := generator.MergeAssessments(existing, generated)

	if len(merged) != 3 {
		t.Errorf("expected 3 merged assessments, got %d", len(merged))
	}
	if report.Added != 1 {
		t.Errorf("expected 1 added, got %d", report.Added)
	}
	if report.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Skipped)
	}
}

func TestMergeAssessments_NoExisting(t *testing.T) {
	generated := []generator.Assessment{
		{ID: "Q1", Text: "What is x in 2x=8?", Difficulty: "easy"},
		{ID: "Q2", Text: "Expand (x+2)(x-3)", Difficulty: "medium"},
	}

	merged, report := generator.MergeAssessments(nil, generated)

	if len(merged) != 2 {
		t.Errorf("expected 2 assessments, got %d", len(merged))
	}
	if report.Added != 2 {
		t.Errorf("expected 2 added, got %d", report.Added)
	}
	if report.Skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", report.Skipped)
	}
}

func TestMergeAssessments_AllDuplicates(t *testing.T) {
	existing := []generator.Assessment{
		{ID: "Q1", Text: "What is 2+2?", Difficulty: "easy"},
	}
	// Identical question text
	generated := []generator.Assessment{
		{ID: "Q1-new", Text: "What is 2+2?", Difficulty: "easy"},
	}

	merged, report := generator.MergeAssessments(existing, generated)

	if len(merged) != 1 {
		t.Errorf("expected 1 (no additions), got %d", len(merged))
	}
	if report.Added != 0 {
		t.Errorf("expected 0 added, got %d", report.Added)
	}
	if report.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Skipped)
	}
}

// --- MergeExamples ---

func TestMergeExamples_AppendDedupSort(t *testing.T) {
	existing := []generator.Example{
		{ID: "WE-01", Topic: "Basic algebra", Difficulty: "easy"},
	}
	generated := []generator.Example{
		{ID: "WE-01-dup", Topic: "Basic algebra", Difficulty: "easy"}, // duplicate by topic
		{ID: "WE-02", Topic: "Quadratic formula", Difficulty: "hard"},
		{ID: "WE-03", Topic: "Factoring", Difficulty: "medium"},
	}

	merged, report := generator.MergeExamples(existing, generated)

	if len(merged) != 3 {
		t.Errorf("expected 3 merged examples, got %d", len(merged))
	}
	if report.Added != 2 {
		t.Errorf("expected 2 added, got %d", report.Added)
	}
	if report.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Skipped)
	}
	// Verify sort order: easy → medium → hard
	wantOrder := []string{"easy", "medium", "hard"}
	for i, ex := range merged {
		if ex.Difficulty != wantOrder[i] {
			t.Errorf("merged[%d].Difficulty = %q, want %q", i, ex.Difficulty, wantOrder[i])
		}
	}
}

func TestMergeExamples_NoExisting(t *testing.T) {
	generated := []generator.Example{
		{ID: "WE-01", Topic: "Algebra basics", Difficulty: "easy"},
		{ID: "WE-02", Topic: "Quadratics", Difficulty: "hard"},
	}

	merged, report := generator.MergeExamples(nil, generated)

	if len(merged) != 2 {
		t.Errorf("expected 2 examples, got %d", len(merged))
	}
	if report.Added != 2 {
		t.Errorf("expected 2 added, got %d", report.Added)
	}
}

// --- MergeTeachingNotes ---

func TestMergeTeachingNotes_Additive(t *testing.T) {
	existing := generator.TeachingNotes(`# Topic — Teaching Notes

## Overview
Existing overview.

## Teaching Sequence & Strategy
Existing strategy.
`)
	generated := generator.TeachingNotes(`# Topic — Teaching Notes

## Overview
New overview (should NOT replace existing).

## High Alert Misconceptions
New misconceptions section.
`)

	merged, report := generator.MergeTeachingNotes(existing, generated)

	if !strings.Contains(string(merged), "Existing overview") {
		t.Error("merged should preserve existing Overview section")
	}
	if strings.Contains(string(merged), "New overview") {
		t.Error("merged should NOT replace existing Overview with generated one")
	}
	if !strings.Contains(string(merged), "High Alert Misconceptions") {
		t.Error("merged should add new sections from generated")
	}
	if !strings.Contains(string(merged), "Teaching Sequence") {
		t.Error("merged should preserve existing Teaching Sequence section")
	}
	if report.Added < 1 {
		t.Errorf("expected at least 1 added section, got %d", report.Added)
	}
}

func TestMergeTeachingNotes_NoExisting(t *testing.T) {
	generated := generator.TeachingNotes("# Notes\n\n## Overview\nContent.\n")

	merged, report := generator.MergeTeachingNotes("", generated)

	if string(merged) != string(generated) {
		t.Error("with no existing content, merged should equal generated")
	}
	_ = report
}

func TestMergeTeachingNotes_NoNewSections(t *testing.T) {
	existing := generator.TeachingNotes("## Overview\nExists.\n\n## Assessment Guidance\nExists.\n")
	generated := generator.TeachingNotes("## Overview\nNew.\n\n## Assessment Guidance\nNew.\n")

	merged, report := generator.MergeTeachingNotes(existing, generated)

	if report.Added != 0 {
		t.Errorf("expected 0 sections added (all exist), got %d", report.Added)
	}
	// Existing content must be unchanged
	if !strings.Contains(string(merged), "Overview") {
		t.Error("existing Overview must be present")
	}
}

// --- MergeReport.String ---

func TestMergeReport_String(t *testing.T) {
	r := generator.MergeReport{Added: 3, Skipped: 1}
	s := r.String()
	if !strings.Contains(s, "3") || !strings.Contains(s, "1") {
		t.Errorf("MergeReport.String() = %q, expected counts", s)
	}

	empty := generator.MergeReport{}
	if empty.String() == "" {
		t.Error("MergeReport.String() for empty report should not be empty")
	}
}

// --- MergeAssessmentsYAML ---

func TestMergeAssessmentsYAML(t *testing.T) {
	existingYAML := `topic_id: F1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
`
	generatedYAML := `topic_id: F1-01
provenance: ai-generated
questions:
  - id: Q1
    text: "What is 2+2?"
    difficulty: easy
  - id: Q2
    text: "Solve x^2 - 4 = 0"
    difficulty: hard
`

	merged, report, err := generator.MergeAssessmentsYAML(existingYAML, generatedYAML)
	if err != nil {
		t.Fatalf("MergeAssessmentsYAML() error = %v", err)
	}
	if report.Added != 1 {
		t.Errorf("expected 1 added, got %d", report.Added)
	}
	if report.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Skipped)
	}
	if !strings.Contains(merged, "Solve x^2") {
		t.Error("merged YAML should contain the new question")
	}
	if !strings.Contains(merged, "What is 2+2") {
		t.Error("merged YAML should preserve the existing question")
	}
}

// --- MergeExamplesYAML ---

func TestMergeExamplesYAML(t *testing.T) {
	existingYAML := `topic_id: F1-01
provenance: ai-generated
worked_examples:
  - id: WE-01
    topic: "Linear equations"
    difficulty: easy
    scenario: "Solve 2x = 8"
`
	generatedYAML := `topic_id: F1-01
provenance: ai-generated
worked_examples:
  - id: WE-01
    topic: "Linear equations"
    difficulty: easy
    scenario: "Solve 2x = 8"
  - id: WE-02
    topic: "Quadratics"
    difficulty: hard
    scenario: "Find roots of x^2-5x+6=0"
`

	merged, report, err := generator.MergeExamplesYAML(existingYAML, generatedYAML)
	if err != nil {
		t.Fatalf("MergeExamplesYAML() error = %v", err)
	}
	if report.Added != 1 {
		t.Errorf("expected 1 added, got %d", report.Added)
	}
	if report.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", report.Skipped)
	}
	if !strings.Contains(merged, "Quadratics") {
		t.Error("merged YAML should contain the new example")
	}
}
