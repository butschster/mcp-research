package service

import "testing"

func TestNormalizeContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "literal backslash-n becomes newline",
			input: `line1\nline2`,
			want:  "line1\nline2",
		},
		{
			name:  "literal backslash-t becomes tab",
			input: `col1\tcol2`,
			want:  "col1\tcol2",
		},
		{
			name:  "multiple occurrences",
			input: `a\nb\nc\nd`,
			want:  "a\nb\nc\nd",
		},
		{
			name:  "mixed newlines and tabs",
			input: `header\n\tindented\n\tindented2`,
			want:  "header\n\tindented\n\tindented2",
		},
		{
			name:  "real newlines preserved",
			input: "already\nnormal",
			want:  "already\nnormal",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no escape sequences",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "markdown with literal escapes",
			input: `# Title\n\nParagraph with **bold**\n\n- item1\n- item2`,
			want:  "# Title\n\nParagraph with **bold**\n\n- item1\n- item2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeContent(tt.input)
			if got != tt.want {
				t.Errorf("normalizeContent(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.want)
			}
		})
	}
}
