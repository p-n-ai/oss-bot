package pipeline

import (
	"regexp"
	"strings"
)

// FixYAMLColonSpacing fixes YAML mapping values where the AI omitted the space
// after the colon (e.g. `text:hubungkaitkan` → `text: hubungkaitkan`).
// Only fixes lines that look like a YAML key-value pair: the colon must follow
// a word character and be immediately followed by a non-space, non-newline char.
// Does NOT touch colons inside quoted strings or URLs.
var yamlColonNoSpace = regexp.MustCompile(`(?m)^(\s*[a-zA-Z_][a-zA-Z0-9_]*):([^\s\n"'{])`)

func FixYAMLColonSpacing(s string) string {
	return yamlColonNoSpace.ReplaceAllString(s, "${1}: ${2}")
}

// RemoveDuplicateKeys removes duplicate YAML mapping keys, keeping only the
// first occurrence. This fixes AI output that repeats a key (e.g. official_ref
// appearing twice), which is technically valid YAML but causes some parsers to
// error or silently use the last value.
func RemoveDuplicateKeys(s string) string {
	lines := strings.Split(s, "\n")
	seen := make(map[string]bool)
	var result []string
	inNested := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track nesting: if the line starts with spaces/tabs, it's nested
		// and we don't check for duplicate keys at the nested level
		// (duplicates across different nested objects are fine).
		isTopLevel := len(line) > 0 && line[0] != ' ' && line[0] != '\t'

		if isTopLevel && strings.Contains(trimmed, ":") {
			// Reset nested tracking when we hit a new top-level key
			inNested = false
			key := strings.SplitN(trimmed, ":", 2)[0]
			if seen[key] {
				// Skip duplicate top-level key and any nested content under it
				inNested = true
				continue
			}
			seen[key] = true
		} else if inNested && !isTopLevel {
			// Skip nested content under a duplicate key
			continue
		} else {
			inNested = false
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// SanitizeYAMLQuoting converts double-quoted YAML strings that contain
// backslash-letter sequences to single-quoted strings. In YAML, double-quoted
// strings process escape sequences (\t → tab, \n → newline, \a → bell), which
// silently corrupts LaTeX notation like \text, \sqrt, \alpha, \neq. Single-quoted
// strings treat backslashes as literal characters, preserving LaTeX as intended.
//
// This is applied to AI-generated YAML before parsing or writing to disk.
func SanitizeYAMLQuoting(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '"' {
			buf.WriteByte(s[i])
			i++
			continue
		}

		// Found a double quote. Scan forward to find the matching closing quote,
		// tracking whether the content contains backslash+letter sequences.
		start := i
		i++ // skip opening "
		hasBackslashLetter := false
		contentStart := i

		for i < len(s) {
			if s[i] == '\\' && i+1 < len(s) {
				next := s[i+1]
				if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
					hasBackslashLetter = true
				}
				i += 2 // skip escape pair
			} else if s[i] == '"' {
				break
			} else if s[i] == '\n' {
				// Unescaped newline inside double quotes — likely not a quoted
				// scalar (could be a YAML comment or block context). Emit what
				// we have so far unchanged and move on.
				buf.WriteString(s[start : i+1])
				i++
				start = i
				contentStart = i
				hasBackslashLetter = false
				continue
			} else {
				i++
			}
		}

		if i >= len(s) {
			// Unterminated quote — emit remainder as-is.
			buf.WriteString(s[start:])
			break
		}

		contentEnd := i
		i++ // skip closing "

		if !hasBackslashLetter {
			// No problematic backslashes — keep original double-quoted form.
			buf.WriteString(s[start:i])
			continue
		}

		// Convert to single-quoted. Take the raw content between the double
		// quotes (which the AI intended as literal text) and wrap in single
		// quotes. In single-quoted YAML, the only special sequence is '' for
		// a literal single quote.
		content := s[contentStart:contentEnd]
		singleContent := strings.ReplaceAll(content, "'", "''")
		buf.WriteByte('\'')
		buf.WriteString(singleContent)
		buf.WriteByte('\'')
	}

	return buf.String()
}
