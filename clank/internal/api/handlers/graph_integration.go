package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"clank/config"
	"clank/internal/db"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// GraphIntegrationRequest represents the request to integrate extraction results into the graph
type GraphIntegrationRequest struct {
	URL           string                  `json:"url" binding:"required"`
	Source        string                  `json:"source,omitempty"`
	Entities      []ExtractedEntity       `json:"entities"`
	Relationships []ExtractedRelationship `json:"relationships"`
	Events        []ExtractedEvent        `json:"events"`
	Metadata      map[string]string       `json:"metadata,omitempty"`
}

// GraphIntegrationResponse represents the response after integration
type GraphIntegrationResponse struct {
	NodesCreated         int               `json:"nodes_created"`
	NodesUpdated         int               `json:"nodes_updated"`
	RelationshipsCreated int               `json:"relationships_created"`
	EventsCreated        int               `json:"events_created"`
	ProcessingTime       time.Duration     `json:"processing_time"`
	EntityMapping        map[string]string `json:"entity_mapping"` // name -> node_id
	Warnings             []string          `json:"warnings,omitempty"`
	Errors               []string          `json:"errors,omitempty"`
}

// GraphIntegrator handles the integration of extraction results into Neo4j
type GraphIntegrator struct {
	debug bool
}

// NewGraphIntegrator creates a new graph integrator
func NewGraphIntegrator(debug bool) *GraphIntegrator {
	return &GraphIntegrator{debug: debug}
}

// IntegrateExtraction integrates extraction results into the graph database
func (gi *GraphIntegrator) IntegrateExtraction(ctx context.Context, req *GraphIntegrationRequest) (*GraphIntegrationResponse, error) {
	startTime := time.Now()
	response := &GraphIntegrationResponse{
		EntityMapping: make(map[string]string),
		Warnings:      make([]string, 0),
		Errors:        make([]string, 0),
	}

	log.Printf("[GRAPH_INTEGRATION] Starting integration for URL: %s", req.URL)
	log.Printf("[GRAPH_INTEGRATION] Input: %d entities, %d relationships, %d events",
		len(req.Entities), len(req.Relationships), len(req.Events))

	// Process in transaction
	result, err := db.ExecuteWrite(func(tx neo4j.Transaction) (interface{}, error) {
		txResponse := &GraphIntegrationResponse{
			EntityMapping: make(map[string]string),
			Warnings:      make([]string, 0),
			Errors:        make([]string, 0),
		}

		// Step 1: Create or update entities
		entityMapping, nodesCreated, nodesUpdated, warnings, errors := gi.processEntities(tx, req.Entities, req.URL)
		txResponse.NodesCreated = nodesCreated
		txResponse.NodesUpdated = nodesUpdated
		txResponse.EntityMapping = entityMapping
		txResponse.Warnings = append(txResponse.Warnings, warnings...)
		txResponse.Errors = append(txResponse.Errors, errors...)

		if len(errors) > 0 {
			log.Printf("[GRAPH_INTEGRATION] Entity processing had errors: %v", errors)
		}

		// Step 2: Create relationships
		relsCreated, relWarnings, relErrors := gi.processRelationships(tx, req.Relationships, entityMapping, req.URL)
		txResponse.RelationshipsCreated = relsCreated
		txResponse.Warnings = append(txResponse.Warnings, relWarnings...)
		txResponse.Errors = append(txResponse.Errors, relErrors...)

		// Step 3: Create events as special nodes
		eventsCreated, eventWarnings, eventErrors := gi.processEvents(tx, req.Events, entityMapping, req.URL)
		txResponse.EventsCreated = eventsCreated
		txResponse.NodesCreated += eventsCreated // Events are nodes too
		txResponse.Warnings = append(txResponse.Warnings, eventWarnings...)
		txResponse.Errors = append(txResponse.Errors, eventErrors...)

		return txResponse, nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	response = result.(*GraphIntegrationResponse)
	response.ProcessingTime = time.Since(startTime)

	log.Printf("[GRAPH_INTEGRATION] Integration completed - Nodes: %d created, %d updated, Relationships: %d, Events: %d, Time: %v",
		response.NodesCreated, response.NodesUpdated, response.RelationshipsCreated, response.EventsCreated, response.ProcessingTime)

	return response, nil
}

// processEntities creates or updates entity nodes
func (gi *GraphIntegrator) processEntities(tx neo4j.Transaction, entities []ExtractedEntity, sourceURL string) (
	entityMapping map[string]string, nodesCreated, nodesUpdated int, warnings, errors []string) {

	entityMapping = make(map[string]string)
	warnings = make([]string, 0)
	errors = make([]string, 0)

	for _, entity := range entities {
		if entity.Name == "" {
			warnings = append(warnings, "Skipping entity with empty name")
			continue
		}

		// Normalize entity name for consistency
		normalizedName := strings.TrimSpace(entity.Name)

		// Check if entity already exists
		nodeID, exists, err := gi.findExistingEntity(tx, normalizedName, entity.Type)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error checking existing entity %s: %v", normalizedName, err))
			continue
		}

		if exists {
			// Update existing entity
			updatedID, err := gi.updateEntity(tx, nodeID, entity, sourceURL)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error updating entity %s: %v", normalizedName, err))
				continue
			}
			entityMapping[normalizedName] = updatedID
			nodesUpdated++

			if gi.debug {
				log.Printf("[GRAPH_INTEGRATION] Updated existing entity: %s (ID: %s)", normalizedName, updatedID)
			}
		} else {
			// Create new entity
			newID, err := gi.createEntity(tx, entity, sourceURL)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error creating entity %s: %v", normalizedName, err))
				continue
			}
			entityMapping[normalizedName] = newID
			nodesCreated++

			if gi.debug {
				log.Printf("[GRAPH_INTEGRATION] Created new entity: %s (ID: %s)", normalizedName, newID)
			}
		}
	}

	return entityMapping, nodesCreated, nodesUpdated, warnings, errors
}

// findExistingEntity checks if an entity already exists in the database
func (gi *GraphIntegrator) findExistingEntity(tx neo4j.Transaction, name, entityType string) (string, bool, error) {
	query := `
		MATCH (n:%s)
		WHERE toLower(n.name) = toLower($name)
		RETURN ID(n) as id
		LIMIT 1
	`

	result, err := tx.Run(fmt.Sprintf(query, entityType), map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return "", false, err
	}

	if result.Next() {
		record := result.Record()
		nodeID := fmt.Sprint(record.Values[0])
		return nodeID, true, nil
	}

	return "", false, nil
}

// createEntity creates a new entity node
func (gi *GraphIntegrator) createEntity(tx neo4j.Transaction, entity ExtractedEntity, sourceURL string) (string, error) {
	now := time.Now()

	// Prepare properties
	props := make(map[string]interface{})
	props["name"] = strings.TrimSpace(entity.Name)
	props["created_at"] = now
	props["updated_at"] = now
	props["source_url"] = sourceURL
	props["confidence"] = entity.Confidence

	// Add entity-specific properties
	if entity.Properties != nil {
		for k, v := range entity.Properties {
			props[k] = v
		}
	}

	// Add mentions if available
	if len(entity.Mentions) > 0 {
		props["mentions"] = entity.Mentions
	}

	query := `
		CREATE (n:%s $props)
		RETURN ID(n) as id
	`

	result, err := tx.Run(fmt.Sprintf(query, entity.Type), map[string]interface{}{
		"props": props,
	})
	if err != nil {
		return "", err
	}

	if !result.Next() {
		return "", fmt.Errorf("failed to create entity node")
	}

	record := result.Record()
	return fmt.Sprint(record.Values[0]), nil
}

// updateEntity updates an existing entity node
func (gi *GraphIntegrator) updateEntity(tx neo4j.Transaction, nodeID string, entity ExtractedEntity, sourceURL string) (string, error) {
	now := time.Now()

	// Prepare update properties
	updateProps := make(map[string]interface{})
	updateProps["updated_at"] = now

	// Add new source URL to sources array
	updateProps["source_urls"] = []string{sourceURL} // Will be merged with existing

	// Update confidence if higher
	if entity.Confidence > 0 {
		updateProps["confidence"] = entity.Confidence
	}

	// Merge entity-specific properties
	if entity.Properties != nil {
		for k, v := range entity.Properties {
			updateProps[k] = v
		}
	}

	query := `
		MATCH (n)
		WHERE ID(n) = $nodeId
		SET n.updated_at = $updated_at
		SET n.source_urls = 
			CASE 
				WHEN n.source_urls IS NULL THEN [$source_url]
				WHEN NOT $source_url IN n.source_urls THEN n.source_urls + [$source_url]
				ELSE n.source_urls
			END
		SET n.confidence = 
			CASE 
				WHEN $confidence > COALESCE(n.confidence, 0) THEN $confidence
				ELSE COALESCE(n.confidence, $confidence)
			END
		SET n += $props
		RETURN ID(n) as id
	`

	params := map[string]interface{}{
		"nodeId":     nodeID,
		"updated_at": now,
		"source_url": sourceURL,
		"confidence": entity.Confidence,
		"props":      entity.Properties,
	}

	result, err := tx.Run(query, params)
	if err != nil {
		return "", err
	}

	if !result.Next() {
		return "", fmt.Errorf("failed to update entity node")
	}

	return nodeID, nil
}

// processRelationships creates relationships between entities
func (gi *GraphIntegrator) processRelationships(tx neo4j.Transaction, relationships []ExtractedRelationship,
	entityMapping map[string]string, sourceURL string) (int, []string, []string) {

	created := 0
	warnings := make([]string, 0)
	errors := make([]string, 0)

	for _, rel := range relationships {
		// Normalize entity names
		fromName := strings.TrimSpace(rel.FromEntity)
		toName := strings.TrimSpace(rel.ToEntity)

		// Check if both entities exist in our mapping
		fromID, fromExists := entityMapping[fromName]
		toID, toExists := entityMapping[toName]

		if !fromExists {
			warnings = append(warnings, fmt.Sprintf("From entity '%s' not found for relationship", fromName))
			continue
		}
		if !toExists {
			warnings = append(warnings, fmt.Sprintf("To entity '%s' not found for relationship", toName))
			continue
		}

		// Create relationship
		err := gi.createRelationship(tx, fromID, toID, rel, sourceURL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error creating relationship %s -> %s: %v", fromName, toName, err))
			continue
		}

		created++

		if gi.debug {
			log.Printf("[GRAPH_INTEGRATION] Created relationship: %s -%s-> %s", fromName, rel.Type, toName)
		}
	}

	return created, warnings, errors
}

// createRelationship creates a relationship between two nodes
func (gi *GraphIntegrator) createRelationship(tx neo4j.Transaction, fromID, toID string,
	rel ExtractedRelationship, sourceURL string) error {

	now := time.Now()

	// Prepare relationship properties
	props := make(map[string]interface{})
	props["created_at"] = now
	props["source_url"] = sourceURL
	props["confidence"] = rel.Confidence

	if rel.Evidence != "" {
		props["evidence"] = rel.Evidence
	}

	// Add relationship-specific properties
	if rel.Properties != nil {
		for k, v := range rel.Properties {
			props[k] = v
		}
	}

	query := `
		MATCH (from), (to)
		WHERE ID(from) = $fromId AND ID(to) = $toId
		CREATE (from)-[r:%s $props]->(to)
		RETURN ID(r) as id
	`

	result, err := tx.Run(fmt.Sprintf(query, rel.Type), map[string]interface{}{
		"fromId": fromID,
		"toId":   toID,
		"props":  props,
	})
	if err != nil {
		return err
	}

	if !result.Next() {
		return fmt.Errorf("failed to create relationship")
	}

	return nil
}

// processEvents creates event nodes and connects them to related entities
func (gi *GraphIntegrator) processEvents(tx neo4j.Transaction, events []ExtractedEvent,
	entityMapping map[string]string, sourceURL string) (int, []string, []string) {

	created := 0
	warnings := make([]string, 0)
	errors := make([]string, 0)

	for _, event := range events {
		if event.Description == "" {
			warnings = append(warnings, "Skipping event with empty description")
			continue
		}

		// Create event node
		eventID, err := gi.createEventNode(tx, event, sourceURL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error creating event node: %v", err))
			continue
		}

		// Connect event to related entities
		for _, entityName := range event.Entities {
			normalizedName := strings.TrimSpace(entityName)
			entityID, exists := entityMapping[normalizedName]
			if !exists {
				warnings = append(warnings, fmt.Sprintf("Entity '%s' not found for event connection", normalizedName))
				continue
			}

			// Create relationship from entity to event
			err := gi.createEventRelationship(tx, entityID, eventID, sourceURL)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Error connecting entity to event: %v", err))
			}
		}

		created++

		if gi.debug {
			log.Printf("[GRAPH_INTEGRATION] Created event: %s (connected to %d entities)",
				event.Description, len(event.Entities))
		}
	}

	return created, warnings, errors
}

// createEventNode creates an event node
func (gi *GraphIntegrator) createEventNode(tx neo4j.Transaction, event ExtractedEvent, sourceURL string) (string, error) {
	now := time.Now()

	// Prepare properties
	props := make(map[string]interface{})
	props["description"] = event.Description
	props["type"] = event.Type
	props["created_at"] = now
	props["source_url"] = sourceURL

	if event.Date != "" {
		props["date"] = event.Date
	}

	if event.Evidence != "" {
		props["evidence"] = event.Evidence
	}

	// Add event-specific properties
	if event.Properties != nil {
		for k, v := range event.Properties {
			props[k] = v
		}
	}

	query := `
		CREATE (e:EVENT $props)
		RETURN ID(e) as id
	`

	result, err := tx.Run(query, map[string]interface{}{
		"props": props,
	})
	if err != nil {
		return "", err
	}

	if !result.Next() {
		return "", fmt.Errorf("failed to create event node")
	}

	record := result.Record()
	return fmt.Sprint(record.Values[0]), nil
}

// createEventRelationship creates a relationship between an entity and an event
func (gi *GraphIntegrator) createEventRelationship(tx neo4j.Transaction, entityID, eventID, sourceURL string) error {
	now := time.Now()

	props := map[string]interface{}{
		"created_at": now,
		"source_url": sourceURL,
	}

	query := `
		MATCH (entity), (event)
		WHERE ID(entity) = $entityId AND ID(event) = $eventId
		CREATE (entity)-[r:INVOLVED_IN $props]->(event)
		RETURN ID(r) as id
	`

	result, err := tx.Run(query, map[string]interface{}{
		"entityId": entityID,
		"eventId":  eventID,
		"props":    props,
	})
	if err != nil {
		return err
	}

	if !result.Next() {
		return fmt.Errorf("failed to create event relationship")
	}

	return nil
}

// GraphIntegrationHandler handles requests to integrate extraction results into the graph
func GraphIntegrationHandler(c *gin.Context) {
	var req GraphIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate database availability
	if !db.IsAvailable() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service is currently unavailable",
		})
		return
	}

	integrator := NewGraphIntegrator(true) // Enable debug logging

	result, err := integrator.IntegrateExtraction(c.Request.Context(), &req)
	if err != nil {
		log.Printf("[GRAPH_INTEGRATION] Integration failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to integrate extraction results",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExtractAndIntegrateHandler combines extraction and graph integration in one call
func ExtractAndIntegrateHandler(cfg *config.Config) gin.HandlerFunc {
	scraper := NewArticleScraper()
	llmClient := llm.NewClient(cfg)
	integrator := NewGraphIntegrator(true)

	return func(c *gin.Context) {
		var req ExtractorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		log.Printf("[EXTRACT_AND_INTEGRATE] Starting combined extraction and integration for URL: %s", req.URL)

		// Step 1: Extract from article (reusing extractor logic)
		startTime := time.Now()

		// Scrape article
		title, content, _, _, err := scraper.ScrapeArticle(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to scrape article: %v", err),
			})
			return
		}

		// Generate extraction prompt and call LLM
		prompt := generateExtractionPrompt(title, content, req.ExtraPrompt)
		messages := []llm.Message{{Role: "user", Content: prompt}}

		llmResult, err := llmClient.Generate(c.Request.Context(), messages)
		if err != nil || len(llmResult.Choices) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process article with LLM",
			})
			return
		}

		// Parse extraction results
		var extractionResult struct {
			Entities      []ExtractedEntity       `json:"entities"`
			Relationships []ExtractedRelationship `json:"relationships"`
			Events        []ExtractedEvent        `json:"events"`
		}

		llmResponse := llmResult.Choices[0].Message.Content
		if err := json.Unmarshal([]byte(llmResponse), &extractionResult); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":        "Failed to parse LLM response",
				"llm_response": llmResponse,
			})
			return
		}

		// Step 2: Integrate into graph database
		integrationReq := &GraphIntegrationRequest{
			URL:           req.URL,
			Source:        "article_extraction",
			Entities:      extractionResult.Entities,
			Relationships: extractionResult.Relationships,
			Events:        extractionResult.Events,
			Metadata:      req.Metadata,
		}

		integrationResult, err := integrator.IntegrateExtraction(c.Request.Context(), integrationReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":             "Failed to integrate into graph database",
				"details":           err.Error(),
				"extraction_result": extractionResult,
			})
			return
		}

		// Combined response
		response := map[string]interface{}{
			"url":   req.URL,
			"title": title,
			"extraction": map[string]interface{}{
				"entities":      extractionResult.Entities,
				"relationships": extractionResult.Relationships,
				"events":        extractionResult.Events,
			},
			"integration":           integrationResult,
			"total_processing_time": time.Since(startTime),
		}

		log.Printf("[EXTRACT_AND_INTEGRATE] Completed successfully - Total time: %v", time.Since(startTime))
		c.JSON(http.StatusOK, response)
	}
}
