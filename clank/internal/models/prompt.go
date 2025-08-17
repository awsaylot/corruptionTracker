package models

import (
	"time"
)

// PromptArgument represents an argument that can be passed to a prompt template
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type,omitempty"`    // string, number, boolean, object
	Default     any    `json:"default,omitempty"` // default value if not provided
}

// PromptMetadata contains metadata about the prompt
type PromptMetadata struct {
	Version     string    `json:"version"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Tags        []string  `json:"tags"`
	Author      string    `json:"author,omitempty"`
	Category    string    `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
}

// Prompt represents a prompt template that can be loaded from JSON
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments"`
	Template    string           `json:"template"`
	Metadata    PromptMetadata   `json:"metadata"`

	// Runtime fields (not from JSON)
	FilePath   string    `json:"-"`
	LoadedAt   time.Time `json:"-"`
	CompiledAt time.Time `json:"-"`
}

// PromptContext represents the context data passed to a prompt
type PromptContext struct {
	Arguments map[string]any `json:"arguments"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	NodeData  map[string]any `json:"node_data,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// PromptResult represents the result of rendering a prompt
type PromptResult struct {
	PromptName   string         `json:"prompt_name"`
	RenderedText string         `json:"rendered_text"`
	Arguments    map[string]any `json:"arguments"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	RenderTime   time.Duration  `json:"render_time"`
	Timestamp    time.Time      `json:"timestamp"`
}

// ValidationError represents a prompt validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (ve ValidationError) Error() string {
	return ve.Message
}
