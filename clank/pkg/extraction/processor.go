package extraction

import (
	"regexp"
	"strings"
)

// ContentProcessor handles cleaning and formatting article content
type ContentProcessor struct {
	// Common patterns to clean
	patterns map[string]*regexp.Regexp
}

// NewContentProcessor creates a new content processor
func NewContentProcessor() *ContentProcessor {
	cp := &ContentProcessor{
		patterns: make(map[string]*regexp.Regexp),
	}

	// Initialize common cleanup patterns
	cp.patterns["whitespace"] = regexp.MustCompile(`\s+`)
	cp.patterns["html"] = regexp.MustCompile(`<[^>]*>`)
	cp.patterns["newlines"] = regexp.MustCompile(`\n{3,}`)
	cp.patterns["advertisements"] = regexp.MustCompile(`(?i)(advertisement|sponsored content|promoted content)`)
	cp.patterns["social"] = regexp.MustCompile(`(?i)(share on|follow us|subscribe to our)`)
	cp.patterns["urls"] = regexp.MustCompile(`https?://\S+`)

	return cp
}

// CleanContent removes unwanted elements and normalizes the text
func (cp *ContentProcessor) CleanContent(content string) string {
	// Remove HTML tags
	content = cp.patterns["html"].ReplaceAllString(content, "")

	// Remove advertisements and social media text
	content = cp.patterns["advertisements"].ReplaceAllString(content, "")
	content = cp.patterns["social"].ReplaceAllString(content, "")

	// Remove URLs
	content = cp.patterns["urls"].ReplaceAllString(content, "")

	// Normalize whitespace
	content = cp.patterns["whitespace"].ReplaceAllString(content, " ")
	content = cp.patterns["newlines"].ReplaceAllString(content, "\n\n")

	// Trim whitespace
	content = strings.TrimSpace(content)

	return content
}

// FormatForLLM prepares the content for LLM processing
func (cp *ContentProcessor) FormatForLLM(title, content, source, date string) string {
	var builder strings.Builder

	// Build a structured format for the LLM
	builder.WriteString("TITLE: ")
	builder.WriteString(title)
	builder.WriteString("\n\n")

	builder.WriteString("SOURCE: ")
	builder.WriteString(source)
	builder.WriteString("\n")

	if date != "" {
		builder.WriteString("DATE: ")
		builder.WriteString(date)
		builder.WriteString("\n")
	}

	builder.WriteString("\nCONTENT:\n")
	builder.WriteString(cp.CleanContent(content))

	return builder.String()
}

// ExtractRelevantParagraphs finds paragraphs most likely to contain entity information
func (cp *ContentProcessor) ExtractRelevantParagraphs(content string) []string {
	// Split into paragraphs
	paragraphs := strings.Split(content, "\n\n")

	// Filter out short or irrelevant paragraphs
	var relevant []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if len(p) < 50 { // Skip very short paragraphs
			continue
		}

		// Look for paragraphs likely to contain entity information
		if containsEntityIndicators(p) {
			relevant = append(relevant, p)
		}
	}

	return relevant
}

// containsEntityIndicators checks if a paragraph likely contains entity information
func containsEntityIndicators(text string) bool {
	indicators := []string{
		"(?i)(mr\\.|mrs\\.|ms\\.|dr\\.|prof\\.|senator|representative)",                                           // People titles
		"(?i)(company|corporation|inc\\.|ltd\\.|LLC)",                                                             // Company indicators
		"(?i)(department|agency|ministry|office)",                                                                 // Organization indicators
		"(?i)(\\$|€|£|¥)\\s?\\d+",                                                                                 // Money amounts
		"(?i)(january|february|march|april|may|june|july|august|september|october|november|december)\\s+\\d{1,2}", // Dates
	}

	for _, pattern := range indicators {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			return true
		}
	}

	return false
}
