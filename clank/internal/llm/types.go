package llm

import (
	"context"
	"time"
)

// Message represents a chat message in a conversation
type Message struct {
	ID        string                 `json:"id,omitempty"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Name      string                 `json:"name,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
}

// Response represents a response from the LLM
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Error   string   `json:"error,omitempty"`
}

// Choice represents a single choice in an LLM response
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Content      string  `json:"content,omitempty"` // For raw text responses
}

// NewErrorResponse creates a new Response with an error
func NewErrorResponse(err string) *Response {
	return &Response{
		Error: err,
		Choices: []Choice{
			{Content: err},
		},
	}
}

// LLMProvider defines the methods that any LLM client must implement
type LLMProvider interface {
	Generate(ctx context.Context, messages []Message) (*Response, error)
	GenerateStream(ctx context.Context, messages []Message, respChan chan<- string) error
}
