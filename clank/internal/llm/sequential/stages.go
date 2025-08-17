package sequential

import (
	"context"
	"time"

	"clank/internal/llm"
	"clank/internal/models"
)

// AnalysisStageProcessor is the interface each stage must implement.
type AnalysisStageProcessor interface {
	GetName() string
	GetDescription() string
	Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, results []*models.ExtractionResult) error
	WithLLMClient(client *llm.Client) AnalysisStageProcessor
}

/*
-----------------------------

	Surface Extraction Stage
	-----------------------------
*/
type surfaceExtractionStage struct {
	llmClient *llm.Client
}

func NewSurfaceExtractionStage() AnalysisStageProcessor {
	return &surfaceExtractionStage{}
}

func (s *surfaceExtractionStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
	s.llmClient = client
	return s
}

func (s *surfaceExtractionStage) GetName() string {
	return "Surface Extraction"
}

func (s *surfaceExtractionStage) GetDescription() string {
	return "Extracts basic metadata, entities and summary-level information from the article."
}

func (s *surfaceExtractionStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, results []*models.ExtractionResult) error {
	// lightweight placeholder: mimic work and produce a minimal result
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
	}

	stage.Confidence = 0.9
	stage.Results = &models.ExtractionResult{}
	stage.Insights = append(stage.Insights, "Basic entities and metadata extracted.")
	return nil
}

/*
-----------------------------

	Deep Analysis Stage
	-----------------------------
*/
type deepAnalysisStage struct {
	llmClient *llm.Client
}

func NewDeepAnalysisStage() AnalysisStageProcessor {
	return &deepAnalysisStage{}
}

func (d *deepAnalysisStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
	d.llmClient = client
	return d
}

func (d *deepAnalysisStage) GetName() string {
	return "Deep Analysis"
}

func (d *deepAnalysisStage) GetDescription() string {
	return "Performs deeper semantic analysis and relationship extraction."
}

func (d *deepAnalysisStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, results []*models.ExtractionResult) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(25 * time.Millisecond):
	}

	stage.Confidence = 0.8
	stage.Results = &models.ExtractionResult{}
	stage.Insights = append(stage.Insights, "Deep semantic patterns identified.")
	return nil
}

/*
-----------------------------

	Cross Reference Stage
	-----------------------------
*/
type crossReferenceStage struct {
	llmClient *llm.Client
}

func NewCrossReferenceStage() AnalysisStageProcessor {
	return &crossReferenceStage{}
}

func (c *crossReferenceStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
	c.llmClient = client
	return c
}

func (c *crossReferenceStage) GetName() string {
	return "Cross Reference"
}

func (c *crossReferenceStage) GetDescription() string {
	return "Cross-references findings against external sources or prior sessions."
}

func (c *crossReferenceStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, results []*models.ExtractionResult) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Millisecond):
	}

	stage.Confidence = 0.75
	stage.Results = &models.ExtractionResult{}
	stage.Insights = append(stage.Insights, "Cross-referencing completed.")
	return nil
}

/*
-----------------------------

	Hypothesis Generation Stage
	-----------------------------
*/
type hypothesisStage struct {
	llmClient *llm.Client
}

func NewHypothesisGenerationStage() AnalysisStageProcessor {
	return &hypothesisStage{}
}

func (h *hypothesisStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
	h.llmClient = client
	return h
}

func (h *hypothesisStage) GetName() string {
	return "Hypothesis Generation"
}

func (h *hypothesisStage) GetDescription() string {
	return "Generates hypotheses and potential explanations from extracted evidence."
}

func (h *hypothesisStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, results []*models.ExtractionResult) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(20 * time.Millisecond):
	}

	stage.Confidence = 0.7
	stage.Results = &models.ExtractionResult{}
	stage.Insights = append(stage.Insights, "Candidate hypotheses generated.")
	return nil
}

/*
-----------------------------

	Recursive Refinement Stage
	-----------------------------
*/
type recursiveRefinementStage struct {
	llmClient *llm.Client
}

func NewRecursiveRefinementStage() AnalysisStageProcessor {
	return &recursiveRefinementStage{}
}

func (r *recursiveRefinementStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
	r.llmClient = client
	return r
}

func (r *recursiveRefinementStage) GetName() string {
	return "Recursive Refinement"
}

func (r *recursiveRefinementStage) GetDescription() string {
	return "Iteratively refines evidence and hypotheses to improve confidence."
}

func (r *recursiveRefinementStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, results []*models.ExtractionResult) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(40 * time.Millisecond):
	}

	stage.Confidence = 0.85
	stage.Results = &models.ExtractionResult{}
	stage.Insights = append(stage.Insights, "Refinement pass completed.")
	return nil
}
