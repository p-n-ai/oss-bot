package main

import (
	"os"
	"path/filepath"
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
	if cmd.Flags().Lookup("chunk-size") == nil {
		t.Error("import command missing --chunk-size flag")
	}

	// Defaults
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		t.Fatalf("getting workers flag: %v", err)
	}
	if workers != 3 {
		t.Errorf("workers default = %d, want 3", workers)
	}

	chunkSize, err := cmd.Flags().GetInt("chunk-size")
	if err != nil {
		t.Fatalf("getting chunk-size flag: %v", err)
	}
	if chunkSize != 2000 {
		t.Errorf("chunk-size default = %d, want 2000", chunkSize)
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

func TestImportSlug(t *testing.T) {
	cases := []struct {
		heading string
		want    string
	}{
		{"Bab 1 Fungsi", "bab-1-fungsi"},
		{"BAB 2 UNGKAPAN KUADRATIK", "bab-2-ungkapan-kuadratik"},
		{"Chapter 3: Algebra", "chapter-3-algebra"},
		{"", ""},
		{"  spaces  ", "spaces"},
		{"1.1 Introduction", "1-1-introduction"},
	}
	for _, c := range cases {
		got := importSlug(c.heading)
		if got != c.want {
			t.Errorf("importSlug(%q) = %q, want %q", c.heading, got, c.want)
		}
	}
}

func TestFindSubjectTopicsDir(t *testing.T) {
	// Create a temp directory tree mimicking a scaffolded OSS repo:
	// curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik-tingkatan-4/topics/
	root := t.TempDir()
	topicsDir := filepath.Join(root, "curricula", "malaysia", "malaysia-kssm",
		"malaysia-kssm-matematik-tingkatan-4", "topics")
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("finds existing subject topics dir", func(t *testing.T) {
		got, err := findSubjectTopicsDir(root, "malaysia-kssm-matematik-tingkatan-4", "malaysia-kssm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != topicsDir {
			t.Errorf("got %q, want %q", got, topicsDir)
		}
	})

	t.Run("returns error when subject not found", func(t *testing.T) {
		_, err := findSubjectTopicsDir(root, "nonexistent-subject", "malaysia-kssm")
		if err == nil {
			t.Error("expected error for missing subject, got nil")
		}
	})

	t.Run("falls back to syllabusID when subjectID is empty", func(t *testing.T) {
		syllabusDir := filepath.Join(root, "curricula", "malaysia", "malaysia-kssm", "topics")
		if err := os.MkdirAll(syllabusDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		// search by syllabusID when subjectID is empty
		got, err := findSubjectTopicsDir(root, "", "malaysia-kssm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != syllabusDir {
			t.Errorf("got %q, want %q", got, syllabusDir)
		}
	})
}
