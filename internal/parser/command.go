// Package parser handles input parsing and content extraction for the bot.
package parser

import (
	"fmt"
	"strings"
)

// BotCommand represents a parsed @oss-bot command from an issue comment.
type BotCommand struct {
	Action       string            // "add", "translate", "scaffold", "import", "quality", "enrich"
	ContentType  string            // "teaching notes", "assessments", "examples"
	TopicPath    string            // Path or ID of the target topic
	Options      map[string]string // Additional options (count, difficulty, language, url, etc.)
	User         string            // GitHub username who issued the command
	IssueNum     int               // Issue number where the command was posted
	RepoFullName string            // "owner/repo"
	CommentBody  string            // Full comment body
	Attachments  []Attachment      // File attachments (for import command)
}

// Attachment represents a file attached to a GitHub comment.
type Attachment struct {
	URL      string // Download URL of the attachment
	FileName string // Original filename (e.g., "syllabus.pdf")
	MimeType string // Detected MIME type
}

// ParseCommand extracts a BotCommand from a comment body containing @oss-bot.
func ParseCommand(body string) (*BotCommand, error) {
	if !strings.Contains(body, "@oss-bot") {
		return nil, fmt.Errorf("no @oss-bot mention found")
	}

	idx := strings.Index(body, "@oss-bot")
	rest := strings.TrimSpace(body[idx+len("@oss-bot"):])

	if rest == "" {
		return nil, fmt.Errorf("no command after @oss-bot")
	}

	cmd := &BotCommand{
		Options: make(map[string]string),
	}

	// Parse key:value options and strip them from the command text.
	parts := strings.Fields(rest)
	var cleanParts []string
	for _, p := range parts {
		if kv, ok := parseKeyValue(p); ok {
			cmd.Options[kv[0]] = kv[1]
		} else {
			cleanParts = append(cleanParts, p)
		}
	}
	rest = strings.Join(cleanParts, " ")

	switch {
	case strings.HasPrefix(rest, "add teaching notes"):
		cmd.Action = "add"
		cmd.ContentType = "teaching notes"
		cmd.TopicPath = extractTopicAfter(rest, "for")

	case strings.Contains(rest, "assessments"):
		cmd.Action = "add"
		cmd.ContentType = "assessments"
		cmd.TopicPath = extractTopicAfter(rest, "for")
		// Extract count if present (e.g., "add 5 assessments")
		for _, p := range cleanParts {
			var n int
			if _, err := fmt.Sscanf(p, "%d", &n); err == nil {
				cmd.Options["count"] = p
			}
		}

	case strings.Contains(rest, "examples"):
		cmd.Action = "add"
		cmd.ContentType = "examples"
		cmd.TopicPath = extractTopicAfter(rest, "for")

	case strings.HasPrefix(rest, "translate"):
		cmd.Action = "translate"
		remaining := strings.TrimSpace(strings.TrimPrefix(rest, "translate "))
		if idx := strings.Index(remaining, " to "); idx >= 0 {
			cmd.TopicPath = strings.TrimSpace(remaining[:idx])
			cmd.Options["to"] = strings.TrimSpace(remaining[idx+4:])
		} else {
			cmd.TopicPath = remaining
		}

	case strings.HasPrefix(rest, "scaffold"):
		cmd.Action = "scaffold"
		rest = strings.TrimPrefix(rest, "scaffold ")
		rest = strings.TrimPrefix(rest, "syllabus ")
		rest = strings.TrimPrefix(rest, "subject ")
		cmd.TopicPath = strings.TrimSpace(rest)

	case strings.HasPrefix(rest, "quality"):
		cmd.Action = "quality"
		cmd.TopicPath = strings.TrimSpace(strings.TrimPrefix(rest, "quality "))

	case strings.HasPrefix(rest, "import"):
		cmd.Action = "import"
		remaining := strings.TrimSpace(strings.TrimPrefix(rest, "import"))
		if strings.HasPrefix(remaining, "http://") || strings.HasPrefix(remaining, "https://") {
			cmd.Options["url"] = remaining
		}

	case strings.HasPrefix(rest, "enrich"):
		cmd.Action = "enrich"
		cmd.TopicPath = strings.TrimSpace(strings.TrimPrefix(rest, "enrich "))

	default:
		return nil, fmt.Errorf("unrecognized command: %s", rest)
	}

	return cmd, nil
}

// parseKeyValue checks if p has the form "key:value" (not a URL) and returns the pair.
func parseKeyValue(p string) ([2]string, bool) {
	if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
		return [2]string{}, false
	}
	if idx := strings.Index(p, ":"); idx > 0 {
		return [2]string{p[:idx], p[idx+1:]}, true
	}
	return [2]string{}, false
}

// extractTopicAfter returns the text after the given keyword in text.
func extractTopicAfter(text, keyword string) string {
	if idx := strings.Index(text, " "+keyword+" "); idx >= 0 {
		return strings.TrimSpace(text[idx+len(keyword)+2:])
	}
	// Fallback: last word
	parts := strings.Fields(text)
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
