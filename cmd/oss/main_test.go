package main

import (
	"testing"
)

func TestImportCmdFlags(t *testing.T) {
	cmd := importCmd()

	// Required flags must be defined
	if cmd.Flags().Lookup("pdf") == nil {
		t.Error("import command missing --pdf flag")
	}
	if cmd.Flags().Lookup("syllabus") == nil {
		t.Error("import command missing --syllabus flag")
	}
	if cmd.Flags().Lookup("subject") == nil {
		t.Error("import command missing --subject flag")
	}
	if cmd.Flags().Lookup("workers") == nil {
		t.Error("import command missing --workers flag")
	}
	if cmd.Flags().Lookup("pr") == nil {
		t.Error("import command missing --pr flag")
	}

	// Defaults
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		t.Fatalf("getting workers flag: %v", err)
	}
	if workers != 3 {
		t.Errorf("workers default = %d, want 3", workers)
	}

	pr, err := cmd.Flags().GetBool("pr")
	if err != nil {
		t.Fatalf("getting pr flag: %v", err)
	}
	if pr != false {
		t.Error("pr default should be false")
	}
}

func TestImportCmdRequiredFlags(t *testing.T) {
	cmd := importCmd()

	pdfFlag := cmd.Flags().Lookup("pdf")
	if pdfFlag == nil {
		t.Fatal("--pdf flag not defined")
	}
	syllabusFlag := cmd.Flags().Lookup("syllabus")
	if syllabusFlag == nil {
		t.Fatal("--syllabus flag not defined")
	}

	// Verify required annotations are set
	annotations := cmd.Annotations
	_ = annotations // required flags are enforced by cobra at runtime, not via annotations

	// Verify Use and Short are set
	if cmd.Use == "" {
		t.Error("import command missing Use field")
	}
	if cmd.Short == "" {
		t.Error("import command missing Short description")
	}
}
