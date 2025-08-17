package db

import (
	"fmt"
	"time"

	"clank/internal/models"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// ArticleStore handles Neo4j operations for articles and extracted data
type ArticleStore struct {
	driver neo4j.Driver
}

// NewArticleStore creates a new article store
func NewArticleStore() *ArticleStore {
	return &ArticleStore{
		driver: driver,
	}
}

// StoreArticle stores an article and its extracted entities in Neo4j
func (s *ArticleStore) StoreArticle(article *models.Article, extraction *models.ExtractionResult) error {
	session := s.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// Create article node
		params := map[string]interface{}{
			"id":          article.ID,
			"url":         article.URL,
			"title":       article.Title,
			"content":     article.Content,
			"source":      article.Source,
			"author":      article.Author,
			"publishDate": article.PublishDate.Format(time.RFC3339),
			"extractedAt": article.ExtractedAt.Format(time.RFC3339),
			"metadata":    article.Metadata,
		}

		_, err := tx.Run(`
			MERGE (a:Article {id: $id})
			SET a += {
				url: $url,
				title: $title,
				content: $content,
				source: $source,
				author: $author,
				publishDate: datetime($publishDate),
				extractedAt: datetime($extractedAt),
				metadata: $metadata
			}
		`, params)
		if err != nil {
			return nil, fmt.Errorf("failed to create article node: %w", err)
		}

		// Create entity nodes and relationships
		for _, entity := range extraction.Entities {
			params := map[string]interface{}{
				"id":          entity.ID,
				"type":        entity.Type,
				"name":        entity.Name,
				"properties":  entity.Properties,
				"confidence":  entity.Confidence,
				"articleId":   article.ID,
				"extractedAt": entity.ExtractedAt.Format(time.RFC3339),
			}

			_, err := tx.Run(`
				MERGE (e:Entity {id: $id})
				SET e += {
					type: $type,
					name: $name,
					properties: $properties,
					confidence: $confidence,
					extractedAt: datetime($extractedAt)
				}
				WITH e
				MATCH (a:Article {id: $articleId})
				MERGE (a)-[r:MENTIONS]->(e)
				SET r.confidence = $confidence
			`, params)
			if err != nil {
				return nil, fmt.Errorf("failed to create entity node: %w", err)
			}

			// Store entity mentions
			for _, mention := range entity.Mentions {
				params["mentionText"] = mention.Text
				params["mentionContext"] = mention.Context
				params["startPos"] = mention.Position.Start
				params["endPos"] = mention.Position.End

				_, err := tx.Run(`
					MATCH (e:Entity {id: $id})
					CREATE (m:Mention {
						text: $mentionText,
						context: $mentionContext,
						position: {start: $startPos, end: $endPos}
					})-[:IN]->(e)
				`, params)
				if err != nil {
					return nil, fmt.Errorf("failed to create mention: %w", err)
				}
			}
		}

		// Create relationships between entities
		for _, rel := range extraction.Relationships {
			params := map[string]interface{}{
				"id":          rel.ID,
				"type":        rel.Type,
				"fromId":      rel.FromID,
				"toId":        rel.ToID,
				"properties":  rel.Properties,
				"confidence":  rel.Confidence,
				"context":     rel.Context,
				"articleId":   article.ID,
				"extractedAt": rel.ExtractedAt.Format(time.RFC3339),
			}

			_, err := tx.Run(`
				MATCH (from:Entity {id: $fromId}), (to:Entity {id: $toId})
				CREATE (from)-[r:RELATES_TO {
					id: $id,
					type: $type,
					properties: $properties,
					confidence: $confidence,
					context: $context,
					articleId: $articleId,
					extractedAt: datetime($extractedAt)
				}]->(to)
			`, params)
			if err != nil {
				return nil, fmt.Errorf("failed to create relationship: %w", err)
			}
		}

		return nil, nil
	})

	return err
}
