package db

import (
	"clank/internal/models"
	"context"
)

// Store defines the interface for database operations
type Store interface {
	// Article operations
	SaveArticle(ctx context.Context, article *models.Article) error
	GetArticle(ctx context.Context, id string) (*models.Article, error)
	DeleteArticle(ctx context.Context, id string) error

	// Entity operations
	SaveEntity(ctx context.Context, entity *models.ExtractedEntity) error
	GetEntity(ctx context.Context, id string) (*models.ExtractedEntity, error)
	DeleteEntity(ctx context.Context, id string) error

	// Relationship operations
	SaveRelationship(ctx context.Context, rel *models.ExtractedRelationship) error
	GetRelationship(ctx context.Context, id string) (*models.ExtractedRelationship, error)
	DeleteRelationship(ctx context.Context, id string) error

	// Analysis operations
	GetEntitiesByArticle(ctx context.Context, articleID string) ([]*models.ExtractedEntity, error)
	GetRelationshipsByArticle(ctx context.Context, articleID string) ([]*models.ExtractedRelationship, error)
}
