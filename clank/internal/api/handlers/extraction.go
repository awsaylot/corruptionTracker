package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"clank/config"
	"clank/internal/db"
	"clank/internal/llm"
	"clank/internal/models"
	"clank/internal/tools/browser"
	"clank/pkg/extraction"

	"github.com/google/uuid"
)

// ExtractionHandler handles article content extraction requests
type ExtractionHandler struct {
	scraper   *browser.ArticleScraper
	processor *extraction.ContentProcessor
	llm       *llm.Client
	db        *db.ArticleStore
}

// NewExtractionHandler creates a new extraction handler
func NewExtractionHandler(cfg *config.Config) *ExtractionHandler {
	return &ExtractionHandler{
		scraper:   browser.NewArticleScraper(),
		processor: extraction.NewContentProcessor(),
		llm:       llm.NewClient(cfg),
		db:        db.NewArticleStore(),
	}
}

// ExtractionRequest represents the request to extract information from a URL
type ExtractionRequest struct {
	URL string `json:"url"`
}

// HandleURLExtraction processes a URL for article extraction
func (h *ExtractionHandler) HandleURLExtraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExtractionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Create context for the request
	ctx := r.Context()

	// Scrape the article
	rawContent, metadata, err := h.scraper.ScrapeArticle(req.URL)
	if err != nil {
		http.Error(w, "Failed to scrape article: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Clean and process the content
	cleanedContent := h.processor.CleanContent(rawContent)

	// Create article model
	article := &models.Article{
		ID:          uuid.New().String(),
		URL:         req.URL,
		Title:       metadata["title"].(string),
		Content:     cleanedContent,
		Source:      metadata["source"].(string),
		Author:      metadata["author"].(string),
		PublishDate: metadata["publishDate"].(time.Time),
		ExtractedAt: time.Now(),
		Metadata:    metadata,
	}

	// Process with LLM for entity extraction
	extraction, err := h.llm.ProcessArticle(ctx, article)
	if err != nil {
		http.Error(w, "Failed to process article with LLM: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store article and extracted entities in Neo4j
	if err := h.db.StoreArticle(article, extraction); err != nil {
		http.Error(w, "Failed to store article data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := struct {
		Article    *models.Article          `json:"article"`
		Extraction *models.ExtractionResult `json:"extraction"`
	}{
		Article:    article,
		Extraction: extraction,
	}

	// Return the processed article and extraction results
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
