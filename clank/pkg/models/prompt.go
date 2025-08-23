package models

// Prompt represents a template prompt for Claude to use
type Prompt struct {
	Name        string `json:"name"`
	Text        string `json:"text"`
	Description string `json:"description,omitempty"`
}
