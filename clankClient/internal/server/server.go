package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clankClient/internal/client"
	"clankClient/internal/config"

	"github.com/gorilla/mux"
	"github.com/mark3labs/mcp-go/mcp"
)

type Server struct {
	client     *client.ClankClient
	httpServer *http.Server
	port       string
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ToolCallRequest struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ResourceRequest struct {
	ResourceURI string `json:"resource_uri"`
}

type PromptRequest struct {
	PromptName string            `json:"prompt_name"`
	Arguments  map[string]string `json:"arguments,omitempty"`
}

func New(port string) *Server {
	if port == "" {
		port = "8081" // Default port for the client daemon
	}

	return &Server{
		port: port,
	}
}

func (s *Server) Start() error {
	// Create and connect clank client
	log.Printf("Creating Clank MCP client connection...")
	var err error
	s.client, err = client.New(config.DefaultServerURL)
	if err != nil {
		return fmt.Errorf("failed to create clank client: %w", err)
	}

	if err := s.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to clank server: %w", err)
	}

	// Setup HTTP routes
	router := mux.NewRouter()
	
	// Health check
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// Info endpoints
	router.HandleFunc("/api/info", s.handleInfo).Methods("GET")
	router.HandleFunc("/api/tools", s.handleListTools).Methods("GET")
	router.HandleFunc("/api/resources", s.handleListResources).Methods("GET")
	router.HandleFunc("/api/prompts", s.handleListPrompts).Methods("GET")
	
	// Action endpoints
	router.HandleFunc("/api/tools/call", s.handleCallTool).Methods("POST")
	router.HandleFunc("/api/resources/get", s.handleGetResource).Methods("POST")
	router.HandleFunc("/api/prompts/get", s.handleGetPrompt).Methods("POST")

	// CORS middleware
	router.Use(s.corsMiddleware)

	s.httpServer = &http.Server{
		Addr:         ":" + s.port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Clank Client Daemon starting on port %s", s.port)
	log.Printf("📡 Connected to Clank MCP Server at %s", config.DefaultServerURL)
	log.Println("Available endpoints:")
	log.Println("  GET  /health              - Health check")
	log.Println("  GET  /api/info            - Server information") 
	log.Println("  GET  /api/tools           - List available tools")
	log.Println("  GET  /api/resources       - List available resources")
	log.Println("  GET  /api/prompts         - List available prompts")
	log.Println("  POST /api/tools/call      - Call a tool")
	log.Println("  POST /api/resources/get   - Get a resource")
	log.Println("  POST /api/prompts/get     - Get a prompt")

	// Handle graceful shutdown
	go s.handleShutdown()

	// Start server
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	<-sigChan
	log.Println("\n🛑 Shutdown signal received, gracefully stopping...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during HTTP server shutdown: %v", err)
	}

	// Disconnect clank client
	if s.client != nil {
		if err := s.client.Disconnect(); err != nil {
			log.Printf("Error disconnecting clank client: %v", err)
		}
	}

	log.Println("✅ Shutdown completed")
	os.Exit(0)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) sendResponse(w http.ResponseWriter, success bool, data interface{}, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	
	response := APIResponse{
		Success: success,
		Data:    data,
		Error:   errorMsg,
	}
	
	statusCode := http.StatusOK
	if !success {
		statusCode = http.StatusInternalServerError
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendResponse(w, true, map[string]string{
		"status":  "healthy",
		"service": "clank-client-daemon",
		"version": config.ClientVersion,
	}, "")
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	s.sendResponse(w, true, map[string]string{
		"message": "Connected to Clank MCP Server",
		"server":  config.DefaultServerURL,
		"version": config.ClientVersion,
	}, "")
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	// We need to capture the output from ListTools since it prints to stdout
	// For now, we'll create a simpler version that returns data instead of printing
	tools, err := s.getToolsList()
	if err != nil {
		s.sendResponse(w, false, nil, fmt.Sprintf("Failed to list tools: %v", err))
		return
	}
	
	s.sendResponse(w, true, map[string]interface{}{
		"tools": tools,
	}, "")
}

func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request) {
	resources, err := s.getResourcesList()
	if err != nil {
		s.sendResponse(w, false, nil, fmt.Sprintf("Failed to list resources: %v", err))
		return
	}
	
	s.sendResponse(w, true, map[string]interface{}{
		"resources": resources,
	}, "")
}

func (s *Server) handleListPrompts(w http.ResponseWriter, r *http.Request) {
	prompts, err := s.getPromptsList()
	if err != nil {
		s.sendResponse(w, false, nil, fmt.Sprintf("Failed to list prompts: %v", err))
		return
	}
	
	s.sendResponse(w, true, map[string]interface{}{
		"prompts": prompts,
	}, "")
}

func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	var req ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON request body")
		return
	}
	
	if req.ToolName == "" {
		s.sendResponse(w, false, nil, "tool_name is required")
		return
	}
	
	result, err := s.callToolAndGetResult(req.ToolName, req.Arguments)
	if err != nil {
		s.sendResponse(w, false, nil, fmt.Sprintf("Failed to call tool: %v", err))
		return
	}
	
	s.sendResponse(w, true, result, "")
}

func (s *Server) handleGetResource(w http.ResponseWriter, r *http.Request) {
	var req ResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON request body")
		return
	}
	
	if req.ResourceURI == "" {
		s.sendResponse(w, false, nil, "resource_uri is required")
		return
	}
	
	result, err := s.getResourceAndGetResult(req.ResourceURI)
	if err != nil {
		s.sendResponse(w, false, nil, fmt.Sprintf("Failed to get resource: %v", err))
		return
	}
	
	s.sendResponse(w, true, result, "")
}

func (s *Server) handleGetPrompt(w http.ResponseWriter, r *http.Request) {
	var req PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON request body")
		return
	}
	
	if req.PromptName == "" {
		s.sendResponse(w, false, nil, "prompt_name is required")
		return
	}
	
	result, err := s.getPromptAndGetResult(req.PromptName, req.Arguments)
	if err != nil {
		s.sendResponse(w, false, nil, fmt.Sprintf("Failed to get prompt: %v", err))
		return
	}
	
	s.sendResponse(w, true, result, "")
}

// Helper methods to get data instead of printing to stdout

func (s *Server) getToolsList() (interface{}, error) {
	result, err := s.client.client.ListTools(s.client.ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	return result.Tools, nil
}

func (s *Server) getResourcesList() (interface{}, error) {
	result, err := s.client.client.ListResources(s.client.ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return nil, err
	}
	return result.Resources, nil
}

func (s *Server) getPromptsList() (interface{}, error) {
	result, err := s.client.client.ListPrompts(s.client.ctx, mcp.ListPromptsRequest{})
	if err != nil {
		return nil, err
	}
	return result.Prompts, nil
}

func (s *Server) callToolAndGetResult(toolName string, arguments map[string]interface{}) (interface{}, error) {
	result, err := s.client.client.CallTool(s.client.ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: arguments,
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) getResourceAndGetResult(resourceURI string) (interface{}, error) {
	result, err := s.client.client.ReadResource(s.client.ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: resourceURI,
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) getPromptAndGetResult(promptName string, arguments map[string]string) (interface{}, error) {
	result, err := s.client.client.GetPrompt(s.client.ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      promptName,
			Arguments: arguments,
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
