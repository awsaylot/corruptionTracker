package llm

// Message represents a message in the LLM conversation
type Message struct {
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Name      string                 `json:"name,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt int64                  `json:"createdAt"`
}
