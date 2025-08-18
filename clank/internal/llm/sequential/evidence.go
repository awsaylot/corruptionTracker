package sequential

import (
	"time"
)

// Evidence represents a piece of evidence from the analysis
type Evidence struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // entity, relationship, pattern, inference
	Stage      int                    `json:"stage"`
	EntityID   string                 `json:"entityId,omitempty"`
	Text       string                 `json:"text"`
	Context    string                 `json:"context"`
	Source     string                 `json:"source"` // stage name or external reference
	Confidence float64                `json:"confidence"`
	CreatedAt  time.Time              `json:"createdAt"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Hypothesis represents a potential explanation or theory
type Hypothesis struct {
	ID          string                 `json:"id"`
	Stage       int                    `json:"stage"`
	Description string                 `json:"description"`
	Evidence    []string               `json:"evidence"` // IDs of supporting evidence
	Confidence  float64                `json:"confidence"`
	Status      string                 `json:"status"` // proposed, supported, refuted, uncertain
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EvidenceChain represents a sequence of connected evidence
type EvidenceChain struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Evidence    []string  `json:"evidence"` // Ordered list of evidence IDs
	Strength    float64   `json:"strength"` // Overall chain strength
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CrossReference represents a connection between pieces of evidence
type CrossReference struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // support, contradict, relate
	FromID      string    `json:"fromId"`
	ToID        string    `json:"toId"`
	Description string    `json:"description"`
	Strength    float64   `json:"strength"`
	CreatedAt   time.Time `json:"createdAt"`
}
