package tools

import (
	"context"
	"fmt"
	"time"
)

// ArticleScraper extends BrowserAutomation for article-specific scraping
type ArticleScraper struct {
	*BrowserAutomation
}

// ScrapedArticle contains all the information extracted from a news article
type ScrapedArticle struct {
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Author      string            `json:"author"`
	PublishDate time.Time         `json:"publishDate"`
	Source      string            `json:"source"`
	Metadata    map[string]string `json:"metadata"`
}

// NewArticleScraper creates a new ArticleScraper instance
func NewArticleScraper() (*ArticleScraper, error) {
	ba, err := NewBrowserAutomation()
	if err != nil {
		return nil, fmt.Errorf("failed to create browser automation: %w", err)
	}

	return &ArticleScraper{
		BrowserAutomation: ba,
	}, nil
}

// ScrapeArticle extracts article content from the given URL
func (as *ArticleScraper) ScrapeArticle(ctx context.Context, url string) (*ScrapedArticle, error) {
	if err := as.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}
	defer as.Close()

	if err := as.NavigateTo(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Wait for the main content to load
	if err := as.WaitForSelector(ctx, "article, [role='article'], .article, .post, .story"); err != nil {
		return nil, fmt.Errorf("failed to find article content: %w", err)
	}

	article := &ScrapedArticle{
		URL:      url,
		Metadata: make(map[string]string),
	}

	// Extract title
	title, err := as.page.Title()
	if err == nil {
		article.Title = title
	}

	// Try multiple selectors for the main content
	content, err := as.getArticleContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get article content: %w", err)
	}
	article.Content = content

	// Try to extract author
	author, _ := as.getArticleAuthor(ctx)
	article.Author = author

	// Try to extract publish date
	date, _ := as.getArticleDate(ctx)
	if !date.IsZero() {
		article.PublishDate = date
	}

	// Extract source from URL or site name
	article.Source = as.extractSourceFromURL(url)

	return article, nil
}

// getArticleContent tries various selectors to find the main article content
func (as *ArticleScraper) getArticleContent(ctx context.Context) (string, error) {
	selectors := []string{
		"article",
		"[role='article']",
		".article-content",
		".post-content",
		".story-content",
		"#article-body",
	}

	for _, selector := range selectors {
		element, err := as.page.QuerySelector(selector)
		if err != nil {
			continue
		}
		if element == nil {
			continue
		}

		// Get the text content
		content, err := element.TextContent()
		if err != nil {
			continue
		}

		if content != "" {
			return content, nil
		}
	}

	return "", fmt.Errorf("could not find article content")
}

// getArticleAuthor tries to find the article author
func (as *ArticleScraper) getArticleAuthor(ctx context.Context) (string, error) {
	selectors := []string{
		"[rel='author']",
		".author",
		".byline",
		"[itemprop='author']",
	}

	for _, selector := range selectors {
		element, err := as.page.QuerySelector(selector)
		if err != nil {
			continue
		}
		if element == nil {
			continue
		}

		author, err := element.TextContent()
		if err != nil {
			continue
		}

		if author != "" {
			return author, nil
		}
	}

	return "", nil
}

// getArticleDate tries to find the article publish date
func (as *ArticleScraper) getArticleDate(ctx context.Context) (time.Time, error) {
	selectors := []string{
		"[itemprop='datePublished']",
		"time",
		".date",
		".published",
		"meta[property='article:published_time']",
	}

	for _, selector := range selectors {
		element, err := as.page.QuerySelector(selector)
		if err != nil {
			continue
		}
		if element == nil {
			continue
		}

		// Try getting the datetime attribute first
		datetime, err := element.GetAttribute("datetime")
		if err == nil && datetime != "" {
			if t, err := time.Parse(time.RFC3339, datetime); err == nil {
				return t, nil
			}
		}

		// Fall back to text content
		dateStr, err := element.TextContent()
		if err == nil && dateStr != "" {
			// Try parsing various date formats
			layouts := []string{
				"2006-01-02",
				"January 2, 2006",
				"Jan 2, 2006",
				time.RFC3339,
			}

			for _, layout := range layouts {
				if t, err := time.Parse(layout, dateStr); err == nil {
					return t, nil
				}
			}
		}
	}

	return time.Time{}, nil
}

// extractSourceFromURL extracts the source name from the URL
func (as *ArticleScraper) extractSourceFromURL(url string) string {
	// TODO: Implement proper domain extraction
	return url
}
