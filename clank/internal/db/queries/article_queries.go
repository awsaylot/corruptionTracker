package queries

import (
	"fmt"

	"clank/internal/models"
)

// Article-related Cypher queries
const (
	// Store article and create initial nodes
	CreateArticleQuery = `
		CREATE (a:Article {
			id: $id,
			url: $url,
			title: $title,
			content: $content,
			source: $source,
			author: $author,
			publishDate: datetime($publishDate),
			extractedAt: datetime($extractedAt)
		})
		RETURN a
	`

	// Create extracted entity and link to article
	CreateEntityQuery = `
		MERGE (e:%s {id: $id})
		ON CREATE SET 
			e += $properties,
			e.name = $name,
			e.confidence = $confidence,
			e.extractedAt = datetime($extractedAt)
		ON MATCH SET
			e.confidence = CASE 
				WHEN $confidence > e.confidence THEN $confidence 
				ELSE e.confidence 
			END
		WITH e
		MATCH (a:Article {id: $articleId})
		MERGE (e)-[r:MENTIONED_IN]->(a)
		ON CREATE SET r.mentions = $mentions
		RETURN e
	`

	// Create relationship between entities
	CreateRelationshipQuery = `
		MATCH (from {id: $fromId})
		MATCH (to {id: $toId})
		MATCH (a:Article {id: $articleId})
		MERGE (from)-[r:%s]->(to)
		ON CREATE SET 
			r += $properties,
			r.confidence = $confidence,
			r.extractedAt = datetime($extractedAt)
		ON MATCH SET
			r.confidence = CASE 
				WHEN $confidence > r.confidence THEN $confidence 
				ELSE r.confidence 
			END
		MERGE (r)-[m:MENTIONED_IN]->(a)
		ON CREATE SET m.context = $context
		RETURN r
	`
)

// CreateArticleParams returns params for creating an article
func CreateArticleParams(article *models.Article) map[string]interface{} {
	return map[string]interface{}{
		"id":          article.ID,
		"url":         article.URL,
		"title":       article.Title,
		"content":     article.Content,
		"source":      article.Source,
		"author":      article.Author,
		"publishDate": article.PublishDate.Format("2006-01-02T15:04:05Z"),
		"extractedAt": article.ExtractedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateEntityParams returns params for creating an entity
func CreateEntityParams(entity *models.ExtractedEntity) map[string]interface{} {
	return map[string]interface{}{
		"id":          entity.ID,
		"name":        entity.Name,
		"properties":  entity.Properties,
		"confidence":  entity.Confidence,
		"mentions":    entity.Mentions,
		"articleId":   entity.ArticleID,
		"extractedAt": entity.ExtractedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GetCreateEntityQuery returns the query for creating an entity of a specific type
func GetCreateEntityQuery(entityType string) string {
	return fmt.Sprintf(CreateEntityQuery, entityType)
}

// CreateRelationshipParams returns params for creating a relationship
func CreateRelationshipParams(rel *models.ExtractedRelationship) map[string]interface{} {
	return map[string]interface{}{
		"fromId":      rel.FromID,
		"toId":        rel.ToID,
		"properties":  rel.Properties,
		"confidence":  rel.Confidence,
		"context":     rel.Context,
		"articleId":   rel.ArticleID,
		"extractedAt": rel.ExtractedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GetCreateRelationshipQuery returns the query for creating a relationship of a specific type
func GetCreateRelationshipQuery(relType string) string {
	return fmt.Sprintf(CreateRelationshipQuery, relType)
}
