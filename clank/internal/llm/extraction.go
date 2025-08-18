package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"clank/internal/interfaces"
	"clank/internal/models"
)

// ProcessArticle sends an article to the LLM for entity and relationship extraction
func (c *Client) ProcessArticle(ctx context.Context, article *models.Article) (*models.ExtractionResult, error) {
	prompt := fmt.Sprintf(`Analyze the following article and extract entities and relationships related to corruption:

Title: %s
Source: %s
Date: %s
Content:
%s

Extract the following information:
1. People involved (names, roles, organizations)
2. Organizations mentioned (companies, government agencies, NGOs)
3. Money or value amounts mentioned
4. Locations relevant to the corruption
5. Time periods or dates
6. Relationships between entities (who paid whom, who is affiliated with what)

Format your response as a valid JSON object with the following structure:
{
  "entities": [
    {
      "id": "string",
      "type": "person|organization|location|money|time",
      "name": "string",
      "properties": {},
      "confidence": 0.0-1.0,
      "mentions": [
        {
          "text": "exact text from article",
          "context": "surrounding sentence"
        }
      ]
    }
  ],
  "relationships": [
    {
      "id": "string", 
      "type": "payment|affiliation|ownership|involvement",
      "fromId": "entity_id",
      "toId": "entity_id",
      "properties": {},
      "confidence": 0.0-1.0,
      "context": "relevant quote from article"
    }
  ],
  "confidence": 0.0-1.0
}`,
		article.Title,
		article.Source,
		article.PublishDate.Format("2006-01-02"),
		article.Content,
	)

	// Create completion request
	messages := []interfaces.Message{
		{
			Role:    "system",
			Content: "You are a precise entity extraction system specializing in analyzing corruption-related news articles.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Convert interface messages to llm.Message
	llmMessages := make([]Message, len(messages))
	for i, msg := range messages {
		llmMessages[i] = Message{
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: time.Now(),
		}
	}

	// Send request to LLM
	resp, err := c.Generate(ctx, llmMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to process article: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("LLM error: %s", resp.Error)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	var content string
	if resp.Choices[0].Content != "" {
		content = resp.Choices[0].Content
	} else {
		content = resp.Choices[0].Message.Content
	}

	// Parse LLM response
	var result models.ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Set extraction timestamp for all entities and relationships
	now := time.Now()
	for i := range result.Entities {
		result.Entities[i].ArticleID = article.ID
		result.Entities[i].ExtractedAt = now
	}

	for i := range result.Relationships {
		result.Relationships[i].ArticleID = article.ID
		result.Relationships[i].ExtractedAt = now
	}

	return &result, nil
}
