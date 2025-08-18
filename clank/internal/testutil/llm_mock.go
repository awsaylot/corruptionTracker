package testutil

import (
	"clank/internal/llm"
	"context"
	"time"
)

// MockLLMClient implements a mock LLM client for testing
type MockLLMClient struct {
	GenerateResponse    string
	GenerateError       error
	Responses           []string
	Errors              []error
	CallCount           int
	GenerateStreamItems []string
	GenerateStreamError error
}

// NewMockLLMClient creates a new mock LLM client
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{}
}

// Generate implements the llm.LLMProvider interface with string return
func (m *MockLLMClient) Generate(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if m.GenerateError != nil {
		return nil, m.GenerateError
	}

	var content string
	if len(m.Responses) > m.CallCount {
		content = m.Responses[m.CallCount]
		m.CallCount++
	} else {
		content = m.GenerateResponse
	}

	return &llm.Response{
		ID:      "mock-response",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock-model",
		Choices: []llm.Choice{
			{
				Content:      content,
				FinishReason: "stop",
			},
		},
	}, nil
}

// GenerateStream implements the LLM client interface
func (m *MockLLMClient) GenerateStream(ctx context.Context, messages []llm.Message, respChan chan<- string) error {
	defer close(respChan)

	if m.GenerateError != nil {
		return m.GenerateError
	}

	if len(m.Responses) > 0 {
		for _, resp := range m.Responses {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case respChan <- resp:
			}
		}
		return nil
	}

	if m.GenerateResponse != "" {
		respChan <- m.GenerateResponse
	}
	return nil
}
