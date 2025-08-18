package handlers

import (
	"bytes"
	"clank/config"
	"clank/internal/models"
	"clank/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractionHandler_HandleURLExtraction(t *testing.T) {
	// Create test configurations
	cfg := &config.Config{}
	cfg.LLM.URL = "http://localhost:8080"
	cfg.LLM.Model = "test-model"

	tests := []struct {
		name           string
		requestBody    ExtractionRequest
		mockSetup      func(*testutil.MockLLMClient, *testutil.MockDB, *testutil.MockBrowserAutomation)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful extraction",
			requestBody: ExtractionRequest{
				URL:   "https://example.com/article",
				Depth: 3,
			},
			mockSetup: func(llm *testutil.MockLLMClient, db *testutil.MockDB, browser *testutil.MockBrowserAutomation) {
				browser.ScrapeResponse = "Test article content"
				llm.GenerateResponse = `{
					"entities": [
						{
							"type": "person",
							"name": "John Doe",
							"confidence": 0.95
						}
					]
				}`
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var resp ExtractionResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.SessionID)
				assert.NotNil(t, resp.Article)
				assert.Contains(t, resp.Article.Content, "Test article content")
			},
		},
		{
			name: "invalid url",
			requestBody: ExtractionRequest{
				URL:   "not-a-url",
				Depth: 3,
			},
			mockSetup:      func(llm *testutil.MockLLMClient, db *testutil.MockDB, browser *testutil.MockBrowserAutomation) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "Invalid URL")
			},
		},
		{
			name: "scraping error",
			requestBody: ExtractionRequest{
				URL:   "https://example.com/article",
				Depth: 3,
			},
			mockSetup: func(llm *testutil.MockLLMClient, db *testutil.MockDB, browser *testutil.MockBrowserAutomation) {
				browser.ScrapeErr = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "Failed to scrape article")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockLLM := testutil.NewMockLLMClient()
			mockDB := testutil.NewMockDB()
			mockBrowser := testutil.NewMockBrowserAutomation()

			// Setup mocks
			tt.mockSetup(mockLLM, mockDB, mockBrowser)

			// Create handler with mocks
			handler := NewExtractionHandler(cfg)
			handler.llm = mockLLM
			handler.db = mockDB
			handler.scraper = mockBrowser

			// Create request
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/extraction/url", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.HandleURLExtraction(rr, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestExtractionHandler_ProcessArticle(t *testing.T) {
	handler := NewExtractionHandler(&config.Config{})
	mockLLM := testutil.NewMockLLMClient()
	handler.llm = mockLLM

	tests := []struct {
		name          string
		article       *models.Article
		mockResponse  string
		mockError     error
		expectError   bool
		checkEntities func(*testing.T, *models.Article)
	}{
		{
			name: "successful processing",
			article: &models.Article{
				ID:      "test-id",
				Content: "Test content about John Doe and Jane Smith",
			},
			mockResponse: `{
				"entities": [
					{
						"type": "person",
						"name": "John Doe",
						"confidence": 0.95
					},
					{
						"type": "person",
						"name": "Jane Smith",
						"confidence": 0.9
					}
				]
			}`,
			checkEntities: func(t *testing.T, article *models.Article) {
				assert.Len(t, article.Entities, 2)
				assert.Equal(t, "John Doe", article.Entities[0].Name)
				assert.Equal(t, "Jane Smith", article.Entities[1].Name)
			},
		},
		{
			name: "llm error",
			article: &models.Article{
				ID:      "test-id",
				Content: "Test content",
			},
			mockError:   assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM.GenerateResponse = tt.mockResponse
			mockLLM.GenerateError = tt.mockError

			err := handler.processArticle(tt.article)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.checkEntities != nil {
				tt.checkEntities(t, tt.article)
			}
		})
	}
}

func TestExtractionHandler_ValidateDepth(t *testing.T) {
	tests := []struct {
		name        string
		depth       int
		expectError bool
	}{
		{
			name:        "valid depth",
			depth:       5,
			expectError: false,
		},
		{
			name:        "too low depth",
			depth:       1,
			expectError: true,
		},
		{
			name:        "too high depth",
			depth:       11,
			expectError: true,
		},
		{
			name:        "zero depth becomes default",
			depth:       0,
			expectError: false,
		},
	}

	handler := NewExtractionHandler(&config.Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ExtractionRequest{
				URL:   "https://example.com",
				Depth: tt.depth,
			}

			err := handler.validateRequest(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.depth == 0 {
					assert.Equal(t, 3, req.Depth) // Check default value
				}
			}
		})
	}
}
