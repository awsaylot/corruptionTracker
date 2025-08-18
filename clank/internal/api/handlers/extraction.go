package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"clank/config"
	"clank/internal/db"
	"clank/internal/llm"
	"clank/internal/llm/sequential"
	"clank/internal/models"
	browser "clank/internal/tools/browser"
	"clank/pkg/extraction"
)

// ExtractionHandler handles article content extraction requests with sequential analysis
type ExtractionHandler struct {
	scraper            Scraper
	processor          Processor
	llm                LLMClient
	db                 Store
	analysisController *sequential.AnalysisController
}

// NewExtractionHandler creates a new extraction handler with sequential analysis
func NewExtractionHandler(cfg *config.Config) *ExtractionHandler {
	llmClient := llm.NewClient(cfg)
	return &ExtractionHandler{
		scraper:            browser.NewArticleScraper(),
		processor:          extraction.NewContentProcessor(),
		llm:                llmClient,
		db:                 db.NewArticleStore(),
		analysisController: sequential.NewAnalysisController(llmClient),
	}
}

// ExtractionRequest represents the request to extract information from a URL
type ExtractionRequest struct {
	URL   string `json:"url"`
	Depth int    `json:"depth,omitempty"` // Analysis depth (2-10)
}

// ExtractionResponse represents the complete extraction response
type ExtractionResponse struct {
	SessionID string                      `json:"sessionId"`
	Article   *models.Article             `json:"article"`
	Session   *sequential.AnalysisSession `json:"session"`
	Status    string                      `json:"status"`
}

// HandleURLExtraction processes a URL for article extraction with sequential analysis
func (h *ExtractionHandler) HandleURLExtraction(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Extraction] Received request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Printf("[Extraction] Invalid method %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExtractionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Extraction] Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[Extraction] Processing URL: %s", req.URL)

	// Validate URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if _, err := url.Parse(req.URL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Validate depth (default to 3 if not specified)
	if req.Depth == 0 {
		req.Depth = 3
	}
	if req.Depth < 2 || req.Depth > 10 {
		http.Error(w, "Depth must be between 2 and 10", http.StatusBadRequest)
		return
	}

	// Create context for the request
	ctx := r.Context()

	// Initialize scraper if needed
	log.Println("[Extraction] Initializing scraper...")
	if err := h.scraper.Initialize(); err != nil {
		log.Printf("[Extraction] Scraper initialization failed: %v", err)
		http.Error(w, "Failed to initialize scraper: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Scrape the article
	log.Printf("[Extraction] Starting article scraping for URL: %s", req.URL)
	article, err := h.scraper.ScrapeArticle(req.URL)
	if err != nil {
		log.Printf("[Extraction] Article scraping failed: %v", err)
		http.Error(w, "Failed to scrape article: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[Extraction] Successfully scraped article, length: %d characters", len(article.Content))

	// Process the content
	log.Println("[Extraction] Starting article processing...")
	result, err := h.processor.ProcessArticle(article.Content)
	if err != nil {
		log.Printf("[Extraction] Article processing failed: %v", err)
		http.Error(w, "Failed to process article: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("[Extraction] Article processing completed successfully")

	// Update article with processed content
	article.Content = result.Content
	if result.Title != "" {
		article.Title = result.Title
	}
	if result.Metadata != nil {
		article.Metadata = result.Metadata
	}

	// Save the article
	log.Println("[Extraction] Saving article to database...")
	if err := h.db.SaveArticle(article); err != nil {
		log.Printf("[Extraction] Failed to save article: %v", err)
		http.Error(w, "Failed to save article: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[Extraction] Article saved successfully with ID: %s", article.ID)

	// Set timestamps
	article.ExtractedAt = time.Now()
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()

	// Configure sequential analysis
	log.Printf("[Extraction] Configuring analysis with depth %d...", req.Depth)
	config := &sequential.AnalysisConfig{
		Depth:                req.Depth,
		MaxStages:            5,
		ConfidenceThreshold:  0.6,
		TimeoutPerStage:      60 * time.Second,
		EnableCrossReference: true,
		EnableHypotheses:     true,
	}
	log.Println("[Extraction] Analysis configuration complete")

	// Start sequential analysis
	session, err := h.analysisController.StartAnalysis(ctx, article, config)
	if err != nil {
		http.Error(w, "Failed to start analysis: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Store article in database
	if err := h.db.SaveArticle(article); err != nil {
		http.Error(w, "Failed to store article data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return immediate response with session info
	response := &ExtractionResponse{
		SessionID: session.ID,
		Article:   article,
		Session:   session,
		Status:    "started",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleAnalysisProgress returns the current progress of an analysis session
func (h *ExtractionHandler) HandleAnalysisProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId parameter is required", http.StatusBadRequest)
		return
	}

	session, err := h.analysisController.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleAnalysisTermination terminates an ongoing analysis session
func (h *ExtractionHandler) HandleAnalysisTermination(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}

	if err := h.analysisController.TerminateSession(req.SessionID); err != nil {
		http.Error(w, "Failed to terminate session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "terminated"}`))
}

// HandleAnalysisConfiguration updates analysis configuration
func (h *ExtractionHandler) HandleAnalysisConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Depth                int  `json:"depth"`
		EnableCrossReference bool `json:"enableCrossReference"`
		EnableHypotheses     bool `json:"enableHypotheses"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate depth
	if req.Depth < 2 || req.Depth > 10 {
		http.Error(w, "Depth must be between 2 and 10", http.StatusBadRequest)
		return
	}

	// Return the validated configuration
	config := map[string]interface{}{
		"depth":                req.Depth,
		"enableCrossReference": req.EnableCrossReference,
		"enableHypotheses":     req.EnableHypotheses,
		"status":               "configured",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleAnalysisEvidence returns evidence chains for a session
func (h *ExtractionHandler) HandleAnalysisEvidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId parameter is required", http.StatusBadRequest)
		return
	}

	session, err := h.analysisController.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"sessionId": sessionID,
		"evidence":  session.Evidence,
		"count":     len(session.Evidence),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleAnalysisHypotheses returns hypotheses for a session
func (h *ExtractionHandler) HandleAnalysisHypotheses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId parameter is required", http.StatusBadRequest)
		return
	}

	session, err := h.analysisController.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"sessionId":  sessionID,
		"hypotheses": session.Hypotheses,
		"count":      len(session.Hypotheses),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleAnalysisDepth updates the analysis depth for a session (if still running)
func (h *ExtractionHandler) HandleAnalysisDepth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID string `json:"sessionId"`
		Depth     int    `json:"depth"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "sessionId is required", http.StatusBadRequest)
		return
	}

	if req.Depth < 2 || req.Depth > 10 {
		http.Error(w, "Depth must be between 2 and 10", http.StatusBadRequest)
		return
	}

	session, err := h.analysisController.GetSession(req.SessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Only allow depth updates for running sessions
	if session.Status != "running" {
		http.Error(w, "Cannot update depth for completed session", http.StatusBadRequest)
		return
	}

	// Update the session config (this would need to be implemented in the controller)
	session.Config.Depth = req.Depth

	response := map[string]interface{}{
		"sessionId": req.SessionID,
		"depth":     req.Depth,
		"status":    "updated",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleAnalysisConfidenceScores returns confidence scores for a session
func (h *ExtractionHandler) HandleAnalysisConfidenceScores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "sessionId parameter is required", http.StatusBadRequest)
		return
	}

	session, err := h.analysisController.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Calculate confidence scores for each stage
	scores := make([]map[string]interface{}, 0)
	for _, stage := range session.Stages {
		stageScore := map[string]interface{}{
			"stage":      stage.Stage,
			"name":       stage.Name,
			"confidence": stage.Confidence,
			"status":     stage.Status,
		}

		if stage.Results != nil {
			stageScore["entitiesCount"] = len(stage.Results.Entities)
			stageScore["relationshipsCount"] = len(stage.Results.Relationships)
		}

		scores = append(scores, stageScore)
	}

	overallConfidence := 0.0
	completedStages := 0
	for _, stage := range session.Stages {
		if stage.Status == "completed" {
			overallConfidence += stage.Confidence
			completedStages++
		}
	}

	if completedStages > 0 {
		overallConfidence /= float64(completedStages)
	}

	response := map[string]interface{}{
		"sessionId":         sessionID,
		"overallConfidence": overallConfidence,
		"stageScores":       scores,
		"completedStages":   completedStages,
		"totalStages":       len(session.Stages),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
