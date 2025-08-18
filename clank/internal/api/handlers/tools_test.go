package handlers

import (
	"bytes"
	"clank/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockSetup      func(*testutil.MockLLMClient, *testutil.MockDB)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful tool execution",
			requestBody: map[string]interface{}{
				"tool":      "graph_search",
				"query":     "Find connections between John Doe and Acme Corp",
				"maxDepth":  3,
				"entityIds": []string{"entity1", "entity2"},
			},
			mockSetup: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateResponse = `{
					"path": [
						{"id": "entity1", "name": "John Doe"},
						{"id": "rel1", "type": "OWNS"},
						{"id": "entity2", "name": "Acme Corp"}
					],
					"confidence": 0.95
				}`
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "result")
				assert.Contains(t, response, "tool")
			},
		},
		{
			name: "invalid tool name",
			requestBody: map[string]interface{}{
				"tool":  "invalid_tool",
				"query": "test query",
			},
			mockSetup:      func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing required parameters",
			requestBody: map[string]interface{}{
				"tool": "graph_search",
				// missing query
			},
			mockSetup:      func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "llm error",
			requestBody: map[string]interface{}{
				"tool":      "graph_search",
				"query":     "test query",
				"entityIds": []string{"entity1"},
			},
			mockSetup: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupTestRouter()

			// Create mocks
			mockLLM := testutil.NewMockLLMClient()
			mockDB := testutil.NewMockDB()

			// Setup mocks
			tt.mockSetup(mockLLM, mockDB)

			// Register handler
			r.POST("/tools", func(c *gin.Context) {
				c.Set("llm", mockLLM)
				c.Set("db", mockDB)
				ToolHandler(c)
			})

			// Create request
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/tools", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Execute request
			r.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestValidateToolParameters(t *testing.T) {
	tests := []struct {
		name        string
		tool        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid graph search parameters",
			tool: "graph_search",
			params: map[string]interface{}{
				"query":     "test query",
				"maxDepth":  3,
				"entityIds": []string{"entity1", "entity2"},
			},
			expectError: false,
		},
		{
			name: "valid entity analysis parameters",
			tool: "entity_analysis",
			params: map[string]interface{}{
				"entityId": "entity1",
				"depth":    2,
			},
			expectError: false,
		},
		{
			name: "missing required parameter",
			tool: "graph_search",
			params: map[string]interface{}{
				"maxDepth": 3,
				// missing query
			},
			expectError: true,
		},
		{
			name: "invalid parameter type",
			tool: "graph_search",
			params: map[string]interface{}{
				"query":     123, // should be string
				"maxDepth":  3,
				"entityIds": []string{"entity1"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToolParameters(tt.tool, tt.params)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteGraphSearch(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		mockSetup   func(*testutil.MockDB)
		expectError bool
		checkResult func(*testing.T, interface{})
	}{
		{
			name: "successful search",
			params: map[string]interface{}{
				"query":     "Find connections",
				"maxDepth":  3,
				"entityIds": []string{"entity1", "entity2"},
			},
			mockSetup: func(db *testutil.MockDB) {
				entity1 := testutil.MockEntity("person", "John Doe")
				entity2 := testutil.MockEntity("company", "Acme Corp")
				entity1.ID = "entity1"
				entity2.ID = "entity2"
				db.Entities[entity1.ID] = entity1
				db.Entities[entity2.ID] = entity2
			},
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, resultMap, "paths")
				assert.Contains(t, resultMap, "entities")
			},
		},
		{
			name: "invalid entity IDs",
			params: map[string]interface{}{
				"query":     "Find connections",
				"maxDepth":  3,
				"entityIds": []string{"invalid1", "invalid2"},
			},
			mockSetup:   func(db *testutil.MockDB) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := testutil.NewMockDB()
			tt.mockSetup(mockDB)

			result, err := executeGraphSearch(tt.params, mockDB)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestExecuteEntityAnalysis(t *testing.T) {
	tests := []struct {
		name        string
		params      map[string]interface{}
		mockSetup   func(*testutil.MockDB)
		expectError bool
		checkResult func(*testing.T, interface{})
	}{
		{
			name: "successful analysis",
			params: map[string]interface{}{
				"entityId": "entity1",
				"depth":    2,
			},
			mockSetup: func(db *testutil.MockDB) {
				entity := testutil.MockEntity("person", "John Doe")
				entity.ID = "entity1"
				db.Entities[entity.ID] = entity
			},
			checkResult: func(t *testing.T, result interface{}) {
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Contains(t, resultMap, "entity")
				assert.Contains(t, resultMap, "relationships")
				assert.Contains(t, resultMap, "score")
			},
		},
		{
			name: "entity not found",
			params: map[string]interface{}{
				"entityId": "invalid",
				"depth":    2,
			},
			mockSetup:   func(db *testutil.MockDB) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := testutil.NewMockDB()
			tt.mockSetup(mockDB)

			result, err := executeEntityAnalysis(tt.params, mockDB)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
