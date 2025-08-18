package extraction

import (
	"clank/internal/models"
	"context"
)

// Scraper defines the interface for article scraping operations
type Scraper interface {
	ScrapeArticle(ctx context.Context, url string) (*models.Article, error)
	ExtractMetadata(ctx context.Context, url string) (map[string]interface{}, error)
}

// Processor defines the interface for content processing
type Processor interface {
	CleanContent(content string) (string, error)
	ValidateContent(content string) error
	ExtractMainText(content string) (string, error)
}

// Analyzer defines the interface for article analysis
type Analyzer interface {
	AnalyzeContent(ctx context.Context, article *models.Article) (*models.AnalysisResult, error)
	GenerateHypotheses(ctx context.Context, article *models.Article) ([]*models.Hypothesis, error)
	BuildEvidenceChain(ctx context.Context, article *models.Article) (*models.EvidenceChain, error)
}

// Pipeline defines the interface for the complete extraction pipeline
type Pipeline interface {
	Process(ctx context.Context, url string, depth int) (*models.ExtractionResult, error)
	GetProgress(ctx context.Context, sessionID string) (*models.ProcessingStatus, error)
	Terminate(ctx context.Context, sessionID string) error
}
