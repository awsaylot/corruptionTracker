package models

import (
	"time"
)

// Role type for chat participants
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleFunction  Role = "function"
)

// MessageType represents different types of chat messages
type MessageType string

const (
	TypeText     MessageType = "text"
	TypeCommand  MessageType = "command"
	TypeFunction MessageType = "function"
	TypeSystem   MessageType = "system"
	TypeError    MessageType = "error"
)

// ChatMessage represents a chat message structure
type ChatMessage struct {
	ID             string                 `json:"id"`
	Role           Role                   `json:"role"`
	Type           MessageType            `json:"type"`
	Content        string                 `json:"content"`
	ConversationID string                 `json:"conversationId,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Context        *GraphContext          `json:"context,omitempty"`
}

// GraphContext represents the graph-related context for a chat message
type GraphContext struct {
	NodeIDs      []string               `json:"nodeIds,omitempty"`
	RelationIDs  []string               `json:"relationIds,omitempty"`
	Subgraph     *NodeWithConnections   `json:"subgraph,omitempty"`
	ContextNodes map[string]interface{} `json:"contextNodes,omitempty"`
}

// StreamResponse represents a streaming message from the LLM
type StreamResponse struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Done      bool                   `json:"done"`
	Timestamp time.Time              `json:"timestamp"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// LLMResponse represents a response from the LLM service
type LLMResponse struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversationId"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	Status         string                 `json:"status"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Context        *GraphContext          `json:"context,omitempty"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID        string        `json:"id"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	Metadata  struct {
		Title       string                 `json:"title,omitempty"`
		Description string                 `json:"description,omitempty"`
		Tags        []string               `json:"tags,omitempty"`
		Custom      map[string]interface{} `json:"custom,omitempty"`
	} `json:"metadata"`
}
