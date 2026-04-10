package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractSchemaDescriptions parses a JSON Schema and extracts all field
// descriptions into a human-readable guide. This helps AI models understand
// what data each field expects, beyond just the structural/type constraints
// in the raw schema.
//
// Example output:
//
//	- topic_id (string, required): The unique identifier for this topic
//	- questions (array, required): List of assessment questions
//	  - id (string, required): Unique question identifier like Q1, Q2
//	  - text (string, required): The question text presented to students
func ExtractSchemaDescriptions(schemaJSON string) string {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return ""
	}

	required := toStringSet(schema["required"])
	props, ok := schema["properties"].(map[string]interface{})
	if !ok || len(props) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Field Descriptions (use these as guidance for what data to generate)\n")

	for _, key := range sortedKeys(props) {
		prop, ok := props[key].(map[string]interface{})
		if !ok {
			continue
		}
		writeFieldDescription(&sb, key, prop, required, 0)
	}

	result := sb.String()
	if strings.Count(result, "\n") <= 1 {
		// No descriptions found.
		return ""
	}
	return result
}

// writeFieldDescription writes a single field's description line and recurses
// into nested objects/arrays.
func writeFieldDescription(sb *strings.Builder, key string, prop map[string]interface{}, parentRequired map[string]bool, depth int) {
	indent := strings.Repeat("  ", depth)
	typeName := jsonSchemaType(prop)
	reqLabel := ""
	if parentRequired[key] {
		reqLabel = ", required"
	}

	desc, _ := prop["description"].(string)
	enumVals := formatEnum(prop)

	if desc != "" || enumVals != "" {
		sb.WriteString(fmt.Sprintf("%s- %s (%s%s)", indent, key, typeName, reqLabel))
		if desc != "" {
			sb.WriteString(": " + desc)
		}
		if enumVals != "" {
			if desc != "" {
				sb.WriteString(". ")
			} else {
				sb.WriteString(": ")
			}
			sb.WriteString("Values: " + enumVals)
		}
		sb.WriteString("\n")
	} else if depth == 0 || parentRequired[key] {
		// Always show required or top-level fields even without descriptions.
		sb.WriteString(fmt.Sprintf("%s- %s (%s%s)\n", indent, key, typeName, reqLabel))
	}

	// Recurse into object properties.
	if nested, ok := prop["properties"].(map[string]interface{}); ok {
		nestedReq := toStringSet(prop["required"])
		for _, nk := range sortedKeys(nested) {
			np, ok := nested[nk].(map[string]interface{})
			if !ok {
				continue
			}
			writeFieldDescription(sb, nk, np, nestedReq, depth+1)
		}
	}

	// Recurse into array items.
	if items, ok := prop["items"].(map[string]interface{}); ok {
		if itemProps, ok := items["properties"].(map[string]interface{}); ok {
			itemReq := toStringSet(items["required"])
			for _, ik := range sortedKeys(itemProps) {
				ip, ok := itemProps[ik].(map[string]interface{})
				if !ok {
					continue
				}
				writeFieldDescription(sb, ik, ip, itemReq, depth+1)
			}
		} else if itemDesc, ok := items["description"].(string); ok && itemDesc != "" {
			sb.WriteString(fmt.Sprintf("%s  (each item: %s)\n", indent, itemDesc))
		}
	}
}

// jsonSchemaType returns a human-readable type label from a JSON schema property.
func jsonSchemaType(prop map[string]interface{}) string {
	if t, ok := prop["type"].(string); ok {
		return t
	}
	if _, ok := prop["oneOf"]; ok {
		return "oneOf"
	}
	if _, ok := prop["anyOf"]; ok {
		return "anyOf"
	}
	return "any"
}

// formatEnum formats an enum array into a readable string.
func formatEnum(prop map[string]interface{}) string {
	vals, ok := prop["enum"].([]interface{})
	if !ok || len(vals) == 0 {
		return ""
	}
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(parts, " | ")
}

// toStringSet converts a JSON array ([]interface{}) to a set of strings.
func toStringSet(v interface{}) map[string]bool {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	set := make(map[string]bool, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			set[s] = true
		}
	}
	return set
}

// sortedKeys returns the keys of a map in sorted order.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Sort for deterministic output.
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
