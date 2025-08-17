package sequential

import (
	"context"
	"fmt"
	"time"

	"clank/internal/llm"
	"clank/internal/models"
)

// AnalysisConfig defines configuration for sequential analysis
type AnalysisConfig struct {
	Depth                int           `json:"depth"`               // Analysis depth (2-10)
	MaxStages            int           `json:"maxStages"`           // Maximum number of stages
	ConfidenceThreshold  float64       `json:"confidenceThreshold"` // Minimum confidence to continue
	TimeoutPerStage      time.Duration `json:"timeoutPerStage"`     // Timeout for each stage
	EnableCrossReference bool          `json:"enableCrossReference"`
	EnableHypotheses     bool          `json:"enableHypotheses"`
}

// DefaultConfig returns sensible defaults for analysis
func DefaultConfig() *AnalysisConfig {
	return &AnalysisConfig{
		Depth:                3,
		MaxStages:            5,
		ConfidenceThreshold:  0.6,
		TimeoutPerStage:      30 * time.Second,
		EnableCrossReference: true,
		EnableHypotheses:     true,
	}
}

// AnalysisStage represents a single stage of analysis
type AnalysisStage struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Stage       int                      `json:"stage"`
	Status      string                   `json:"status"` // pending, running, completed, failed
	StartedAt   time.Time                `json:"startedAt"`
	CompletedAt *time.Time               `json:"completedAt"`
	Results     *models.ExtractionResult `json:"results"`
	Confidence  float64                  `json:"confidence"`
	Insights    []string                 `json:"insights"`
	Questions   []string                 `json:"questions"`
	Error       string                   `json:"error,omitempty"`
}

// AnalysisSession tracks the entire sequential analysis process
type AnalysisSession struct {
	ID           string                   `json:"id"`
	ArticleID    string                   `json:"articleId"`
	Config       *AnalysisConfig          `json:"config"`
	Stages       []*AnalysisStage         `json:"stages"`
	CurrentStage int                      `json:"currentStage"`
	Status       string                   `json:"status"` // running, completed, failed, terminated
	StartedAt    time.Time                `json:"startedAt"`
	CompletedAt  *time.Time               `json:"completedAt"`
	FinalResults *models.ExtractionResult `json:"finalResults"`
	Evidence     []*EvidenceChain         `json:"evidence"`
	Hypotheses   []*Hypothesis            `json:"hypotheses"`
}

// EvidenceChain represents a chain of evidence supporting a claim
type EvidenceChain struct {
	ID         string    `json:"id"`
	Claim      string    `json:"claim"`
	Evidence   []string  `json:"evidence"`
	Confidence float64   `json:"confidence"`
	Sources    []string  `json:"sources"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Hypothesis represents a possible explanation or theory
type Hypothesis struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	Supporting  []string  `json:"supporting"`
	Conflicting []string  `json:"conflicting"`
	Questions   []string  `json:"questions"`
	CreatedAt   time.Time `json:"createdAt"`
}

// AnalysisController manages the sequential analysis process
type AnalysisController struct {
	sessions map[string]*AnalysisSession
	stages   []StageProcessor
}

// StageProcessor defines the interface for processing analysis stages
type StageProcessor interface {
	Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error
	GetName() string
	GetDescription() string
}

// NewAnalysisController creates a new analysis controller
func NewAnalysisController(llmClient *llm.Client) *AnalysisController {
	return &AnalysisController{
		sessions: make(map[string]*AnalysisSession),
		stages: []StageProcessor{
			NewSurfaceExtractionStage().WithLLMClient(llmClient),
			NewDeepAnalysisStage().WithLLMClient(llmClient),
			NewCrossReferenceStage().WithLLMClient(llmClient),
			NewHypothesisGenerationStage().WithLLMClient(llmClient),
			NewRecursiveRefinementStage().WithLLMClient(llmClient),
		},
	}
}

// StartAnalysis begins a new sequential analysis session
func (ac *AnalysisController) StartAnalysis(ctx context.Context, article *models.Article, config *AnalysisConfig) (*AnalysisSession, error) {
	if config == nil {
		config = DefaultConfig()
	}

	session := &AnalysisSession{
		ID:           generateSessionID(),
		ArticleID:    article.ID,
		Config:       config,
		Stages:       make([]*AnalysisStage, 0),
		CurrentStage: 0,
		Status:       "running",
		StartedAt:    time.Now(),
		Evidence:     make([]*EvidenceChain, 0),
		Hypotheses:   make([]*Hypothesis, 0),
	}

	// Create stages based on configuration
	maxStages := min(config.MaxStages, len(ac.stages))
	for i := 0; i < maxStages; i++ {
		processor := ac.stages[i]
		stage := &AnalysisStage{
			ID:          fmt.Sprintf("%s-stage-%d", session.ID, i+1),
			Name:        processor.GetName(),
			Description: processor.GetDescription(),
			Stage:       i + 1,
			Status:      "pending",
		}
		session.Stages = append(session.Stages, stage)
	}

	ac.sessions[session.ID] = session

	// Start processing in a goroutine
	go ac.processSession(ctx, session, article)

	return session, nil
}

// GetSession retrieves an analysis session by ID
func (ac *AnalysisController) GetSession(sessionID string) (*AnalysisSession, error) {
	session, exists := ac.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// TerminateSession stops an analysis session
func (ac *AnalysisController) TerminateSession(sessionID string) error {
	session, exists := ac.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Status = "terminated"
	now := time.Now()
	session.CompletedAt = &now

	return nil
}

// processSession runs the sequential analysis process
func (ac *AnalysisController) processSession(ctx context.Context, session *AnalysisSession, article *models.Article) {
	var previousResults []*models.ExtractionResult

	for i, stage := range session.Stages {
		// Check if session was terminated
		if session.Status == "terminated" {
			break
		}

		session.CurrentStage = i
		stage.Status = "running"
		stage.StartedAt = time.Now()

		// Create stage context with timeout
		stageCtx, cancel := context.WithTimeout(ctx, session.Config.TimeoutPerStage)

		// Process the stage
		processor := ac.stages[i]
		err := processor.Process(stageCtx, session, stage, article, previousResults)
		cancel()

		if err != nil {
			stage.Status = "failed"
			stage.Error = err.Error()
			now := time.Now()
			stage.CompletedAt = &now
			session.Status = "failed"
			session.CompletedAt = &now
			return
		}

		// Mark stage as completed
		stage.Status = "completed"
		now := time.Now()
		stage.CompletedAt = &now

		// Add results to previous results for next stage
		if stage.Results != nil {
			previousResults = append(previousResults, stage.Results)
		}

		// Check confidence threshold
		if stage.Confidence < session.Config.ConfidenceThreshold {
			break
		}

		// Check if we've reached the configured depth
		if i+1 >= session.Config.Depth {
			break
		}
	}

	// Finalize session
	session.Status = "completed"
	now := time.Now()
	session.CompletedAt = &now

	// Combine all results
	session.FinalResults = ac.combineResults(previousResults)
}

// combineResults merges results from all stages
func (ac *AnalysisController) combineResults(results []*models.ExtractionResult) *models.ExtractionResult {
	if len(results) == 0 {
		return nil
	}

	combined := &models.ExtractionResult{
		Entities:      make([]models.ExtractedEntity, 0),
		Relationships: make([]models.ExtractedRelationship, 0),
		Confidence:    0,
	}

	// Merge entities and relationships, avoiding duplicates
	entityMap := make(map[string]models.ExtractedEntity)
	relationshipMap := make(map[string]models.ExtractedRelationship)

	totalConfidence := 0.0
	for _, result := range results {
		totalConfidence += result.Confidence

		for _, entity := range result.Entities {
			if existing, exists := entityMap[entity.ID]; exists {
				// Update confidence if higher
				if entity.Confidence > existing.Confidence {
					entityMap[entity.ID] = entity
				}
			} else {
				entityMap[entity.ID] = entity
			}
		}

		for _, rel := range result.Relationships {
			if existing, exists := relationshipMap[rel.ID]; exists {
				if rel.Confidence > existing.Confidence {
					relationshipMap[rel.ID] = rel
				}
			} else {
				relationshipMap[rel.ID] = rel
			}
		}
	}

	// Convert maps back to slices
	for _, entity := range entityMap {
		combined.Entities = append(combined.Entities, entity)
	}

	for _, rel := range relationshipMap {
		combined.Relationships = append(combined.Relationships, rel)
	}

	// Average confidence
	combined.Confidence = totalConfidence / float64(len(results))

	return combined
}

// Helper functions
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
