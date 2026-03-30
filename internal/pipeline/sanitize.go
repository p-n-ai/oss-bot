package pipeline

import "strings"

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
