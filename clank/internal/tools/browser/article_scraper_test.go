package tools

import (
	"context"
	"testing"
)

func TestArticleScraper(t *testing.T) {
	ctx := context.Background()

	scraper, err := NewArticleScraper()
	if err != nil {
		t.Fatalf("Failed to create article scraper: %v", err)
	}

	// Test with a known news article
	article, err := scraper.ScrapeArticle(ctx, "https://example.com/article")
	if err != nil {
		t.Fatalf("Failed to scrape article: %v", err)
	}

	if article.URL == "" {
		t.Error("Expected URL to be set")
	}

	if article.Content == "" {
		t.Error("Expected content to be extracted")
	}

	if article.Title == "" {
		t.Error("Expected title to be extracted")
	}
}

func TestArticleExtraction_MultipleSelectors(t *testing.T) {
	ctx := context.Background()

	scraper, err := NewArticleScraper()
	if err != nil {
		t.Fatalf("Failed to create article scraper: %v", err)
	}

	// Test cases for different news sites
	testCases := []string{
		"https://www.theguardian.com/world/2023/aug/17/example-article",
		"https://www.bbc.com/news/world-12345678",
		"https://www.reuters.com/world/example-article-2023-08-17/",
	}

	for _, url := range testCases {
		t.Run(url, func(t *testing.T) {
			article, err := scraper.ScrapeArticle(ctx, url)
			if err != nil {
				t.Skipf("Skipping test for %s: %v", url, err)
				return
			}

			if article.Content == "" {
				t.Errorf("Failed to extract content from %s", url)
			}
		})
	}
}
