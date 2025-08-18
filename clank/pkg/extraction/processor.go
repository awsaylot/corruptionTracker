// pkg/extraction/processor.go
package extraction

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"clank/internal/models"
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

// ProcessArticle processes raw article content into a structured result
func (cp *ContentProcessor) ProcessArticle(content string) (*models.ProcessingResult, error) {
	// Split into title and content
	parts := strings.SplitN(content, "\n", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid content format")
	}

	title := strings.TrimPrefix(strings.TrimSpace(parts[0]), "TITLE:")
	mainContent := strings.TrimSpace(parts[1])

	// Clean content
	cleanedContent := cp.CleanContent(mainContent)
	relevantParagraphs := cp.ExtractRelevantParagraphs(cleanedContent)

	result := &models.ProcessingResult{
		Title:     strings.TrimSpace(title),
		Content:   strings.Join(relevantParagraphs, "\n\n"),
		WordCount: len(strings.Fields(cleanedContent)),
		Metadata:  make(map[string]interface{}),
	}

	return result, nil
}

// FormatForLLM formats article content for LLM processing
func (cp *ContentProcessor) FormatForLLM(title, content, source, date string) string {
	var parts []string
	parts = append(parts, "TITLE: "+title)
	parts = append(parts, "SOURCE: "+source)
	parts = append(parts, "DATE: "+date)
	parts = append(parts, "CONTENT:\n"+content)
	return strings.Join(parts, "\n")
}

// ExtractRelevantParagraphs extracts paragraphs that are likely to contain relevant content
func (cp *ContentProcessor) ExtractRelevantParagraphs(content string) []string {
	// Split by double newlines to get paragraphs
	paragraphs := strings.Split(content, "\n\n")

	var relevant []string
	for _, p := range paragraphs {
		// Remove extra whitespace
		p = cp.whitespaceRegex.ReplaceAllString(p, " ")
		p = strings.TrimSpace(p)

		// Skip very short paragraphs
		if len(p) < 50 {
			continue
		}

		// Skip paragraphs that match ad patterns
		isAd := false
		for _, pattern := range cp.adPatterns {
			if pattern.MatchString(p) {
				isAd = true
				break
			}
		}
		if !isAd {
			relevant = append(relevant, p)
		}
	}
	return relevant
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
