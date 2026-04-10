package validator

import (
	"os"
	"path/filepath"
)

// SchemaResolver resolves schema files with per-subject overrides and global fallback.
type SchemaResolver struct {
	globalDir string
}

// NewSchemaResolver creates a resolver rooted at the global schema directory.
func NewSchemaResolver(globalSchemaDir string) *SchemaResolver {
	return &SchemaResolver{globalDir: globalSchemaDir}
}

// GlobalDir returns the global schema directory path.
func (r *SchemaResolver) GlobalDir() string {
	return r.globalDir
}

// ResolveSchemaPath checks subjectSchemasDir first, then globalDir for a schema file.
// Returns the resolved path and true if found, or ("", false) if not found in either location.
func (r *SchemaResolver) ResolveSchemaPath(schemaType, subjectSchemasDir string) (string, bool) {
	fileName := schemaType + ".schema.json"

	// Check subject-level override first
	if subjectSchemasDir != "" {
		subjectPath := filepath.Join(subjectSchemasDir, fileName)
		if _, err := os.Stat(subjectPath); err == nil {
			return subjectPath, true
		}
	}

	// Fall back to global
	globalPath := filepath.Join(r.globalDir, fileName)
	if _, err := os.Stat(globalPath); err == nil {
		return globalPath, true
	}

	return "", false
}

// FindSubjectDir walks up from a YAML file path to find the nearest ancestor
// directory containing "subject.yaml". Returns "" if not found within 5 levels.
func FindSubjectDir(yamlFilePath string) string {
	dir := filepath.Dir(yamlFilePath)

	for i := 0; i < 5; i++ {
		candidate := filepath.Join(dir, "subject.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return ""
}

// SubjectSchemaDir returns the "schema" subdirectory of the given subject directory,
// or "" if the directory does not exist or has no schema/ subfolder.
func SubjectSchemaDir(subjectDir string) string {
	if subjectDir == "" {
		return ""
	}
	schemaDir := filepath.Join(subjectDir, "schema")
	info, err := os.Stat(schemaDir)
	if err != nil || !info.IsDir() {
		return ""
	}
	return schemaDir
}
