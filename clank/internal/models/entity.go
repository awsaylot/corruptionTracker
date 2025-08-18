package models

// Entity represents an entity in the system
type Entity struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}
