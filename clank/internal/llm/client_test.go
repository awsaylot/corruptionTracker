package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"clank/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Generate(t *testing.T) {
	cfg := &config.Config{}
	cfg.LLM.URL = "http://localhost:8080"
	cfg.LLM.Model = "test-model"

	tests := []struct {
		name           string
		messages       []Message
		serverResponse string
		serverStatus   int
		expectError    bool
		expectedOutput string
	}{
		{
			name: "successful generation",
			messages: []Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Test prompt"},
			},
			serverResponse: `{
				"response": "Generated text"
			}`,
			serverStatus:   http.StatusOK,
			expectedOutput: "Generated text",
		},
		{
			name: "successful generation with MCP context",
			messages: []Message{
				{Role: "system", Content: "Using Model Context Protocol"},
				{Role: "user", Content: "Test with MCP"},
				{Role: "assistant", Content: "Previous context"},
			},
			serverResponse: `{
				"response": "MCP-aware response",
				"context": {
					"tools": ["graph_search", "entity_analysis"]
				}
			}`,
			serverStatus:   http.StatusOK,
			expectedOutput: "MCP-aware response",
		},
		{
			name: "server error",
			messages: []Message{
				{Role: "user", Content: "Test prompt"},
			},
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
		{
			name: "invalid json response",
			messages: []Message{
				{Role: "user", Content: "Test prompt"},
			},
			serverResponse: "invalid json",
			serverStatus:   http.StatusOK,
			expectError:    true,
		},
		{
			name:        "empty messages",
			messages:    []Message{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			// Create client with test server URL
			cfg := &config.Config{}
			cfg.LLM.URL = server.URL
			cfg.LLM.Model = "test-model"
			client := NewClient(cfg)

			// Test generation
			ctx := context.Background()
			output, err := client.Generate(ctx, tt.messages)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

func TestClient_GenerateStream(t *testing.T) {
	tests := []struct {
		name            string
		messages        []Message
		serverResponses []string
		serverDelay     time.Duration
		serverError     bool
		expectedOutputs []string
		expectError     bool
	}{
		{
			name: "successful streaming",
			messages: []Message{
				{Role: "user", Content: "Test prompt"},
			},
			serverResponses: []string{"First", "Second", "Third"},
			serverDelay:     10 * time.Millisecond,
			expectedOutputs: []string{"First", "Second", "Third"},
		},
		{
			name: "successful streaming with MCP",
			messages: []Message{
				{Role: "system", Content: "Using MCP"},
				{Role: "user", Content: "Test with context"},
			},
			serverResponses: []string{
				`{"type": "text", "content": "First"}`,
				`{"type": "tool_call", "tool": "graph_search"}`,
				`{"type": "text", "content": "Final"}`,
			},
			serverDelay:     10 * time.Millisecond,
			expectedOutputs: []string{`{"type": "text", "content": "First"}`, `{"type": "tool_call", "tool": "graph_search"}`, `{"type": "text", "content": "Final"}`},
		},
		{
			name: "server error",
			messages: []Message{
				{Role: "user", Content: "Test prompt"},
			},
			serverError: true,
			expectError: true,
		},
		{
			name:        "empty messages",
			messages:    []Message{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				flusher, ok := w.(http.Flusher)
				require.True(t, ok)

				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")

				if tt.serverError {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				for _, resp := range tt.serverResponses {
					_, err := w.Write([]byte("data: " + resp + "\n\n"))
					require.NoError(t, err)
					flusher.Flush()
					time.Sleep(tt.serverDelay)
				}
			}))
			defer server.Close()

			// Create client with test server URL
			cfg := &config.Config{}
			cfg.LLM.URL = server.URL
			cfg.LLM.Model = "test-model"
			client := NewClient(cfg)

			// Test streaming
			ctx := context.Background()
			respChan := make(chan string, len(tt.serverResponses))

			err := client.GenerateStream(ctx, tt.messages, respChan)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			var outputs []string
			for resp := range respChan {
				outputs = append(outputs, resp)
			}

			assert.Equal(t, tt.expectedOutputs, outputs)
		})
	}
}

func TestClient_Generate_Request(t *testing.T) {
	cfg := &config.Config{}
	cfg.LLM.URL = "http://localhost:8080"
	cfg.LLM.Model = "test-model"

	tests := []struct {
		name         string
		messages     []Message
		checkRequest func(*testing.T, *http.Request)
		expectError  bool
	}{
		{
			name: "valid request",
			messages: []Message{
				{Role: "user", Content: "Test prompt"},
			},
			checkRequest: func(t *testing.T, req *http.Request) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			},
		},
		{
			name: "request with MCP headers",
			messages: []Message{
				{Role: "system", Content: "Using MCP"},
				{Role: "user", Content: "Test"},
			},
			checkRequest: func(t *testing.T, req *http.Request) {
				assert.Equal(t, http.MethodPost, req.Method)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.NotEmpty(t, req.Header.Get("MCP-Version"))
				assert.NotEmpty(t, req.Header.Get("MCP-Agent"))
			},
		},
		{
			name:        "empty messages",
			messages:    []Message{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkRequest != nil {
					tt.checkRequest(t, r)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response":"test"}`))
			}))
			defer server.Close()

			cfg.LLM.URL = server.URL
			client := NewClient(cfg)

			ctx := context.Background()
			_, err := client.Generate(ctx, tt.messages)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_Generate_Response(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectError   bool
		expectedValue string
	}{
		{
			name:          "valid JSON response",
			response:      `{"response": "Test", "usage": {"total_tokens": 10}}`,
			expectError:   false,
			expectedValue: "Test",
		},
		{
			name:          "valid MCP response",
			response:      `{"response": "Test", "context": {"tools": ["search"]}, "usage": {"total_tokens": 10}}`,
			expectError:   false,
			expectedValue: "Test",
		},
		{
			name:        "invalid JSON",
			response:    `{invalid json`,
			expectError: true,
		},
		{
			name:        "empty response",
			response:    "",
			expectError: true,
		},
		{
			name:        "missing required fields",
			response:    `{"someField": "value"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			cfg := &config.Config{}
			cfg.LLM.URL = server.URL
			cfg.LLM.Model = "test-model"
			client := NewClient(cfg)

			resp, err := client.Generate(context.Background(), []Message{{Role: "user", Content: "test"}})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, resp)
			}
		})
	}
}

func TestClient_ProcessTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		setupServer func(http.ResponseWriter, *http.Request)
		expectError bool
	}{
		{
			name:    "successful completion before timeout",
			timeout: time.Second,
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response": "Quick response"}`))
			},
			expectError: false,
		},
		{
			name:    "timeout occurs",
			timeout: time.Millisecond * 50,
			setupServer: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Millisecond * 100)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"response": "Slow response"}`))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.setupServer))
			defer server.Close()

			cfg := &config.Config{}
			cfg.LLM.URL = server.URL
			cfg.LLM.Model = "test-model"
			cfg.LLM.Timeout = tt.timeout
			client := NewClient(cfg)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			_, err := client.Generate(ctx, []Message{{Role: "user", Content: "Test"}})

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context deadline exceeded")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
