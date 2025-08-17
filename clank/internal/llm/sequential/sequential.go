package sequential

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"clank/internal/llm"
	"clank/internal/models"
)

// SurfaceExtractionStage handles initial entity and relationship extraction
type SurfaceExtractionStage struct {
	llmClient *llm.Client
}

func NewSurfaceExtractionStage() *SurfaceExtractionStage {
	return &SurfaceExtractionStage{
		llmClient: nil, // Will be injected by the controller
	}
}

func (s *SurfaceExtractionStage) WithLLMClient(client *llm.Client) *SurfaceExtractionStage {
	s.llmClient = client
	return s
}

func (s *SurfaceExtractionStage) GetName() string {
	return "Surface Extraction"
}

func (s *SurfaceExtractionStage) GetDescription() string {
	return "Extract basic entities, relationships, and temporal information from the article"
}

func (s *SurfaceExtractionStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
	// Get LLM client from session context or create one
	if s.llmClient == nil {
		return fmt.Errorf("LLM client not initialized")
	}

	prompt := fmt.Sprintf(`Perform initial entity extraction from this article. Focus on identifying:

1. PEOPLE: Names, roles, positions, organizations they're affiliated with
2. ORGANIZATIONS: Companies, government agencies, institutions
3. LOCATIONS: Cities, countries, specific addresses or venues
4. MONEY: Amounts, currencies, contracts, payments
5. TIME: Dates, time periods, sequences of events

Article: %s
Title: %s
Content: %s

Extract entities and relationships in this JSON format:
{
  "entities": [
    {
      "id": "unique_id",
      "type": "person|organization|location|money|time",
      "name": "entity_name",
      "properties": {
        "role": "string",
        "description": "string",
        "context": "where mentioned in article"
      },
      "confidence": 0.0-1.0,
      "mentions": [
        {
          "text": "exact text from article",
          "context": "surrounding sentence"
        }
      ]
    }
  ],
  "relationships": [
    {
      "id": "unique_id",
      "type": "payment|employment|ownership|investigation|accusation",
      "fromId": "source_entity_id",
      "toId": "target_entity_id",
      "properties": {
        "amount": "money amount if applicable",
        "date": "when relationship occurred",
        "details": "additional context"
      },
      "confidence": 0.0-1.0,
      "context": "relevant quote from article"
    }
  ],
  "confidence": 0.0-1.0
}`, article.URL, article.Title, article.Content)

	// Use LLM to extract entities
	messages := []llm.Message{
		{Role: "system", Content: "You are an expert entity extraction system for corruption analysis."},
		{Role: "user", Content: prompt},
	}

	resp, err := s.llmClient.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no response from LLM")
	}

	// Parse the response
	var result models.ExtractionResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Set metadata
	now := time.Now()
	for i := range result.Entities {
		result.Entities[i].ArticleID = article.ID
		result.Entities[i].ExtractedAt = now
	}

	for i := range result.Relationships {
		result.Relationships[i].ArticleID = article.ID
		result.Relationships[i].ExtractedAt = now
	}

	stage.Results = &result
	stage.Confidence = result.Confidence
	stage.Insights = []string{
		fmt.Sprintf("Extracted %d entities and %d relationships", len(result.Entities), len(result.Relationships)),
		"Initial extraction complete with basic entity recognition",
	}

	return nil
}

// DeepAnalysisStage performs deeper analysis of extracted entities
type DeepAnalysisStage struct {
	llmClient *llm.Client
}

func NewDeepAnalysisStage() *DeepAnalysisStage {
	return &DeepAnalysisStage{
		llmClient: nil,
	}
}

func (s *DeepAnalysisStage) WithLLMClient(client *llm.Client) *DeepAnalysisStage {
	s.llmClient = client
	return s
}

func (s *DeepAnalysisStage) GetName() string {
	return "Deep Analysis"
}

func (s *DeepAnalysisStage) GetDescription() string {
	return "Analyze entity roles, relationship strengths, motivations, and patterns"
}

func (s *DeepAnalysisStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
	if len(previousResults) == 0 {
		return fmt.Errorf("no previous results to analyze")
	}

	lastResult := previousResults[len(previousResults)-1]

	// Serialize previous results for analysis
	prevData, _ := json.Marshal(lastResult)

	prompt := fmt.Sprintf(`Perform deep analysis of the extracted entities and relationships. Focus on:

1. ROLE ANALYSIS: What roles do people play? Are they victims, perpetrators, investigators, witnesses?
2. RELATIONSHIP STRENGTH: How strong/direct are the connections? What's the evidence quality?
3. MOTIVATION ANALYSIS: What might motivate the relationships and actions described?
4. PATTERN RECOGNITION: Do you see corruption patterns, conflict of interest, quid pro quo?
5. POWER DYNAMICS: Who has power/influence over whom?

Previous extraction results:
%s

Original article:
%s

Provide enhanced analysis in this JSON format:
{
  "entities": [
    {
      "id": "entity_id_from_previous_stage",
      "type": "same_as_before",
      "name": "same_as_before", 
      "properties": {
        "role_analysis": "detailed role description",
        "influence_level": "high|medium|low",
        "corruption_risk": "high|medium|low",
        "motivations": ["list", "of", "possible", "motivations"],
        "power_indicators": ["signs", "of", "power", "or", "influence"]
      },
      "confidence": 0.0-1.0,
      "mentions": "same_as_before"
    }
  ],
  "relationships": [
    {
      "id": "relationship_id_from_previous_stage", 
      "type": "same_or_refined_type",
      "fromId": "same",
      "toId": "same",
      "properties": {
        "strength": "strong|medium|weak",
        "corruption_indicators": ["red", "flags", "identified"],
        "pattern_match": "type of corruption pattern if any",
        "evidence_quality": "high|medium|low",
        "timeline_importance": "critical|important|minor"
      },
      "confidence": 0.0-1.0,
      "context": "same_or_enhanced"
    }
  ],
  "insights": ["key insights from deep analysis"],
  "patterns": ["corruption patterns identified"],
  "confidence": 0.0-1.0
}`, string(prevData), article.Content)

	messages := []llm.Message{
		{Role: "system", Content: "You are an expert corruption analyst with deep knowledge of corruption patterns, power dynamics, and investigative techniques."},
		{Role: "user", Content: prompt},
	}

	resp, err := s.llmClient.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	var analysisResult struct {
		Entities      []models.ExtractedEntity       `json:"entities"`
		Relationships []models.ExtractedRelationship `json:"relationships"`
		Insights      []string                       `json:"insights"`
		Patterns      []string                       `json:"patterns"`
		Confidence    float64                        `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysisResult); err != nil {
		return fmt.Errorf("failed to parse analysis response: %w", err)
	}

	// Create enhanced result
	result := &models.ExtractionResult{
		Entities:      analysisResult.Entities,
		Relationships: analysisResult.Relationships,
		Confidence:    analysisResult.Confidence,
	}

	stage.Results = result
	stage.Confidence = result.Confidence
	stage.Insights = append(analysisResult.Insights, analysisResult.Patterns...)

	return nil
}

// CrossReferenceStage validates consistency and identifies conflicts
type CrossReferenceStage struct {
	llmClient *llm.Client
}

func NewCrossReferenceStage() *CrossReferenceStage {
	return &CrossReferenceStage{
		llmClient: nil,
	}
}

func (s *CrossReferenceStage) WithLLMClient(client *llm.Client) *CrossReferenceStage {
	s.llmClient = client
	return s
}

func (s *CrossReferenceStage) GetName() string {
	return "Cross-Reference Validation"
}

func (s *CrossReferenceStage) GetDescription() string {
	return "Check internal consistency, identify conflicts, and validate against known patterns"
}

func (s *CrossReferenceStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
	if len(previousResults) < 2 {
		// Skip if we don't have enough previous results
		stage.Results = previousResults[len(previousResults)-1]
		stage.Confidence = 0.8
		stage.Insights = []string{"Cross-reference validation skipped - insufficient previous stages"}
		return nil
	}

	// Compare the last two results
	prevData, _ := json.Marshal(previousResults)

	prompt := fmt.Sprintf(`Cross-reference and validate the analysis results for consistency and accuracy:

Previous analysis stages:
%s

Article content:
%s

Perform these validation checks:
1. INTERNAL CONSISTENCY: Do the entities and relationships make sense together?
2. FACT CHECKING: Are claims supported by evidence in the article?
3. TIMELINE COHERENCE: Do temporal relationships make sense?
4. LOGICAL CONSISTENCY: Are there contradictions in the analysis?
5. PATTERN VALIDATION: Do identified corruption patterns actually match the evidence?

Return validation results in JSON format:
{
  "validation_results": {
    "consistency_score": 0.0-1.0,
    "fact_check_score": 0.0-1.0, 
    "timeline_coherence": 0.0-1.0,
    "logical_consistency": 0.0-1.0,
    "pattern_validation": 0.0-1.0
  },
  "issues_found": [
    {
      "type": "inconsistency|unsupported_claim|timeline_error|logical_error",
      "description": "description of the issue",
      "severity": "high|medium|low",
      "affected_entities": ["entity_ids"],
      "suggested_fix": "how to resolve this issue"
    }
  ],
  "validated_entities": [], // corrected entities
  "validated_relationships": [], // corrected relationships
  "confidence": 0.0-1.0
}`, string(prevData), article.Content)

	messages := []llm.Message{
		{Role: "system", Content: "You are a meticulous fact-checker and validation expert specializing in corruption investigations."},
		{Role: "user", Content: prompt},
	}

	resp, err := s.llmClient.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse validation response
	var validation struct {
		ValidationResults struct {
			ConsistencyScore   float64 `json:"consistency_score"`
			FactCheckScore     float64 `json:"fact_check_score"`
			TimelineCoherence  float64 `json:"timeline_coherence"`
			LogicalConsistency float64 `json:"logical_consistency"`
			PatternValidation  float64 `json:"pattern_validation"`
		} `json:"validation_results"`
		IssuesFound            []map[string]interface{}       `json:"issues_found"`
		ValidatedEntities      []models.ExtractedEntity       `json:"validated_entities"`
		ValidatedRelationships []models.ExtractedRelationship `json:"validated_relationships"`
		Confidence             float64                        `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &validation); err != nil {
		return fmt.Errorf("failed to parse validation response: %w", err)
	}

	result := &models.ExtractionResult{
		Entities:      validation.ValidatedEntities,
		Relationships: validation.ValidatedRelationships,
		Confidence:    validation.Confidence,
	}

	stage.Results = result
	stage.Confidence = result.Confidence
	stage.Insights = []string{
		fmt.Sprintf("Consistency score: %.2f", validation.ValidationResults.ConsistencyScore),
		fmt.Sprintf("Found %d validation issues", len(validation.IssuesFound)),
		"Cross-reference validation completed",
	}

	return nil
}

// HypothesisGenerationStage generates possible explanations and theories
type HypothesisGenerationStage struct {
	llmClient *llm.Client
}

func NewHypothesisGenerationStage() *HypothesisGenerationStage {
	return &HypothesisGenerationStage{
		llmClient: nil,
	}
}

func (s *HypothesisGenerationStage) WithLLMClient(client *llm.Client) *HypothesisGenerationStage {
	s.llmClient = client
	return s
}

func (s *HypothesisGenerationStage) GetName() string {
	return "Hypothesis Generation"
}

func (s *HypothesisGenerationStage) GetDescription() string {
	return "Generate possible explanations, identify missing information, and suggest follow-up investigations"
}

func (s *HypothesisGenerationStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
	allResults, _ := json.Marshal(previousResults)

	prompt := fmt.Sprintf(`Based on the analysis results, generate hypotheses and identify gaps in information:

Analysis results so far:
%s

Article:
%s

Generate hypotheses and analysis in JSON format:
{
  "hypotheses": [
    {
      "id": "hypothesis_1", 
      "description": "Detailed hypothesis description",
      "type": "corruption|conflict_of_interest|fraud|bribery|other",
      "confidence": 0.0-1.0,
      "supporting_evidence": ["evidence that supports this hypothesis"],
      "contradicting_evidence": ["evidence that contradicts this"],
      "required_evidence": ["what evidence would confirm/deny this"],
      "implications": ["what this would mean if true"]
    }
  ],
  "missing_information": [
    {
      "type": "financial_records|witness_statements|timeline_gaps|relationship_details",
      "description": "What information is missing",
      "importance": "critical|important|minor",
      "potential_sources": ["where this info might be found"]
    }
  ],
  "follow_up_questions": ["Questions that should be investigated further"],
  "confidence": 0.0-1.0
}`, string(allResults), article.Content)

	messages := []llm.Message{
		{Role: "system", Content: "You are an investigative analyst expert at generating theories and identifying information gaps in corruption cases."},
		{Role: "user", Content: prompt},
	}

	resp, err := s.llmClient.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	var hypothesesResult struct {
		Hypotheses         []Hypothesis             `json:"hypotheses"`
		MissingInformation []map[string]interface{} `json:"missing_information"`
		FollowUpQuestions  []string                 `json:"follow_up_questions"`
		Confidence         float64                  `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &hypothesesResult); err != nil {
		return fmt.Errorf("failed to parse hypotheses response: %w", err)
	}

	// Add hypotheses to session
	session.Hypotheses = append(session.Hypotheses, hypothesesResult.Hypotheses...)

	// Use the last result as base and enhance it
	lastResult := previousResults[len(previousResults)-1]
	stage.Results = lastResult
	stage.Confidence = hypothesesResult.Confidence
	stage.Insights = append([]string{
		fmt.Sprintf("Generated %d hypotheses", len(hypothesesResult.Hypotheses)),
		fmt.Sprintf("Identified %d information gaps", len(hypothesesResult.MissingInformation)),
	}, hypothesesResult.FollowUpQuestions...)
	stage.Questions = hypothesesResult.FollowUpQuestions

	return nil
}

// RecursiveRefinementStage performs additional refinement passes
type RecursiveRefinementStage struct {
	llmClient *llm.Client
}

func NewRecursiveRefinementStage() *RecursiveRefinementStage {
	return &RecursiveRefinementStage{
		llmClient: nil,
	}
}

func (s *RecursiveRefinementStage) WithLLMClient(client *llm.Client) *RecursiveRefinementStage {
	s.llmClient = client
	return s
}

func (s *RecursiveRefinementStage) GetName() string {
	return "Recursive Refinement"
}

func (s *RecursiveRefinementStage) GetDescription() string {
	return "Perform additional refinement and synthesis of all previous analysis"
}

func (s *RecursiveRefinementStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
	// Synthesize all previous results
	allData, _ := json.Marshal(map[string]interface{}{
		"previous_results": previousResults,
		"hypotheses":       session.Hypotheses,
		"evidence":         session.Evidence,
	})

	prompt := fmt.Sprintf(`Perform final synthesis and refinement of the complete analysis:

Complete analysis data:
%s

Article:
%s

Provide final refined analysis with:
1. Most confident entities and relationships
2. Key insights and conclusions  
3. Confidence assessment
4. Summary of corruption indicators
5. Recommended next steps

JSON format:
{
  "final_entities": [], // most confident/important entities
  "final_relationships": [], // most confident/important relationships  
  "key_insights": ["List of the most important insights"],
  "corruption_indicators": {
    "financial_irregularities": ["list"],
    "conflict_of_interest": ["list"], 
    "abuse_of_power": ["list"],
    "lack_of_transparency": ["list"]
  },
  "confidence_assessment": {
    "overall_confidence": 0.0-1.0,
    "data_quality": "high|medium|low",
    "evidence_strength": "strong|moderate|weak"
  },
  "next_steps": ["Recommended follow-up actions"],
  "confidence": 0.0-1.0
}`, string(allData), article.Content)

	messages := []llm.Message{
		{Role: "system", Content: "You are a senior investigative analyst providing final synthesis of a complex corruption analysis."},
		{Role: "user", Content: prompt},
	}

	resp, err := s.llmClient.Generate(ctx, messages)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	var finalResult struct {
		FinalEntities        []models.ExtractedEntity       `json:"final_entities"`
		FinalRelationships   []models.ExtractedRelationship `json:"final_relationships"`
		KeyInsights          []string                       `json:"key_insights"`
		CorruptionIndicators map[string][]string            `json:"corruption_indicators"`
		ConfidenceAssessment map[string]interface{}         `json:"confidence_assessment"`
		NextSteps            []string                       `json:"next_steps"`
		Confidence           float64                        `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &finalResult); err != nil {
		return fmt.Errorf("failed to parse final refinement response: %w", err)
	}

	result := &models.ExtractionResult{
		Entities:      finalResult.FinalEntities,
		Relationships: finalResult.FinalRelationships,
		Confidence:    finalResult.Confidence,
	}

	stage.Results = result
	stage.Confidence = result.Confidence
	stage.Insights = append(finalResult.KeyInsights, finalResult.NextSteps...)

	return nil
}

// package sequential

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"clank/internal/llm"
// 	"clank/internal/models"
// )

// // SurfaceExtractionStage handles initial entity and relationship extraction
// type SurfaceExtractionStage struct {
// 	llmClient *llm.Client
// }

// func NewSurfaceExtractionStage() *SurfaceExtractionStage {
// 	return &SurfaceExtractionStage{
// 		llmClient: nil, // Will be injected by the controller
// 	}
// }

// func (s *SurfaceExtractionStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
// 	return &SurfaceExtractionStage{llmClient: client}
// }

// func (s *SurfaceExtractionStage) GetName() string {
// 	return "Surface Extraction"
// }

// func (s *SurfaceExtractionStage) GetDescription() string {
// 	return "Extract basic entities, relationships, and temporal information from the article"
// }

// func (s *SurfaceExtractionStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
// 	// Get LLM client from session context or create one
// 	if s.llmClient == nil {
// 		return fmt.Errorf("LLM client not initialized")
// 	}

// 	prompt := fmt.Sprintf(`Perform initial entity extraction from this article. Focus on identifying:

// 1. PEOPLE: Names, roles, positions, organizations they're affiliated with
// 2. ORGANIZATIONS: Companies, government agencies, institutions
// 3. LOCATIONS: Cities, countries, specific addresses or venues
// 4. MONEY: Amounts, currencies, contracts, payments
// 5. TIME: Dates, time periods, sequences of events

// Article: %s
// Title: %s
// Content: %s

// Extract entities and relationships in this JSON format:
// {
//   "entities": [
//     {
//       "id": "unique_id",
//       "type": "person|organization|location|money|time",
//       "name": "entity_name",
//       "properties": {
//         "role": "string",
//         "description": "string",
//         "context": "where mentioned in article"
//       },
//       "confidence": 0.0-1.0,
//       "mentions": [
//         {
//           "text": "exact text from article",
//           "context": "surrounding sentence"
//         }
//       ]
//     }
//   ],
//   "relationships": [
//     {
//       "id": "unique_id",
//       "type": "payment|employment|ownership|investigation|accusation",
//       "fromId": "source_entity_id",
//       "toId": "target_entity_id",
//       "properties": {
//         "amount": "money amount if applicable",
//         "date": "when relationship occurred",
//         "details": "additional context"
//       },
//       "confidence": 0.0-1.0,
//       "context": "relevant quote from article"
//     }
//   ],
//   "confidence": 0.0-1.0
// }`, article.URL, article.Title, article.Content)

// 	// Use LLM to extract entities
// 	messages := []llm.Message{
// 		{Role: "system", Content: "You are an expert entity extraction system for corruption analysis."},
// 		{Role: "user", Content: prompt},
// 	}

// 	resp, err := s.llmClient.Generate(ctx, messages)
// 	if err != nil {
// 		return fmt.Errorf("LLM generation failed: %w", err)
// 	}

// 	if len(resp.Choices) == 0 {
// 		return fmt.Errorf("no response from LLM")
// 	}

// 	// Parse the response
// 	var result models.ExtractionResult
// 	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
// 		return fmt.Errorf("failed to parse LLM response: %w", err)
// 	}

// 	// Set metadata
// 	now := time.Now()
// 	for i := range result.Entities {
// 		result.Entities[i].ArticleID = article.ID
// 		result.Entities[i].ExtractedAt = now
// 	}

// 	for i := range result.Relationships {
// 		result.Relationships[i].ArticleID = article.ID
// 		result.Relationships[i].ExtractedAt = now
// 	}

// 	stage.Results = &result
// 	stage.Confidence = result.Confidence
// 	stage.Insights = []string{
// 		fmt.Sprintf("Extracted %d entities and %d relationships", len(result.Entities), len(result.Relationships)),
// 		"Initial extraction complete with basic entity recognition",
// 	}

// 	return nil
// }

// // DeepAnalysisStage performs deeper analysis of extracted entities
// type DeepAnalysisStage struct {
// 	llmClient *llm.Client
// }

// func NewDeepAnalysisStage() *DeepAnalysisStage {
// 	return &DeepAnalysisStage{
// 		llmClient: nil,
// 	}
// }

// func (s *DeepAnalysisStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
// 	return &DeepAnalysisStage{llmClient: client}
// }

// func (s *DeepAnalysisStage) GetName() string {
// 	return "Deep Analysis"
// }

// func (s *DeepAnalysisStage) GetDescription() string {
// 	return "Analyze entity roles, relationship strengths, motivations, and patterns"
// }

// func (s *DeepAnalysisStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
// 	if len(previousResults) == 0 {
// 		return fmt.Errorf("no previous results to analyze")
// 	}

// 	lastResult := previousResults[len(previousResults)-1]

// 	// Serialize previous results for analysis
// 	prevData, _ := json.Marshal(lastResult)

// 	prompt := fmt.Sprintf(`Perform deep analysis of the extracted entities and relationships. Focus on:

// 1. ROLE ANALYSIS: What roles do people play? Are they victims, perpetrators, investigators, witnesses?
// 2. RELATIONSHIP STRENGTH: How strong/direct are the connections? What's the evidence quality?
// 3. MOTIVATION ANALYSIS: What might motivate the relationships and actions described?
// 4. PATTERN RECOGNITION: Do you see corruption patterns, conflict of interest, quid pro quo?
// 5. POWER DYNAMICS: Who has power/influence over whom?

// Previous extraction results:
// %s

// Original article:
// %s

// Provide enhanced analysis in this JSON format:
// {
//   "entities": [
//     {
//       "id": "entity_id_from_previous_stage",
//       "type": "same_as_before",
//       "name": "same_as_before",
//       "properties": {
//         "role_analysis": "detailed role description",
//         "influence_level": "high|medium|low",
//         "corruption_risk": "high|medium|low",
//         "motivations": ["list", "of", "possible", "motivations"],
//         "power_indicators": ["signs", "of", "power", "or", "influence"]
//       },
//       "confidence": 0.0-1.0,
//       "mentions": "same_as_before"
//     }
//   ],
//   "relationships": [
//     {
//       "id": "relationship_id_from_previous_stage",
//       "type": "same_or_refined_type",
//       "fromId": "same",
//       "toId": "same",
//       "properties": {
//         "strength": "strong|medium|weak",
//         "corruption_indicators": ["red", "flags", "identified"],
//         "pattern_match": "type of corruption pattern if any",
//         "evidence_quality": "high|medium|low",
//         "timeline_importance": "critical|important|minor"
//       },
//       "confidence": 0.0-1.0,
//       "context": "same_or_enhanced"
//     }
//   ],
//   "insights": ["key insights from deep analysis"],
//   "patterns": ["corruption patterns identified"],
//   "confidence": 0.0-1.0
// }`, string(prevData), article.Content)

// 	messages := []llm.Message{
// 		{Role: "system", Content: "You are an expert corruption analyst with deep knowledge of corruption patterns, power dynamics, and investigative techniques."},
// 		{Role: "user", Content: prompt},
// 	}

// 	resp, err := s.llmClient.Generate(ctx, messages)
// 	if err != nil {
// 		return fmt.Errorf("LLM generation failed: %w", err)
// 	}

// 	// Parse response
// 	var analysisResult struct {
// 		Entities      []models.ExtractedEntity       `json:"entities"`
// 		Relationships []models.ExtractedRelationship `json:"relationships"`
// 		Insights      []string                       `json:"insights"`
// 		Patterns      []string                       `json:"patterns"`
// 		Confidence    float64                        `json:"confidence"`
// 	}

// 	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysisResult); err != nil {
// 		return fmt.Errorf("failed to parse analysis response: %w", err)
// 	}

// 	// Create enhanced result
// 	result := &models.ExtractionResult{
// 		Entities:      analysisResult.Entities,
// 		Relationships: analysisResult.Relationships,
// 		Confidence:    analysisResult.Confidence,
// 	}

// 	stage.Results = result
// 	stage.Confidence = result.Confidence
// 	stage.Insights = append(analysisResult.Insights, analysisResult.Patterns...)

// 	return nil
// }

// // CrossReferenceStage validates consistency and identifies conflicts
// type CrossReferenceStage struct {
// 	llmClient *llm.Client
// }

// func NewCrossReferenceStage() *CrossReferenceStage {
// 	return &CrossReferenceStage{
// 		llmClient: nil,
// 	}
// }

// func (s *CrossReferenceStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
// 	return &CrossReferenceStage{llmClient: client}
// }

// func (s *CrossReferenceStage) GetName() string {
// 	return "Cross-Reference Validation"
// }

// func (s *CrossReferenceStage) GetDescription() string {
// 	return "Check internal consistency, identify conflicts, and validate against known patterns"
// }

// func (s *CrossReferenceStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
// 	if len(previousResults) < 2 {
// 		// Skip if we don't have enough previous results
// 		stage.Results = previousResults[len(previousResults)-1]
// 		stage.Confidence = 0.8
// 		stage.Insights = []string{"Cross-reference validation skipped - insufficient previous stages"}
// 		return nil
// 	}

// 	// Compare the last two results
// 	prevData, _ := json.Marshal(previousResults)

// 	prompt := fmt.Sprintf(`Cross-reference and validate the analysis results for consistency and accuracy:

// Previous analysis stages:
// %s

// Article content:
// %s

// Perform these validation checks:
// 1. INTERNAL CONSISTENCY: Do the entities and relationships make sense together?
// 2. FACT CHECKING: Are claims supported by evidence in the article?
// 3. TIMELINE COHERENCE: Do temporal relationships make sense?
// 4. LOGICAL CONSISTENCY: Are there contradictions in the analysis?
// 5. PATTERN VALIDATION: Do identified corruption patterns actually match the evidence?

// Return validation results in JSON format:
// {
//   "validation_results": {
//     "consistency_score": 0.0-1.0,
//     "fact_check_score": 0.0-1.0,
//     "timeline_coherence": 0.0-1.0,
//     "logical_consistency": 0.0-1.0,
//     "pattern_validation": 0.0-1.0
//   },
//   "issues_found": [
//     {
//       "type": "inconsistency|unsupported_claim|timeline_error|logical_error",
//       "description": "description of the issue",
//       "severity": "high|medium|low",
//       "affected_entities": ["entity_ids"],
//       "suggested_fix": "how to resolve this issue"
//     }
//   ],
//   "validated_entities": [],
//   "validated_relationships": [],
//   "confidence": 0.0-1.0
// }`, string(prevData), article.Content)

// 	messages := []llm.Message{
// 		{Role: "system", Content: "You are a meticulous fact-checker and validation expert specializing in corruption investigations."},
// 		{Role: "user", Content: prompt},
// 	}

// 	resp, err := s.llmClient.Generate(ctx, messages)
// 	if err != nil {
// 		return fmt.Errorf("LLM generation failed: %w", err)
// 	}

// 	// Parse validation response
// 	var validation struct {
// 		ValidationResults struct {
// 			ConsistencyScore   float64 `json:"consistency_score"`
// 			FactCheckScore     float64 `json:"fact_check_score"`
// 			TimelineCoherence  float64 `json:"timeline_coherence"`
// 			LogicalConsistency float64 `json:"logical_consistency"`
// 			PatternValidation  float64 `json:"pattern_validation"`
// 		} `json:"validation_results"`
// 		IssuesFound            []map[string]interface{}       `json:"issues_found"`
// 		ValidatedEntities      []models.ExtractedEntity       `json:"validated_entities"`
// 		ValidatedRelationships []models.ExtractedRelationship `json:"validated_relationships"`
// 		Confidence             float64                        `json:"confidence"`
// 	}

// 	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &validation); err != nil {
// 		return fmt.Errorf("failed to parse validation response: %w", err)
// 	}

// 	result := &models.ExtractionResult{
// 		Entities:      validation.ValidatedEntities,
// 		Relationships: validation.ValidatedRelationships,
// 		Confidence:    validation.Confidence,
// 	}

// 	stage.Results = result
// 	stage.Confidence = result.Confidence
// 	stage.Insights = []string{
// 		fmt.Sprintf("Consistency score: %.2f", validation.ValidationResults.ConsistencyScore),
// 		fmt.Sprintf("Found %d validation issues", len(validation.IssuesFound)),
// 		"Cross-reference validation completed",
// 	}

// 	return nil
// }

// // HypothesisGenerationStage generates possible explanations and theories
// type HypothesisGenerationStage struct {
// 	llmClient *llm.Client
// }

// func NewHypothesisGenerationStage() *HypothesisGenerationStage {
// 	return &HypothesisGenerationStage{
// 		llmClient: nil,
// 	}
// }

// func (s *HypothesisGenerationStage) WithLLMClient(client *llm.Client) AnalysisStageProcessor {
// 	return &HypothesisGenerationStage{llmClient: client}
// }

// func (s *HypothesisGenerationStage) GetName() string {
// 	return "Hypothesis Generation"
// }

// func (s *HypothesisGenerationStage) GetDescription() string {
// 	return "Generate possible explanations, identify missing information, and suggest follow-up investigations"
// }

// func (s *HypothesisGenerationStage) Process(ctx context.Context, session *AnalysisSession, stage *AnalysisStage, article *models.Article, previousResults []*models.ExtractionResult) error {
// 	allResults, _ := json.Marshal(previousResults)

// 	prompt := fmt.Sprintf(`Based on the analysis results, generate hypotheses and identify gaps in information:

// Analysis results so far:
// %s

// Article:
// %s

// Generate hypotheses and analysis in JSON format:
// {
//   "hypotheses": [
//     {
//       "id": "hypothesis_1",
//       "description": "Detailed hypothesis description",
//       "type": "corruption|conflict_of_interest|fraud|bribery|other",
//       "confidence": 0.0-1.0,
//       "supporting_evidence": ["evidence that supports this hypothesis"],
//       "contradicting_evidence": ["evidence that contradicts this"],
//       "required_evidence": ["what evidence would confirm/deny this"],
//       "implications": ["what this would mean if true"]
//     }
//   ],
//   "missing_information": [
//     {
//       "type": "financial_records|witness_statements|timeline_gaps|relationship_details",
//       "description": "What information is missing",
//       "importance": "critical|important|minor",
//       "potential_sources": ["where this info might be found"]
//     }
//   ],
//   "follow_up_questions": ["Questions that should be investigated further"],
