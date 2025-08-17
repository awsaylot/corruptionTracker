package models

import (
	"encoding/json"
	"time"
)

// Node represents a node in the graph database
type Node struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Props     map[string]any         `json:"properties"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Relationship represents a relationship between two nodes
type Relationship struct {
	ID       string                 `json:"id"`
	FromID   string                 `json:"fromId"`
	ToID     string                 `json:"toId"`
	Type     string                 `json:"type"`
	Props    map[string]any         `json:"properties"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Connection represents a node connected to another node through a relationship
type Connection struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Properties   map[string]interface{} `json:"properties"`
	Relationship struct {
		ID         string                 `json:"id"`
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
		Direction  string                 `json:"direction"`
		CreatedAt  time.Time              `json:"createdAt"`
	} `json:"relationship"`
}

// NodeWithConnections represents a node with its connections to other nodes
type NodeWithConnections struct {
	ID          string       `json:"id"`
	Type        string       `json:"type"`
	Properties  interface{}  `json:"properties"`
	Connections []Connection `json:"connections"`
	Metadata    struct {
		CreatedAt       time.Time `json:"createdAt"`
		UpdatedAt       time.Time `json:"updatedAt"`
		ConnectionCount int       `json:"connectionCount"`
	} `json:"metadata"`
}

// NodeType represents the schema for a type of node
type NodeType struct {
	Name       string                 `json:"name"`
	Schema     map[string]string      `json:"schema"`
	Validators map[string]interface{} `json:"validators,omitempty"`
}

// ToJSON converts a struct to JSON string
func ToJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromJSON converts JSON string to a struct
func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}
