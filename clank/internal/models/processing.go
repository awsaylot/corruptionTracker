package models

import "time"

// ProcessingStatus represents the current status of article processing
type ProcessingStatus struct {
	SessionID string    `json:"sessionId"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	Stage     string    `json:"stage"`
	StartedAt time.Time `json:"startedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Error     string    `json:"error,omitempty"`
}

// Processing status constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// Stage constants
const (
	StageScraping           = "scraping"
	StageContentCleaning    = "cleaning"
	StageEntityExtraction   = "entity_extraction"
	StageRelationExtraction = "relation_extraction"
	StageAnalysis           = "analysis"
	StageStorage            = "storage"
)
