package models

// ProcessingResult contains the results of article content processing
type ProcessingResult struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Summary     string                 `json:"summary"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error,omitempty"`
	SourceURL   string                 `json:"sourceUrl"`
	PublishedAt string                 `json:"publishedAt,omitempty"`
	WordCount   int                    `json:"wordCount"`
}
