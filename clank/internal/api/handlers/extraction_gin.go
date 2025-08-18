package handlers

import (
	"log"
	"net/url"

	"clank/config"
	"clank/internal/db"
	"clank/internal/llm"
	"clank/internal/llm/sequential"
	browser "clank/internal/tools/browser"
	"clank/pkg/extraction"

	"github.com/gin-gonic/gin"
)

// ExtractionGinHandler handles article content extraction requests with sequential analysis for Gin
type ExtractionGinHandler struct {
	scraper            Scraper
	processor          Processor
	llm                LLMClient
	db                 Store
	analysisController *sequential.AnalysisController
}

// NewExtractionGinHandler creates a new extraction handler with sequential analysis for Gin
func NewExtractionGinHandler(cfg *config.Config) *ExtractionGinHandler {
	llmClient := llm.NewClient(cfg)
	return &ExtractionGinHandler{
		scraper:            browser.NewArticleScraper(),
		processor:          extraction.NewContentProcessor(),
		llm:                llmClient,
		db:                 db.NewArticleStore(),
		analysisController: sequential.NewAnalysisController(llmClient),
	}
}

// HandleURLExtraction processes a URL for article extraction with sequential analysis
func (h *ExtractionGinHandler) HandleURLExtraction(c *gin.Context) {
	var req struct {
		URL   string `json:"url"`
		Depth int    `json:"depth,omitempty"` // Analysis depth (2-10)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Extraction] Invalid request body: %v", err)
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("[Extraction] Processing URL: %s", req.URL)

	// Validate URL
	if req.URL == "" {
		c.JSON(400, gin.H{"error": "URL is required"})
		return
	}

	if _, err := url.Parse(req.URL); err != nil {
		c.JSON(400, gin.H{"error": "Invalid URL"})
		return
	}

	// Validate depth (default to 3 if not specified)
	if req.Depth == 0 {
		req.Depth = 3
	}
	if req.Depth < 2 || req.Depth > 10 {
		c.JSON(400, gin.H{"error": "Depth must be between 2 and 10"})
		return
	}

	// Initialize scraper if needed
	log.Println("[Extraction] Initializing scraper...")
	if err := h.scraper.Initialize(); err != nil {
		log.Printf("[Extraction] Scraper initialization failed: %v", err)
		c.JSON(500, gin.H{"error": "Failed to initialize scraper: " + err.Error()})
		return
	}

	// Scrape the article
	log.Printf("[Extraction] Starting article scraping for URL: %s", req.URL)
	article, err := h.scraper.ScrapeArticle(req.URL)
	if err != nil {
		log.Printf("[Extraction] Article scraping failed: %v", err)
		c.JSON(500, gin.H{"error": "Failed to scrape article: " + err.Error()})
		return
	}
	log.Printf("[Extraction] Successfully scraped article, length: %d characters", len(article.Content))

	// Process the content
	log.Println("[Extraction] Starting article processing...")
	result, err := h.processor.ProcessArticle(article.Content)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to process article: " + err.Error()})
		return
	}
	log.Println("[Extraction] Article processing completed successfully")

	// Update article with processed content
	article.Content = result.Content
	if result.Title != "" {
		article.Title = result.Title
	}
	if result.Metadata != nil {
		article.Metadata = result.Metadata
	}

	// Save the article
	log.Println("[Extraction] Saving article to database...")
	if err := h.db.SaveArticle(article); err != nil {
		log.Printf("[Extraction] Failed to save article: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save article: " + err.Error()})
		return
	}
	log.Printf("[Extraction] Article saved successfully with ID: %s", article.ID)

	// Return successful response
	c.JSON(200, gin.H{
		"articleId": article.ID,
		"title":     article.Title,
		"content":   article.Content,
		"url":       article.URL,
		"status":    "success",
	})
}
