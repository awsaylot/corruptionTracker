package models

import (
	"time"
)

// Evidence represents a piece of evidence in the investigation
type Evidence struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Stage      int                    `json:"stage"`
	EntityID   string                 `json:"entityId"`
	Text       string                 `json:"text"`
	Context    string                 `json:"context"`
	Source     string                 `json:"source"`
	Confidence float64                `json:"confidence"`
	CreatedAt  time.Time              `json:"createdAt"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Hypothesis represents an investigative hypothesis
type Hypothesis struct {
	ID          string                 `json:"id"`
	Stage       int                    `json:"stage"`
	Description string                 `json:"description"`
	Evidence    []string               `json:"evidence"`
	Confidence  float64                `json:"confidence"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AnalysisResult represents the result of analyzing an article
type AnalysisResult struct {
	ID            string         `json:"id"`
	ArticleID     string         `json:"articleId"`
	Hypotheses    []*Hypothesis  `json:"hypotheses"`
	Evidence      []*Evidence    `json:"evidence"`
	EvidenceChain *EvidenceChain `json:"evidenceChain,omitempty"`
	Confidence    float64        `json:"confidence"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// EvidenceChain represents a chain of linked evidence
type EvidenceChain struct {
	ID        string         `json:"id"`
	ArticleID string         `json:"articleId"`
	Evidence  []*Evidence    `json:"evidence"`
	Links     []EvidenceLink `json:"links"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// EvidenceLink represents a connection between two pieces of evidence
type EvidenceLink struct {
	FromID     string  `json:"fromId"`
	ToID       string  `json:"toId"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
}
