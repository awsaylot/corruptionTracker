package tools

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ArticleScraper extends BrowserAutomation for article-specific scraping
type ArticleScraper struct {
	*BrowserAutomation
}

// NewArticleScraper creates a new ArticleScraper instance
func NewArticleScraper() *ArticleScraper {
	ba, _ := NewBrowserAutomation()
	return &ArticleScraper{
		BrowserAutomation: ba,
	}
}

// ScrapeArticle extracts article content from the given URL
func (as *ArticleScraper) ScrapeArticle(articleURL string) (string, map[string]interface{}, error) {
	ctx := context.Background()

	if err := as.Initialize(ctx); err != nil {
		return "", nil, fmt.Errorf("failed to initialize browser: %w", err)
	}
	defer as.Close()

	if err := as.NavigateTo(ctx, articleURL); err != nil {
		return "", nil, fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Wait for the main content to load
	contentSelectors := []string{
		"article", "[role='article']", ".article", ".post", ".story",
		".content", ".article-body", ".story-body",
	}

	var content string
	for _, selector := range contentSelectors {
		if err := as.WaitForSelector(ctx, selector); err == nil {
			element, err := as.page.QuerySelector(selector)
			if err == nil && element != nil {
				if text, err := element.TextContent(); err == nil && text != "" {
					content = text
					break
				}
			}
		}
	}

	if content == "" {
		return "", nil, fmt.Errorf("could not extract article content")
	}

	// Get title
	title, _ := as.page.Title()

	// Extract metadata
	metadata := map[string]interface{}{
		"title":       title,
		"source":      as.extractSourceFromURL(articleURL),
		"author":      as.extractAuthor(ctx),
		"publishDate": as.extractPublishDate(ctx),
	}

	return content, metadata, nil
}

// extractSourceFromURL extracts the source name from the URL
func (as *ArticleScraper) extractSourceFromURL(articleURL string) string {
	if u, err := url.Parse(articleURL); err == nil {
		parts := strings.Split(u.Hostname(), ".")
		if len(parts) >= 2 {
			// Get the domain name without TLD
			return strings.Title(parts[len(parts)-2])
		}
		return u.Hostname()
	}
	return "Unknown"
}

// extractAuthor tries to find the article author
func (as *ArticleScraper) extractAuthor(ctx context.Context) string {
	selectors := []string{
		"[rel='author']", ".author", ".byline", "[itemprop='author']",
		".author-name", ".article-author", ".story-author",
	}

	for _, selector := range selectors {
		if element, err := as.page.QuerySelector(selector); err == nil && element != nil {
			if text, err := element.TextContent(); err == nil && text != "" {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

// extractPublishDate tries to find the article publish date
func (as *ArticleScraper) extractPublishDate(ctx context.Context) time.Time {
	selectors := []string{
		"[itemprop='datePublished']", "time", ".date", ".published",
		"meta[property='article:published_time']", ".publish-date",
	}

	for _, selector := range selectors {
		if element, err := as.page.QuerySelector(selector); err == nil && element != nil {
			// Try datetime attribute first
			if datetime, err := element.GetAttribute("datetime"); err == nil && datetime != "" {
				if t, err := time.Parse(time.RFC3339, datetime); err == nil {
					return t
				}
			}

			// Try content attribute for meta tags
			if content, err := element.GetAttribute("content"); err == nil && content != "" {
				if t, err := time.Parse(time.RFC3339, content); err == nil {
					return t
				}
			}

			// Try text content
			if text, err := element.TextContent(); err == nil && text != "" {
				layouts := []string{
					"2006-01-02", "January 2, 2006", "Jan 2, 2006",
					"2006-01-02T15:04:05Z07:00", time.RFC3339,
				}
				for _, layout := range layouts {
					if t, err := time.Parse(layout, strings.TrimSpace(text)); err == nil {
						return t
					}
				}
			}
		}
	}
	return time.Now() // Fallback to current time
}
