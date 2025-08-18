package browser

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"clank/internal/models"

	"github.com/google/uuid"
)

// ArticleScraper extends BrowserAutomation for article-specific scraping
var (
	// Regular expressions for content cleaning
	removeScriptStyleRegex = regexp.MustCompile(`(?s)<(script|style).*?</(?:script|style)>`)
	multipleSpacesRegex    = regexp.MustCompile(`\s+`)
	multipleNewlinesRegex  = regexp.MustCompile(`\n\s*\n`)
)

type ArticleScraper struct {
	*BrowserAutomation
	initialized bool
}

// NewArticleScraper creates a new ArticleScraper instance
func NewArticleScraper() *ArticleScraper {
	ba, _ := NewBrowserAutomation()
	return &ArticleScraper{
		BrowserAutomation: ba,
		initialized:       false,
	}
}

// Initialize prepares the scraper for use
func (as *ArticleScraper) Initialize() error {
	if as.initialized {
		return nil
	}

	if err := as.BrowserAutomation.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize browser automation: %w", err)
	}

	as.initialized = true
	return nil
}

// cleanContent removes unwanted elements and normalizes whitespace
func (as *ArticleScraper) cleanContent(content string) string {
	// Remove script and style content
	content = removeScriptStyleRegex.ReplaceAllString(content, "")

	// Replace multiple newlines and spaces with single instances
	content = multipleSpacesRegex.ReplaceAllString(content, " ")
	content = multipleNewlinesRegex.ReplaceAllString(content, "\n")

	// Trim whitespace
	return strings.TrimSpace(content)
}

// ScrapeArticle scrapes content and metadata from the given URL
func (as *ArticleScraper) ScrapeArticle(urlStr string) (*models.Article, error) {
	if !as.initialized {
		if err := as.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize scraper: %w", err)
		}
	}

	ctx := context.Background()
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Navigate to the URL
	if err := as.navigateWithRetry(ctx, urlStr); err != nil {
		return nil, fmt.Errorf("failed to navigate to URL: %w", err)
	}

	title, author, pubDate := as.extractMetadata(ctx)
	content, err := as.extractMainContent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content: %w", err)
	}

	article := &models.Article{
		ID:          uuid.New().String(),
		URL:         urlStr,
		Title:       title,
		Content:     content,
		Source:      parsed.Host,
		Author:      author,
		PublishDate: pubDate,
		ExtractedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	return article, nil
}

// extractMetadata extracts title, author, and publication date from HTML
func (as *ArticleScraper) extractMetadata(ctx context.Context) (string, string, time.Time) {
	var title, author string
	var pubDate time.Time

	// Extract title using common meta tags and selectors
	titleSelectors := []string{
		"meta[property='og:title']",
		"meta[name='twitter:title']",
		"h1.article-title",
		"h1.entry-title",
		"h1",
	}

	for _, selector := range titleSelectors {
		element, err := as.page.QuerySelector(selector)
		if err == nil && element != nil {
			if t, err := element.GetAttribute("content"); err == nil && t != "" {
				title = t
				break
			}
		}
	}

	// Extract author
	authorSelectors := []string{
		"meta[name='author']",
		".author",
		".byline",
		"[rel='author']",
	}

	for _, selector := range authorSelectors {
		element, err := as.page.QuerySelector(selector)
		if err == nil && element != nil {
			if a, err := element.GetAttribute("content"); err == nil && a != "" {
				author = a
				break
			}
		}
	}

	// Extract date
	dateSelectors := []string{
		"meta[property='article:published_time']",
		"time[pubdate]",
		".published-date",
		".post-date",
	}

	for _, selector := range dateSelectors {
		element, err := as.page.QuerySelector(selector)
		if err == nil && element != nil {
			if d, err := element.GetAttribute("datetime"); err == nil && d != "" {
				if t, err := time.Parse(time.RFC3339, d); err == nil {
					pubDate = t
					break
				}
			}
		}
	}

	return title, author, pubDate
}

// navigateWithRetry attempts to navigate to a URL with retries
func (as *ArticleScraper) navigateWithRetry(ctx context.Context, urlStr string) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := as.BrowserAutomation.NavigateTo(ctx, urlStr); err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to navigate after retries: %w", lastErr)
}

// extractMainContent extracts the main content from the current page
func (as *ArticleScraper) extractMainContent(ctx context.Context) (string, error) {
	// Find main article content using common selectors
	selectors := []string{
		"article",
		"[role='main']",
		"#content",
		"#main-content",
		".article-content",
		".post-content",
	}

	for _, selector := range selectors {
		content, err := as.BrowserAutomation.GetElementTextBySelector(ctx, selector)
		if err == nil && content != "" {
			return as.cleanContent(content), nil
		}
	}

	// Fallback: try to get body content
	content, err := as.BrowserAutomation.GetElementTextBySelector(ctx, "body")
	if err != nil {
		return "", fmt.Errorf("failed to extract content: %w", err)
	}

	return as.cleanContent(content), nil
}

// extractSourceFromURL extracts the source name from the URL
func (as *ArticleScraper) extractSourceFromURL(urlStr string) string {
	if u, err := url.Parse(urlStr); err == nil {
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
		if text, err := as.BrowserAutomation.GetElementTextBySelector(ctx, selector); err == nil && text != "" {
			return strings.TrimSpace(text)
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
		if val, err := as.BrowserAutomation.GetElementAttributeBySelector(ctx, selector, "datetime"); err == nil && val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				return t
			}
		}

		if val, err := as.BrowserAutomation.GetElementAttributeBySelector(ctx, selector, "content"); err == nil && val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				return t
			}
		}

		if text, err := as.BrowserAutomation.GetElementTextBySelector(ctx, selector); err == nil && text != "" {
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
	return time.Now() // Fallback to current time
}
