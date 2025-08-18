package handlers

import (
	"bytes"
	"clank/internal/prompts"
	"clank/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestListPrompts(t *testing.T) {
	tests := []struct {
		name           string
		mockPrompts    map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful prompt listing",
			mockPrompts: map[string]string{
				"entity_extraction":     "Extract entities from {{.text}}",
				"relationship_analysis": "Analyze relationships in {{.context}}",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Len(t, response["prompts"], 2)
			},
		},
		{
			name:           "empty prompt list",
			mockPrompts:    map[string]string{},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Empty(t, response["prompts"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			mockPromptService := &testutil.MockPromptService{
				Prompts: tt.mockPrompts,
			}

			r.GET("/prompts", func(c *gin.Context) {
				c.Set("promptService", mockPromptService)
				ListPrompts(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/prompts", nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestGetPrompt(t *testing.T) {
	tests := []struct {
		name           string
		promptName     string
		mockPrompts    map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "get existing prompt",
			promptName: "entity_extraction",
			mockPrompts: map[string]string{
				"entity_extraction": "Extract entities from {{.text}}",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response prompts.Prompt
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "entity_extraction", response.Name)
				assert.Contains(t, response.Template, "Extract entities")
			},
		},
		{
			name:           "prompt not found",
			promptName:     "non_existent",
			mockPrompts:    map[string]string{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			mockPromptService := &testutil.MockPromptService{
				Prompts: tt.mockPrompts,
			}

			r.GET("/prompts/:name", func(c *gin.Context) {
				c.Set("promptService", mockPromptService)
				GetPrompt(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/prompts/"+tt.promptName, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestRenderPrompt(t *testing.T) {
	tests := []struct {
		name           string
		promptName     string
		requestBody    map[string]interface{}
		mockPrompts    map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful render",
			promptName: "entity_extraction",
			requestBody: map[string]interface{}{
				"text": "John Doe is CEO of Acme Corp",
			},
			mockPrompts: map[string]string{
				"entity_extraction": "Extract entities from {{.text}}",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["rendered"].(string), "John Doe")
			},
		},
		{
			name:       "missing required variable",
			promptName: "entity_extraction",
			requestBody: map[string]interface{}{
				"wrong_var": "test",
			},
			mockPrompts: map[string]string{
				"entity_extraction": "Extract entities from {{.text}}",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "prompt not found",
			promptName: "non_existent",
			requestBody: map[string]interface{}{
				"text": "test",
			},
			mockPrompts:    map[string]string{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			mockPromptService := &testutil.MockPromptService{
				Prompts: tt.mockPrompts,
			}

			r.POST("/prompts/:name/render", func(c *gin.Context) {
				c.Set("promptService", mockPromptService)
				RenderPrompt(c)
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/prompts/"+tt.promptName+"/render", bytes.NewBuffer(body))
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

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name           string
		promptName     string
		requestBody    map[string]interface{}
		mockPrompts    map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "valid arguments",
			promptName: "entity_extraction",
			requestBody: map[string]interface{}{
				"text": "test content",
			},
			mockPrompts: map[string]string{
				"entity_extraction": "Extract entities from {{.text}}",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.True(t, response["valid"].(bool))
			},
		},
		{
			name:       "missing required argument",
			promptName: "entity_extraction",
			requestBody: map[string]interface{}{
				"wrong_var": "test",
			},
			mockPrompts: map[string]string{
				"entity_extraction": "Extract entities from {{.text}}",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.False(t, response["valid"].(bool))
				assert.Contains(t, response["errors"], "text")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()
			mockPromptService := &testutil.MockPromptService{
				Prompts: tt.mockPrompts,
			}

			r.POST("/prompts/:name/validate", func(c *gin.Context) {
				c.Set("promptService", mockPromptService)
				ValidatePrompt(c)
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/prompts/"+tt.promptName+"/validate", bytes.NewBuffer(body))
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
