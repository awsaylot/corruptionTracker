package llm

import (
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

// Generate implements the LLM client interface
func (m *MockLLMClient) Generate(ctx context.Context, messages []Message) (*Response, error) {
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

	return &Response{
		ID:      "mock-response",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock-model",
		Choices: []Choice{
			{
				Content:      content,
				FinishReason: "stop",
			},
		},
	}, nil
}

// GenerateStream implements the LLM client interface
func (m *MockLLMClient) GenerateStream(ctx context.Context, messages []Message, respChan chan<- string) error {
	defer close(respChan)

	if m.GenerateStreamError != nil {
		return m.GenerateStreamError
	}

	if len(m.GenerateStreamItems) > 0 {
		for _, item := range m.GenerateStreamItems {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case respChan <- item:
			}
		}
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case respChan <- m.GenerateResponse:
	}

	return nil
}
