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

func TestBloomLevel_CrossSubjectVerbs(t *testing.T) {
	tests := []struct {
		name     string
		verb     string
		expected string
	}{
		// Science verbs
		{"science-hypothesize", "hypothesize", "analyze"},
		{"science-experiment", "experiment", "apply"},
		{"science-observe", "observe", "understand"},
		{"science-measure", "measure", "apply"},
		{"science-predict", "predict", "analyze"},
		// Humanities verbs
		{"humanities-contextualize", "contextualize", "analyze"},
		{"humanities-empathize", "empathize", "understand"},
		// General verbs
		{"general-synthesize", "synthesize", "create"},
		{"general-reflect", "reflect", "evaluate"},
		{"general-research", "research", "analyze"},
		{"general-collaborate", "collaborate", "apply"},
		{"general-present", "present", "apply"},
		// Existing verbs should still work
		{"math-solve", "solve", "apply"},
		{"math-calculate", "calculate", "apply"},
		{"math-define", "define", "remember"},
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
