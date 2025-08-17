package sequential

import (
	"context"
	"fmt"
	"sync"
	"time"

	"clank/internal/llm"
	"clank/internal/models"

	"github.com/google/uuid"
)

// AnalysisController manages sequential analysis sessions
type AnalysisController struct {
	llmClient *llm.Client
	sessions  map[string]*AnalysisSession
	mu        sync.RWMutex
	stages    []AnalysisStageProcessor
}

// NewAnalysisController creates a new analysis controller
func NewAnalysisController(llmClient *llm.Client) *AnalysisController {
	controller := &AnalysisController{
		llmClient: llmClient,
		sessions:  make(map[string]*AnalysisSession),
		stages: []AnalysisStageProcessor{
			NewSurfaceExtractionStage().WithLLMClient(llmClient),
			NewDeepAnalysisStage().WithLLMClient(llmClient),
			NewCrossReferenceStage().WithLLMClient(llmClient),
			NewHypothesisGenerationStage().WithLLMClient(llmClient),
			NewRecursiveRefinementStage().WithLLMClient(llmClient),
		},
	}
	return controller
}

// StartAnalysis starts a new sequential analysis session
func (c *AnalysisController) StartAnalysis(ctx context.Context, article *models.Article, config *AnalysisConfig) (*AnalysisSession, error) {
	sessionID := uuid.New().String()

	session := &AnalysisSession{
		ID:         sessionID,
		ArticleID:  article.ID,
		Config:     config,
		Status:     "running",
		StartedAt:  time.Now(),
		Evidence:   make([]Evidence, 0),
		Hypotheses: make([]Hypothesis, 0),
		Results:    make([]*models.ExtractionResult, 0),
		Stages:     make([]*AnalysisStage, 0),
	}

	// Initialize stages based on depth
	maxStages := config.MaxStages
	if config.Depth < maxStages {
		maxStages = config.Depth
	}

	for i := 0; i < maxStages && i < len(c.stages); i++ {
		processor := c.stages[i]
		stage := &AnalysisStage{
			Stage:       i + 1,
			Name:        processor.GetName(),
			Description: processor.GetDescription(),
			Status:      "pending",
			Confidence:  0.0,
			Insights:    make([]string, 0),
		}
		session.Stages = append(session.Stages, stage)
	}

	c.mu.Lock()
	c.sessions[sessionID] = session
	c.mu.Unlock()

	// Start processing in background
	go c.processSession(ctx, session, article)

	return session, nil
}

// GetSession retrieves a session by ID
func (c *AnalysisController) GetSession(sessionID string) (*AnalysisSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// TerminateSession terminates a running session
func (c *AnalysisController) TerminateSession(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	if session.Status == "running" {
		session.Status = "terminated"
		now := time.Now()
		session.CompletedAt = &now
	}

	return nil
}

// processSession processes all stages of an analysis session
func (c *AnalysisController) processSession(ctx context.Context, session *AnalysisSession, article *models.Article) {
	defer func() {
		if session.Status == "running" {
			session.Status = "completed"
			now := time.Now()
			session.CompletedAt = &now
		}
	}()

	for i, stage := range session.Stages {
		// Check if session was terminated
		c.mu.RLock()
		if session.Status == "terminated" {
			c.mu.RUnlock()
			return
		}
		c.mu.RUnlock()

		// Process stage with timeout
		stageCtx, cancel := context.WithTimeout(ctx, session.Config.TimeoutPerStage)

		now := time.Now()
		stage.StartedAt = &now
		stage.Status = "running"

		processor := c.stages[i]
		err := processor.Process(stageCtx, session, stage, article, session.Results)

		cancel()

		completedAt := time.Now()
		stage.CompletedAt = &completedAt

		if err != nil {
			stage.Status = "failed"
			stage.Error = err.Error()
			session.Status = "failed"
			session.Error = fmt.Sprintf("Stage %d failed: %v", stage.Stage, err)
			return
		}

		stage.Status = "completed"

		// Add results if available
		if stage.Results != nil {
			session.Results = append(session.Results, stage.Results)
		}

		// Check confidence threshold
		if stage.Confidence < session.Config.ConfidenceThreshold {
			// Could continue or stop based on policy
			// For now, continue but log low confidence
			stage.Insights = append(stage.Insights,
				fmt.Sprintf("Low confidence: %.2f (threshold: %.2f)",
					stage.Confidence, session.Config.ConfidenceThreshold))
		}
	}
}

// ListSessions returns all sessions (you might want to add pagination)
func (c *AnalysisController) ListSessions() []*AnalysisSession {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]*AnalysisSession, 0, len(c.sessions))
	for _, session := range c.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// CleanupSessions removes old completed sessions
func (c *AnalysisController) CleanupSessions(maxAge time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, session := range c.sessions {
		if session.CompletedAt != nil && session.CompletedAt.Before(cutoff) {
			delete(c.sessions, id)
		}
	}
}
