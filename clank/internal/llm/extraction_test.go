package llm

import (
	"context"
	"testing"

	"clank/internal/models"
	"clank/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractEntities(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		setupMocks  func(*testutil.MockLLMClient)
		expectError bool
		expected    []*models.ExtractedEntity
	}{
		{
			name: "successful entity extraction",
			text: "John Smith from Acme Corp paid $500,000 to Official Jane",
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{
					"entities": [
						{
							"name": "John Smith",
							"type": "person",
							"properties": {
								"role": "executive",
								"company": "Acme Corp"
							}
						},
						{
							"name": "Acme Corp",
							"type": "company",
							"properties": {
								"industry": "technology"
							}
						},
						{
							"name": "Official Jane",
							"type": "person",
							"properties": {
								"role": "government official"
							}
						}
					]
				}`
			},
			expectError: false,
			expected: []*models.Entity{
				{
					Name: "John Smith",
					Type: "person",
					Properties: map[string]interface{}{
						"role":    "executive",
						"company": "Acme Corp",
					},
				},
				{
					Name: "Acme Corp",
					Type: "company",
					Properties: map[string]interface{}{
						"industry": "technology",
					},
				},
				{
					Name: "Official Jane",
					Type: "person",
					Properties: map[string]interface{}{
						"role": "government official",
					},
				},
			},
		},
		{
			name:        "empty text",
			text:        "",
			setupMocks:  func(mock *testutil.MockLLMClient) {},
			expectError: true,
		},
		{
			name: "llm error",
			text: "Valid text",
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateError = assert.AnError
			},
			expectError: true,
		},
		{
			name: "invalid response format",
			text: "Valid text",
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

			entities, err := ExtractEntities(context.Background(), tt.text, mockLLM)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.expected), len(entities))

			for i, expected := range tt.expected {
				assert.Equal(t, expected.Name, entities[i].Name)
				assert.Equal(t, expected.Type, entities[i].Type)
				assert.Equal(t, expected.Properties, entities[i].Properties)
			}
		})
	}
}

func TestExtractRelationships(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		entities    []*models.Entity
		setupMocks  func(*testutil.MockLLMClient)
		expectError bool
		expected    []*models.Relationship
	}{
		{
			name: "successful relationship extraction",
			text: "John Smith from Acme Corp paid $500,000 to Official Jane",
			entities: []*models.Entity{
				{ID: "1", Name: "John Smith", Type: "person"},
				{ID: "2", Name: "Acme Corp", Type: "company"},
				{ID: "3", Name: "Official Jane", Type: "person"},
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{
					"relationships": [
						{
							"from": "1",
							"to": "2",
							"type": "WORKS_FOR",
							"properties": {}
						},
						{
							"from": "1",
							"to": "3",
							"type": "PAID",
							"properties": {
								"amount": "500000",
								"currency": "USD",
								"date": "2025-08-17"
							}
						}
					]
				}`
			},
			expectError: false,
			expected: []*models.Relationship{
				{
					FromID:     "1",
					ToID:       "2",
					Type:       "WORKS_FOR",
					Properties: map[string]interface{}{},
				},
				{
					FromID: "1",
					ToID:   "3",
					Type:   "PAID",
					Properties: map[string]interface{}{
						"amount":   "500000",
						"currency": "USD",
						"date":     "2025-08-17",
					},
				},
			},
		},
		{
			name:        "no entities provided",
			text:        "Valid text",
			entities:    []*models.Entity{},
			setupMocks:  func(mock *testutil.MockLLMClient) {},
			expectError: true,
		},
		{
			name: "llm error",
			text: "Valid text",
			entities: []*models.Entity{
				{ID: "1", Name: "Entity", Type: "person"},
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

			relationships, err := ExtractRelationships(context.Background(), tt.text, tt.entities, mockLLM)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.expected), len(relationships))

			for i, expected := range tt.expected {
				assert.Equal(t, expected.FromID, relationships[i].FromID)
				assert.Equal(t, expected.ToID, relationships[i].ToID)
				assert.Equal(t, expected.Type, relationships[i].Type)
				assert.Equal(t, expected.Properties, relationships[i].Properties)
			}
		})
	}
}

func TestValidateEntityType(t *testing.T) {
	tests := []struct {
		name      string
		entity    *models.Entity
		expectErr bool
	}{
		{
			name: "valid person type",
			entity: &models.Entity{
				Name: "John Doe",
				Type: "person",
			},
			expectErr: false,
		},
		{
			name: "valid company type",
			entity: &models.Entity{
				Name: "Acme Corp",
				Type: "company",
			},
			expectErr: false,
		},
		{
			name: "valid organization type",
			entity: &models.Entity{
				Name: "NGO",
				Type: "organization",
			},
			expectErr: false,
		},
		{
			name: "invalid type",
			entity: &models.Entity{
				Name: "Invalid",
				Type: "invalid_type",
			},
			expectErr: true,
		},
		{
			name: "empty type",
			entity: &models.Entity{
				Name: "No Type",
				Type: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEntityType(tt.entity)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRelationshipType(t *testing.T) {
	tests := []struct {
		name         string
		relationship *models.Relationship
		expectErr    bool
	}{
		{
			name: "valid ownership type",
			relationship: &models.Relationship{
				Type: "OWNS",
			},
			expectErr: false,
		},
		{
			name: "valid payment type",
			relationship: &models.Relationship{
				Type: "PAID",
			},
			expectErr: false,
		},
		{
			name: "valid employment type",
			relationship: &models.Relationship{
				Type: "WORKS_FOR",
			},
			expectErr: false,
		},
		{
			name: "invalid type",
			relationship: &models.Relationship{
				Type: "INVALID_TYPE",
			},
			expectErr: true,
		},
		{
			name: "empty type",
			relationship: &models.Relationship{
				Type: "",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRelationshipType(tt.relationship)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
