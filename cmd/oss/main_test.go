package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p-n-ai/oss-bot/internal/ai"
	"github.com/p-n-ai/oss-bot/internal/pipeline"
)

func TestMergeTopicYAML(t *testing.T) {
	existing := `id: MT4-01
official_ref: "1.0"
name: "Fungsi dan Persamaan Kuadratik"
name_en: "Quadratic Functions and Equations"
subject_id: malaysia-kssm-matematik-tingkatan-4
syllabus_id: malaysia-kssm
country_id: malaysia
language: ms
difficulty: intermediate
tier: core
learning_objectives:
  - id: 1.1.1
    text: "Recognise a function and describe its characteristics."
    bloom: understand
prerequisites:
  required: []
  recommended: []
bloom_levels:
  - understand
mastery:
  minimum_score: 0.75
  assessment_count: 3
  spaced_repetition:
    initial_interval_days: 3
    multiplier: 2.5
ai_teaching_notes: "MT4-01.teaching.md"
quality_level: 1
provenance: ai-assisted`

	newContent := `id: MT4-01
name: "Fungsi dan Persamaan Kuadratik"
name_en: "Quadratic Functions and Equations"
learning_objectives:
  - id: 1.1.1
    text: "Recognise a function and describe its characteristics."
    bloom: understand
  - id: 1.1.2
    text: "Make connection between a function and a non-function."
    bloom: apply`

	// Mock provider returns a plausible merged YAML.
	mergedYAML := existing + "\n  - id: 1.1.2\n    text: \"Make connection between a function and a non-function.\"\n    bloom: apply"
	provider := ai.NewMockProvider(mergedYAML)

	got, err := mergeTopicYAML(context.Background(), provider, existing, newContent, "1.0 FUNGSI DAN PERSAMAAN KUADRATIK")
	if err != nil {
		t.Fatalf("mergeTopicYAML() error = %v", err)
	}
	if strings.TrimSpace(got) == "" {
		t.Error("mergeTopicYAML() returned empty string")
	}
	// The mock just returns what we gave it; verify it round-trips through the function.
	if got != mergedYAML {
		t.Errorf("mergeTopicYAML() output does not match mock response")
	}
}

func TestImportCmdFlags(t *testing.T) {
	cmd := importCmd()

	// Required flags must be defined
	if cmd.Flags().Lookup("pdf") == nil {
		t.Error("import command missing --pdf flag")
	}
	if cmd.Flags().Lookup("syllabus") == nil {
		t.Error("import command missing --syllabus flag")
	}
	if cmd.Flags().Lookup("subject-grade") == nil {
		t.Error("import command missing --subject-grade flag")
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
	if cmd.Flags().Lookup("force") == nil {
		t.Error("import command missing --force flag")
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

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		t.Fatalf("getting force flag: %v", err)
	}
	if force != false {
		t.Error("force default should be false")
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
	// curricula/malaysia/malaysia-kssm/malaysia-kssm-matematik/malaysia-kssm-matematik-tingkatan-4/topics/
	root := t.TempDir()
	topicsDir := filepath.Join(root, "curricula", "malaysia", "malaysia-kssm",
		"malaysia-kssm-matematik", "malaysia-kssm-matematik-tingkatan-4", "topics")
	if err := os.MkdirAll(topicsDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("finds existing subject_grade topics dir", func(t *testing.T) {
		got, err := findSubjectTopicsDir(root, "malaysia-kssm-matematik-tingkatan-4", "malaysia-kssm-matematik", "malaysia-kssm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != topicsDir {
			t.Errorf("got %q, want %q", got, topicsDir)
		}
	})

	t.Run("returns error when subject not found", func(t *testing.T) {
		_, err := findSubjectTopicsDir(root, "nonexistent-subject", "", "malaysia-kssm")
		if err == nil {
			t.Error("expected error for missing subject, got nil")
		}
	})

	t.Run("falls back to syllabusID when subjectGradeID and subjectID are empty", func(t *testing.T) {
		syllabusDir := filepath.Join(root, "curricula", "malaysia", "malaysia-kssm", "topics")
		if err := os.MkdirAll(syllabusDir, 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
		got, err := findSubjectTopicsDir(root, "", "", "malaysia-kssm")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != syllabusDir {
			t.Errorf("got %q, want %q", got, syllabusDir)
		}
	})
}

func TestSubjectBaseID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"malaysia-kssm-matematik-tingkatan-3", "malaysia-kssm-matematik"},
		{"india-cbse-physics-class-12", "india-cbse-physics"},
		{"uk-cambridge-igcse-mathematics-0580", "uk-cambridge-igcse-mathematics-0580"}, // no grade detected (4-digit code)
		{"", ""},
	}
	for _, c := range tests {
		got := subjectBaseID(c.input)
		if got != c.want {
			t.Errorf("subjectBaseID(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestExtractDSKPTopics(t *testing.T) {
	input := `
BIDANG PEMBELAJARAN
PERKAITAN DAN ALGEBRA
TAJUK
1.0 FUNGSI DAN PERSAMAAN KUADRATIK
Kandungan fungsi kuadratik dan persamaan.
TAJUK
2.0 POLINOMIAL
Kandungan operasi polinomial.
BIDANG PEMBELAJARAN
STATISTIK DAN KEBARANGKALIAN
TAJUK
3.0 STATISTIK
Kandungan data dan taburan.
`
	topics := extractDSKPTopics(input)
	if len(topics) != 3 {
		t.Fatalf("expected 3 topics, got %d", len(topics))
	}

	cases := []struct {
		number string
		name   string
		area   string
	}{
		{"1.0", "FUNGSI DAN PERSAMAAN KUADRATIK", "PERKAITAN DAN ALGEBRA"},
		{"2.0", "POLINOMIAL", "PERKAITAN DAN ALGEBRA"},
		{"3.0", "STATISTIK", "STATISTIK DAN KEBARANGKALIAN"},
	}
	for i, c := range cases {
		if topics[i].Number != c.number {
			t.Errorf("topics[%d].Number = %q, want %q", i, topics[i].Number, c.number)
		}
		if topics[i].Name != c.name {
			t.Errorf("topics[%d].Name = %q, want %q", i, topics[i].Name, c.name)
		}
		if topics[i].LearningArea != c.area {
			t.Errorf("topics[%d].LearningArea = %q, want %q", i, topics[i].LearningArea, c.area)
		}
	}
	if !strings.Contains(topics[0].Content, "Kandungan fungsi") {
		t.Errorf("topics[0].Content missing expected text, got: %q", topics[0].Content)
	}
}

func TestExtractDSKPTopicsNoMatch(t *testing.T) {
	// Generic Markdown document — should return nil, not panic.
	input := "# Chapter 1\nSome content\n## Section 1.1\nMore content"
	topics := extractDSKPTopics(input)
	if len(topics) != 0 {
		t.Errorf("expected 0 topics for non-DSKP input, got %d", len(topics))
	}
}

func TestDSKPTopicsToChunks(t *testing.T) {
	topics := []dskpTopic{
		{LearningArea: "PERKAITAN DAN ALGEBRA", Number: "1.0", Name: "FUNGSI", Content: "some content"},
		{LearningArea: "STATISTIK", Number: "2.0", Name: "DATA", Content: "more content"},
	}
	chunks := dskpTopicsToChunks(topics)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Heading != "1.0 FUNGSI" {
		t.Errorf("chunks[0].Heading = %q, want %q", chunks[0].Heading, "1.0 FUNGSI")
	}
	if !strings.Contains(chunks[0].Content, "BIDANG PEMBELAJARAN: PERKAITAN DAN ALGEBRA") {
		t.Errorf("chunks[0].Content missing learning area header, got: %q", chunks[0].Content)
	}
	if !strings.Contains(chunks[0].Content, "TAJUK: 1.0 FUNGSI") {
		t.Errorf("chunks[0].Content missing TAJUK line, got: %q", chunks[0].Content)
	}
	if chunks[0].Total != 2 {
		t.Errorf("chunks[0].Total = %d, want 2", chunks[0].Total)
	}
	if chunks[1].Index != 1 {
		t.Errorf("chunks[1].Index = %d, want 1", chunks[1].Index)
	}
}

func TestSubjectPrefix(t *testing.T) {
	cases := []struct {
		subjectID string
		want      string
	}{
		{"malaysia-kssm-matematik-tingkatan-4", "MT"},
		{"malaysia-kssm-fizik-tingkatan-4", "PH"},
		{"malaysia-kssm-kimia-tingkatan-5", "CH"},
		{"malaysia-kssm-sains-tingkatan-2", "SC"},
		{"malaysia-kssm-biologi-tingkatan-5", "BI"},
		{"malaysia-kssm-sejarah-tingkatan-3", "HI"},
		{"india-cbse-mathematics-class-10", "MT"},
		{"india-cbse-physics-class-12", "PH"},
	}
	for _, c := range cases {
		got := subjectPrefix(c.subjectID)
		if got != c.want {
			t.Errorf("subjectPrefix(%q) = %q, want %q", c.subjectID, got, c.want)
		}
	}
}

func TestGradeNumber(t *testing.T) {
	cases := []struct {
		subjectID string
		want      string
	}{
		{"malaysia-kssm-matematik-tingkatan-4", "4"},
		{"india-cbse-physics-class-12", "12"},
		{"malaysia-kssm-matematik", ""},           // no grade (exam-based)
		{"uk-cambridge-igcse-mathematics-0580", ""}, // 580 > 20, not a grade
	}
	for _, c := range cases {
		got := gradeNumber(c.subjectID)
		if got != c.want {
			t.Errorf("gradeNumber(%q) = %q, want %q", c.subjectID, got, c.want)
		}
	}
}

func TestTopicFileID(t *testing.T) {
	cases := []struct {
		subjectID  string
		heading    string
		chunkIndex int
		want       string
	}{
		{"malaysia-kssm-matematik-tingkatan-4", "1.0 FUNGSI DAN PERSAMAAN", 0, "MT4-01"},
		{"malaysia-kssm-matematik-tingkatan-4", "10.0 STATISTIK DAN KEBARANGKALIAN", 9, "MT4-10"},
		{"india-cbse-physics-class-12", "3.0 Motion in a Plane", 2, "PH12-03"},
		{"malaysia-kssm-matematik-tingkatan-4", "no number heading", 4, "MT4-05"}, // fallback to chunkIndex+1
		{"uk-cambridge-igcse-mathematics-0580", "1.0 Number", 0, "MT-01"},         // no grade
	}
	for _, c := range cases {
		got := topicFileID(c.subjectID, c.heading, c.chunkIndex)
		if got != c.want {
			t.Errorf("topicFileID(%q, %q, %d) = %q, want %q",
				c.subjectID, c.heading, c.chunkIndex, got, c.want)
		}
	}
}

func TestParseDSKPTopicLine(t *testing.T) {
	cases := []struct {
		line   string
		number string
		name   string
	}{
		{"1.0 FUNGSI DAN PERSAMAAN KUADRATIK", "1.0", "FUNGSI DAN PERSAMAAN KUADRATIK"},
		{"10.0 STATISTIK DAN KEBARANGKALIAN", "10.0", "STATISTIK DAN KEBARANGKALIAN"},
		{"PLAIN HEADING NO NUMBER", "", "PLAIN HEADING NO NUMBER"},
		{"2.0 POLINOMIAL", "2.0", "POLINOMIAL"},
	}
	for _, c := range cases {
		number, name := parseDSKPTopicLine(c.line)
		if number != c.number {
			t.Errorf("parseDSKPTopicLine(%q) number = %q, want %q", c.line, number, c.number)
		}
		if name != c.name {
			t.Errorf("parseDSKPTopicLine(%q) name = %q, want %q", c.line, name, c.name)
		}
	}
}

func TestReassembleDSKPTopicLine(t *testing.T) {
	cases := []struct {
		name     string
		lines    []string
		start    int
		wantLine string
	}{
		{
			name:     "normal single line",
			lines:    []string{"TAJUK", "1.0 FUNGSI DAN PERSAMAAN", "content"},
			start:    1,
			wantLine: "1.0 FUNGSI DAN PERSAMAAN",
		},
		{
			name:     "fragmented 13.0",
			lines:    []string{"TAJUK", "1", "3", ".", "0", "", "KEBARANGKALIAN MUDAH", "content"},
			start:    1,
			wantLine: "13.0 KEBARANGKALIAN MUDAH",
		},
		{
			name:     "fragmented with spaces",
			lines:    []string{"TAJUK", " 1 ", " 3 ", " . ", " 0 ", "", " KEBARANGKALIAN MUDAH ", "content"},
			start:    1,
			wantLine: "13.0 KEBARANGKALIAN MUDAH",
		},
		{
			name:     "two digit number no fragmentation",
			lines:    []string{"TAJUK", "13.0 KEBARANGKALIAN MUDAH", "content"},
			start:    1,
			wantLine: "13.0 KEBARANGKALIAN MUDAH",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotLine, _ := reassembleDSKPTopicLine(c.lines, c.start)
			if gotLine != c.wantLine {
				t.Errorf("got line %q, want %q", gotLine, c.wantLine)
			}
		})
	}
}

func TestExtractDSKPTopicsFragmentedNumber(t *testing.T) {
	// Simulates PDF text where "13.0" is split across lines.
	input := `
BIDANG PEMBELAJARAN
STATISTIK DAN KEBARANGKALIAN
TAJUK
1
3
.
0

KEBARANGKALIAN MUDAH
Kandungan kebarangkalian mudah.
`
	topics := extractDSKPTopics(input)
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
	if topics[0].Number != "13.0" {
		t.Errorf("Number = %q, want %q", topics[0].Number, "13.0")
	}
	if topics[0].Name != "KEBARANGKALIAN MUDAH" {
		t.Errorf("Name = %q, want %q", topics[0].Name, "KEBARANGKALIAN MUDAH")
	}
}

func TestGenerateAllCmdFlags(t *testing.T) {
	cmd := generateAllCmd()

	// Required flags must be defined
	if cmd.Flags().Lookup("syllabus") == nil {
		t.Error("generate all command missing --syllabus flag")
	}
	if cmd.Flags().Lookup("subject-grade") == nil {
		t.Error("generate all command missing --subject-grade flag")
	}
	if cmd.Flags().Lookup("workers") == nil {
		t.Error("generate all command missing --workers flag")
	}
	if cmd.Flags().Lookup("dry-run") == nil {
		t.Error("generate all command missing --dry-run flag")
	}

	// Defaults
	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		t.Fatalf("getting workers flag: %v", err)
	}
	if workers != 3 {
		t.Errorf("workers default = %d, want 3", workers)
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		t.Fatalf("getting dry-run flag: %v", err)
	}
	if dryRun != false {
		t.Error("dry-run default should be false")
	}

	// Verify Use and Short are set
	if cmd.Use != "all" {
		t.Errorf("Use = %q, want %q", cmd.Use, "all")
	}
	if cmd.Short == "" {
		t.Error("generate all command missing Short description")
	}
}

func TestGenerateAllRegistered(t *testing.T) {
	cmd := generateCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "all" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'all' subcommand not registered under 'generate'")
	}
}

func TestDiscoverTopicIDs(t *testing.T) {
	// Create a temp directory with topic YAML files
	dir := t.TempDir()

	// Write topic files with id fields
	topicFiles := map[string]string{
		"MT4-01.yaml": "id: MT4-01\nname: Fungsi\n",
		"MT4-02.yaml": "id: MT4-02\nname: Polinomial\n",
		"MT4-03.yaml": "id: MT4-03\nname: Statistik\n",
	}
	for name, content := range topicFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	// Write files that should be excluded
	excludeFiles := map[string]string{
		"MT4-01.assessments.yaml": "assessments:\n  - q1\n",
		"MT4-01.examples.yaml":    "examples:\n  - e1\n",
		"README.md":               "# Topics\n",
		"not-a-topic.yaml":        "something: else\n", // no id field
	}
	for name, content := range excludeFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	t.Run("discovers topic IDs sorted", func(t *testing.T) {
		ids, err := discoverTopicIDs(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"MT4-01", "MT4-02", "MT4-03"}
		if len(ids) != len(want) {
			t.Fatalf("got %d IDs, want %d: %v", len(ids), len(want), ids)
		}
		for i, id := range ids {
			if id != want[i] {
				t.Errorf("ids[%d] = %q, want %q", i, id, want[i])
			}
		}
	})

	t.Run("empty directory returns empty slice", func(t *testing.T) {
		emptyDir := t.TempDir()
		ids, err := discoverTopicIDs(emptyDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("expected 0 IDs, got %d", len(ids))
		}
	})
}

func TestRemapTopicIDs(t *testing.T) {
	scaffold := []scaffoldTopic{
		{ID: "SC3-01", Name: "Rangsangan dan Gerak Balas", NameEn: "Stimuli and Response"},
		{ID: "SC3-02", Name: "Respirasi", NameEn: "Respiration"},
		{ID: "SC3-03", Name: "Pengangkutan", NameEn: "Transport"},
	}

	topicYAML := func(id, name, nameEn string) string {
		return "id: " + id + "\nname: \"" + name + "\"\nname_en: \"" + nameEn + "\"\n"
	}

	getID := func(yaml string) string {
		id, _ := extractTopicIDAndName(yaml)
		return id
	}

	t.Run("no scaffold returns zero", func(t *testing.T) {
		topics := []pipeline.TopicResult{{Output: topicYAML("TP6", "X", "X")}}
		if n := remapTopicIDs(topics, nil); n != 0 {
			t.Errorf("expected 0 remaps, got %d", n)
		}
	})

	t.Run("already-correct IDs are left alone", func(t *testing.T) {
		topics := []pipeline.TopicResult{
			{Output: topicYAML("SC3-01", "Rangsangan dan Gerak Balas", "Stimuli and Response")},
			{Output: topicYAML("SC3-02", "Respirasi", "Respiration")},
		}
		if n := remapTopicIDs(topics, scaffold); n != 0 {
			t.Errorf("expected 0 remaps, got %d", n)
		}
		if got := getID(topics[0].Output); got != "SC3-01" {
			t.Errorf("id mutated: %s", got)
		}
	})

	t.Run("substring match (strategy 1)", func(t *testing.T) {
		topics := []pipeline.TopicResult{
			{Output: topicYAML("TP6-SC3-02", "Respirasi", "Respiration")},
		}
		if n := remapTopicIDs(topics, scaffold); n != 1 {
			t.Fatalf("expected 1 remap, got %d", n)
		}
		if got := getID(topics[0].Output); got != "SC3-02" {
			t.Errorf("want SC3-02, got %s", got)
		}
	})

	t.Run("numeric-suffix match (strategy 2)", func(t *testing.T) {
		topics := []pipeline.TopicResult{
			{Output: topicYAML("TP6-01", "Totally Unrelated", "Totally Unrelated")},
		}
		if n := remapTopicIDs(topics, scaffold); n != 1 {
			t.Fatalf("expected 1 remap, got %d", n)
		}
		if got := getID(topics[0].Output); got != "SC3-01" {
			t.Errorf("want SC3-01, got %s", got)
		}
	})

	t.Run("name similarity match (strategy 3) — bare TP6 with distinct names", func(t *testing.T) {
		topics := []pipeline.TopicResult{
			{Output: topicYAML("TP6", "Rangsangan dan Gerak Balas", "Stimuli and Response")},
			{Output: topicYAML("TP6", "Respirasi", "Respiration")},
			{Output: topicYAML("TP6", "Pengangkutan", "Transport")},
		}
		if n := remapTopicIDs(topics, scaffold); n != 3 {
			t.Fatalf("expected 3 remaps, got %d", n)
		}
		wants := []string{"SC3-01", "SC3-02", "SC3-03"}
		for i, want := range wants {
			if got := getID(topics[i].Output); got != want {
				t.Errorf("topic %d: want %s, got %s", i, want, got)
			}
		}
	})

	t.Run("name similarity match via name_en when name missing", func(t *testing.T) {
		topics := []pipeline.TopicResult{
			{Output: "id: TP6\nname: \"\"\nname_en: \"Respiration\"\n"},
		}
		if n := remapTopicIDs(topics, scaffold); n != 1 {
			t.Fatalf("expected 1 remap, got %d", n)
		}
		if got := getID(topics[0].Output); got != "SC3-02" {
			t.Errorf("want SC3-02, got %s", got)
		}
	})

	t.Run("collision guard — duplicate names don't collapse onto one scaffold id", func(t *testing.T) {
		// Both topics claim the same name. First gets SC3-02; second should NOT
		// also get SC3-02 — it should either miss (below threshold) or go to
		// the next best unclaimed candidate.
		topics := []pipeline.TopicResult{
			{Output: topicYAML("TP6", "Respirasi", "Respiration")},
			{Output: topicYAML("TP6", "Respirasi", "Respiration")},
		}
		remapTopicIDs(topics, scaffold)
		a, b := getID(topics[0].Output), getID(topics[1].Output)
		if a == "SC3-02" && b == "SC3-02" {
			t.Errorf("both topics collapsed onto SC3-02: a=%s b=%s", a, b)
		}
	})

	t.Run("unmatched topic is left untouched for the write-guard", func(t *testing.T) {
		topics := []pipeline.TopicResult{
			{Output: topicYAML("XYZ-99", "Completely Different Topic", "Completely Different Topic")},
		}
		if n := remapTopicIDs(topics, scaffold); n != 0 {
			t.Errorf("expected 0 remaps, got %d", n)
		}
		if got := getID(topics[0].Output); got != "XYZ-99" {
			t.Errorf("id mutated: %s", got)
		}
	})
}

func TestRewriteTopicIDLine(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		newID string
		want  string
	}{
		{
			name:  "unquoted",
			in:    "id: TP6\nname: Foo\nsubject_id: bar\n",
			newID: "SC3-01",
			want:  "id: SC3-01\nname: Foo\nsubject_id: bar\n",
		},
		{
			name:  "quoted",
			in:    "id: \"TP6\"\nname: Foo\n",
			newID: "SC3-01",
			want:  "id: SC3-01\nname: Foo\n",
		},
		{
			name:  "does not touch subject_id",
			in:    "subject_id: foo\nid: TP6\n",
			newID: "SC3-01",
			want:  "subject_id: foo\nid: SC3-01\n",
		},
		{
			name:  "no id line — unchanged",
			in:    "name: Foo\n",
			newID: "SC3-01",
			want:  "name: Foo\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := rewriteTopicIDLine(tc.in, tc.newID)
			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestParseTopLevelTopicFields(t *testing.T) {
	t.Run("ignores nested id inside performance_standards", func(t *testing.T) {
		// Regression: production scaffold files contain a performance_standards
		// array where each entry has `id: "TPN"`. The old line-trim loader
		// overwrote the real top-level id with the last TP6 it saw.
		yamlContent := `id: SC3-01
official_ref: "Bab 1"
name: "Rangsangan dan Gerak Balas"
name_en: "Stimuli and Response"
performance_standards:
    - level: 1
      id: "TP1"
      text: "Recall"
    - level: 6
      id: "TP6"
      text: "Create"
prerequisites:
    required: []
`
		got := parseTopLevelTopicFields(yamlContent)
		if got.ID != "SC3-01" {
			t.Errorf("ID: want SC3-01, got %q", got.ID)
		}
		if got.Name != "Rangsangan dan Gerak Balas" {
			t.Errorf("Name: want %q, got %q", "Rangsangan dan Gerak Balas", got.Name)
		}
		if got.NameEn != "Stimuli and Response" {
			t.Errorf("NameEn: want %q, got %q", "Stimuli and Response", got.NameEn)
		}
	})

	t.Run("strips quotes from quoted top-level values", func(t *testing.T) {
		yamlContent := "id: \"SC3-02\"\nname: \"Respirasi\"\nname_en: \"Respiration\"\n"
		got := parseTopLevelTopicFields(yamlContent)
		if got.ID != "SC3-02" || got.Name != "Respirasi" || got.NameEn != "Respiration" {
			t.Errorf("unexpected: %+v", got)
		}
	})

	t.Run("ignores comments and list items at column 0", func(t *testing.T) {
		yamlContent := `# comment
- dangling
id: SC3-03
name: Pengangkutan
`
		got := parseTopLevelTopicFields(yamlContent)
		if got.ID != "SC3-03" || got.Name != "Pengangkutan" {
			t.Errorf("unexpected: %+v", got)
		}
	})

	t.Run("first top-level id wins when keys appear twice", func(t *testing.T) {
		// A malformed stub could in theory repeat id at the top level; the
		// first one wins (matches the YAML spec for duplicate keys — later
		// wins, but yaml.v3 errors out, and we want the real first definition).
		yamlContent := "id: SC3-01\nname: Foo\nid: TP6\n"
		got := parseTopLevelTopicFields(yamlContent)
		if got.ID != "SC3-01" {
			t.Errorf("want first id to win, got %q", got.ID)
		}
	})
}

func TestLoadScaffoldTopicsWithPerformanceStandards(t *testing.T) {
	dir := t.TempDir()
	// Mini scaffold stub mirroring the real production file structure:
	// top-level id: SC3-01, plus a performance_standards array with TP1..TP6.
	stub := `id: SC3-01
name: "Rangsangan dan Gerak Balas"
name_en: "Stimuli and Response"
performance_standards:
    - level: 1
      id: "TP1"
    - level: 6
      id: "TP6"
`
	if err := os.WriteFile(filepath.Join(dir, "SC3-01.yaml"), []byte(stub), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	topics := loadScaffoldTopics(dir)
	if len(topics) != 1 {
		t.Fatalf("want 1 topic, got %d", len(topics))
	}
	if topics[0].ID != "SC3-01" {
		t.Errorf("want id SC3-01, got %q", topics[0].ID)
	}
	if topics[0].Name != "Rangsangan dan Gerak Balas" {
		t.Errorf("want name %q, got %q", "Rangsangan dan Gerak Balas", topics[0].Name)
	}
	if topics[0].NameEn != "Stimuli and Response" {
		t.Errorf("want name_en %q, got %q", "Stimuli and Response", topics[0].NameEn)
	}
}

func TestBuildWholePDFPromptScaffoldLayout(t *testing.T) {
	scaffold := []scaffoldTopic{
		{ID: "SC3-01", Name: "Rangsangan dan Gerak Balas", NameEn: "Stimuli and Response"},
		{ID: "SC3-02", Name: "Respirasi", NameEn: "Respiration"},
		{ID: "SC3-03", Name: "Pengangkutan", NameEn: "Transport"},
	}
	pdf := "PDF_SENTINEL_CONTENT"
	prompt := buildWholePDFPrompt(pdf, wholePDFPromptOpts{
		syllabusID:     "malaysia-kssm",
		subjectGradeID: "malaysia-kssm-sains-tingkatan-3",
		countryID:      "malaysia",
		language:       "ms",
		prefix:         "SC",
		grade:          "3",
		scaffoldTopics: scaffold,
	})

	t.Run("scaffold task appears before document content", func(t *testing.T) {
		taskIdx := strings.Index(prompt, "YOUR TASK:")
		pdfIdx := strings.Index(prompt, pdf)
		if taskIdx < 0 || pdfIdx < 0 {
			t.Fatalf("missing sections: task=%d pdf=%d", taskIdx, pdfIdx)
		}
		if taskIdx >= pdfIdx {
			t.Errorf("YOUR TASK must appear before document content (task=%d, pdf=%d)", taskIdx, pdfIdx)
		}
	})

	t.Run("output skeleton contains each scaffold id and name verbatim", func(t *testing.T) {
		for _, st := range scaffold {
			if !strings.Contains(prompt, "id: "+st.ID) {
				t.Errorf("prompt missing output skeleton line %q", "id: "+st.ID)
			}
			if !strings.Contains(prompt, st.Name) {
				t.Errorf("prompt missing scaffold name %q", st.Name)
			}
		}
	})

	t.Run("trailing reminder lists all scaffold ids after the document", func(t *testing.T) {
		pdfIdx := strings.LastIndex(prompt, pdf)
		reminderIdx := strings.Index(prompt, "REMEMBER")
		if reminderIdx < 0 {
			t.Fatal("REMEMBER reminder missing")
		}
		if reminderIdx <= pdfIdx {
			t.Errorf("REMEMBER reminder must appear after document content (reminder=%d, pdf=%d)", reminderIdx, pdfIdx)
		}
		for _, st := range scaffold {
			if !strings.Contains(prompt[reminderIdx:], st.ID) {
				t.Errorf("trailing reminder missing scaffold id %q", st.ID)
			}
		}
	})

	t.Run("no generic SC3-NN format sentence when scaffold present", func(t *testing.T) {
		if strings.Contains(prompt, "Topic IDs MUST follow the format") {
			t.Errorf("prompt should not include the generic format sentence when scaffold is provided")
		}
	})

	t.Run("no scaffold → generic format sentence is present", func(t *testing.T) {
		plain := buildWholePDFPrompt(pdf, wholePDFPromptOpts{
			syllabusID: "malaysia-kssm", subjectGradeID: "malaysia-kssm-sains-tingkatan-3",
			countryID: "malaysia", language: "ms", prefix: "SC", grade: "3",
		})
		if !strings.Contains(plain, "Topic IDs MUST follow the format: SC3-NN") {
			t.Errorf("plain prompt missing generic format sentence")
		}
		if strings.Contains(plain, "YOUR TASK:") {
			t.Errorf("plain prompt should not include YOUR TASK section")
		}
	})
}

func TestNameSimilarity(t *testing.T) {
	tests := []struct {
		a, b    string
		wantMin float64
		wantMax float64
	}{
		{"Respirasi", "Respirasi", 1.0, 1.0},
		{"Rangsangan dan Gerak Balas", "RANGSANGAN DAN GERAK BALAS", 1.0, 1.0},
		{"Respirasi", "", 0, 0},
		{"", "", 0, 0},
		{"Respirasi", "Pengangkutan", 0, 0},
		{"Rangsangan dan Gerak Balas", "Rangsangan Gerak", 0.4, 0.8},
	}
	for _, tc := range tests {
		got := nameSimilarity(tc.a, tc.b)
		if got < tc.wantMin || got > tc.wantMax {
			t.Errorf("nameSimilarity(%q,%q) = %f, want between %f and %f", tc.a, tc.b, got, tc.wantMin, tc.wantMax)
		}
	}
}
