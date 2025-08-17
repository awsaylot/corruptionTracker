package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"clank/config"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPService handles MCP server functionality alongside llama.cpp
type MCPService struct {
	llmClient *llm.Client
	server    *mcp.Server
}

// MCPRequest represents an incoming MCP request
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	ID     string          `json:"id"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  *MCPError       `json:"error,omitempty"`
	ID     string          `json:"id"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPService creates a new MCP service integrated with llama.cpp
func NewMCPService(cfg *config.Config) *MCPService {
	service := &MCPService{
		llmClient: llm.NewClient(cfg),
	}

	// Create MCP implementation for tool calling and context injection
	impl := &mcp.Implementation{
		Name:    "clank-llm-proxy",
		Version: "v0.1.0",
	}

	// Initialize MCP server
	service.server = mcp.NewServer(impl, &mcp.ServerOptions{
		// Configure server options as needed
	})

	return service
}

// ProcessWithMCP processes a message through MCP before sending to llama.cpp
func (s *MCPService) ProcessWithMCP(ctx context.Context, messages []llm.Message) ([]llm.Message, error) {
	// This is where you can add MCP processing:
	// 1. Tool calling
	// 2. Context injection from graph database
	// 3. Message preprocessing

	log.Printf("Processing %d messages through MCP", len(messages))

	// For now, just pass through the messages
	// You can add MCP tool calling logic here later
	processedMessages := make([]llm.Message, len(messages))
	copy(processedMessages, messages)

	// Example: Add system context from graph database
	if len(processedMessages) > 0 {
		// You could inject context here based on the conversation
		// systemContext := s.getContextFromGraph(messages)
		// if systemContext != "" {
		//     processedMessages = append([]llm.Message{{
		//         Role: "system",
		//         Content: systemContext,
		//     }}, processedMessages...)
		// }
	}

	return processedMessages, nil
}

// GenerateWithMCP generates a response using llama.cpp with MCP preprocessing
func (s *MCPService) GenerateWithMCP(ctx context.Context, messages []llm.Message, responseChan chan<- string) error {
	// Process messages through MCP first
	processedMessages, err := s.ProcessWithMCP(ctx, messages)
	if err != nil {
		return fmt.Errorf("MCP processing failed: %w", err)
	}

	// Send to llama.cpp
	return s.llmClient.GenerateStream(ctx, processedMessages, responseChan)
}

// MCPHandlerSSE handles MCP requests via Server-Sent Events
func MCPHandlerSSE(c *gin.Context) {
	cfg := config.LoadConfig()
	mcpService := NewMCPService(cfg)

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	log.Printf("Starting MCP SSE connection")

	// Create SSE transport for MCP
	transport := mcp.NewSSEServerTransport("mcp-llm-proxy", c.Writer)

	// Run the MCP server
	ctx := c.Request.Context()
	if err := mcpService.server.Run(ctx, transport); err != nil {
		log.Printf("MCP server error: %v", err)
		c.String(http.StatusInternalServerError, "MCP server error: %v", err)
		return
	}
}

// MCPChatHandler handles chat requests with MCP integration
func MCPChatHandler(cfg *config.Config) gin.HandlerFunc {
	mcpService := NewMCPService(cfg)

	return func(c *gin.Context) {
		var request struct {
			Messages []llm.Message `json:"messages"`
			Stream   bool          `json:"stream"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if request.Stream {
			// Handle streaming response
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")

			responseChan := make(chan string, 100)

			go func() {
				defer close(responseChan)
				if err := mcpService.GenerateWithMCP(c.Request.Context(), request.Messages, responseChan); err != nil {
					log.Printf("MCP generation error: %v", err)
					responseChan <- fmt.Sprintf("data: {\"error\": \"%s\"}\n\n", err.Error())
				}
			}()

			// Stream responses back
			for chunk := range responseChan {
				if chunk != "" {
					fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
					c.Writer.Flush()
				}
			}
		} else {
			// Handle non-streaming response
			result, err := mcpService.llmClient.Generate(c.Request.Context(), request.Messages)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		}
	}
}
