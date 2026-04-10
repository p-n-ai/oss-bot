package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// schemaInfo holds parsed JSON Schema metadata used for post-processing.
type schemaInfo struct {
	Required   map[string]bool
	Properties map[string]propertyInfo
}

// propertyInfo holds type and nested schema info for a single property.
type propertyInfo struct {
	Type       string
	Default    interface{}
	Enum       []string
	Required   map[string]bool
	Properties map[string]propertyInfo
	ItemType   string // type of array items
	ItemProps  map[string]propertyInfo
	ItemReq    map[string]bool
}

// parseSchemaInfo extracts required fields and type information from a JSON Schema.
func parseSchemaInfo(schemaJSON string) (*schemaInfo, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &raw); err != nil {
		return nil, err
	}

	info := &schemaInfo{
		Required:   toStringSet(raw["required"]),
		Properties: parseProperties(raw),
	}
	return info, nil
}

// parseProperties extracts property info from a schema object.
func parseProperties(schema map[string]interface{}) map[string]propertyInfo {
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}

	result := make(map[string]propertyInfo, len(props))
	for key, val := range props {
		propMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}
		pi := propertyInfo{}
		if t, ok := propMap["type"].(string); ok {
			pi.Type = t
		}
		if d, ok := propMap["default"]; ok {
			pi.Default = d
		}
		if enumArr, ok := propMap["enum"].([]interface{}); ok {
			for _, e := range enumArr {
				if s, ok := e.(string); ok {
					pi.Enum = append(pi.Enum, s)
				}
			}
		}
		// Nested object properties.
		if pi.Type == "object" {
			pi.Required = toStringSet(propMap["required"])
			pi.Properties = parseProperties(propMap)
		}
		// Array item properties.
		if pi.Type == "array" {
			if items, ok := propMap["items"].(map[string]interface{}); ok {
				if it, ok := items["type"].(string); ok {
					pi.ItemType = it
				}
				if pi.ItemType == "object" {
					pi.ItemReq = toStringSet(items["required"])
					pi.ItemProps = parseProperties(items)
				}
			}
		}
		result[key] = pi
	}
	return result
}

// EnforceSchemaRequiredFields checks the generated YAML against the schema's
// required fields and adds any missing ones with sensible placeholder values.
// This ensures AI output always contains every required field.
func EnforceSchemaRequiredFields(content, schemaJSON string) string {
	info, err := parseSchemaInfo(schemaJSON)
	if err != nil || info == nil {
		return content
	}
	if len(info.Required) == 0 {
		return content
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		return content
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return content
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return content
	}

	changed := enforceRequired(mapping, info.Required, info.Properties)
	if !changed {
		return content
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return content
	}
	return string(out)
}

// enforceRequired adds missing required fields to a YAML mapping node.
// Returns true if any fields were added.
func enforceRequired(mapping *yaml.Node, required map[string]bool, props map[string]propertyInfo) bool {
	changed := false
	existing := mappingKeys(mapping)

	for field := range required {
		if existing[field] {
			continue
		}
		pi := props[field]
		valNode := defaultValueNode(pi)
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: field, Tag: "!!str"},
			valNode,
		)
		changed = true
	}

	// Also enforce required fields within nested objects and array items.
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		key := mapping.Content[i].Value
		val := mapping.Content[i+1]
		pi, ok := props[key]
		if !ok {
			continue
		}

		// Nested object: enforce required sub-fields.
		if pi.Type == "object" && val.Kind == yaml.MappingNode && len(pi.Required) > 0 {
			if enforceRequired(val, pi.Required, pi.Properties) {
				changed = true
			}
		}

		// Array: enforce required fields in each item.
		if pi.Type == "array" && val.Kind == yaml.SequenceNode && len(pi.ItemReq) > 0 {
			for _, item := range val.Content {
				if item.Kind == yaml.MappingNode {
					if enforceRequired(item, pi.ItemReq, pi.ItemProps) {
						changed = true
					}
				}
			}
		}
	}

	return changed
}

// defaultValueNode creates a YAML node with a sensible default value for a property.
func defaultValueNode(pi propertyInfo) *yaml.Node {
	// Use schema default if available.
	if pi.Default != nil {
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", pi.Default), Tag: "!!str"}
	}
	// Use first enum value if available.
	if len(pi.Enum) > 0 {
		return &yaml.Node{Kind: yaml.ScalarNode, Value: pi.Enum[0], Tag: "!!str"}
	}

	switch pi.Type {
	case "string":
		return &yaml.Node{Kind: yaml.ScalarNode, Value: "", Tag: "!!str", Style: yaml.DoubleQuotedStyle}
	case "integer", "number":
		return &yaml.Node{Kind: yaml.ScalarNode, Value: "0", Tag: "!!int"}
	case "boolean":
		return &yaml.Node{Kind: yaml.ScalarNode, Value: "false", Tag: "!!bool"}
	case "array":
		return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	case "object":
		return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Value: "", Tag: "!!str", Style: yaml.DoubleQuotedStyle}
	}
}

// mappingKeys returns a set of all keys in a yaml.MappingNode.
func mappingKeys(mapping *yaml.Node) map[string]bool {
	keys := make(map[string]bool)
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		keys[mapping.Content[i].Value] = true
	}
	return keys
}

// EnforceStringQuoting ensures that all YAML values corresponding to
// "type": "string" fields in the schema are double-quoted. This prevents
// YAML parsers from interpreting values as numbers, booleans, or other types.
func EnforceStringQuoting(content, schemaJSON string) string {
	info, err := parseSchemaInfo(schemaJSON)
	if err != nil || info == nil || len(info.Properties) == 0 {
		return content
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		return content
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return content
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return content
	}

	changed := enforceQuoting(mapping, info.Properties)
	if !changed {
		return content
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		return content
	}
	return string(out)
}

// enforceQuoting sets double-quoted style on scalar nodes that the schema
// declares as "type": "string". Returns true if any nodes were modified.
func enforceQuoting(mapping *yaml.Node, props map[string]propertyInfo) bool {
	changed := false

	for i := 0; i < len(mapping.Content)-1; i += 2 {
		key := mapping.Content[i].Value
		val := mapping.Content[i+1]
		pi, ok := props[key]
		if !ok {
			continue
		}

		// Scalar string fields: ensure double-quoted.
		if pi.Type == "string" && val.Kind == yaml.ScalarNode {
			if val.Style != yaml.DoubleQuotedStyle && val.Style != yaml.SingleQuotedStyle {
				// Don't quote empty block scalars or multiline content.
				if !strings.Contains(val.Value, "\n") {
					val.Style = yaml.DoubleQuotedStyle
					changed = true
				}
			}
		}

		// Nested object: recurse.
		if pi.Type == "object" && val.Kind == yaml.MappingNode && len(pi.Properties) > 0 {
			if enforceQuoting(val, pi.Properties) {
				changed = true
			}
		}

		// Array items: enforce quoting on each item's fields.
		if pi.Type == "array" && val.Kind == yaml.SequenceNode && len(pi.ItemProps) > 0 {
			for _, item := range val.Content {
				if item.Kind == yaml.MappingNode {
					if enforceQuoting(item, pi.ItemProps) {
						changed = true
					}
				}
			}
		}
	}

	return changed
}
