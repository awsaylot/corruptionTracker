package prompts

import (
	"context"
	"encoding/json"
	"fmt"

	"clank/internal/interfaces"
	"clank/internal/models"
)

// ArticleExtractionPrompt represents a prompt for extracting information from news articles
type ArticleExtractionPrompt struct{}

// NewArticleExtractionPrompt creates a new ArticleExtractionPrompt instance
func NewArticleExtractionPrompt() *ArticleExtractionPrompt {
	return &ArticleExtractionPrompt{}
}

// Generate generates entity and relationship data from an article
func (p *ArticleExtractionPrompt) Generate(ctx context.Context, article *models.Article, llm interfaces.LLMProvider) ([]models.ExtractedEntity, []models.ExtractedRelationship, error) {
	if article == nil {
		return nil, nil, fmt.Errorf("article is nil")
	}

	messages := []interfaces.Message{
		{
			Role:    "system",
			Content: "You are an expert entity extraction system for analyzing corruption-related news articles.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf(promptTemplate, article.Content),
		},
	}

	resp, err := llm.Generate(ctx, messages)
	if err != nil {
		return nil, nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	var result struct {
		Entities      []models.ExtractedEntity
		Relationships []models.ExtractedRelationship
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result.Entities, result.Relationships, nil
}

const (
	// promptTemplate is the main template for extracting entities and relationships
	promptTemplate = `Extract entities and relationships from the following news article. Focus on people, organizations, locations, and events related to corruption or misconduct.

Please analyze this article content and provide information in JSON format:
%s

Response should be in this format:
{
  "entities": [
    {
      "type": "PERSON | ORGANIZATION | LOCATION",
      "name": "string",
      "properties": {
        "role": "string",
        "sector": "string",
        "details": "string"
      },
      "confidence": 0.95,
      "mentions": [
        {
          "text": "exact text from article",
          "context": "surrounding sentence or paragraph"
        }
      ]
    }
  ],
  "relationships": [
    {
      "type": "INVOLVED_IN | ACCUSED_OF | CONNECTED_TO",
      "from": "entity name",
      "to": "entity name",
      "properties": {
        "date": "string",
        "amount": "number",
        "details": "string"
      },
      "confidence": 0.95,
      "context": "relevant quote from article"
    }
  ]
}`

	// validationTemplate is used to validate extracted entities
	validationTemplate = `Validate these extracted entities:
%s

Original text:
%s

For each entity:
1. Verify name accuracy
2. Confirm role and properties
3. Check for missing important information
4. Evaluate confidence score

Response format:
{
  "validations": [
    {
      "entityId": "string",
      "isValid": true,
      "confidence": 0.95,
      "suggestedCorrections": {},
      "missingInfo": []
    }
  ]
}`
)

// ValidateResults validates extracted entities and relationships
func (p *ArticleExtractionPrompt) ValidateResults(entities []models.ExtractedEntity, relationships []models.ExtractedRelationship) error {
	if len(entities) == 0 {
		return fmt.Errorf("no entities extracted")
	}

	for _, entity := range entities {
		if entity.Type == "" {
			return fmt.Errorf("entity '%s' has no type", entity.Name)
		}
		if entity.Name == "" {
			return fmt.Errorf("found entity with empty name")
		}
	}

	for _, rel := range relationships {
		if rel.Type == "" {
			return fmt.Errorf("relationship between '%s' and '%s' has no type", rel.FromID, rel.ToID)
		}
		if rel.FromID == "" || rel.ToID == "" {
			return fmt.Errorf("relationship missing FromID or ToID")
		}
	}

	return nil
}

// EnrichArticle enriches an article with additional metadata
func (p *ArticleExtractionPrompt) EnrichArticle(ctx context.Context, article *models.Article, llm interfaces.LLMProvider) error {
	if article == nil {
		return fmt.Errorf("article is nil")
	}

	prompt := fmt.Sprintf(`Analyze the following article and provide enriched metadata:

Title: %s
Content: %s

Please provide:
1. A brief summary
2. Key topics/themes
3. Overall sentiment
4. Corruption risk score (0-1)

Format response as JSON:
{
  "summary": "string",
  "topics": ["topic1", "topic2"],
  "sentiment": "positive|neutral|negative",
  "risk_score": float
}`, article.Title, article.Content)

	// Process with LLM
	messages := []interfaces.Message{
		{
			Role:    "system",
			Content: "You are an expert in analyzing news articles for corruption-related content.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := llm.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to enrich article: %w", err)
	}

	if resp == "" {
		return fmt.Errorf("no response from LLM")
	}

	var enrichment map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &enrichment); err != nil {
		return fmt.Errorf("failed to parse enrichment data: %w", err)
	}

	// Update article metadata
	if article.Metadata == nil {
		article.Metadata = make(map[string]interface{})
	}
	for k, v := range enrichment {
		article.Metadata[k] = v
	}

	return nil
}
