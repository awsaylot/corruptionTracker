package handlers

import (
	"clank/internal/llm"
	"clank/internal/models"
	"context"
	"time"
)

// Scraper defines the interface for article scraping
type Scraper interface {
	Initialize() error
	ScrapeArticle(url string) (*models.Article, error)
}

// Processor defines the interface for content processing
type Processor interface {
	ProcessArticle(content string) (*models.ProcessingResult, error)
}

// LLMClient defines the interface for LLM operations
type LLMClient interface {
	Generate(ctx context.Context, messages []llm.Message) (*llm.Response, error)
	GenerateStream(ctx context.Context, messages []llm.Message, respChan chan<- string) error
}

// Store defines the interface for database operations
type Store interface {
	SaveArticle(article *models.Article) error
	GetArticleByID(id string) (*models.Article, error)
	GetArticlesByTimeRange(startTime, endTime time.Time) ([]*models.Article, error)
	UpdateArticle(article *models.Article) error
}
