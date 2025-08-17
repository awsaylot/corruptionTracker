package models

import (
	"time"
)

// Article represents a news article and its extracted information
type Article struct {
	ID          string                 `json:"id"`
	URL         string                 `json:"url"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Source      string                 `json:"source"`
	Author      string                 `json:"author,omitempty"`
	PublishDate time.Time              `json:"publishDate"`
	ExtractedAt time.Time              `json:"extractedAt"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExtractedEntity represents an entity found in an article
type ExtractedEntity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Properties  map[string]interface{} `json:"properties"`
	Confidence  float64                `json:"confidence"`
	Mentions    []EntityMention        `json:"mentions"`
	ArticleID   string                 `json:"articleId"`
	ExtractedAt time.Time              `json:"extractedAt"`
}

// EntityMention represents a specific mention of an entity in the text
type EntityMention struct {
	Text     string `json:"text"`
	Context  string `json:"context"`
	Position struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"position"`
}

// ExtractedRelationship represents a relationship between entities found in an article
// ExtractionResult represents the structured output from article processing
type ExtractionResult struct {
	Entities      []ExtractedEntity       `json:"entities"`
	Relationships []ExtractedRelationship `json:"relationships"`
	Confidence    float64                 `json:"confidence"`
}

type ExtractedRelationship struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	FromID      string                 `json:"fromId"`
	ToID        string                 `json:"toId"`
	Properties  map[string]interface{} `json:"properties"`
	Confidence  float64                `json:"confidence"`
	Context     string                 `json:"context"`
	ArticleID   string                 `json:"articleId"`
	ExtractedAt time.Time              `json:"extractedAt"`
}

// ExtractionResult contains all information extracted from an article
type ExtractionResult struct {
	Article        Article                 `json:"article"`
	Entities       []ExtractedEntity       `json:"entities"`
	Relationships  []ExtractedRelationship `json:"relationships"`
	ProcessingTime time.Duration           `json:"processingTime"`
	Error          string                  `json:"error,omitempty"`
}
