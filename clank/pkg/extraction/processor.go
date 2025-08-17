// pkg/extraction/processor.go
package extraction

import (
	"html"
	"regexp"
	"strings"
)

// ContentProcessor handles cleaning and processing of article content
type ContentProcessor struct {
	// HTML tag remover
	htmlTagRegex *regexp.Regexp
	// Multiple whitespace reducer
	whitespaceRegex *regexp.Regexp
	// Ad/navigation text patterns
	adPatterns []*regexp.Regexp
}

// NewContentProcessor creates a new content processor
func NewContentProcessor() *ContentProcessor {
	return &ContentProcessor{
		htmlTagRegex:    regexp.MustCompile(`<[^>]*>`),
		whitespaceRegex: regexp.MustCompile(`\s+`),
		adPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(advertisement|subscribe|newsletter|cookie policy|privacy policy)`),
			regexp.MustCompile(`(?i)(click here|read more|follow us|share this)`),
			regexp.MustCompile(`(?i)(trending now|most popular|recommended|related articles)`),
		},
	}
}

// CleanContent removes HTML tags, ads, and cleans up article content
func (cp *ContentProcessor) CleanContent(rawContent string) string {
	// Remove HTML tags
	content := cp.htmlTagRegex.ReplaceAllString(rawContent, " ")

	// Decode HTML entities
	content = html.UnescapeString(content)

	// Remove common ad/navigation text
	for _, pattern := range cp.adPatterns {
		content = pattern.ReplaceAllString(content, "")
	}

	// Normalize whitespace
	content = cp.whitespaceRegex.ReplaceAllString(content, " ")

	// Trim and return
	return strings.TrimSpace(content)
}
