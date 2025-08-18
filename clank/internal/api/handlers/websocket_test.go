package handlers

import (
	"clank/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketHandler(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*testutil.MockLLMClient, *testutil.MockDB)
		messages      []string
		expectedCodes []int
		checkResponse func(*testing.T, []string)
	}{
		{
			name: "successful websocket connection and message exchange",
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateResponse = `{"type": "response", "content": "Hello, how can I help?"}`
			},
			messages: []string{
				`{"type": "message", "content": "Hi there"}`,
			},
			expectedCodes: []int{http.StatusSwitchingProtocols},
			checkResponse: func(t *testing.T, responses []string) {
				require.Len(t, responses, 1)
				var resp map[string]interface{}
				err := json.Unmarshal([]byte(responses[0]), &resp)
				require.NoError(t, err)
				assert.Equal(t, "response", resp["type"])
				assert.Contains(t, resp["content"], "Hello")
			},
		},
		{
			name:       "invalid message format",
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			messages: []string{
				`invalid json`,
			},
			expectedCodes: []int{http.StatusSwitchingProtocols},
			checkResponse: func(t *testing.T, responses []string) {
				require.Len(t, responses, 1)
				assert.Contains(t, responses[0], "error")
			},
		},
		{
			name: "llm error handling",
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateError = assert.AnError
			},
			messages: []string{
				`{"type": "message", "content": "Generate error"}`,
			},
			expectedCodes: []int{http.StatusSwitchingProtocols},
			checkResponse: func(t *testing.T, responses []string) {
				require.Len(t, responses, 1)
				assert.Contains(t, responses[0], "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test server
			r := setupTestRouter()
			mockLLM := testutil.NewMockLLMClient()
			mockDB := testutil.NewMockDB()
			tt.setupMocks(mockLLM, mockDB)

			// Register websocket handler
			r.GET("/ws", func(c *gin.Context) {
				c.Set("llm", mockLLM)
				c.Set("db", mockDB)
				WebSocketHandler(c)
			})

			// Create test server
			server := httptest.NewServer(r)
			defer server.Close()

			// Replace http with ws in server URL
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

			// Connect to websocket
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			defer ws.Close()

			responses := make([]string, 0)

			// Send test messages
			for _, msg := range tt.messages {
				err = ws.WriteMessage(websocket.TextMessage, []byte(msg))
				require.NoError(t, err)

				// Read response
				_, message, err := ws.ReadMessage()
				if err != nil {
					break
				}
				responses = append(responses, string(message))
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, responses)
			}
		})
	}
}

func TestHandleWebSocketMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     []byte
		setupMocks  func(*testutil.MockLLMClient, *testutil.MockDB)
		expectError bool
		checkResult func(*testing.T, []byte)
	}{
		{
			name:    "valid message processing",
			message: []byte(`{"type": "message", "content": "test message"}`),
			setupMocks: func(llm *testutil.MockLLMClient, db *testutil.MockDB) {
				llm.GenerateResponse = `{"type": "response", "content": "processed message"}`
			},
			expectError: false,
			checkResult: func(t *testing.T, result []byte) {
				var resp map[string]interface{}
				err := json.Unmarshal(result, &resp)
				require.NoError(t, err)
				assert.Equal(t, "response", resp["type"])
				assert.Equal(t, "processed message", resp["content"])
			},
		},
		{
			name:        "invalid json message",
			message:     []byte(`invalid json`),
			setupMocks:  func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			expectError: true,
		},
		{
			name:        "missing required fields",
			message:     []byte(`{}`),
			setupMocks:  func(llm *testutil.MockLLMClient, db *testutil.MockDB) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := testutil.NewMockLLMClient()
			mockDB := testutil.NewMockDB()
			tt.setupMocks(mockLLM, mockDB)

			result, err := handleWebSocketMessage(tt.message, mockLLM, mockDB)

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
