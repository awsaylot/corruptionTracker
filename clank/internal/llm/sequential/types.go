package sequential

import (
	"context"
	"time"

	"clank/internal/llm"
	"clank/internal/models"
)

// AnalysisConfig configures the sequential analysis process
type AnalysisConfig struct {
	Depth                int           `json:"depth"`
	MaxStages            int           `json:"maxStages"`
	ConfidenceThreshold  float64       `json:"confidenceThreshold"`
	TimeoutPerStage      time.Duration `json:"timeoutPerStage"`
	EnableCrossReference bool          `json:"enableCrossReference"`
	EnableHypotheses     bool          `json:"enableHypotheses"`
}

// AnalysisSession represents a sequential analysis session
type AnalysisSession struct {
	ID          string                     `json:"id"`
	ArticleID   string                     `json:"articleId"`
	Config      *AnalysisConfig            `json:"config"`
	Stages      []*AnalysisStage           `json:"stages"`
	Status      string                     `json:"status"` // "running", "completed", "failed", "terminated"
	StartedAt   time.Time                  `json:"startedAt"`
	CompletedAt *time.Time                 `json:"completedAt,omitempty"`
	Error       string                     `json:"error,omitempty"`
	Evidence    []Evidence                 `json:"evidence"`
	Hypotheses  []Hypothesis               `json:"hypotheses"`
	Results     []*models.ExtractionResult `json:"results"`
}

// AnalysisStage represents a single stage in the sequential analysis
type AnalysisStage struct {
	Stage       int                      `json:"stage"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Status      string                   `json:"status"` // "pending", "running", "completed", "failed"
	StartedAt   *time.Time               `json:"startedAt,omitempty"`
	CompletedAt *time.Time               `json:"completedAt,omitempty"`
	Results     *models.ExtractionResult `json:"results,omitempty"`
	Confidence  float64                  `json:"confidence"`
	Insights    []string                 `json:"insights"`
	Questions   []string                 `json:"questions,omitempty"`
	Error       string                   `json:"error,omitempty"`
}

// Evidence represents a piece of evidence in the analysis
type Evidence struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// Hypothesis represents a generated hypothesis
type Hypothesis struct {
	ID                    string    `json:"id"`
	Description           string    `json:"description"`
	Type                  string    `json:"type"`
	Confidence            float64   `json:"confidence"`
	SupportingEvidence    []string  `json:"supporting_evidence"`
	ContradictingEvidence []string  `json:"contradicting_evidence"`
	RequiredEvidence      []string  `json:"required_evidence"`
	Implications          []string  `json:"implications"`
	CreatedAt             time.Time `json:"createdAt"`
}

// AnalysisStageProcessor defines the interface for analysis stage processors
type AnalysisStageProcessor interface {
	GetName() string
	GetDescription() string
	Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error
	WithLLMClient(client *llm.Client) AnalysisStageProcessor
}
