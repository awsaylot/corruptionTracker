package extraction

import (
	"strings"
	"testing"
)

func TestContentProcessor_CleanContent(t *testing.T) {
	cp := NewContentProcessor()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove HTML tags",
			input:    "<p>This is a test</p><div>More text</div>",
			expected: "This is a test More text",
		},
		{
			name:     "Remove advertisements",
			input:    "Article content. Advertisement: Buy now! More content.",
			expected: "Article content. More content.",
		},
		{
			name:     "Remove social media text",
			input:    "Important news. Share on Facebook! More news. Follow us on Twitter!",
			expected: "Important news. More news.",
		},
		{
			name:     "Normalize whitespace",
			input:    "Line one.\n\n\n\nLine two.   Extra  spaces.",
			expected: "Line one.\n\nLine two. Extra spaces.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cp.CleanContent(tt.input)
			result = strings.TrimSpace(result)
			if result != strings.TrimSpace(tt.expected) {
				t.Errorf("CleanContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContentProcessor_FormatForLLM(t *testing.T) {
	cp := NewContentProcessor()

	title := "Test Article"
	content := "This is the article content.\nIt has multiple lines."
	source := "News Site"
	date := "2025-08-17"

	result := cp.FormatForLLM(title, content, source, date)

	expectedParts := []string{
		"TITLE: Test Article",
		"SOURCE: News Site",
		"DATE: 2025-08-17",
		"CONTENT:",
		"This is the article content.",
		"It has multiple lines.",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("FormatForLLM() missing expected part: %v", part)
		}
	}
}

func TestContentProcessor_ExtractRelevantParagraphs(t *testing.T) {
	cp := NewContentProcessor()

	input := `Short line.

Mr. John Smith, CEO of Tech Corp Inc., announced a $50 million investment.

This is a regular paragraph without entities.

The Department of Justice investigation began in January 2025.

Another short line.`

	relevant := cp.ExtractRelevantParagraphs(input)

	if len(relevant) != 2 {
		t.Errorf("Expected 2 relevant paragraphs, got %d", len(relevant))
	}

	expectedPhrases := []string{
		"Mr. John Smith",
		"Department of Justice",
	}

	for _, phrase := range expectedPhrases {
		found := false
		for _, p := range relevant {
			if strings.Contains(p, phrase) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find paragraph containing %q", phrase)
		}
	}
}
