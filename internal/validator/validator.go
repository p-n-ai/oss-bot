// Package validator provides JSON Schema validation for OSS curriculum YAML files.
package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// ValidationResult holds the result of validating a single file.
type ValidationResult struct {
	File   string
	Type   string
	Valid  bool
	Errors []string
}

// Validator validates YAML files against JSON Schemas.
type Validator struct {
	schemas       map[string]*jsonschema.Schema
	resolver      *SchemaResolver                // nil for legacy New() path
	resolvedCache map[string]*jsonschema.Schema   // keyed by abs schema file path
}

// New creates a Validator by loading all schemas from the given directory.
func New(schemaDir string) (*Validator, error) {
	v := &Validator{
		schemas: make(map[string]*jsonschema.Schema),
	}

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("reading schema directory: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".schema.json") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".schema.json")
		path := filepath.Join(schemaDir, entry.Name())

		schema, err := compiler.Compile(path)
		if err != nil {
			return nil, fmt.Errorf("compiling schema %s: %w", name, err)
		}

		v.schemas[name] = schema
	}

	return v, nil
}

// ValidateFile validates a single YAML file against the specified schema type.
func (v *Validator) ValidateFile(filePath, schemaType string) (*ValidationResult, error) {
	result := &ValidationResult{
		File: filePath,
		Type: schemaType,
	}

	schema, ok := v.schemas[schemaType]
	if !ok {
		return nil, fmt.Errorf("unknown schema type: %s (available: %v)", schemaType, v.SchemaTypes())
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Parse YAML to generic interface
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		result.Valid = false
		result.Errors = []string{fmt.Sprintf("YAML parse error: %v", err)}
		return result, nil
	}

	// Convert YAML types to JSON-compatible types
	jsonData := convertYAMLToJSON(yamlData)

	// Validate against schema
	if err := schema.Validate(jsonData); err != nil {
		result.Valid = false
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			result.Errors = flattenValidationErrors(ve)
		} else {
			result.Errors = []string{err.Error()}
		}
		return result, nil
	}

	result.Valid = true
	return result, nil
}

// ValidateDir validates all YAML files in a directory tree, auto-detecting schema types.
func (v *Validator) ValidateDir(dir string) ([]ValidationResult, error) {
	var results []ValidationResult

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		// Skip teaching notes markdown files
		if strings.HasSuffix(path, ".teaching.md") {
			return nil
		}

		schemaType := DetectSchemaType(path)
		if schemaType == "" {
			return nil // Skip files we can't classify
		}

		if _, ok := v.schemas[schemaType]; !ok {
			return nil // Skip if schema not loaded
		}

		result, err := v.ValidateFile(path, schemaType)
		if err != nil {
			results = append(results, ValidationResult{
				File:   path,
				Type:   schemaType,
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return nil
		}
		results = append(results, *result)
		return nil
	})

	return results, err
}

// SchemaTypes returns the list of loaded schema type names.
func (v *Validator) SchemaTypes() []string {
	types := make([]string, 0, len(v.schemas))
	for t := range v.schemas {
		types = append(types, t)
	}
	return types
}

// DetectSchemaType determines the schema type from a file path.
func DetectSchemaType(path string) string {
	base := filepath.Base(path)

	switch {
	case base == "syllabus.yaml" || base == "syllabus.yml":
		return "syllabus"
	case strings.HasSuffix(base, ".assessments.yaml"):
		return "assessments"
	case strings.HasSuffix(base, ".examples.yaml"):
		return "examples"
	case base == "subject.yaml" || base == "subject.yml":
		return "subject"
	case base == "subject-grade.yaml" || base == "subject-grade.yml":
		return "subject_grade"
	case strings.Contains(path, "subjects/") || strings.Contains(path, "subjects\\"):
		return "subject"
	case (strings.Contains(path, "topics/") || strings.Contains(path, "topics\\")) &&
		(strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".yml")) &&
		!strings.Contains(base, ".assessments.") &&
		!strings.Contains(base, ".examples.") &&
		!strings.Contains(base, ".teaching."):
		return "topic"
	case strings.Contains(path, "concepts/"):
		return "concept"
	case strings.Contains(path, "taxonomy/"):
		return "taxonomy"
	default:
		return ""
	}
}

// convertYAMLToJSON converts YAML-parsed data to JSON-compatible types.
// YAML parses integers as int, but JSON Schema expects float64.
// YAML parses maps as map[string]interface{}, which is already JSON-compatible.
func convertYAMLToJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = convertYAMLToJSON(v)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[fmt.Sprint(k)] = convertYAMLToJSON(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = convertYAMLToJSON(v)
		}
		return result
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	default:
		return val
	}
}

// flattenValidationErrors extracts human-readable error messages from validation errors.
func flattenValidationErrors(ve *jsonschema.ValidationError) []string {
	var errs []string

	if ve.Message != "" {
		location := ve.InstanceLocation
		if location == "" {
			location = "(root)"
		}
		errs = append(errs, fmt.Sprintf("%s: %s", location, ve.Message))
	}

	for _, cause := range ve.Causes {
		errs = append(errs, flattenValidationErrors(cause)...)
	}

	return errs
}

// MarshalYAMLToJSON converts a YAML file's content to a JSON-compatible map.
func MarshalYAMLToJSON(data []byte) (interface{}, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("YAML parse error: %w", err)
	}
	return convertYAMLToJSON(yamlData), nil
}

// PrettyJSON converts an interface to formatted JSON string (for debugging).
func PrettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(b)
}

// NewWithResolver creates a Validator that resolves schemas per-file using the
// given SchemaResolver. Schemas are compiled and cached lazily.
func NewWithResolver(resolver *SchemaResolver) *Validator {
	return &Validator{
		schemas:       make(map[string]*jsonschema.Schema),
		resolver:      resolver,
		resolvedCache: make(map[string]*jsonschema.Schema),
	}
}

// compileAndCache compiles and caches a schema by its absolute file path.
func (v *Validator) compileAndCache(schemaPath string) (*jsonschema.Schema, error) {
	absPath, err := filepath.Abs(schemaPath)
	if err != nil {
		absPath = schemaPath
	}

	if cached, ok := v.resolvedCache[absPath]; ok {
		return cached, nil
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	schema, err := compiler.Compile(absPath)
	if err != nil {
		return nil, fmt.Errorf("compiling schema %s: %w", schemaPath, err)
	}

	v.resolvedCache[absPath] = schema
	return schema, nil
}

// ValidateFileResolved validates a YAML file, resolving the schema based on
// the file's location (subject-level override or global fallback).
func (v *Validator) ValidateFileResolved(filePath, schemaType string) (*ValidationResult, error) {
	if v.resolver == nil {
		return nil, fmt.Errorf("ValidateFileResolved requires a resolver (use NewWithResolver)")
	}

	result := &ValidationResult{
		File: filePath,
		Type: schemaType,
	}

	// Resolve the schema path
	subjectDir := FindSubjectDir(filePath)
	schemasDir := SubjectSchemasDir(subjectDir)
	schemaPath, found := v.resolver.ResolveSchemaPath(schemaType, schemasDir)
	if !found {
		return nil, fmt.Errorf("schema %q not found in subject or global directories", schemaType)
	}

	schema, err := v.compileAndCache(schemaPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return v.validateData(data, schema, result)
}

// ValidateContentResolved validates YAML content (bytes) against a resolved schema.
// Used by the generate pipeline to validate in-memory content without writing temp files.
func (v *Validator) ValidateContentResolved(content []byte, schemaType, subjectSchemasDir string) (*ValidationResult, error) {
	if v.resolver == nil {
		return nil, fmt.Errorf("ValidateContentResolved requires a resolver (use NewWithResolver)")
	}

	result := &ValidationResult{
		Type: schemaType,
	}

	schemaPath, found := v.resolver.ResolveSchemaPath(schemaType, subjectSchemasDir)
	if !found {
		return nil, fmt.Errorf("schema %q not found in subject or global directories", schemaType)
	}

	schema, err := v.compileAndCache(schemaPath)
	if err != nil {
		return nil, err
	}

	return v.validateData(content, schema, result)
}

// ValidateDirResolved walks a directory tree, validating each YAML file
// with per-subject schema resolution.
func (v *Validator) ValidateDirResolved(dir string) ([]ValidationResult, error) {
	if v.resolver == nil {
		return nil, fmt.Errorf("ValidateDirResolved requires a resolver (use NewWithResolver)")
	}

	var results []ValidationResult

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		if strings.HasSuffix(path, ".teaching.md") {
			return nil
		}

		schemaType := DetectSchemaType(path)
		if schemaType == "" {
			return nil
		}

		result, err := v.ValidateFileResolved(path, schemaType)
		if err != nil {
			results = append(results, ValidationResult{
				File:   path,
				Type:   schemaType,
				Valid:  false,
				Errors: []string{err.Error()},
			})
			return nil
		}
		results = append(results, *result)
		return nil
	})

	return results, err
}

// validateData validates parsed YAML data against a compiled schema.
func (v *Validator) validateData(data []byte, schema *jsonschema.Schema, result *ValidationResult) (*ValidationResult, error) {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		result.Valid = false
		result.Errors = []string{fmt.Sprintf("YAML parse error: %v", err)}
		return result, nil
	}

	jsonData := convertYAMLToJSON(yamlData)

	if err := schema.Validate(jsonData); err != nil {
		result.Valid = false
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			result.Errors = flattenValidationErrors(ve)
		} else {
			result.Errors = []string{err.Error()}
		}
		return result, nil
	}

	result.Valid = true
	return result, nil
}
