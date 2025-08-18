package testutil

import (
	"clank/internal/models"
	"time"

	"github.com/google/uuid"
)

// MockArticle creates a test article with given parameters
func MockArticle(url, title, content string) *models.Article {
	return &models.Article{
		ID:          uuid.New().String(),
		URL:         url,
		Title:       title,
		Content:     content,
		Source:      "Test Source",
		Author:      "Test Author",
		PublishDate: time.Now(),
		ExtractedAt: time.Now(),
		Metadata:    map[string]interface{}{"test": true},
	}
}

// MockEntity creates a test entity with given parameters
func MockEntity(entityType, name string) *models.ExtractedEntity {
	return &models.ExtractedEntity{
		ID:   uuid.New().String(),
		Type: entityType,
		Name: name,
		Properties: map[string]interface{}{
			"test": true,
		},
		Confidence: 0.95,
		Mentions: []models.EntityMention{
			{
				Text:    name,
				Context: "Test context containing " + name,
				Position: struct {
					Start int `json:"start"`
					End   int `json:"end"`
				}{
					Start: 0,
					End:   len(name),
				},
			},
		},
		ArticleID:   uuid.New().String(),
		ExtractedAt: time.Now(),
	}
}

// MockBrowserAutomation is a mock implementation of browser automation for testing
type MockBrowserAutomation struct {
	ScrapeResponse string
	InitializeErr  error
	ScrapeErr      error
}

// NewMockBrowserAutomation creates a new mock browser automation with default values
func NewMockBrowserAutomation() *MockBrowserAutomation {
	return &MockBrowserAutomation{
		ScrapeResponse: "Test content",
	}
}

// Initialize implements browser.Browser
func (m *MockBrowserAutomation) Initialize() error {
	return m.InitializeErr
}

// ScrapeArticle implements browser.Browser
func (m *MockBrowserAutomation) ScrapeArticle(url string) (*models.Article, error) {
	if m.ScrapeErr != nil {
		return nil, m.ScrapeErr
	}

	return &models.Article{
		ID:          uuid.New().String(),
		URL:         url,
		Content:     m.ScrapeResponse,
		Title:       "Test Article",
		Source:      "Test Source",
		Author:      "Test Author",
		PublishDate: time.Now(),
		ExtractedAt: time.Now(),
		Metadata:    map[string]interface{}{"test": true},
	}, nil
}

// MockRelationship creates a test relationship between two entities
func MockRelationship(fromID, toID, relType string) *models.ExtractedRelationship {
	return &models.ExtractedRelationship{
		ID:         uuid.New().String(),
		Type:       relType,
		FromID:     fromID,
		ToID:       toID,
		Properties: map[string]interface{}{"test": true},
		Confidence: 0.9,
		Context:    "Test relationship context",
	}
}

// MockEvidence creates a test evidence
func MockEvidence(entityID string) *models.Evidence {
	return &models.Evidence{
		ID:         uuid.New().String(),
		Type:       "entity",
		Stage:      1,
		EntityID:   entityID,
		Text:       "Test evidence text",
		Context:    "Test evidence context",
		Source:     "Test source",
		Confidence: 0.85,
		CreatedAt:  time.Now(),
		Metadata:   map[string]interface{}{"test": true},
	}
}

// MockHypothesis creates a test hypothesis
func MockHypothesis() *models.Hypothesis {
	return &models.Hypothesis{
		ID:          uuid.New().String(),
		Stage:       2,
		Description: "Test hypothesis",
		Evidence:    []string{uuid.New().String(), uuid.New().String()},
		Confidence:  0.75,
		Status:      "proposed",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"test": true},
	}
}
