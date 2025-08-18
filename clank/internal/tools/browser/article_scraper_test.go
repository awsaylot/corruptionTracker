package browser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticleScraper_ScrapeArticle(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		mockHTML       string
		expectedTitle  string
		expectedSource string
		expectedAuthor string
		expectError    bool
	}{
		{
			name: "successful scrape",
			url:  "https://example.com/article",
			mockHTML: `
				<html>
					<head>
						<title>Test Article</title>
						<meta name="author" content="John Doe">
					</head>
					<body>
						<article>
							<h1>Test Article</h1>
							<div class="author">By John Doe</div>
							<div class="content">Test content</div>
						</article>
					</body>
				</html>
			`,
			expectedTitle:  "Test Article",
			expectedSource: "example.com",
			expectedAuthor: "John Doe",
		},
		{
			name:        "invalid url",
			url:         "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := NewArticleScraper()

			if tt.mockHTML != "" {
				// TODO: Mock browser automation to return mockHTML
			}

			article, err := scraper.ScrapeArticle(tt.url)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, article)
			assert.NotEmpty(t, article.Content)
			assert.Equal(t, tt.expectedTitle, article.Title)
			assert.Equal(t, tt.expectedSource, article.Source)
			assert.Equal(t, tt.expectedAuthor, article.Author)
		})
	}
}

func TestArticleScraper_CleanContent(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "remove HTML tags",
			input:          "<p>Test content</p><script>alert('test')</script>",
			expectedOutput: "Test content",
		},
		{
			name:           "handle empty input",
			input:          "",
			expectedOutput: "",
		},
		{
			name:           "preserve meaningful whitespace",
			input:          "<p>First paragraph</p>\n<p>Second paragraph</p>",
			expectedOutput: "First paragraph\nSecond paragraph",
		},
	}

	scraper := NewArticleScraper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := scraper.cleanContent(tt.input)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

func TestArticleScraper_ExtractMetadata(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		expectedTitle  string
		expectedAuthor string
		expectedDate   string
	}{
		{
			name: "extract from meta tags",
			html: `
				<html>
					<head>
						<meta property="og:title" content="Test Title">
						<meta name="author" content="John Doe">
						<meta property="article:published_time" content="2025-08-17">
					</head>
				</html>
			`,
			expectedTitle:  "Test Title",
			expectedAuthor: "John Doe",
			expectedDate:   "2025-08-17",
		},
		{
			name: "extract from schema.org",
			html: `
				<html>
					<script type="application/ld+json">
					{
						"@type": "Article",
						"headline": "Test Title",
						"author": {
							"name": "John Doe"
						},
						"datePublished": "2025-08-17"
					}
					</script>
				</html>
			`,
			expectedTitle:  "Test Title",
			expectedAuthor: "John Doe",
			expectedDate:   "2025-08-17",
		},
	}

	scraper := NewArticleScraper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, author, date := scraper.extractMetadata(context.Background())
			assert.Equal(t, tt.expectedTitle, title)
			assert.Equal(t, tt.expectedAuthor, author)
			assert.Equal(t, tt.expectedDate, date)
		})
	}
}
