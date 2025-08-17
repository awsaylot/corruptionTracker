package handlers

import (
	"fmt"
	"log"
	"net/http"

	"clank/config"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
)

// LLMRequest represents a request to the LLM
type LLMRequest struct {
	Messages []llm.Message `json:"messages"`
	Model    string        `json:"model,omitempty"`
	Stream   bool          `json:"stream,omitempty"`
}

// LLMHandler handles direct HTTP requests to llama.cpp (non-streaming)
func LLMHandler(c *gin.Context) {
	cfg := config.LoadConfig()
	client := llm.NewClient(cfg)

	var req LLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No messages provided",
		})
		return
	}

	log.Printf("Processing LLM request with %d messages", len(req.Messages))

	// Generate response from llama.cpp
	result, err := client.Generate(c.Request.Context(), req.Messages)
	if err != nil {
		log.Printf("Error generating response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate response from llama.cpp",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// LLMStreamHandler handles streaming HTTP requests to llama.cpp
func LLMStreamHandler(cfg *config.Config) gin.HandlerFunc {
	client := llm.NewClient(cfg)

	return func(c *gin.Context) {
		var req LLMRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		if len(req.Messages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No messages provided",
			})
			return
		}

		log.Printf("Processing streaming LLM request with %d messages", len(req.Messages))

		// Set SSE headers
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// Create response channel
		responseChan := make(chan string, 100)

		// Start streaming generation
		go func() {
			defer close(responseChan)
			if err := client.GenerateStream(c.Request.Context(), req.Messages, responseChan); err != nil {
				log.Printf("Streaming generation error: %v", err)
				responseChan <- fmt.Sprintf(`{"error": "%s"}`, err.Error())
			}
		}()

		// Stream responses back to client
		for chunk := range responseChan {
			if chunk != "" {
				// Send as SSE format
				fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
				c.Writer.Flush()
			}
		}

		// Send completion marker
		fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
		c.Writer.Flush()
	}
}

// HealthHandler checks if the service is healthy and can reach llama.cpp
func HealthHandler(c *gin.Context) {
	cfg := config.LoadConfig()

	// Basic health check
	health := gin.H{
		"status":  "ok",
		"service": "clank-backend",
		"llm_url": cfg.LLM.URL,
	}

	// Test connection to llama.cpp
	client := llm.NewClient(cfg)
	testMessages := []llm.Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	// Try a simple request to verify llama.cpp is accessible
	_, err := client.Generate(c.Request.Context(), testMessages)
	if err != nil {
		health["llm_status"] = "error"
		health["llm_error"] = err.Error()
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	health["llm_status"] = "ok"
	c.JSON(http.StatusOK, health)
}
