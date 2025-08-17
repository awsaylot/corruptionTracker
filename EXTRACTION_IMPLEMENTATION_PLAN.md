# Article Extraction System Implementation Plan### 1. LLM Integration

### 1. Sequential Analysis System
- [ ] Multi-stage Analysis Pipeline
  - Initial entity extraction
  - Relationship ### 4. Neo4j Schema
- [ ] Update schema for
  - Article sources
  - Extraction metadata
  - Confidence scores
  - Temporal relationships
  - Analysis stages
  - Evidence chains
  - Hypothesis trees
  - Cross-reference links

### 5. API Endpoints
- [ ] Update /api/extraction/url
  - Add depth parameter
  - Add stage configuration
  - Support SSE for stage progress
  - Add analysis termination control

- [ ] Create /api/analysis/
  - POST /stages - Configure analysis stages
  - GET /progress - Get analysis progress
  - POST /terminate - Stop analysis
  - GET /evidence - Get evidence chains
  - GET /hypotheses - Get hypothesis trees
  - POST /depth - Update analysis depth
  - GET /confidence - Get confidence scores

## Browser Integrationcation
  - Deep context analysis
  - Cross-reference verification
  - Hypothesis generation and testing
  - Confidence scoring with evidence

### 2. Prompt Engineering
- [x] Design base extraction prompt
  - Entity identification
  - Relationship extraction
  - Temporal information
  - Location information
  - Money/value information

- [ ] Design sequential prompts
  - Stage 1: Surface extraction
    * Basic entities and relationships
    * Initial temporal sequence
    * Geographic scope identification
  
  - Stage 2: Deep Analysis
    * Entity role analysis
    * Relationship strength assessment
    * Motivation inference
    * Pattern recognition
  
  - Stage 3: Cross-Reference
    * Internal consistency check
    * Historical context comparison
    * Pattern matching with known corruption schemes
  
  - Stage 4: Hypothesis Generation
    * Generate possible explanations
    * Identify missing information
    * Suggest additional sources
  
  - Stage N: Recursive Refinement
    * Depth configurable from frontend
    * Each level builds on previous insights
    * Confidence threshold checks

### 3. Response Processing
- [x] Design base response format
  - JSON structure
  - Entity validation
  - Relationship validation
  - Confidence scoring

- [ ] Enhanced Response Structure
  - Analysis depth tracking
  - Stage-specific outputs
  - Confidence evolution tracking
  - Evidence chains
  - Hypothesis trees
  - Missing information markers

### 4. Analysis Flow Control
- [ ] Create AnalysisController
  - Depth configuration
  - Stage progression management
  - Confidence threshold gates
  - Resource usage monitoring
  - Analysis termination conditions

Files Created:
- [x] `clank/internal/llm/extraction.go`: Article processing with LLM
- [x] `clank/internal/db/article_store.go`: Neo4j storage for articles and entities

Files to Create:
- [ ] `clank/internal/llm/sequential/`
  - `controller.go`: Analysis flow management
  - `stages.go`: Stage-specific prompt templates
  - `prompts.go`: Dynamic prompt generation
  - `responses.go`: Enhanced response structures
  - `evidence.go`: Evidence chain tracking
  - `hypotheses.go`: Hypothesis management

- [ ] `clank/internal/models/analysis.go`
  - Sequential analysis data structures
  - Stage progression tracking
  - Evidence chain models
  - Hypothesis tree models

- [ ] `clank/pkg/extraction/sequential/`
  - `analyzer.go`: Sequential analysis orchestration
  - `validator.go`: Multi-stage validation
  - `confidence.go`: Confidence scoring system
  - `synthesis.go`: Cross-stage insight synthesis
Build a system to extract structured information from news articles about corruption. The system will accept a URL, use Playwright to scrape the article content, process it through an LLM, and store the extracted entities and relationships in Neo4j.

## Frontend Components

### 1. URL Input Interface
- [ ] Create URLInput component
  - Clean URL input field
  - URL validation (news sites, valid format)
  - Submit button
  - Loading state handling

### 2. Extraction Status Component
- [ ] Create ExtractionStatus component
  - Show scraping status
  - Show LLM processing status
  - Show database update status
  - Error display

### 3. Extraction Results Interface
- [ ] Create ExtractionResults component
  - Visual representation of extracted entities
  - Graph preview of relationships
  - Ability to edit/correct extracted data
  - Confirmation interface before saving

### 4. Error Handling
- [ ] Create error boundary for extraction page
- [ ] Add error states for each component
- [ ] Add retry mechanisms
- [ ] Add user feedback mechanisms

## Backend API Endpoints

### 1. URL Processing Endpoint
- [ ] POST /api/extraction/url
  - Accept URL
  - Validate URL (news sites whitelist)
  - Use browser_automation to scrape content
  - Clean and format article text
  - Send to LLM for processing
  - Store results in Neo4j
  - Return status/results
  - Support progress updates (SSE/WebSocket)

## LLM Integration

### 1. Prompt Engineering
- [ ] Design extraction prompt
  - Entity identification
  - Relationship extraction
  - Temporal information
  - Location information
  - Money/value information

### 2. Response Processing
- [ ] Design response format
  - JSON structure
  - Entity validation
  - Relationship validation
  - Confidence scoring

## Data Models

### 1. Frontend Types
- [ ] Create types/extraction.ts
  - Article input types
  - Processing state types
  - Result types
  - Error types
  - Analysis depth configuration
  - Stage progression tracking
  - Hypothesis visualization types
  - Evidence chain types

- [ ] Create types/analysis.ts
  - Sequential analysis types
  - Stage configuration
  - Evidence tracking
  - Hypothesis tree types
  - Confidence scoring types
  - Analysis termination conditions

### 2. Backend Models
- [ ] Update models/basic.go
  - Add article source tracking
  - Add confidence scores
  - Add extraction metadata
  - Add sequential analysis tracking
  - Add evidence chains
  - Add hypothesis trees

### 3. Frontend Components
- [ ] Create components/analysis/
  - AnalysisDepthSelector.tsx
  - StageProgressTracker.tsx
  - HypothesisTree.tsx
  - EvidenceChainVisualizer.tsx
  - ConfidenceTracker.tsx
  - AnalysisTerminator.tsx

### 4. Neo4j Schema
- [ ] Update schema for
  - Article sources
  - Extraction metadata
  - Confidence scores
  - Temporal relationships

## Browser Integration

### 1. Article Scraping
- [ ] Extend browser_automation.go
  - Add article-specific selectors (content, title, date, author)
  - Add site-specific handling (different news sites)
  - Implement content cleaning
  - Extract metadata (publish date, source, author)
  - Handle paywalls/cookies/popups

### 2. Content Processing
- [ ] Create article_processor.go
  - Remove ads and navigation
  - Extract main article text
  - Handle different article formats
  - Prepare text for LLM processing

## Implementation Order

1. Sequential Analysis System
   - [ ] Create base analysis framework
     * Implement stage controller
     * Define stage interfaces
     * Set up pipeline orchestration
     * Add depth configuration

   Files to Create:
   - [ ] `clank/internal/llm/sequential/controller.go`: Analysis flow control
   - [ ] `clank/internal/llm/sequential/stages.go`: Stage definitions
   - [ ] `clank/internal/llm/sequential/pipeline.go`: Pipeline orchestration
   - [ ] `clank/internal/models/analysis.go`: Analysis models

2. Evidence and Hypothesis System
   - [ ] Implement evidence tracking
   - [ ] Add hypothesis generation
   - [ ] Create confidence scoring
   - [ ] Set up cross-referencing

   Files to Create:
   - [ ] `clank/pkg/extraction/sequential/evidence.go`: Evidence tracking
   - [ ] `clank/pkg/extraction/sequential/hypothesis.go`: Hypothesis management
   - [ ] `clank/pkg/extraction/sequential/confidence.go`: Confidence scoring
   - [ ] `clank/pkg/extraction/sequential/crossref.go`: Cross-reference system

3. Frontend Integration
   - [ ] Create analysis configuration UI
   - [ ] Add progress visualization
   - [ ] Implement evidence display
   - [ ] Add hypothesis explorer

   Files to Create:
   - [ ] `frontend/components/analysis/AnalysisConfig.tsx`
   - [ ] `frontend/components/analysis/ProgressTracker.tsx`
   - [ ] `frontend/components/analysis/EvidenceChain.tsx`
   - [ ] `frontend/components/analysis/HypothesisTree.tsx`

4. Browser Automation Enhancement (Completed)
   - [x] Add article-specific selectors to browser_automation.go
   - [x] Implement content cleaning
   - [x] Add metadata extraction
   - [x] Handle common news sites

   Files Created:
   - [x] `clank/internal/tools/browser/article_scraper.go`: New article-specific scraping functions
   - [x] `clank/pkg/extraction/sites/news_sites.go`: News site-specific selectors and rules
   - [x] `clank/pkg/extraction/sites/generic.go`: Generic article extraction rules
   - [x] `clank/pkg/extraction/processor.go`: Content cleaning and processing utilities

   Files Modified:
   - [x] `clank/internal/tools/browser/browser_automation.go`: Added article-specific methods

2. Backend Processing
   - [x] Create URL processing endpoint
   - [x] Integrate with browser automation
   - [ ] Add LLM processing
   - [ ] Implement Neo4j storage

   Files Created:
   - [x] `clank/internal/api/handlers/extraction.go`: URL processing endpoint handler
   - [x] `clank/internal/api/handlers/health.go`: Health check endpoint
   - [x] `clank/internal/api/routes/routes.go`: API route configuration
   - [x] `clank/internal/models/article.go`: Article data structures
   - [x] `clank/internal/llm/prompts/article_extraction.go`: LLM prompt templates
   - [x] `clank/internal/db/queries/article_queries.go`: Neo4j queries for articles

   Files to Modify (Next):
   - [ ] `clank/internal/llm/client.go`: Add article processing method
   - [ ] `clank/internal/db/neo4j.go`: Add article storage methods

3. Frontend Development
   - [ ] Create URL input component
   - [ ] Add validation
   - [ ] Show extraction progress
   - [ ] Display results/errors

   Files to Create:
   - `frontend/components/extraction/URLInput.tsx`: URL input and validation
   - `frontend/components/extraction/ExtractionStatus.tsx`: Progress tracking
   - `frontend/components/extraction/ResultsDisplay.tsx`: Results display
   - `frontend/types/extraction.ts`: TypeScript types for extraction

   Files to Modify:
   - `frontend/pages/extraction/index.tsx`: Implement extraction page
   - `frontend/utils/api.ts`: Add extraction API calls

4. Error Handling
   - [ ] Handle invalid URLs
   - [ ] Handle scraping failures
   - [ ] Handle LLM failures
   - [ ] Add retry mechanisms

   Files to Create:
   - `clank/pkg/extraction/errors/extraction_errors.go`: Custom error types
   - `frontend/components/extraction/ErrorDisplay.tsx`: Error display component

   Files to Modify:
   - All previously created files to add error handling

5. Testing & Validation
   - [ ] Test with various news sites
   - [ ] Test error scenarios
   - [ ] Performance testing
   - [ ] Load testing

   Files to Create:
   - `clank/internal/tools/browser/article_scraper_test.go`: Scraper tests
   - `clank/pkg/extraction/processor_test.go`: Content processor tests
   - `clank/test/e2e/extraction_test.go`: E2E tests
   - `frontend/components/extraction/__tests__/`: Frontend component tests

## File Structure

```
frontend/
  ├── components/
  │   └── extraction/
  │       ├── URLInput.tsx
  │       ├── ExtractionStatus.tsx
  │       └── ResultsDisplay.tsx
  ├── types/
  │   └── extraction.ts
  └── pages/
      └── extraction/
          └── index.tsx

clank/
  ├── internal/
  │   ├── api/
  │   │   └── handlers/
  │   │       └── extraction.go
  │   ├── models/
  │   │   └── article.go
  │   └── tools/
  │       └── browser/
  │           ├── browser_automation.go
  │           ├── article_scraper.go
  │           └── content_processor.go
  └── pkg/
      └── extraction/
          ├── sites/
          │   ├── generic.go
          │   └── news_sites.go
          └── processor.go
```

## Milestones

1. Browser Automation (2 days)
   - Enhance browser automation for article scraping
   - Implement content cleaning
   - Handle different news sites

2. Basic Pipeline (2 days)
   - Create URL processing endpoint
   - Implement LLM processing
   - Basic Neo4j storage

3. Frontend Integration (1 day)
   - URL input implementation
   - Progress tracking
   - Error handling

4. Testing & Refinement (2 days)
   - Test with various news sites
   - Add retry mechanisms
   - Performance optimization
