package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"clank/config"
	"clank/internal/llm"
	"clank/internal/models"
	"clank/internal/prompts"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPService handles MCP server functionality alongside llama.cpp
type MCPService struct {
	llmClient    *llm.Client
	server       *mcp.Server
	promptLoader *prompts.PromptLoader
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
		llmClient:    llm.NewClient(cfg),
		promptLoader: prompts.NewPromptLoader("./prompts"), // adjust path to your prompts folder
	}

	// Load all prompts at startup
	if err := service.promptLoader.LoadPrompts(); err != nil {
		log.Printf("Error loading prompts: %v", err)
	}

	// Create MCP implementation for tool calling and context injection
	impl := &mcp.Implementation{
		Name:    "clank-llm-proxy",
		Version: "v0.1.0",
	}

	service.server = mcp.NewServer(impl, &mcp.ServerOptions{})

	return service
}

// ProcessWithMCP processes messages through MCP, injecting system prompts
func (s *MCPService) ProcessWithMCP(ctx context.Context, messages []llm.Message) ([]llm.Message, error) {
	log.Printf("Processing %d messages through MCP", len(messages))

	var processed []llm.Message

	// Reload prompts if files have changed
	if err := s.promptLoader.ReloadIfChanged(); err != nil {
		log.Printf("Error reloading prompts: %v", err)
	}

	// Inject system prompt
	systemContext := &models.PromptContext{
		Arguments: map[string]any{},
		Timestamp: time.Now(),
	}
	if sysPrompt, err := s.promptLoader.RenderPrompt("system", systemContext); err == nil {
		processed = append(processed, llm.Message{
			Role:    "system",
			Content: sysPrompt.RenderedText,
		})
	} else {
		log.Printf("Failed to render system prompt: %v", err)
	}

	// TODO: optionally add prompts for tools or context analysis here based on message content

	// Append user messages
	processed = append(processed, messages...)

	return processed, nil
}

// GenerateWithMCP generates a response using llama.cpp with MCP preprocessing
func (s *MCPService) GenerateWithMCP(ctx context.Context, messages []llm.Message, responseChan chan<- string) error {
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

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	log.Printf("Starting MCP SSE connection")

	transport := mcp.NewSSEServerTransport("mcp-llm-proxy", c.Writer)

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

			for chunk := range responseChan {
				if chunk != "" {
					fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
					c.Writer.Flush()
				}
			}
		} else {
			processedMessages, err := mcpService.ProcessWithMCP(c.Request.Context(), request.Messages)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			result, err := mcpService.llmClient.Generate(c.Request.Context(), processedMessages)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		}
	}
}
