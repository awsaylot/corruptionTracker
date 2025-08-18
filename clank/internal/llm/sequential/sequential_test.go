package sequential

import (
	"context"
	"testing"
	"time"

	"clank/internal/llm"
	"clank/internal/models"
	"clank/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisController_StartAnalysis(t *testing.T) {
	tests := []struct {
		name           string
		config         *AnalysisConfig
		mockResponses  []string
		mockErrors     []error
		expectedStages int
		expectError    bool
	}{
		{
			name: "successful analysis",
			config: &AnalysisConfig{
				Depth:                3,
				MaxStages:            5,
				ConfidenceThreshold:  0.7,
				TimeoutPerStage:      time.Second,
				EnableCrossReference: true,
				EnableHypotheses:     true,
			},
			mockResponses: []string{
				`{"result": {"entities": [{"type": "person", "name": "John Doe"}]}}`,
				`{"result": {"relationships": [{"type": "owns", "from": "John Doe", "to": "Company A"}]}}`,
				`{"result": {"insights": ["Found suspicious pattern"]}}`,
			},
			expectedStages: 3,
		},
		{
			name: "analysis with error",
			config: &AnalysisConfig{
				Depth:               2,
				MaxStages:           3,
				ConfidenceThreshold: 0.7,
				TimeoutPerStage:     time.Second,
			},
			mockResponses: []string{`{"result": {"entities": []}}`},
			mockErrors:    []error{assert.AnError},
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockLLM := testutil.NewMockLLMClient()

			// Setup mock responses
			mockLLM.Responses = tt.mockResponses
			if len(tt.mockErrors) > 0 {
				mockLLM.GenerateError = tt.mockErrors[0]
			}

			// Create controller with mock
			controller := NewAnalysisController(&llm.Client{})

			// Create test article
			article := testutil.MockArticle("https://example.com", "Test Article", "Test content")

			// Start analysis
			session, err := controller.StartAnalysis(context.Background(), article, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, session)
			assert.Equal(t, tt.expectedStages, len(session.Stages))
		})
	}
}

func TestAnalysisStages_Process(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		mockError    error
		expectError  bool
		checkStage   func(*testing.T, *AnalysisStage)
	}{
		{
			name: "surface extraction stage",
			mockResponse: `{
				"result": {
					"entities": [
						{"type": "person", "name": "John Doe"},
						{"type": "company", "name": "Acme Corp"}
					]
				}
			}`,
			checkStage: func(t *testing.T, stage *AnalysisStage) {
				require.NotNil(t, stage.Results)
				assert.Len(t, stage.Results.Entities, 2)
				assert.Equal(t, "John Doe", stage.Results.Entities[0].Name)
				assert.Equal(t, "Acme Corp", stage.Results.Entities[1].Name)
			},
		},
		{
			name: "deep analysis stage",
			mockResponse: `{
				"result": {
					"relationships": [
						{"type": "owns", "from": "John Doe", "to": "Acme Corp"}
					]
				}
			}`,
			checkStage: func(t *testing.T, stage *AnalysisStage) {
				require.NotNil(t, stage.Results)
				assert.Len(t, stage.Results.Relationships, 1)
				assert.Equal(t, "owns", stage.Results.Relationships[0].Type)
			},
		},
		{
			name:        "stage with error",
			mockError:   assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock LLM client
			mockLLM := testutil.NewMockLLMClient()
			mockLLM.GenerateResponse = tt.mockResponse
			mockLLM.GenerateError = tt.mockError

			// Create test article
			article := testutil.MockArticle("https://example.com", "Test Article", "Test content")

			// Create analysis session and stage
			session := &AnalysisSession{
				ID:        "test-session",
				ArticleID: article.ID,
				Status:    "running",
				StartedAt: time.Now(),
				Config:    &AnalysisConfig{},
				Results:   make([]*models.ExtractionResult, 0),
			}

			stage := &AnalysisStage{
				Stage:       1,
				Name:        tt.name,
				Description: "Test stage",
				Status:      "pending",
			}

			// Create processor
			processor := NewSurfaceExtractionStage().WithLLMClient(&llm.Client{})

			// Process stage
			err := processor.Process(context.Background(), session, stage, article, nil)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkStage != nil {
				tt.checkStage(t, stage)
			}
		})
	}
}
