package parser_test

import (
	"testing"

	"github.com/p-n-ai/oss-bot/internal/parser"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantAction  string
		wantType    string
		wantTopic   string
		wantOption  map[string]string
		wantErr     bool
	}{
		{
			name:       "add-teaching-notes",
			input:      "@oss-bot add teaching notes for F2-01",
			wantAction: "add",
			wantType:   "teaching notes",
			wantTopic:  "F2-01",
		},
		{
			name:       "add-assessments-with-count",
			input:      "@oss-bot add 5 assessments for F1-01 difficulty:medium",
			wantAction: "add",
			wantType:   "assessments",
			wantTopic:  "F1-01",
			wantOption: map[string]string{"difficulty": "medium"},
		},
		{
			name:       "translate",
			input:      "@oss-bot translate F1-01 to ms",
			wantAction: "translate",
			wantTopic:  "F1-01",
			wantOption: map[string]string{"to": "ms"},
		},
		{
			name:       "quality",
			input:      "@oss-bot quality malaysia-kssm-matematik-tingkatan1",
			wantAction: "quality",
			wantTopic:  "malaysia-kssm-matematik-tingkatan1",
		},
		{
			name:       "scaffold",
			input:      "@oss-bot scaffold syllabus india/cbse/mathematics-class10",
			wantAction: "scaffold",
			wantTopic:  "india/cbse/mathematics-class10",
		},
		{
			name:       "import-url",
			input:      "@oss-bot import https://example.org/curriculum-spec",
			wantAction: "import",
			wantTopic:  "",
			wantOption: map[string]string{"url": "https://example.org/curriculum-spec"},
		},
		{
			name:       "import-attachment",
			input:      "@oss-bot import",
			wantAction: "import",
			wantTopic:  "",
		},
		{
			name:       "import-attachment-with-vision",
			input:      "@oss-bot import vision:true",
			wantAction: "import",
			wantTopic:  "",
			wantOption: map[string]string{"vision": "true"},
		},
		{
			name:    "no-bot-mention",
			input:   "Just a regular comment",
			wantErr: true,
		},
		{
			name:    "empty-command",
			input:   "@oss-bot",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parser.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if cmd.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", cmd.Action, tt.wantAction)
			}
			if tt.wantType != "" && cmd.ContentType != tt.wantType {
				t.Errorf("ContentType = %q, want %q", cmd.ContentType, tt.wantType)
			}
			if cmd.TopicPath != tt.wantTopic {
				t.Errorf("TopicPath = %q, want %q", cmd.TopicPath, tt.wantTopic)
			}
			for k, v := range tt.wantOption {
				if cmd.Options[k] != v {
					t.Errorf("Options[%q] = %q, want %q", k, cmd.Options[k], v)
				}
			}
		})
	}
}
