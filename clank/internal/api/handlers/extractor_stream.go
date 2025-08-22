package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"clank/config"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
)

// monitorCloseChan wraps a channel so any close attempt is logged
func monitorCloseChan(ch chan string, closed *int32) {
	if !atomic.CompareAndSwapInt32(closed, 0, 1) {
		log.Printf("[CHANNEL CLOSE ATTEMPT] channel already closed!\nStack:\n%s", debug.Stack())
	} else {
		log.Printf("[CHANNEL CLOSE ATTEMPT] closing channel for the first time\nStack:\n%s", debug.Stack())
		close(ch)
	}
}

// ExtractorStreamHandler streams LLM extraction results as they arrive
func ExtractorStreamHandler(cfg *config.Config) gin.HandlerFunc {
	scraper := NewArticleScraper()
	llmClient := llm.NewClient(cfg)

	return func(c *gin.Context) {
		startTime := time.Now()
		debug := &DebugInfo{
			Errors:   []string{},
			Warnings: []string{},
		}

		var req ExtractorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			debug.Errors = append(debug.Errors, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "debug": debug})
			return
		}

		// Scrape article
		title, content, statusCode, contentType, err := scraper.ScrapeArticle(req.URL)
		debug.ScrapeTime = time.Since(startTime)
		debug.HTTPStatus = statusCode
		debug.ContentType = contentType
		debug.ContentLength = len(content)
		debug.WordCount = len(strings.Fields(content))
		if err != nil {
			debug.Errors = append(debug.Errors, err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to scrape article: %v", err), "debug": debug})
			return
		}

		// Set up Server-Sent Events
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// Send initial status
		c.SSEvent("status", map[string]interface{}{
			"phase":    "scraping_complete",
			"title":    title,
			"length":   len(content),
			"progress": 0,
		})
		c.Writer.Flush()

		prompt := generateExtractionPrompt(title, content, req.ExtraPrompt)
		messages := []llm.Message{{Role: "user", Content: prompt}}

		// Create a context with longer timeout for LLM processing
		llmCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		responseChan := make(chan string, 100)
		errChan := make(chan error, 1)
		var llmBuffer strings.Builder
		progress := 0
		lastUpdate := time.Now()
		var closedFlag int32

		// Start streaming goroutine
		go func() {
			defer func() {
				monitorCloseChan(responseChan, &closedFlag)
			}()

			if err := llmClient.GenerateStream(llmCtx, messages, responseChan); err != nil {
				log.Printf("[EXTRACTOR STREAM] LLM error: %v", err)
				select {
				case errChan <- err:
				default:
				}
			}
		}()

		// Send progress updates
		c.SSEvent("status", map[string]interface{}{
			"phase":    "llm_processing",
			"progress": 0,
		})
		c.Writer.Flush()

		// Accumulate chunks with real-time streaming
		timeout := time.NewTimer(12 * time.Minute)
		defer timeout.Stop()

		for {
			select {
			case chunk, ok := <-responseChan:
				if !ok {
					// Channel closed, streaming is done
					goto ProcessResponse
				}

				llmBuffer.WriteString(chunk)
				progress += len(chunk)

				// Send progress updates every 500ms or every 200 chars
				if time.Since(lastUpdate) > 500*time.Millisecond || len(chunk) > 200 {
					c.SSEvent("progress", map[string]interface{}{
						"phase":       "generating",
						"progress":    progress,
						"partial":     chunk,
						"total_chars": progress,
					})
					c.Writer.Flush()
					lastUpdate = time.Now()
				}

			case err := <-errChan:
				// LLM error occurred
				debug.Errors = append(debug.Errors, err.Error())
				c.SSEvent("error", map[string]interface{}{
					"error": "LLM processing failed",
					"debug": debug,
				})
				return

			case <-timeout.C:
				// Overall timeout
				cancel()
				debug.Errors = append(debug.Errors, "Request timeout after 12 minutes")
				c.SSEvent("error", map[string]interface{}{
					"error": "Request timeout",
					"debug": debug,
				})
				return

			case <-c.Request.Context().Done():
				// Client disconnected
				cancel()
				log.Printf("[EXTRACTOR STREAM] Client disconnected: %v", c.Request.Context().Err())
				return
			}
		}

	ProcessResponse:
		llmResponse := llmBuffer.String()
		log.Printf("[EXTRACTOR STREAM] LLM complete, total length: %d chars", len(llmResponse))

		// Send completion status
		c.SSEvent("status", map[string]interface{}{
			"phase":       "processing_complete",
			"total_chars": len(llmResponse),
		})
		c.Writer.Flush()

		if len(llmResponse) == 0 {
			debug.Errors = append(debug.Errors, "Empty response from LLM")
			c.SSEvent("error", map[string]interface{}{
				"error": "Empty response from LLM",
				"debug": debug,
			})
			return
		}

		// Parse JSON
		var extractionResult struct {
			Entities      []ExtractedEntity       `json:"entities"`
			Relationships []ExtractedRelationship `json:"relationships"`
			Events        []ExtractedEvent        `json:"events"`
		}

		if err := json.Unmarshal([]byte(llmResponse), &extractionResult); err != nil {
			debug.Errors = append(debug.Errors, "Failed to parse LLM JSON: "+err.Error())
			c.SSEvent("error", map[string]interface{}{
				"error":        "Failed to parse LLM response",
				"llm_response": llmResponse[:min(1000, len(llmResponse))],
				"debug":        debug,
				"total_chars":  len(llmResponse),
			})
			return
		}

		// Send final results
		resp := ExtractorResponse{
			URL:            req.URL,
			Title:          title,
			Entities:       extractionResult.Entities,
			Relationships:  extractionResult.Relationships,
			Events:         extractionResult.Events,
			ProcessingTime: time.Since(startTime),
			Debug:          debug,
		}

		if req.Metadata != nil {
			resp.Metadata = make(map[string]interface{})
			for k, v := range req.Metadata {
				resp.Metadata[k] = v
			}
		}

		c.SSEvent("result", resp)
		c.SSEvent("done", nil)
		c.Writer.Flush()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
