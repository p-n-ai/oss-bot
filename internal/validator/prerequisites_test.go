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
