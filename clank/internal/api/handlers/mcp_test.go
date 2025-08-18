package handlers

import (
	"bytes"
	"clank/config"
	"clank/internal/llm"
	"clank/internal/testutil"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMCPService(t *testing.T) {
	cfg := &config.Config{
		LLMEndpoint: "http://localhost:8080",
	}

	service := NewMCPService(cfg)
	assert.NotNil(t, service)
	assert.NotNil(t, service.llmClient)
	assert.NotNil(t, service.server)
	assert.NotNil(t, service.promptLoader)
}

func TestMCPService_ProcessWithMCP(t *testing.T) {
	tests := []struct {
		name           string
		messages       []llm.Message
		setupMocks     func(*testutil.MockLLMClient)
		expectError    bool
		checkProcessed func(*testing.T, []llm.Message)
	}{
		{
			name: "successful processing with system prompt",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{"response": "Hello there!"}`
			},
			expectError: false,
			checkProcessed: func(t *testing.T, processed []llm.Message) {
				require.GreaterOrEqual(t, len(processed), 2) // At least system + user message
				assert.Equal(t, "system", processed[0].Role)
				assert.Equal(t, "user", processed[1].Role)
				assert.Equal(t, "Hello", processed[1].Content)
			},
		},
		{
			name:        "empty message list",
			messages:    []llm.Message{},
			setupMocks:  func(mock *testutil.MockLLMClient) {},
			expectError: false,
			checkProcessed: func(t *testing.T, processed []llm.Message) {
				require.GreaterOrEqual(t, len(processed), 1) // At least system prompt
				assert.Equal(t, "system", processed[0].Role)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				LLMEndpoint: "http://localhost:8080",
			}
			service := NewMCPService(cfg)

			mockLLM := testutil.NewMockLLMClient()
			tt.setupMocks(mockLLM)
			service.llmClient = mockLLM

			processed, err := service.ProcessWithMCP(context.Background(), tt.messages)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.checkProcessed != nil {
				tt.checkProcessed(t, processed)
			}
		})
	}
}

func TestMCPService_GenerateWithMCP(t *testing.T) {
	tests := []struct {
		name        string
		messages    []llm.Message
		setupMocks  func(*testutil.MockLLMClient)
		expectError bool
		expected    []string
	}{
		{
			name: "successful generation",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.StreamResponses = []string{
					"Hello",
					" there!",
				}
			},
			expectError: false,
			expected: []string{
				"Hello",
				" there!",
			},
		},
		{
			name: "llm error",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.StreamError = assert.AnError
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				LLMEndpoint: "http://localhost:8080",
			}
			service := NewMCPService(cfg)

			mockLLM := testutil.NewMockLLMClient()
			tt.setupMocks(mockLLM)
			service.llmClient = mockLLM

			responseChan := make(chan string, 10)
			err := service.GenerateWithMCP(context.Background(), tt.messages, responseChan)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			var responses []string
			for response := range responseChan {
				responses = append(responses, response)
			}

			assert.Equal(t, tt.expected, responses)
		})
	}
}

func TestMCPChatHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		setupMocks     func(*testutil.MockLLMClient)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful non-streaming request",
			request: map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
				"stream": false,
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateResponse = `{"response": "Hello there!"}`
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "response")
			},
		},
		{
			name: "successful streaming request",
			request: map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
				"stream": true,
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.StreamResponses = []string{
					"Hello",
					" there!",
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				responses := strings.Split(strings.TrimSpace(rr.Body.String()), "\n\n")
				assert.Greater(t, len(responses), 0)
				for _, resp := range responses {
					assert.True(t, strings.HasPrefix(resp, "data: "))
				}
			},
		},
		{
			name: "invalid request format",
			request: map[string]interface{}{
				"invalid": "format",
			},
			setupMocks:     func(mock *testutil.MockLLMClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "llm error",
			request: map[string]interface{}{
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
				"stream": false,
			},
			setupMocks: func(mock *testutil.MockLLMClient) {
				mock.GenerateError = assert.AnError
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				LLMEndpoint: "http://localhost:8080",
			}

			r := setupTestRouter()
			mockLLM := testutil.NewMockLLMClient()
			tt.setupMocks(mockLLM)

			handler := MCPChatHandler(cfg)
			r.POST("/chat", handler)

			body, err := json.Marshal(tt.request)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(body))
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
