package handlers

import (
	"bytes"
	"clank/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func(*testutil.MockLLMClient, *testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful completion request",
			requestBody: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 0.7,
			},
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateResponse = `{"text": "Generated response", "usage": {"total_tokens": 50}}`
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "text")
				assert.Contains(t, response, "usage")
			},
		},
		{
			name: "missing prompt",
			requestBody: map[string]interface{}{
				"maxTokens":   100,
				"temperature": 0.7,
			},
			setupMocks:     func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "llm service error",
			requestBody: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 0.7,
			},
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "invalid temperature value",
			requestBody: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 2.0, // Should be between 0 and 1
			},
			setupMocks:     func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			mockLLM := testutil.NewMockLLMClient()
			mockDB := testutil.NewMockDB()
			tt.setupMocks(mockLLM, mockDB)

			r.POST("/llm/complete", func(c *gin.Context) {
				c.Set("llm", mockLLM)
				c.Set("db", mockDB)
				LLMHandler(c)
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/llm/complete", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestStreamLLMHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func(*testutil.MockLLMClient, *testutil.MockDB)
		expectedEvents []string
		expectError    bool
	}{
		{
			name: "successful streaming request",
			requestBody: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 0.7,
			},
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.StreamResponses = []string{
					"First chunk",
					"Second chunk",
					"Final chunk",
				}
			},
			expectedEvents: []string{
				"First chunk",
				"Second chunk",
				"Final chunk",
			},
			expectError: false,
		},
		{
			name: "stream error handling",
			requestBody: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 0.7,
			},
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.StreamError = assert.AnError
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			mockLLM := testutil.NewMockLLMClient()
			mockDB := testutil.NewMockDB()
			tt.setupMocks(mockLLM, mockDB)

			r.POST("/llm/stream", func(c *gin.Context) {
				c.Set("llm", mockLLM)
				c.Set("db", mockDB)
				StreamLLMHandler(c)
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/llm/stream", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "text/event-stream")
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if tt.expectError {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				return
			}

			assert.Equal(t, http.StatusOK, rr.Code)

			// Parse SSE response
			events := parseSSEEvents(rr.Body.String())
			assert.Equal(t, tt.expectedEvents, events)
		})
	}
}

// Helper function to parse SSE events from response
func parseSSEEvents(response string) []string {
	var events []string
	for _, line := range strings.Split(response, "\n") {
		if strings.HasPrefix(line, "data: ") {
			event := strings.TrimPrefix(line, "data: ")
			events = append(events, strings.TrimSpace(event))
		}
	}
	return events
}

func TestValidateLLMRequest(t *testing.T) {
	tests := []struct {
		name    string
		request map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid request",
			request: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 0.7,
			},
			wantErr: false,
		},
		{
			name: "missing prompt",
			request: map[string]interface{}{
				"maxTokens":   100,
				"temperature": 0.7,
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			request: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   100,
				"temperature": 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid maxTokens",
			request: map[string]interface{}{
				"prompt":      "Test prompt",
				"maxTokens":   -1,
				"temperature": 0.7,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLLMRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
