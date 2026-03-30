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

func TestValidateBloomLevels(t *testing.T) {
	tests := []struct {
		name       string
		objectives []validator.LearningObjective
		wantErrors int
	}{
		{
			name: "all valid",
			objectives: []validator.LearningObjective{
				{ID: "1.0.1", Bloom: "understand"},
				{ID: "2.0.1", Bloom: "apply"},
				{ID: "3.0.1", Bloom: "analyze"},
			},
			wantErrors: 0,
		},
		{
			name: "unrecognised level",
			objectives: []validator.LearningObjective{
				{ID: "1.0.1", Bloom: "understand"},
				{ID: "2.0.1", Bloom: "think_hard"}, // not a valid level
			},
			wantErrors: 1,
		},
		{
			name: "missing bloom level",
			objectives: []validator.LearningObjective{
				{ID: "1.0.1", Bloom: ""},
			},
			wantErrors: 1,
		},
		{
			name:       "empty objectives",
			objectives: nil,
			wantErrors: 0,
		},
		{
			name: "cross-subject verbs recognised",
			objectives: []validator.LearningObjective{
				{ID: "1.0.1", Bloom: "analyze"},  // used for hypothesize/predict
				{ID: "2.0.1", Bloom: "create"},   // used for synthesize
				{ID: "3.0.1", Bloom: "evaluate"}, // used for reflect
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateBloomLevels(tt.objectives)
			if len(errs) != tt.wantErrors {
				t.Errorf("ValidateBloomLevels() returned %d errors, want %d: %v",
					len(errs), tt.wantErrors, errs)
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
				{ID: "1.0.1", Bloom: "apply"},
			},
			questions: []validator.AssessmentQuestion{
				{ID: "Q1", LearningObjective: "1.0.1", Text: "Solve the equation 2x + 3 = 7"},
			},
			wantErrors: 0,
		},
		{
			name: "question-exceeds-bloom",
			objectives: []validator.LearningObjective{
				{ID: "1.0.1", Bloom: "remember"},
			},
			questions: []validator.AssessmentQuestion{
				{ID: "Q1", LearningObjective: "1.0.1", Text: "Evaluate and compare the two approaches"},
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
