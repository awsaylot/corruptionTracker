package interfaces

import "context"

// Message represents a message in the conversation with the LLM
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMProvider defines the interface for LLM interactions
type LLMProvider interface {
	Generate(ctx context.Context, messages []Message) (string, error)
	GenerateStream(ctx context.Context, messages []Message, respChan chan<- string) error
}
