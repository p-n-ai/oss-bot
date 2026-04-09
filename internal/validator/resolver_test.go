package validator

import (
	"os"
	"path/filepath"
	"testing"
)

// setupResolverTestDirs creates a temporary directory structure:
//
//	tmp/
//	  schema/                          <- global schemas
//	    assessments.schema.json
//	    topic.schema.json
//	  curricula/country/syllabus/subject/
//	    subject.yaml
//	    schemas/                        <- subject-level overrides
//	      assessments.schema.json      <- different from global
//	    grade/
//	      subject-grade.yaml
//	      topics/
//	        MT1-01.yaml
func setupResolverTestDirs(t *testing.T) (root string) {
	t.Helper()
	root = t.TempDir()

	// Global schemas
	globalDir := filepath.Join(root, "schema")
	os.MkdirAll(globalDir, 0o755)
	os.WriteFile(filepath.Join(globalDir, "assessments.schema.json"), []byte(`{"global":"assessments"}`), 0o644)
	os.WriteFile(filepath.Join(globalDir, "topic.schema.json"), []byte(`{"global":"topic"}`), 0o644)

	// Subject directory with subject.yaml
	subjectDir := filepath.Join(root, "curricula", "country", "syllabus", "subject")
	os.MkdirAll(subjectDir, 0o755)
	os.WriteFile(filepath.Join(subjectDir, "subject.yaml"), []byte("id: subject\n"), 0o644)

	// Subject-level schema override (only assessments)
	subjectSchemas := filepath.Join(subjectDir, "schemas")
	os.MkdirAll(subjectSchemas, 0o755)
	os.WriteFile(filepath.Join(subjectSchemas, "assessments.schema.json"), []byte(`{"subject":"assessments"}`), 0o644)

	// Subject-grade + topics
	topicsDir := filepath.Join(subjectDir, "grade", "topics")
	os.MkdirAll(topicsDir, 0o755)
	os.WriteFile(filepath.Join(subjectDir, "grade", "subject-grade.yaml"), []byte("id: grade\n"), 0o644)
	os.WriteFile(filepath.Join(topicsDir, "MT1-01.yaml"), []byte("id: MT1-01\n"), 0o644)

	return root
}

func TestFindSubjectDir(t *testing.T) {
	root := setupResolverTestDirs(t)
	subjectDir := filepath.Join(root, "curricula", "country", "syllabus", "subject")

	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "from topic file (3 levels up)",
			filePath: filepath.Join(subjectDir, "grade", "topics", "MT1-01.yaml"),
			want:     subjectDir,
		},
		{
			name:     "from subject-grade file (1 level up)",
			filePath: filepath.Join(subjectDir, "grade", "subject-grade.yaml"),
			want:     subjectDir,
		},
		{
			name:     "from subject.yaml itself",
			filePath: filepath.Join(subjectDir, "subject.yaml"),
			want:     subjectDir,
		},
		{
			name:     "not found — root level file",
			filePath: filepath.Join(root, "some-file.yaml"),
			want:     "",
		},
		{
			name:     "not found — nonexistent path",
			filePath: "/nonexistent/path/file.yaml",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSubjectDir(tt.filePath)
			if got != tt.want {
				t.Errorf("FindSubjectDir(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestSubjectSchemasDir(t *testing.T) {
	root := setupResolverTestDirs(t)
	subjectDir := filepath.Join(root, "curricula", "country", "syllabus", "subject")

	tests := []struct {
		name       string
		subjectDir string
		wantEmpty  bool
	}{
		{
			name:       "schemas dir exists",
			subjectDir: subjectDir,
			wantEmpty:  false,
		},
		{
			name:       "no schemas dir",
			subjectDir: filepath.Join(root, "curricula", "country"),
			wantEmpty:  true,
		},
		{
			name:       "empty string",
			subjectDir: "",
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubjectSchemasDir(tt.subjectDir)
			if tt.wantEmpty && got != "" {
				t.Errorf("SubjectSchemasDir(%q) = %q, want empty", tt.subjectDir, got)
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("SubjectSchemasDir(%q) = empty, want non-empty", tt.subjectDir)
			}
		})
	}
}

func TestResolveSchemaPath(t *testing.T) {
	root := setupResolverTestDirs(t)
	globalDir := filepath.Join(root, "schema")
	subjectSchemasDir := filepath.Join(root, "curricula", "country", "syllabus", "subject", "schemas")

	resolver := NewSchemaResolver(globalDir)

	tests := []struct {
		name              string
		schemaType        string
		subjectSchemasDir string
		wantPath          string
		wantFound         bool
	}{
		{
			name:              "subject override exists",
			schemaType:        "assessments",
			subjectSchemasDir: subjectSchemasDir,
			wantPath:          filepath.Join(subjectSchemasDir, "assessments.schema.json"),
			wantFound:         true,
		},
		{
			name:              "global fallback (no subject override)",
			schemaType:        "topic",
			subjectSchemasDir: subjectSchemasDir,
			wantPath:          filepath.Join(globalDir, "topic.schema.json"),
			wantFound:         true,
		},
		{
			name:              "not found anywhere",
			schemaType:        "concept",
			subjectSchemasDir: subjectSchemasDir,
			wantPath:          "",
			wantFound:         false,
		},
		{
			name:              "empty subject schemas dir — global fallback",
			schemaType:        "topic",
			subjectSchemasDir: "",
			wantPath:          filepath.Join(globalDir, "topic.schema.json"),
			wantFound:         true,
		},
		{
			name:              "nonexistent subject schemas dir — global fallback",
			schemaType:        "assessments",
			subjectSchemasDir: "/nonexistent/schemas",
			wantPath:          filepath.Join(globalDir, "assessments.schema.json"),
			wantFound:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotFound := resolver.ResolveSchemaPath(tt.schemaType, tt.subjectSchemasDir)
			if gotFound != tt.wantFound {
				t.Errorf("ResolveSchemaPath(%q, %q) found = %v, want %v", tt.schemaType, tt.subjectSchemasDir, gotFound, tt.wantFound)
			}
			if gotPath != tt.wantPath {
				t.Errorf("ResolveSchemaPath(%q, %q) path = %q, want %q", tt.schemaType, tt.subjectSchemasDir, gotPath, tt.wantPath)
			}
		})
	}
}

func TestResolveSchemaPath_PerFileMixed(t *testing.T) {
	root := setupResolverTestDirs(t)
	globalDir := filepath.Join(root, "schema")
	subjectSchemasDir := filepath.Join(root, "curricula", "country", "syllabus", "subject", "schemas")

	resolver := NewSchemaResolver(globalDir)

	// assessments should come from subject
	aPath, aFound := resolver.ResolveSchemaPath("assessments", subjectSchemasDir)
	if !aFound || aPath != filepath.Join(subjectSchemasDir, "assessments.schema.json") {
		t.Errorf("assessments should resolve to subject schema, got %q (found=%v)", aPath, aFound)
	}

	// topic should fall back to global
	tPath, tFound := resolver.ResolveSchemaPath("topic", subjectSchemasDir)
	if !tFound || tPath != filepath.Join(globalDir, "topic.schema.json") {
		t.Errorf("topic should resolve to global schema, got %q (found=%v)", tPath, tFound)
	}
}
