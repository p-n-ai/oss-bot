package pipeline_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

func TestNoopReporter_ImplementsInterface(t *testing.T) {
	var _ pipeline.ProgressReporter = &pipeline.NoopReporter{}
}

func TestCLIReporter_ImplementsInterface(t *testing.T) {
	var _ pipeline.ProgressReporter = pipeline.NewCLIReporter()
}

func TestNoopReporter_AllMethodsCallable(t *testing.T) {
	r := &pipeline.NoopReporter{}
	r.OnStart(10)
	r.OnProgress(1, 10, "topic-1", "generating")
	r.OnError("topic-2", nil)
	r.OnComplete(nil)
}

func TestCLIReporter_TracksProgress(t *testing.T) {
	r := pipeline.NewCLIReporter()
	r.OnStart(5)

	r.OnProgress(1, 5, "topic-algebra", "generating")
	r.OnProgress(2, 5, "topic-geometry", "generating")
	r.OnProgress(3, 5, "topic-calculus", "validating")

	if r.Current() != 3 {
		t.Errorf("Current() = %d, want 3", r.Current())
	}
	if r.Total() != 5 {
		t.Errorf("Total() = %d, want 5", r.Total())
	}
}

func TestCLIReporter_TrackErrors(t *testing.T) {
	r := pipeline.NewCLIReporter()
	r.OnStart(3)
	r.OnError("topic-bad", nil)
	r.OnError("topic-also-bad", nil)

	if r.ErrorCount() != 2 {
		t.Errorf("ErrorCount() = %d, want 2", r.ErrorCount())
	}
}

func TestCLIReporter_OnComplete(t *testing.T) {
	r := pipeline.NewCLIReporter()
	r.OnStart(2)
	r.OnProgress(1, 2, "topic-a", "done")
	r.OnProgress(2, 2, "topic-b", "done")

	// Should not panic
	r.OnComplete(nil)
}
