package prompts

import (
	"context"
	"testing"

	"clank/internal/models"
	"clank/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticleExtractionPrompt_Generate(t *testing.T) {
	tests := []struct {
		name        string
		article     *models.Article
		setupMocks  func(*testutil.MockLLMClient)
		expectError bool
		checkResult func(*testing.T, []models.ExtractedEntity, []models.ExtractedRelationship)
	}{
		{
			name: "successful extraction",
			article: &models.Article{
				URL:     "https://example.com/article",
				Title:   "Company A in Corruption Scandal",
				Content: "Company A CEO John Smith allegedly paid $1M to Official Jane Doe for contract approval.",
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{
					"entities": [
						{
							"name": "Company A",
							"type": "company",
							"properties": {"industry": "unknown"}
						},
						{
							"name": "John Smith",
							"type": "person",
							"properties": {"role": "CEO", "company": "Company A"}
						},
						{
							"name": "Jane Doe",
							"type": "person",
							"properties": {"role": "government official"}
						}
					],
					"relationships": [
						{
							"from": "John Smith",
							"to": "Company A",
							"type": "WORKS_FOR",
							"properties": {"position": "CEO"}
						},
						{
							"from": "Company A",
							"to": "Jane Doe",
							"type": "PAID",
							"properties": {"amount": "1000000", "currency": "USD"}
						}
					]
				}`
			},
			expectError: false,
			checkResult: func(t *testing.T, entities []models.ExtractedEntity, relationships []models.ExtractedRelationship) {
				// Check entities
				require.Len(t, entities, 3)
				assert.Equal(t, "Company A", entities[0].Name)
				if props := entities[0].Properties; props != nil {
					assert.Equal(t, "unknown", props["industry"])
				}
				assert.Equal(t, "John Smith", entities[1].Name)
				if props := entities[1].Properties; props != nil {
					assert.Equal(t, "CEO", props["role"])
				}
				assert.Equal(t, "Jane Doe", entities[2].Name)

				// Check relationships
				require.Len(t, relationships, 2)
				assert.Equal(t, "WORKS_FOR", relationships[0].Type)
				assert.Equal(t, "PAID", relationships[1].Type)

				// Check properties
				assert.Equal(t, "CEO", entities[1].Properties["role"])
				assert.Equal(t, "1000000", relationships[1].Properties["amount"])
			},
		},
		{
			name: "empty article content",
			article: &models.Article{
				URL:     "https://example.com/article",
				Title:   "Empty Article",
				Content: "",
			},
			setupMocks:  func(mock *testutil.MockLLMClient) {},
			expectError: true,
		},
		{
			name: "llm error",
			article: &models.Article{
				URL:     "https://example.com/article",
				Title:   "Test Article",
				Content: "Valid content",
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateError = assert.AnError
			},
			expectError: true,
		},
		{
			name: "invalid response format",
			article: &models.Article{
				URL:     "https://example.com/article",
				Title:   "Test Article",
				Content: "Valid content",
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{"invalid": "format"}`
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := testutil.NewMockLLMClient()
			tt.setupMocks(mockLLM)

			prompt := NewArticleExtractionPrompt()
			entities, relationships, err := prompt.Generate(context.Background(), tt.article, mockLLM)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, entities, relationships)
			}
		})
	}
}

func TestArticleExtractionPrompt_ValidateResults(t *testing.T) {
	tests := []struct {
		name          string
		entities      []models.ExtractedEntity
		relationships []models.ExtractedRelationship
		expectError   bool
	}{
		{
			name: "valid results",
			entities: []models.ExtractedEntity{
				{
					Name: "Company A",
					Type: "company",
					Properties: map[string]interface{}{
						"industry": "technology",
					},
				},
				{
					Name: "John Smith",
					Type: "person",
					Properties: map[string]interface{}{
						"role": "CEO",
					},
				},
			},
			relationships: []models.ExtractedRelationship{
				{
					FromID: "1",
					ToID:   "2",
					Type:   "WORKS_FOR",
					Properties: map[string]interface{}{
						"position": "CEO",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing entity type",
			entities: []models.Entity{
				{
					Name: "Invalid Entity",
					Type: "",
				},
			},
			relationships: []models.Relationship{},
			expectError:   true,
		},
		{
			name: "invalid relationship type",
			entities: []models.Entity{
				{Name: "Entity A", Type: "person"},
				{Name: "Entity B", Type: "person"},
			},
			relationships: []models.Relationship{
				{
					FromID: "1",
					ToID:   "2",
					Type:   "INVALID_TYPE",
				},
			},
			expectError: true,
		},
		{
			name:          "empty results",
			entities:      []models.Entity{},
			relationships: []models.Relationship{},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := NewArticleExtractionPrompt()
			err := prompt.ValidateResults(tt.entities, tt.relationships)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestArticleExtractionPrompt_Enrichment(t *testing.T) {
	tests := []struct {
		name        string
		article     *models.Article
		setupMocks  func(*testutil.MockLLMClient)
		expectError bool
		checkResult func(*testing.T, *models.Article)
	}{
		{
			name: "successful enrichment",
			article: &models.Article{
				URL:     "https://example.com/article",
				Title:   "Test Article",
				Content: "Basic content",
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{
					"summary": "Enriched summary",
					"topics": ["corruption", "government"],
					"sentiment": "negative",
					"risk_score": 0.85
				}`
			},
			expectError: false,
			checkResult: func(t *testing.T, article *models.Article) {
				assert.Contains(t, article.Properties, "summary")
				assert.Contains(t, article.Properties, "topics")
				assert.Contains(t, article.Properties, "sentiment")
				assert.Contains(t, article.Properties, "risk_score")

				topics, ok := article.Properties["topics"].([]interface{})
				require.True(t, ok)
				assert.Contains(t, topics, "corruption")
			},
		},
		{
			name: "enrichment error",
			article: &models.Article{
				URL:     "https://example.com/article",
				Title:   "Test Article",
				Content: "Basic content",
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateError = assert.AnError
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := testutil.NewMockLLMClient()
			tt.setupMocks(mockLLM)

			prompt := NewArticleExtractionPrompt()
			err := prompt.EnrichArticle(context.Background(), tt.article, mockLLM)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, tt.article)
			}
		})
	}
}
