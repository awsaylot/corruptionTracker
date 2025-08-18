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

// SaveArticle stores an article and its extracted entities in Neo4j
// UpdateArticle updates an existing article in the database
func (s *ArticleStore) UpdateArticle(article *models.Article) error {
	session := s.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
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
			MATCH (a:Article {id: $id})
			SET a += {
				url: $url,
				title: $title,
				content: $content,
				source: $source,
				author: $author,
				publishDate: $publishDate,
				extractedAt: $extractedAt,
				metadata: $metadata
			}
			RETURN a`, params)

		if err != nil {
			return nil, fmt.Errorf("failed to update article: %v", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to execute update transaction: %v", err)
	}

	return nil
}

func (s *ArticleStore) SaveArticle(article *models.Article) error {
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

		// Process entities if present
		if article.Entities != nil {
			for _, entity := range article.Entities {
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
		}

		// Process relationships if present
		if article.Relations != nil {
			for _, rel := range article.Relations {
				params := map[string]interface{}{
					"id":          rel.ID,
					"type":        rel.Type,
					"fromId":      rel.FromID,
					"toId":        rel.ToID,
					"properties":  rel.Properties,
					"confidence":  rel.Confidence,
					"articleId":   article.ID,
					"extractedAt": rel.ExtractedAt.Format(time.RFC3339),
				}

				_, err := tx.Run(`
					MATCH (from:Entity {id: $fromId}), (to:Entity {id: $toId})
					MERGE (from)-[r:RELATES_TO {id: $id}]->(to)
					SET r += {
						type: $type,
						properties: $properties,
						confidence: $confidence,
						extractedAt: datetime($extractedAt)
					}
					WITH r
					MATCH (a:Article {id: $articleId})
					MERGE (a)-[:CONTAINS_RELATION]->(r)
				`, params)

				if err != nil {
					return nil, fmt.Errorf("failed to create relationship: %w", err)
				}
			}
		}

		return nil, nil
	})

	return err
}

// GetArticleByID retrieves an article by its ID
func (s *ArticleStore) GetArticleByID(id string) (*models.Article, error) {
	session := s.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		params := map[string]interface{}{
			"id": id,
		}

		records, err := tx.Run(`
			MATCH (a:Article {id: $id})
			OPTIONAL MATCH (a)-[:MENTIONS]->(e:Entity)
			OPTIONAL MATCH (a)-[:CONTAINS_RELATION]->(r:RELATES_TO)
			RETURN a, collect(e) as entities, collect(r) as relations
		`, params)

		if err != nil {
			return nil, fmt.Errorf("failed to query article: %w", err)
		}

		record, err := records.Single()
		if err != nil {
			return nil, fmt.Errorf("article not found: %w", err)
		}

		articleNode := record.GetByIndex(0).(neo4j.Node)
		article := &models.Article{
			ID:          id,
			URL:         articleNode.Props["url"].(string),
			Title:       articleNode.Props["title"].(string),
			Content:     articleNode.Props["content"].(string),
			Source:      articleNode.Props["source"].(string),
			Author:      articleNode.Props["author"].(string),
			PublishDate: parseTime(articleNode.Props["publishDate"].(string)),
			ExtractedAt: parseTime(articleNode.Props["extractedAt"].(string)),
			Metadata:    articleNode.Props["metadata"].(map[string]interface{}),
		}

		return article, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.Article), nil
}

// GetArticlesByTimeRange retrieves articles within a time range
func (s *ArticleStore) GetArticlesByTimeRange(startTime, endTime time.Time) ([]*models.Article, error) {
	session := s.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		params := map[string]interface{}{
			"startTime": startTime.Format(time.RFC3339),
			"endTime":   endTime.Format(time.RFC3339),
		}

		records, err := tx.Run(`
			MATCH (a:Article)
			WHERE a.publishDate >= datetime($startTime) AND a.publishDate <= datetime($endTime)
			RETURN a
			ORDER BY a.publishDate DESC
		`, params)

		if err != nil {
			return nil, fmt.Errorf("failed to query articles: %w", err)
		}

		var articles []*models.Article
		for records.Next() {
			record := records.Record()
			articleNode := record.GetByIndex(0).(neo4j.Node)
			article := &models.Article{
				ID:          articleNode.Props["id"].(string),
				URL:         articleNode.Props["url"].(string),
				Title:       articleNode.Props["title"].(string),
				Content:     articleNode.Props["content"].(string),
				Source:      articleNode.Props["source"].(string),
				Author:      articleNode.Props["author"].(string),
				PublishDate: parseTime(articleNode.Props["publishDate"].(string)),
				ExtractedAt: parseTime(articleNode.Props["extractedAt"].(string)),
				Metadata:    articleNode.Props["metadata"].(map[string]interface{}),
			}
			articles = append(articles, article)
		}

		return articles, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*models.Article), nil
}

// parseTime parses a time string in RFC3339 format
func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Now() // Default to current time on error
	}
	return t
}
