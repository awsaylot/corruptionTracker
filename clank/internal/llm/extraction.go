package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a precise entity extraction system specializing in analyzing corruption-related news articles.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Send request to LLM
	resp, err := c.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to process article: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse LLM response
	var result models.ExtractionResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
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

// sendRequest is a helper method for making HTTP requests
func (c *Client) sendRequest(ctx context.Context, req *GenerateRequest, result *GenerateResponse) error {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
