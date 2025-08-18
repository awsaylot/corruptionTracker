package testutil

import (
	"clank/internal/models"
	"context"
	"time"
)

// MockDB is a mock implementation of the database interface
type MockDB struct {
	Articles      map[string]*models.Article
	Entities      map[string]*models.ExtractedEntity
	Relationships map[string]*models.ExtractedRelationship
	SaveError     error
	RetrieveError error
	DeleteError   error
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{
		Articles:      make(map[string]*models.Article),
		Entities:      make(map[string]*models.ExtractedEntity),
		Relationships: make(map[string]*models.ExtractedRelationship),
	}
}

// SaveArticle mocks saving an article
func (m *MockDB) SaveArticle(ctx context.Context, article *models.Article) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.Articles[article.ID] = article
	return nil
}

// GetArticleByID mocks retrieving an article by ID
func (m *MockDB) GetArticleByID(id string) (*models.Article, error) {
	if m.RetrieveError != nil {
		return nil, m.RetrieveError
	}
	article, exists := m.Articles[id]
	if !exists {
		return nil, nil
	}
	return article, nil
}

// GetArticle mocks retrieving an article - deprecated, use GetArticleByID instead
func (m *MockDB) GetArticle(ctx context.Context, id string) (*models.Article, error) {
	return m.GetArticleByID(id)
}

// GetArticlesByTimeRange mocks retrieving articles within a time range
func (m *MockDB) GetArticlesByTimeRange(startTime, endTime time.Time) ([]*models.Article, error) {
	if m.RetrieveError != nil {
		return nil, m.RetrieveError
	}

	var articles []*models.Article
	for _, article := range m.Articles {
		if article.CreatedAt.After(startTime) && article.CreatedAt.Before(endTime) {
			articles = append(articles, article)
		}
	}
	return articles, nil
}

// GetArticlesByTimeRange retrieves articles within a time range
func (m *MockDB) GetArticlesByTimeRange(startTime, endTime time.Time) ([]*models.Article, error) {
	if m.RetrieveError != nil {
		return nil, m.RetrieveError
	}

	var articles []*models.Article
	for _, article := range m.Articles {
		if article.CreatedAt.After(startTime) && article.CreatedAt.Before(endTime) {
			articles = append(articles, article)
		}
	}
	return articles, nil
}

// SaveEntity mocks saving an entity
func (m *MockDB) SaveEntity(ctx context.Context, entity *models.ExtractedEntity) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.Entities[entity.ID] = entity
	return nil
}

// GetEntity mocks retrieving an entity
func (m *MockDB) GetEntity(ctx context.Context, id string) (*models.ExtractedEntity, error) {
	if m.RetrieveError != nil {
		return nil, m.RetrieveError
	}
	entity, exists := m.Entities[id]
	if !exists {
		return nil, nil
	}
	return entity, nil
}

// SaveRelationship mocks saving a relationship
func (m *MockDB) SaveRelationship(ctx context.Context, rel *models.ExtractedRelationship) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.Relationships[rel.ID] = rel
	return nil
}

// GetRelationship mocks retrieving a relationship
func (m *MockDB) GetRelationship(ctx context.Context, id string) (*models.ExtractedRelationship, error) {
	if m.RetrieveError != nil {
		return nil, m.RetrieveError
	}
	rel, exists := m.Relationships[id]
	if !exists {
		return nil, nil
	}
	return rel, nil
}

// DeleteArticle mocks deleting an article
func (m *MockDB) DeleteArticle(ctx context.Context, id string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	delete(m.Articles, id)
	return nil
}

// DeleteEntity mocks deleting an entity
func (m *MockDB) DeleteEntity(ctx context.Context, id string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	delete(m.Entities, id)
	return nil
}

// DeleteRelationship mocks deleting a relationship
func (m *MockDB) DeleteRelationship(ctx context.Context, id string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	delete(m.Relationships, id)
	return nil
}
