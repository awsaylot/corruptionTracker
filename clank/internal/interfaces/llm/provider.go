package llm

import (
	"context"
	"io"
)

// Provider defines the interface for LLM operations
type Provider interface {
	// Chat handles direct conversations with the LLM
	Chat(ctx context.Context, messages []Message) (*Message, error)

	// StreamChat handles streaming conversations with the LLM
	StreamChat(ctx context.Context, messages []Message, output io.Writer) error

	// ProcessWithContext processes messages with given context
	ProcessWithContext(ctx context.Context, messages []Message, context string) (*Message, error)

	// GenerateWithTools generates responses using available tools
	GenerateWithTools(ctx context.Context, messages []Message, tools []Tool) (*Message, error)
}

// Tool represents a tool available to the LLM
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Config represents LLM configuration
type Config struct {
	Model       string                 `json:"model"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"maxTokens"`
	Options     map[string]interface{} `json:"options,omitempty"`
}
